package importservice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/observability"
	"github.com/luxus-connect/telefonia/api/internal/services"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"github.com/luxus-connect/telefonia/api/internal/vivo"
)

type ObjectGetter interface {
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)
}

type Processor struct {
	Store   *store.Store
	Storage ObjectGetter
	Log     *slog.Logger
	SM      *statemachine.Engine
}

func (p *Processor) engine() *statemachine.Engine {
	if p.SM == nil {
		return statemachine.NewEngine(p.Store)
	}
	return p.SM
}

func (p *Processor) ProcessImport(ctx context.Context, importRequestID string) error {
	start := time.Now()
	var processErr error
	err := p.Store.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = store.CtxWithTx(ctx, tx)
		// Lock the request until both invoice data and terminal status are committed.
		// A redelivery waits for this transaction and never processes the same request twice.
		req, err := p.Store.GetImportRequestForUpdate(ctx, importRequestID)
		if err != nil {
			return err
		}
		if req == nil {
			return httputil.BusinessError(notifications.ImportRequestNotFound)
		}
		if req.Status == 2 {
			return nil
		}
		if req.Status != 0 {
			return httputil.BusinessError(notifications.ImportRequestNotPending)
		}
		// Serialize account creation and replacement checks for this provider.
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", "invoice-import:"+req.ProviderID); err != nil {
			return err
		}
		work, err := tx.Begin(ctx)
		if err != nil {
			return err
		}
		processErr = p.processInner(store.CtxWithTx(ctx, work), req)
		if processErr != nil {
			if err := work.Rollback(ctx); err != nil {
				return err
			}
		} else if err := work.Commit(ctx); err != nil {
			return err
		}
		now := time.Now().UTC()
		status := 2
		var message *string
		if processErr != nil {
			status = 3
			if isPDFUnparsed(processErr) {
				status = 4
			}
			msg := importErrorMessage(processErr)
			message = &msg
			if p.Log != nil {
				p.Log.Error("invoice import failed", "request_id", importRequestID, "error", processErr)
			}
		}
		return p.Store.UpdateImportRequestStatus(ctx, importRequestID, status, message, &now)
	})
	observability.Observe("import.process", time.Since(start), err == nil && processErr == nil)
	if err != nil {
		return err
	}
	return processErr
}

func importErrorMessage(err error) string {
	var app *httputil.AppError
	if errors.As(err, &app) && app.Kind != httputil.ErrorInternal {
		return app.Error()
	}
	return "Não foi possível processar a fatura. Tente novamente ou contate o suporte."
}

func (p *Processor) processInner(ctx context.Context, req *store.ImportRequestRow) error {
	if p.Storage == nil {
		return httputil.BusinessError(notifications.ObjectStorageUnavailable)
	}
	raw, err := p.Storage.GetObject(ctx, req.StorageBucket, req.StorageObjectKey)
	if err != nil {
		return fmt.Errorf("storage get: %w", err)
	}
	if isPDFBytes(raw) {
		return httputil.BusinessError(notifications.ImportPDFNotParsed)
	}
	sum := sha256.Sum256(raw)
	fileHash := hex.EncodeToString(sum[:])

	parsed, err := vivo.ParseLatin1(raw)
	if err != nil {
		return err
	}

	header := getHeader(parsed)
	if header == nil {
		return httputil.BusinessError(notifications.ImportHeaderMissing)
	}
	customer011 := getCustomer011(parsed)
	if customer011 == nil {
		return fmt.Errorf("missing 011D customer record")
	}

	taxID := httputil.NormalizeDigits(customer011.Document)
	if len(taxID) != 11 && len(taxID) != 14 {
		return httputil.BusinessError(notifications.ImportCustomerDocumentInvalid)
	}

	orgID, _, err := p.Store.GetProviderByID(ctx, req.ProviderID)
	if err != nil {
		return err
	}

	if orgID != req.OrganizationID {
		return httputil.BusinessError(notifications.ProviderNotFound)
	}
	company, account, month, cycle, err := p.resolveContext(ctx, orgID, req, parsed, header, taxID, customer011)
	if err != nil {
		return err
	}

	if existingHash, err := p.Store.FindActiveInvoiceByContentSHA256(ctx, fileHash); err != nil {
		return err
	} else if existingHash != nil {
		return httputil.BusinessError(notifications.InvoiceDuplicateFileHash)
	}

	existingKey, err := p.Store.FindActiveInvoiceByBusinessKey(ctx, account.ID, month.ID, header.DueDate)
	if err != nil {
		return err
	}

	var parentID *string
	var impact *float64
	if existingKey != nil {
		if !req.AllowSubstitute {
			return httputil.BusinessError(notifications.InvoiceDuplicateSameProcessingMonth)
		}
		if err := p.Store.MarkInvoiceSubstituted(ctx, existingKey.ID); err != nil {
			return err
		}
		parentID = &existingKey.ID
		delta := header.TotalAmount - existingKey.TotalAmount
		impact = &delta
	}

	otherMonth, err := p.Store.InvoiceExistsInOtherProcessingMonth(ctx, account.ID, company.ID, month.ID, header.DueDate)
	if err != nil {
		return err
	}
	if otherMonth {
		return httputil.BusinessError(notifications.InvoiceDuplicateOtherProcessingMonth)
	}

	numbersInFile := buildNumbersFrom110D(parsed)
	importCustomer, err := p.resolveCustomer(ctx, orgID, req.ProviderID, company, taxID, parsed)
	if err != nil {
		return err
	}

	invoiceID := uuid.New().String()
	var digitableLine, pixQrCode *string
	for _, record := range parsed {
		if payment, ok := record.(*vivo.Line020DPayment); ok {
			if value := strings.TrimSpace(payment.DigitableLine); value != "" {
				digitableLine = &value
			}
			if value := strings.TrimSpace(payment.PixQrCode); value != "" {
				pixQrCode = &value
			}
			break
		}
	}
	inv := store.ProviderInvoiceInsert{
		ID: invoiceID, Number: header.ReferenceMonth,
		ProviderAccountID: account.ID, ContractingCompanyID: company.ID,
		BillingCycleID: cycle.ID, ProcessingMonthID: month.ID,
		IssueDate: header.IssueDate, DueDate: header.DueDate, TotalAmount: header.TotalAmount,
		SubtotalServices: header.SubtotalServices, SubtotalUsage: header.SubtotalUsageExceeded,
		ParentInvoiceID: parentID, ContentSHA256: &fileHash, SubstitutionImpact: impact,
		DigitableLine: digitableLine, PixQrCode: pixQrCode,
	}

	if err := p.Store.CreateProviderInvoice(ctx, inv); err != nil {
		return err
	}

	if err := p.processInvoiceServices(ctx, req.ProviderID, invoiceID, parsed, header); err != nil {
		return err
	}

	if err := p.processLines(ctx, orgID, req.ProviderID, account.ID, invoiceID, parsed, importCustomer, numbersInFile, header); err != nil {
		return err
	}

	if err := p.applyAbsentLines(ctx, orgID, account.ID, invoiceID, numbersInFile); err != nil {
		return err
	}

	if err := p.applyAutomaticExceedances(ctx, orgID, invoiceID, parsed); err != nil {
		return err
	}

	return nil
}

func (p *Processor) resolveContext(ctx context.Context, orgID string, req *store.ImportRequestRow, parsed []any,
	header *vivo.Line010DHeader, taxID string, customer011 *vivo.Line011DCustomer) (*store.ContractingCompanyRow, *store.ProviderAccountRow, *store.ProcessingMonthRow, *models.ListBillingCycleResponse, error) {

	var company *store.ContractingCompanyRow
	var err error

	if len(taxID) == 14 {
		company, err = p.Store.GetContractingCompanyByTaxID(ctx, req.ProviderID, taxID)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if company == nil {
			id := uuid.New().String()
			legalName := resolveLegalName(customer011)
			if err := p.Store.CreateContractingCompany(ctx, id, req.ProviderID, legalName, taxID); err != nil {
				return nil, nil, nil, nil, err
			}
			company = &store.ContractingCompanyRow{ID: id, ProviderID: req.ProviderID, LegalName: legalName, TaxID: taxID}
		}
	} else {
		customers, err := p.Store.ListCustomersByDocument(ctx, orgID, taxID)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if len(customers) != 1 {
			return nil, nil, nil, nil, httputil.BusinessError(notifications.ImportCPFRequiresExistingCustomer)
		}
		cnpj, err := p.Store.GetCustomerCNPJ(ctx, customers[0])
		if err != nil || cnpj == "" {
			return nil, nil, nil, nil, httputil.BusinessError(notifications.CustomerContractingCompanyMismatch)
		}
		cnpj = httputil.NormalizeDigits(cnpj)
		company, err = p.Store.GetContractingCompanyByTaxID(ctx, req.ProviderID, cnpj)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if company == nil {
			return nil, nil, nil, nil, httputil.BusinessError(notifications.ImportContractingCompanyNotFound)
		}
	}

	account, err := p.Store.GetProviderAccount(ctx, company.ID, header.AccountNumber)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if account == nil {
		id := uuid.New().String()
		if err := p.Store.CreateProviderAccount(ctx, id, company.ID, header.AccountNumber); err != nil {
			return nil, nil, nil, nil, err
		}
		account = &store.ProviderAccountRow{ID: id, ContractingCompanyID: company.ID, AccountNumber: header.AccountNumber}
	}

	cycle, err := p.Store.GetBillingCycleByCode(ctx, req.ProviderID, header.ReferenceMonth)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if cycle == nil {
		blocked, err := p.Store.ExistsClosedProcessingMonthIntersecting(ctx, orgID, req.ProviderID, header.BillingStartDate, header.BillingEndDate)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if blocked {
			return nil, nil, nil, nil, httputil.BusinessError(notifications.ProcessingMonthRetroactiveBlocked)
		}
		id := uuid.New().String()
		name := header.BillingEndDate.Format("January 2006")
		if err := p.Store.CreateBillingCycle(ctx, orgID, id, req.ProviderID, header.ReferenceMonth, name, header.BillingStartDate, header.BillingEndDate); err != nil {
			return nil, nil, nil, nil, err
		}
		cycle, err = p.Store.GetBillingCycleByCode(ctx, req.ProviderID, header.ReferenceMonth)
		if err != nil {
			return nil, nil, nil, nil, err
		}
	}

	month, err := p.Store.GetProcessingMonth(ctx, orgID, req.ProcessingMonthID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if month == nil {
		return nil, nil, nil, nil, httputil.BusinessError(notifications.ProcessingMonthNotFound)
	}
	if month.ProviderID != req.ProviderID {
		return nil, nil, nil, nil, httputil.BusinessError(notifications.ProcessingMonthProviderMismatch)
	}
	if month.Status != "open" {
		return nil, nil, nil, nil, httputil.BusinessError(notifications.ProcessingMonthNotOpen)
	}

	return company, account, month, cycle, nil
}

func (p *Processor) resolveCustomer(ctx context.Context, orgID, providerID string, company *store.ContractingCompanyRow, taxID string, parsed []any) (string, error) {
	customer011 := getCustomer011(parsed)
	if customer011 == nil {
		return "", nil
	}
	doc := httputil.NormalizeDigits(customer011.Document)
	if doc == "" {
		return "", nil
	}
	ids, err := p.Store.ListCustomersByDocument(ctx, orgID, doc)
	if err != nil {
		return "", err
	}
	var matches []string
	for _, id := range ids {
		ok, err := p.Store.CustomerHasActiveProvider(ctx, id, providerID)
		if err != nil {
			return "", err
		}
		if ok {
			matches = append(matches, id)
		}
	}
	if len(matches) == 1 {
		cnpj, err := p.Store.GetCustomerCNPJ(ctx, matches[0])
		if err != nil {
			return "", err
		}
		if cnpj != "" && httputil.NormalizeDigits(cnpj) != company.TaxID {
			return "", httputil.BusinessError(notifications.CustomerContractingCompanyMismatch)
		}
		return matches[0], nil
	}
	return "", nil
}

func (p *Processor) processInvoiceServices(ctx context.Context, providerID, invoiceID string, parsed []any, header *vivo.Line010DHeader) error {
	for _, rec := range parsed {
		var planCode, name string
		var qty int
		var total, franchise, used float64
		switch svc := rec.(type) {
		case *vivo.Line050HService:
			planCode, name, qty, franchise, used, total = svc.PlanCode, svc.ServiceName, svc.Quantity, svc.Franchise, svc.Used, svc.Total
		case *vivo.Line050GService:
			planCode, name, qty, franchise, used, total = svc.PlanCode, svc.ServiceName, svc.Quantity, svc.Franchise, svc.Used, svc.Total
		default:
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		plan, err := p.resolvePlan(ctx, providerID, planCode)
		if err != nil {
			return err
		}
		if plan == nil {
			return httputil.BusinessError(notifications.N("IMPORT_PLAN_MISSING", "Há serviços ou linhas sem plano no arquivo. Revise o TXT da operadora."))
		}
		qtyF := float64(qty)
		if qtyF <= 0 {
			qtyF = 1
		}
		if err := p.Store.CreateInvoiceService(ctx, store.InvoiceServiceInsert{
			ID:             uuid.New().String(),
			InvoiceID:      invoiceID,
			PlanID:         plan.ID,
			Description:    name,
			Quantity:       qtyF,
			TotalPrice:     total,
			QuotaAmount:    &franchise,
			ConsumedAmount: &used,
		}); err != nil {
			return err
		}
	}
	if header.SubtotalUsageExceeded != 0 {
		plan, err := p.resolvePlan(ctx, providerID, "CONSUMO")
		if err != nil {
			return err
		}
		if plan != nil {
			usage := header.SubtotalUsageExceeded
			if err := p.Store.CreateInvoiceService(ctx, store.InvoiceServiceInsert{
				ID:          uuid.New().String(),
				InvoiceID:   invoiceID,
				PlanID:      plan.ID,
				Description: "Consumo",
				Quantity:    1,
				TotalPrice:  usage,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Processor) changeLineStatus(ctx context.Context, orgID, lineID, from, to, trigger string) error {
	if from == to {
		return nil
	}
	if err := validateImportLineTransition(p.engine(), from, to, importEnforceStateMachine()); err != nil {
		return err
	}
	if err := p.Store.UpdatePhoneLineStatus(ctx, lineID, to); err != nil {
		return err
	}
	if err := p.engine().RecordTransition(ctx, orgID, statemachine.EntityPhoneLine, lineID, from, to, trigger, nil, nil, nil); err != nil {
		return err
	}
	return nil
}

func (p *Processor) processLines(ctx context.Context, orgID, providerID, accountID, invoiceID string, parsed []any, customerID string, numbers map[string]struct{}, header *vivo.Line010DHeader) error {
	consumptionByLine := make(map[string]float64)
	for _, rec := range parsed {
		if detail, ok := rec.(*vivo.InvoiceFranchiseLineDetail); ok {
			number := httputil.NormalizeDigits(detail.PhoneNumber)
			if number != "" && detail.UsageAmount > 0 {
				consumptionByLine[number] += detail.UsageAmount
			}
		}
	}
	seen := map[string]struct{}{}
	activation := header.IssueDate
	if !header.BillingStartDate.IsZero() {
		activation = header.BillingStartDate
	}
	for _, rec := range parsed {
		line, ok := rec.(*vivo.Line110DAccountLineDetail)
		if !ok {
			continue
		}
		numberKey := httputil.NormalizeDigits(line.PhoneNumber)
		if numberKey == "" {
			continue
		}
		if _, dup := seen[numberKey]; dup {
			continue
		}
		seen[numberKey] = struct{}{}

		plan, err := p.resolvePlan(ctx, providerID, line.PlanName)
		if err != nil {
			return err
		}
		if plan == nil {
			return httputil.BusinessError(notifications.N("IMPORT_PLAN_MISSING", "Há serviços ou linhas sem plano no arquivo. Revise o TXT da operadora."))
		}

		pl, err := p.Store.GetPhoneLineByNumber(ctx, numberKey)
		if err != nil {
			return err
		}
		if pl != nil && pl.ProviderAccountID != accountID {
			return fmt.Errorf("line %s linked to another account", numberKey)
		}
		created := false
		if pl == nil {
			id := uuid.New().String()
			if err := p.Store.CreatePhoneLine(ctx, id, plan.ID, accountID, numberKey); err != nil {
				return err
			}
			pl = &store.PhoneLineRow{ID: id, Number: numberKey, ProviderAccountID: accountID, ProviderPlanID: plan.ID, Status: "in_stock"}
			created = true
			if err := p.auditImport(ctx, "Create", pl.ID, map[string]any{
				"message": fmt.Sprintf("Linha criada em estoque automaticamente a partir da fatura %s / %s.", header.ReferenceMonth, header.DueDate.Format("02/01/2006")),
				"number":  numberKey,
			}); err != nil {
				return err
			}
		}

		if pl.Status == "cancelled" || pl.Status == "suspended" {
			return httputil.BusinessError(notifications.InvoiceImportedLineOrphanDestination)
		}

		_, activeCustomer, err := p.Store.GetActivePhoneLineCustomerLink(ctx, pl.ID)
		if err != nil {
			return err
		}
		if customerID != "" && activeCustomer == "" {
			if err := p.Store.AssignPhoneLineCustomer(ctx, pl.ID, customerID, header.IssueDate, nil); err != nil {
				return err
			}
			if err := p.Store.AddCustomerProviderLink(ctx, customerID, providerID, header.IssueDate); err != nil {
				return err
			}
			if err := p.Store.ReactivateCustomer(ctx, customerID); err != nil {
				return err
			}
			activeCustomer = customerID
		}

		prevStatus := pl.Status
		target := prevStatus
		switch {
		case prevStatus == "in_transition" || prevStatus == "awaiting_invoice":
			target = "active"
		case activeCustomer != "":
			target = "active"
		case prevStatus == "inactive":
			target = "in_stock"
		default:
			if !created {
				target = "in_stock"
			}
		}
		if target == "active" && prevStatus != "active" {
			if err := validateImportLineTransition(p.engine(), prevStatus, "active", importEnforceStateMachine()); err != nil {
				return err
			}
			if err := p.Store.ActivatePhoneLineFromInvoice(ctx, pl.ID, activation); err != nil {
				return err
			}
			if err := p.engine().RecordTransition(ctx, orgID, statemachine.EntityPhoneLine, pl.ID, prevStatus, "active", "import_invoice", nil, nil, nil); err != nil {
				return err
			}
			if prevStatus == "in_transition" || prevStatus == "awaiting_invoice" {
				if err := p.auditImport(ctx, "Reconcile", pl.ID, map[string]any{
					"message":         fmt.Sprintf("Linha %s conciliada automaticamente. Status: Ativa desde %s.", numberKey, activation.Format("02/01/2006")),
					"previous_status": prevStatus,
					"activation_date": activation.Format("2006-01-02"),
				}); err != nil {
					return err
				}
			}
		} else if target != prevStatus {
			if err := p.changeLineStatus(ctx, orgID, pl.ID, prevStatus, target, "import_invoice"); err != nil {
				return err
			}
			if target == "in_stock" && prevStatus == "inactive" {
				if err := p.auditImport(ctx, "ReactivateStock", pl.ID, map[string]any{
					"message": fmt.Sprintf("Linha %s retornou ao estoque após reaparecer na fatura.", numberKey),
				}); err != nil {
					return err
				}
			}
		}

		withConsumption := line.LineTotal
		consumption := consumptionByLine[numberKey]
		base := withConsumption - consumption
		if base < 0 {
			base = 0
		}
		if err := p.Store.UpdatePhoneLineCosts(ctx, pl.ID, base, withConsumption, invoiceID); err != nil {
			return err
		}
		if err := p.Store.LinkInvoicePhoneLine(ctx, invoiceID, pl.ID); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) applyAbsentLines(ctx context.Context, orgID, accountID, invoiceID string, numbersInFile map[string]struct{}) error {
	lines, err := p.Store.ListPhoneLinesByAccount(ctx, accountID)
	if err != nil {
		return err
	}
	for _, line := range lines {
		key := httputil.NormalizeDigits(line.Number)
		if key == "" {
			continue
		}
		if _, present := numbersInFile[key]; present {
			continue
		}
		_, activeCustomer, err := p.Store.GetActivePhoneLineCustomerLink(ctx, line.ID)
		if err != nil {
			return err
		}
		if activeCustomer == "" {
			if line.Status == "in_stock" || line.Status == "active" {
				if err := p.changeLineStatus(ctx, orgID, line.ID, line.Status, "inactive", "import_absent"); err != nil {
					return err
				}
				if err := p.auditImport(ctx, "InactivateStock", line.ID, map[string]any{
					"message":         fmt.Sprintf("Linha %s inativada no estoque por ausência na fatura. Última fatura preservada.", line.Number),
					"last_invoice_id": invoiceID,
					"previous_status": line.Status,
				}); err != nil {
					return err
				}
			}
		} else {
			if err := p.changeLineStatus(ctx, orgID, line.ID, line.Status, "awaiting_invoice", "import_absent"); err != nil {
				return err
			}
			if err := p.auditImport(ctx, "AwaitingInvoice", line.ID, map[string]any{
				"message":         fmt.Sprintf("Linha %s ausente na fatura. Status: Aguardando fatura.", line.Number),
				"previous_status": line.Status,
				"customer_id":     activeCustomer,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Processor) auditImport(ctx context.Context, changeType, phoneLineID string, payload map[string]any) error {
	var newStr *string
	if b, err := json.Marshal(payload); err == nil {
		s := string(b)
		newStr = &s
	}
	system := "import"
	if err := p.Store.InsertAuditLog(ctx, uuid.New().String(), changeType, "PhoneLine", phoneLineID, &system, nil, newStr, time.Now().UTC(), ""); err != nil {
		return err
	}
	if p.Log != nil {
		p.Log.Info("import matrix", "change", changeType, "phone_line_id", phoneLineID, "payload", payload)
	}
	return nil
}

func (p *Processor) resolvePlan(ctx context.Context, providerID, planCode string) (*store.ProviderPlanRow, error) {
	if strings.TrimSpace(planCode) == "" {
		return nil, nil
	}
	plan, err := p.Store.GetPlanByProviderAndCode(ctx, providerID, planCode)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		return plan, nil
	}
	id := uuid.New().String()
	if err := p.Store.CreateProviderPlan(ctx, id, providerID, planCode, planCode); err != nil {
		return nil, err
	}
	return &store.ProviderPlanRow{ID: id, ProviderID: providerID, Code: planCode, Name: planCode}, nil
}

func getHeader(parsed []any) *vivo.Line010DHeader {
	for _, rec := range parsed {
		if h, ok := rec.(*vivo.Line010DHeader); ok {
			return h
		}
	}
	return nil
}

func getCustomer011(parsed []any) *vivo.Line011DCustomer {
	for _, rec := range parsed {
		if c, ok := rec.(*vivo.Line011DCustomer); ok {
			return c
		}
	}
	return nil
}

func buildNumbersFrom110D(parsed []any) map[string]struct{} {
	set := map[string]struct{}{}
	for _, rec := range parsed {
		if l, ok := rec.(*vivo.Line110DAccountLineDetail); ok {
			n := httputil.NormalizeDigits(l.PhoneNumber)
			if n != "" {
				set[n] = struct{}{}
			}
		}
	}
	return set
}

func resolveLegalName(c *vivo.Line011DCustomer) string {
	if c.LegalName != "" {
		return strings.TrimSpace(c.LegalName)
	}
	if c.Name != "" {
		return strings.TrimSpace(c.Name)
	}
	return "Empresa não identificada"
}

func isPDFBytes(raw []byte) bool {
	trimmed := bytes.TrimSpace(raw)
	return bytes.HasPrefix(trimmed, []byte("%PDF"))
}

func isPDFUnparsed(err error) bool {
	var app *httputil.AppError
	if errors.As(err, &app) {
		for _, n := range app.Notifications {
			if n.Code == "IMPORT_PDF_NOT_PARSED" {
				return true
			}
		}
	}
	return strings.Contains(strings.ToLower(err.Error()), "pdf ainda não parseado")
}

type extraHit struct {
	phoneNumber string
	description string
	amount      float64
}

func collectExtraHits(parsed []any) []extraHit {
	var hits []extraHit
	for _, rec := range parsed {
		switch item := rec.(type) {
		case *vivo.InvoiceFranchiseLineDetail:
			hits = append(hits, extraHit{phoneNumber: item.PhoneNumber, description: item.ServiceDescription, amount: item.UsageAmount})
		case *vivo.Line052DExtraUsageDetail:
			hits = append(hits, extraHit{description: item.Description, amount: item.Amount})
		}
	}
	return hits
}

func (p *Processor) applyAutomaticExceedances(ctx context.Context, orgID, invoiceID string, parsed []any) error {
	terms, err := p.Store.ListExceedanceTerms(ctx, orgID, true)
	if err != nil || len(terms) == 0 {
		return err
	}
	now := time.Now().UTC()
	for _, hit := range collectExtraHits(parsed) {
		term := services.MatchExceedanceTerm(hit.description, terms)
		if term == nil {
			continue
		}
		var phoneLineID *string
		if n := httputil.NormalizeDigits(hit.phoneNumber); n != "" {
			pl, err := p.Store.GetPhoneLineByNumber(ctx, n)
			if err != nil {
				return err
			}
			if pl == nil {
				continue
			}
			phoneLineID = &pl.ID
			settings, err := p.Store.GetPhoneLineExceedanceSettings(ctx, pl.ID)
			if err != nil {
				return err
			}
			if settings == nil || !settings.ChargeExceedances {
				continue
			}
			charged, chargeType := services.ChargedExceedanceAmount(hit.amount, settings.ExceedanceChargeType, term)
			inserted, err := p.Store.InsertDetectedExceedance(ctx, store.DetectedExceedanceRow{
				ID: uuid.New().String(), InvoiceID: invoiceID, PhoneLineID: phoneLineID, TermID: &term.ID,
				Term: term.Term, Description: hit.description, InvoiceAmount: hit.amount, ChargedAmount: charged,
				ChargeType: chargeType, Applied: charged > 0, CreatedAt: now,
			})
			if err != nil {
				return err
			}
			if !inserted || charged <= 0 {
				continue
			}
			procID, err := p.Store.GetPrimaryProcessingIDForLine(ctx, pl.ID)
			if err != nil {
				return err
			}
			if procID == "" {
				continue
			}
			item := store.BillingCompositionItemRow{
				ID: uuid.New().String(), ProcessingID: procID, ItemType: "exceedance",
				Description: term.Term + " — " + strings.TrimSpace(hit.description),
				Amount:      charged, Quantity: 1, Active: true, CreatedAt: now, UpdatedAt: now, Proportional: false,
			}
			if err := p.Store.CreateBillingCompositionItem(ctx, item); err != nil {
				return err
			}
			secondary, err := p.Store.GetMirroredSecondaryProcessingID(ctx, pl.ID)
			if err != nil {
				return err
			}
			if secondary != "" {
				copy := item
				copy.ID = uuid.New().String()
				copy.ProcessingID = secondary
				if err := p.Store.CreateBillingCompositionItem(ctx, copy); err != nil {
					return err
				}
			}
			continue
		}
		_, err := p.Store.InsertDetectedExceedance(ctx, store.DetectedExceedanceRow{
			ID: uuid.New().String(), InvoiceID: invoiceID, TermID: &term.ID,
			Term: term.Term, Description: hit.description, InvoiceAmount: hit.amount, ChargedAmount: 0,
			ChargeType: term.ChargeType, Applied: false, CreatedAt: now,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
