package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/parceiros"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) ImportParceirosSale(ctx context.Context, orgID string, input parceiros.SaleIntake) (parceiros.IntakeResult, error) {
	prepared, err := parceiros.Prepare(input)
	if err != nil {
		return parceiros.IntakeResult{}, err
	}
	existing, err := s.Store.FindParceirosImport(ctx, orgID, prepared.ExternalSaleID)
	if err != nil {
		return parceiros.IntakeResult{}, err
	}
	if existing != nil {
		return intakeResultFromRow(existing), nil
	}

	files, skipped, err := s.storeParceirosFiles(orgID, prepared.ExternalSaleID, input.Attachments, &prepared.Warnings)
	if err != nil {
		return parceiros.IntakeResult{}, err
	}

	payloadInput := input
	for i := range payloadInput.Attachments {
		payloadInput.Attachments[i].ContentBase64 = ""
	}
	payload, err := json.Marshal(payloadInput)
	if err != nil {
		return parceiros.IntakeResult{}, err
	}

	var result parceiros.IntakeResult
	err = s.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = store.CtxWithTx(ctx, tx)
		again, err := s.Store.FindParceirosImport(ctx, orgID, prepared.ExternalSaleID)
		if err != nil {
			return err
		}
		if again != nil {
			result = intakeResultFromRow(again)
			return nil
		}

		customerID, created, err := s.upsertParceirosCustomer(ctx, orgID, &prepared)
		if err != nil {
			return err
		}
		lineID, lineCreated, err := s.upsertParceirosLine(ctx, orgID, customerID, input, &prepared)
		if err != nil {
			return err
		}
		importID := uuid.NewString()
		accepted := len(files)
		if prepared.Warnings == nil {
			prepared.Warnings = []string{}
		}
		row := store.ParceirosImportInsert{
			ID: importID, OrganizationID: orgID, ExternalSaleID: prepared.ExternalSaleID,
			Protocol: clip(&prepared.Warnings, "Protocolo", prepared.Protocol, 64),
			TaskProtocol: clip(&prepared.Warnings, "Protocolo do Task", prepared.TaskProtocol, 64),
			CustomerID: customerID, PhoneLineID: lineID, CustomerCreated: created, PhoneLineCreated: lineCreated,
			AttachmentsAccepted: accepted, AttachmentsSkipped: skipped,
			PartnerName: clip(&prepared.Warnings, "Parceiro", input.Sale.PartnerName, 256),
			BranchName: clip(&prepared.Warnings, "Filial", input.Sale.BranchName, 256),
			OperatorName: clip(&prepared.Warnings, "Operadora", input.Sale.OperatorName, 128),
			PlanName: clip(&prepared.Warnings, "Plano", input.Sale.PlanName, 128),
			LineNumber: clip(&prepared.Warnings, "Número da linha", strings.TrimSpace(input.Sale.LineNumber), 32),
			Value: input.Sale.Value, CommissionValue: input.Sale.CommissionValue,
			SoldAt: prepared.SoldAt, ConcludedAt: prepared.ConcludedAt,
			SimType: prepared.SimType,
			Iccid: clip(&prepared.Warnings, "ICCID", prepared.Iccid, 64),
			ContractFormat: clip(&prepared.Warnings, "Formato do contrato", input.Sale.ContractFormat, 32),
			IsUpgrade: input.Sale.IsUpgrade,
			UpgradeOfProtocol: clip(&prepared.Warnings, "Protocolo do upgrade", input.Sale.UpgradeOfProtocol, 64),
			IsPortability: input.Sale.IsPortability,
			PortabilityNumber: clip(&prepared.Warnings, "Número da portabilidade", input.Sale.PortabilityNumber, 32),
			DonorOperator: clip(&prepared.Warnings, "Operadora de origem", input.Sale.DonorOperator, 32),
			Warnings: prepared.Warnings, Payload: payload,
		}
		if err = s.Store.InsertParceirosImport(ctx, row); err != nil {
			return err
		}
		for _, file := range files {
			title := file.title
			mime := file.contentType
			size := file.size
			if _, err = s.Store.CreateCustomerAttachment(ctx, orgID, customerID, file.attachmentID, models.RegisterCustomerAttachmentInput{
				Title:            &title,
				OriginalFileName: file.fileName,
				StorageBucket:    "local",
				StorageObjectKey: file.objectKey,
				ContentType:      &mime,
				SizeBytes:        &size,
			}); err != nil {
				return err
			}
			if err = s.Store.InsertParceirosAttachmentLink(ctx, importID, file.externalID, file.attachmentID); err != nil {
				return err
			}
		}
		result = parceiros.IntakeResult{
			CustomerID: customerID, CustomerCreated: created, PhoneLineID: lineID, PhoneLineCreated: lineCreated,
			AttachmentsAccepted: accepted, AttachmentsSkipped: skipped, Warnings: prepared.Warnings,
		}
		return nil
	})
	return result, err
}

type parceirosStoredFile struct {
	externalID   string
	attachmentID string
	title        string
	fileName     string
	contentType  string
	size         int64
	objectKey    string
}

func (s *Service) storeParceirosFiles(orgID, saleID string, attachments []parceiros.Attachment, warnings *[]string) ([]parceirosStoredFile, int, error) {
	root := s.ParceirosUploadDir
	if strings.TrimSpace(root) == "" {
		root = "./parceiros-uploads"
	}
	var files []parceirosStoredFile
	skipped := 0
	seen := map[string]bool{}
	for i, attachment := range attachments {
		externalID := strings.TrimSpace(attachment.ID)
		if externalID == "" {
			externalID = fmt.Sprintf("idx-%d", i)
		}
		if seen[externalID] {
			*warnings = append(*warnings, "Anexo repetido na mesma venda. A segunda cópia não entrou.")
			skipped++
			continue
		}
		seen[externalID] = true
		data, warning, ok := parceiros.DecodeAttachment(attachment.ContentBase64)
		if !ok {
			*warnings = append(*warnings, warning)
			skipped++
			continue
		}
		attachmentID := uuid.NewString()
		ext := attachmentExtension(attachment.MimeType, attachment.Name)
		objectKey := path.Join(orgID, saleID, attachmentID+ext)
		full := filepath.Join(root, filepath.FromSlash(objectKey))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return nil, skipped, err
		}
		if err := os.WriteFile(full, data, 0o644); err != nil {
			return nil, skipped, err
		}
		name := strings.TrimSpace(attachment.Name)
		if name == "" {
			name = attachmentID + ext
		}
		title := strings.TrimSpace(attachment.Type)
		if title == "" {
			title = "OTHER"
		}
		mime := strings.TrimSpace(attachment.MimeType)
		if mime == "" {
			mime = "application/octet-stream"
		}
		files = append(files, parceirosStoredFile{
			externalID: externalID, attachmentID: attachmentID, title: title,
			fileName: name, contentType: mime, size: int64(len(data)), objectKey: objectKey,
		})
	}
	return files, skipped, nil
}

func (s *Service) upsertParceirosCustomer(ctx context.Context, orgID string, prepared *parceiros.Prepared) (string, bool, error) {
	if prepared.LookupDocument {
		existing, err := s.Store.FindCustomerIDByDocument(ctx, orgID, prepared.DocumentDigits)
		if err != nil {
			return "", false, err
		}
		if existing != "" {
			prepared.Warnings = append(prepared.Warnings, "Cliente já existia pelo documento. A ficha manual foi mantida.")
			return existing, false, nil
		}
	}
	customerID := uuid.NewString()
	legalName := ""
	if prepared.CustomerType == "pj" {
		legalName = prepared.CustomerName
	}
	if err := s.Store.InsertParceirosCustomer(ctx, orgID, customerID, prepared.CustomerType, prepared.CustomerName, legalName, prepared.RawAddress, prepared.RawDocument, prepared.ReviewPending); err != nil {
		return "", false, err
	}
	if prepared.LookupDocument {
		if err := s.Store.InsertParceirosDocument(ctx, customerID, prepared.DocumentKind, prepared.DocumentDigits); err != nil {
			return "", false, err
		}
	}
	if rg := clip(&prepared.Warnings, "RG", prepared.RG, 128); rg != "" {
		if err := s.Store.InsertParceirosDocument(ctx, customerID, "rg", rg); err != nil {
			return "", false, err
		}
	}
	if prepared.StructuredAddress != nil {
		addr := prepared.StructuredAddress
		var complement *string
		if addr.Complement != "" {
			complement = &addr.Complement
		}
		if err := s.Store.CreateCustomerAddress(ctx, customerID, models.CreateCustomerAddressInput{
			Street: addr.Street, Number: addr.Number, Neighborhood: addr.Neighborhood,
			City: addr.City, State: addr.State, ZipCode: addr.ZipCode, Complement: complement, Country: "Brasil",
		}); err != nil {
			return "", false, err
		}
	}
	profile := &models.CustomerRegistrationProfile{RG: prepared.RG}
	if prepared.Phone != "" {
		profile.Phone = prepared.Phone
		profile.Mobile = prepared.Phone
	}
	if prepared.Email != "" {
		profile.Emails = []string{prepared.Email}
	}
	if prepared.Phone != "" || prepared.Email != "" || prepared.RG != "" {
		if err := s.Store.SetCustomerRegistration(ctx, customerID, profile, nil); err != nil {
			return "", false, err
		}
	}
	return customerID, true, nil
}

func (s *Service) upsertParceirosLine(ctx context.Context, orgID, customerID string, input parceiros.SaleIntake, prepared *parceiros.Prepared) (*string, bool, error) {
	if prepared.LineDigits == "" {
		return nil, false, nil
	}
	found, err := s.Store.FindParceirosPhoneLine(ctx, prepared.LineDigits)
	if err != nil {
		return nil, false, err
	}
	if found != nil && found.OrganizationID != orgID {
		prepared.Warnings = append(prepared.Warnings, "Esse número já existe em outra organização. A linha não foi reaproveitada.")
		return nil, false, nil
	}
	if found != nil {
		owner, err := s.Store.ActivePhoneLineCustomer(ctx, found.ID)
		if err != nil {
			return nil, false, err
		}
		if owner != "" && owner != customerID {
			prepared.Warnings = append(prepared.Warnings, "Linha já vinculada a outro cliente. A venda entrou e o vínculo atual foi mantido.")
			return &found.ID, false, nil
		}
		if owner == "" {
			if err = s.Store.LinkParceirosPhoneLine(ctx, found.ID, customerID, parceirosLineStart(prepared)); err != nil {
				return nil, false, err
			}
		}
		if err = s.updateParceirosLineDetails(ctx, found.ID, input, prepared); err != nil {
			return nil, false, err
		}
		return &found.ID, false, nil
	}

	shell, err := s.Store.EnsureParceirosShell(ctx, orgID)
	if err != nil {
		return nil, false, err
	}
	lineID := uuid.NewString()
	var iccid *string
	if trimmed := clip(&prepared.Warnings, "ICCID", prepared.Iccid, 64); trimmed != "" {
		iccid = &trimmed
	}
	original := clip(&prepared.Warnings, "Número da linha", strings.TrimSpace(input.Sale.LineNumber), 64)
	if err = s.Store.CreatePhoneLineWithIdentity(ctx, lineID, shell.PlanID, shell.AccountID, prepared.LineDigits, original, prepared.LineDigits, prepared.SimType, iccid); err != nil {
		return nil, false, err
	}
	if err = s.updateParceirosLineDetails(ctx, lineID, input, prepared); err != nil {
		return nil, false, err
	}
	if err = s.Store.LinkParceirosPhoneLine(ctx, lineID, customerID, parceirosLineStart(prepared)); err != nil {
		return nil, false, err
	}
	return &lineID, true, nil
}

func (s *Service) updateParceirosLineDetails(ctx context.Context, lineID string, input parceiros.SaleIntake, prepared *parceiros.Prepared) error {
	return s.Store.UpdateParceirosPhoneLine(ctx, lineID,
		clip(&prepared.Warnings, "Operadora", input.Sale.OperatorName, 128),
		clip(&prepared.Warnings, "Plano", input.Sale.PlanName, 128),
		prepared.SimType,
		clip(&prepared.Warnings, "ICCID", prepared.Iccid, 64),
		clip(&prepared.Warnings, "IMEI", prepared.DeviceImei, 32),
		clip(&prepared.Warnings, "EID", prepared.DeviceEid, 64),
		clip(&prepared.Warnings, "Plataforma", prepared.DevicePlatform, 16),
		clip(&prepared.Warnings, "Número da portabilidade", input.Sale.PortabilityNumber, 32),
		clip(&prepared.Warnings, "Operadora de origem", input.Sale.DonorOperator, 32),
		input.Sale.IsVirginChip, input.Sale.IsPortability)
}

func parceirosLineStart(prepared *parceiros.Prepared) time.Time {
	if prepared.ConcludedAt != nil {
		return *prepared.ConcludedAt
	}
	if prepared.SoldAt != nil {
		return *prepared.SoldAt
	}
	return time.Now().UTC()
}

func clip(warnings *[]string, label, value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if utf8.RuneCountInString(trimmed) <= max {
		return trimmed
	}
	*warnings = append(*warnings, label+" foi encurtado para caber no cadastro.")
	return string([]rune(trimmed)[:max])
}

func intakeResultFromRow(row *store.ParceirosImportRow) parceiros.IntakeResult {
	warnings := row.Warnings
	if warnings == nil {
		warnings = []string{}
	}
	return parceiros.IntakeResult{
		CustomerID: row.CustomerID, CustomerCreated: row.CustomerCreated,
		PhoneLineID: row.PhoneLineID, PhoneLineCreated: row.PhoneLineCreated,
		AttachmentsAccepted: row.AttachmentsAccepted, AttachmentsSkipped: row.AttachmentsSkipped,
		Warnings: warnings,
	}
}

func attachmentExtension(mime, name string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" || len(ext) > 8 || strings.Contains(ext, "/") || strings.Contains(ext, "\\") {
		return ""
	}
	return ext
}
