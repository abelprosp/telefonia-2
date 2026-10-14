package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

const intakeMaxFileBytes = 8 << 20
const intakeMaxTotalBytes = 16 << 20

type IntakeObjectStorage interface {
	PutIntakeObject(ctx context.Context, bucket, key, contentType string, data []byte) error
}

type preparedIntakeAttachment struct {
	input     models.CustomerIntakeAttachment
	data      []byte
	extension string
}

func intakeValidation(message string) error {
	return httputil.ValidationError(notifications.N("CUSTOMER_INTAKE_INVALID", message))
}

func validateCustomerIntake(input *models.CustomerIntakeInput) (models.CreateCustomerInput, []preparedIntakeAttachment, string, error) {
	var customer models.CreateCustomerInput
	if input.Client == nil || input.Sale == nil {
		return customer, nil, "", intakeValidation("Informe os blocos client e sale.")
	}
	input.SourceSystem = strings.TrimSpace(input.SourceSystem)
	if input.SourceSystem == "" {
		input.SourceSystem = "external"
	}
	if utf8.RuneCountInString(input.SourceSystem) > 128 || strings.ContainsRune(input.SourceSystem, 0) {
		return customer, nil, "", intakeValidation("sourceSystem deve ter até 128 caracteres.")
	}
	// Bound every free-text input before it can be persisted or rendered.
	for _, object := range []any{input.Client, input.Sale} {
		v := reflect.ValueOf(object).Elem()
		for i := 0; i < v.NumField(); i++ {
			if v.Field(i).Kind() != reflect.String {
				continue
			}
			value := strings.TrimSpace(v.Field(i).String())
			max := 256
			if v.Type().Field(i).Name == "Notes" {
				max = 4000
			}
			if utf8.RuneCountInString(value) > max || strings.ContainsRune(value, 0) {
				return customer, nil, "", intakeValidation("Campo inválido ou acima do tamanho permitido: " + v.Type().Field(i).Name)
			}
			v.Field(i).SetString(value)
		}
	}
	c, sale := input.Client, input.Sale
	c.Document = httputil.NormalizeDigits(c.Document)
	c.DocumentType = strings.ToUpper(c.DocumentType)
	switch c.DocumentType {
	case "CPF":
		customer.Type = "PF"
	case "CNPJ":
		customer.Type = "PJ"
	default:
		return customer, nil, "", intakeValidation("documentType deve ser CPF ou CNPJ.")
	}
	customer.Name, customer.Document = c.Name, c.Document
	if customer.Type == "PJ" {
		customer.LegalName = &c.LegalName
	}
	customer.Profile = &models.CustomerRegistrationProfile{RG: c.RG, Phone: c.Phone}
	if c.Email != "" {
		customer.Profile.Emails = []string{c.Email}
		customer.BillingEmail = &c.Email
	}
	if c.Address != "" || c.AddressNumber != "" || c.City != "" || c.State != "" || c.ZipCode != "" || c.Neighborhood != "" || c.Complement != "" {
		customer.Addresses = []models.CreateCustomerAddressInput{{Street: c.Address, Number: c.AddressNumber, Neighborhood: c.Neighborhood, City: c.City, State: c.State, ZipCode: c.ZipCode, Complement: &c.Complement, Country: "Brasil"}}
	}
	if err := validateCreateCustomer(customer); err != nil {
		return customer, nil, "", err
	}
	if err := normalizeCustomerRegistration(customer.Type, customer.Profile, customer.Addresses, customer.BillingEmail); err != nil {
		return customer, nil, "", err
	}
	if sale.ID == "" || sale.Protocol == "" || sale.SoldAt == nil || sale.Value == nil || sale.CommissionValue == nil {
		return customer, nil, "", intakeValidation("sale exige id, protocol, soldAt, value e commissionValue (valores em reais).")
	}
	if *sale.Value < 0 || *sale.CommissionValue < 0 {
		return customer, nil, "", intakeValidation("Valores monetários não podem ser negativos.")
	}
	if sale.ConcludedAt != nil && sale.ConcludedAt.Before(*sale.SoldAt) {
		return customer, nil, "", intakeValidation("concludedAt não pode ser anterior a soldAt.")
	}
	soldUTC := sale.SoldAt.UTC()
	sale.SoldAt = &soldUTC
	if sale.ConcludedAt != nil {
		concludedUTC := sale.ConcludedAt.UTC()
		sale.ConcludedAt = &concludedUTC
	}
	if sale.IsUpgrade && sale.UpgradeOfProtocol == "" {
		return customer, nil, "", intakeValidation("Informe upgradeOfProtocol para vendas de upgrade.")
	}
	for _, number := range []*string{&sale.LineNumber, &sale.PortabilityNumber} {
		*number = httputil.NormalizeDigits(*number)
		if *number != "" && (len(*number) < 10 || len(*number) > 15) {
			return customer, nil, "", intakeValidation("Telefones da venda devem conter DDD, entre 10 e 15 dígitos.")
		}
	}
	if sale.IsPortability && (sale.PortabilityNumber == "" || sale.DonorOperator == "") {
		return customer, nil, "", intakeValidation("Portabilidade exige portabilityNumber e donorOperator.")
	}
	for _, field := range []struct {
		value    string
		min, max int
	}{{sale.ICCID, 19, 22}, {sale.DeviceIMEI, 15, 15}, {sale.DeviceEID, 32, 32}} {
		if field.value != "" && (httputil.NormalizeDigits(field.value) != field.value || len(field.value) < field.min || len(field.value) > field.max) {
			return customer, nil, "", intakeValidation("ICCID, IMEI ou EID inválido; envie como texto contendo apenas dígitos.")
		}
	}
	if len(input.Attachments) > 10 {
		return customer, nil, "", intakeValidation("Envie no máximo 10 anexos.")
	}
	files := []preparedIntakeAttachment{}
	seen := map[string]bool{}
	total := 0
	for i := range input.Attachments {
		a := &input.Attachments[i]
		a.ID, a.Name, a.Type, a.MimeType = strings.TrimSpace(a.ID), strings.TrimSpace(a.Name), strings.TrimSpace(a.Type), strings.ToLower(strings.TrimSpace(a.MimeType))
		if a.ID == "" || a.Name == "" || a.Type == "" || len(a.ID) > 256 || len(a.Name) > 256 || len(a.Type) > 64 || seen[a.ID] || strings.ContainsRune(a.ID, 0) || strings.ContainsRune(a.Type, 0) || strings.ContainsAny(a.Name, "\\/\x00\r\n") {
			return customer, nil, "", intakeValidation("Cada anexo exige id único, type e name válidos (sem caminhos).")
		}
		seen[a.ID] = true
		extensions := map[string]string{"application/pdf": ".pdf", "image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}
		extension, allowed := extensions[a.MimeType]
		if !allowed || len(a.ContentBase64) > base64.StdEncoding.EncodedLen(intakeMaxFileBytes)+4 {
			return customer, nil, "", intakeValidation("Anexos aceitos: PDF, JPEG, PNG e WebP, até 8 MiB por arquivo.")
		}
		data, err := base64.StdEncoding.Strict().DecodeString(a.ContentBase64)
		if err != nil || len(data) == 0 || len(data) > intakeMaxFileBytes {
			return customer, nil, "", intakeValidation("contentBase64 inválido ou arquivo acima de 8 MiB.")
		}
		total += len(data)
		if total > intakeMaxTotalBytes {
			return customer, nil, "", intakeValidation("Os anexos devem somar no máximo 16 MiB.")
		}
		if http.DetectContentType(data) != a.MimeType {
			return customer, nil, "", intakeValidation("O conteúdo do arquivo não corresponde ao mimeType informado.")
		}
		a.ContentBase64 = base64.StdEncoding.EncodeToString(data)
		files = append(files, preparedIntakeAttachment{input: *a, data: data, extension: extension})
	}
	// Hash the normalized request, including file bytes. Never store Base64 in the database.
	body, err := json.Marshal(input)
	if err != nil {
		return customer, nil, "", err
	}
	hash := sha256.Sum256(body)
	return customer, files, hex.EncodeToString(hash[:]), nil
}

func (s *Service) ImportCustomerSale(ctx context.Context, input models.CustomerIntakeInput) (*models.CustomerIntakeResponse, error) {
	customer, files, fingerprint, err := validateCustomerIntake(&input)
	if err != nil {
		return nil, err
	}
	org, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if strings.ContainsAny(org, "/\\") || org == "." || org == ".." {
		return nil, intakeValidation("Organização de destino inválida.")
	}
	actor, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	if len(files) > 0 && (s.IntakeStorage == nil || s.IntakeBucket == "") {
		return nil, httputil.UnavailableError(notifications.N("INTAKE_STORAGE_UNAVAILABLE", "Armazenamento indisponível para receber anexos."))
	}
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	var response *models.CustomerIntakeResponse
	err = s.Store.WithTx(ctx, func(_ context.Context, tx pgx.Tx) error {
		txCtx := store.CtxWithTx(ctx, tx)
		// Serialize retries and simultaneous sales for the same CPF/CNPJ within this tenant.
		for _, key := range []string{"intake:" + org + ":" + input.SourceSystem + ":" + input.Sale.ID, "customer:" + org + ":" + customer.Document} {
			if _, err := tx.Exec(txCtx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, key); err != nil {
				return err
			}
		}
		existing, err := s.Store.GetCustomerIntake(txCtx, org, input.SourceSystem, input.Sale.ID)
		if err != nil {
			return err
		}
		if existing != nil {
			if existing.PayloadHash != fingerprint {
				return httputil.BusinessError(notifications.N("INTAKE_REPLAY_CONFLICT", "Esta venda externa já foi recebida com outro conteúdo; nenhum dado foi alterado."))
			}
			response = &models.CustomerIntakeResponse{CustomerID: existing.CustomerID, SaleID: existing.SaleID, IntakeID: existing.ID, Replayed: true}
			return nil
		}
		ids, err := s.Store.IntakeCustomersByDocument(txCtx, org, customer.Document)
		if err != nil {
			return err
		}
		if len(ids) > 1 {
			return httputil.BusinessError(notifications.N("INTAKE_CUSTOMER_AMBIGUOUS", "Há mais de um cliente com este documento; revise os cadastros antes de importar."))
		}
		customerID := uuid.NewString()
		if len(ids) == 1 {
			customerID = ids[0]
			existingCustomer, err := s.Store.GetCustomerInOrg(txCtx, org, customerID, nil)
			if err != nil {
				return err
			}
			if existingCustomer == nil || !existingCustomer.Active {
				return httputil.BusinessError(notifications.N("INTAKE_CUSTOMER_INACTIVE", "O cliente está inativo. Reative-o antes de importar uma nova venda."))
			}
			if !strings.EqualFold(existingCustomer.Type, customer.Type) {
				return httputil.BusinessError(notifications.N("INTAKE_CUSTOMER_TYPE_MISMATCH", "O tipo do cliente existente não corresponde ao documento recebido."))
			}
			// Preserve existing registration; the sale's source client snapshot remains auditable.
		} else {
			if err := s.Store.CreateCustomer(txCtx, org, customerID, "", strings.ToLower(customer.Type), customer.Name, customer.Document, httputil.DocumentTypeForCustomer(strings.ToLower(customer.Type)), customer.LegalName, nil, nil, nil); err != nil {
				return err
			}
			if err := s.Store.SetCustomerRegistration(txCtx, customerID, customer.Profile, &customer.Addresses); err != nil {
				return err
			}
			if customer.BillingEmail != nil {
				if err := s.Store.UpdateCustomerBillingEmail(txCtx, org, customerID, *customer.BillingEmail); err != nil {
					return err
				}
			}
		}
		now := time.Now().UTC()
		r := models.CustomerIntakeRecord{ID: uuid.NewString(), CustomerID: customerID, SaleID: uuid.NewString(), SourceSystem: input.SourceSystem, Sale: *input.Sale, Client: *input.Client, Attachments: []models.CustomerIntakeAttachmentInfo{}, ImportedAt: now, PayloadHash: fingerprint}
		if err := s.Store.CreateIntakeSale(txCtx, org, customerID, r.SaleID, "IMP-"+strings.ReplaceAll(r.SaleID, "-", "")[:24], actor.ID, uuid.NewString(), r.Sale, now); err != nil {
			return err
		}
		objectNamespace := sha256.Sum256([]byte(org + "\x00" + input.SourceSystem + "\x00" + input.Sale.ID + "\x00" + fingerprint))
		for i, file := range files {
			key := path.Join(org, "intake", hex.EncodeToString(objectNamespace[:]), fmt.Sprintf("%02d%s", i, file.extension))
			// Deterministic keys make retries safe even after an uncertain DB commit.
			if err := s.IntakeStorage.PutIntakeObject(txCtx, s.IntakeBucket, key, file.input.MimeType, file.data); err != nil {
				return err
			}
			attachmentID := uuid.NewString()
			title := file.input.Type + " — " + input.Sale.Protocol
			if utf8.RuneCountInString(title) > 256 {
				title = string([]rune(title)[:256])
			}
			size := int64(len(file.data))
			if _, err := s.Store.CreateCustomerAttachment(txCtx, org, customerID, attachmentID, models.RegisterCustomerAttachmentInput{Title: &title, OriginalFileName: file.input.Name, StorageBucket: s.IntakeBucket, StorageObjectKey: key, ContentType: &file.input.MimeType, SizeBytes: &size}); err != nil {
				return err
			}
			r.Attachments = append(r.Attachments, models.CustomerIntakeAttachmentInfo{ExternalID: file.input.ID, Type: file.input.Type, Name: file.input.Name, MimeType: file.input.MimeType, AttachmentID: attachmentID, SizeBytes: size})
		}
		if err := s.Store.SaveCustomerIntake(txCtx, org, r); err != nil {
			return err
		}
		changedBy := actor.ID
		if err := s.Store.InsertAuditLog(txCtx, uuid.NewString(), "Create", "CustomerIntakes", r.ID, &changedBy, nil, nil, now, ""); err != nil {
			return err
		}
		response = &models.CustomerIntakeResponse{CustomerID: customerID, SaleID: r.SaleID, IntakeID: r.ID}
		return nil
	})
	if err != nil {
		var appErr *httputil.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		slog.ErrorContext(ctx, "customer intake failed", "error", err)
		return nil, httputil.InternalError(notifications.N("INTAKE_FAILED", "Não foi possível concluir a importação. Reenvie o mesmo conteúdo para tentar novamente."))
	}
	return response, nil
}

func (s *Service) ListCustomerIntakes(ctx context.Context, customerID string) ([]models.CustomerIntakeRecord, error) {
	if _, err := s.requireCustomer(ctx, customerID); err != nil {
		return nil, err
	}
	org, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.ListCustomerIntakes(ctx, org, customerID)
}
