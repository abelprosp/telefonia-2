package importservice

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
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
}

func (p *Processor) ProcessImport(ctx context.Context, importRequestID string) error {
	req, err := p.Store.GetImportRequest(ctx, importRequestID)
	if err != nil {
		return err
	}
	if req == nil {
		return httputil.BusinessError(notifications.ImportRequestNotFound)
	}
	if req.Status != 0 {
		return httputil.BusinessError(notifications.ImportRequestNotPending)
	}
	now := time.Now().UTC()
	_ = p.Store.UpdateImportRequestStatus(ctx, importRequestID, 1, nil, nil)

	processErr := p.process(ctx, req)
	if processErr != nil {
		msg := processErr.Error()
		_ = p.Store.UpdateImportRequestStatus(ctx, importRequestID, 3, &msg, &now)
		return processErr
	}
	_ = p.Store.UpdateImportRequestStatus(ctx, importRequestID, 2, nil, &now)
	return nil
}

func (p *Processor) process(ctx context.Context, req *store.ImportRequestRow) error {
	raw, err := p.Storage.GetObject(ctx, req.StorageBucket, req.StorageObjectKey)
	if err != nil {
		return fmt.Errorf("storage get: %w", err)
	}
	parsed, err := vivo.ParseLatin1(raw)
	if err != nil {
		return err
	}

	header := getHeader(parsed)
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

	company, account, month, cycle, err := p.resolveContext(ctx, orgID, req, parsed, header, taxID, customer011)
	if err != nil {
		return err
	}

	dup, err := p.Store.InvoiceDuplicateExists(ctx, account.ID, company.ID, month.ID, header.DueDate)
	if err != nil {
		return err
	}
	if dup {
		return httputil.BusinessError(notifications.InvoiceDuplicateSameProcessingMonth)
	}

	numbersInFile := buildNumbersFrom110D(parsed)
	importCustomer, err := p.resolveCustomer(ctx, orgID, req.ProviderID, company, taxID, parsed)
	if err != nil {
		return err
	}

	invoiceID := uuid.New().String()
	inv := store.ProviderInvoiceInsert{
		ID: invoiceID, Number: header.ReferenceMonth,
		ProviderAccountID: account.ID, ContractingCompanyID: company.ID,
		BillingCycleID: cycle.ID, ProcessingMonthID: month.ID,
		IssueDate: header.IssueDate, DueDate: header.DueDate, TotalAmount: header.TotalAmount,
		SubtotalServices: header.SubtotalServices, SubtotalUsage: header.SubtotalUsageExceeded,
	}

	if err := p.Store.CreateProviderInvoice(ctx, inv); err != nil {
		return err
	}

	if err := p.processLines(ctx, req.ProviderID, account.ID, invoiceID, parsed, importCustomer, numbersInFile, header); err != nil {
		return err
	}

	if err := p.applyAbsentLines(ctx, account.ID, invoiceID, numbersInFile); err != nil {
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
		cnpj, _ := p.Store.GetCustomerCNPJ(ctx, matches[0])
		if cnpj != "" && httputil.NormalizeDigits(cnpj) != company.TaxID {
			return "", httputil.BusinessError(notifications.CustomerContractingCompanyMismatch)
		}
		return matches[0], nil
	}
	return "", nil
}

func (p *Processor) processLines(ctx context.Context, providerID, accountID, invoiceID string, parsed []any, customerID string, numbers map[string]struct{}, header *vivo.Line010DHeader) error {
	seen := map[string]struct{}{}
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
		if err != nil || plan == nil {
			continue
		}

		pl, err := p.Store.GetPhoneLineByNumber(ctx, numberKey)
		if err != nil {
			return err
		}
		if pl != nil && pl.ProviderAccountID != accountID {
			return fmt.Errorf("line %s linked to another account", numberKey)
		}
		if pl == nil {
			id := uuid.New().String()
			if err := p.Store.CreatePhoneLine(ctx, id, plan.ID, accountID, numberKey); err != nil {
				return err
			}
			pl = &store.PhoneLineRow{ID: id, Number: numberKey, ProviderAccountID: accountID, ProviderPlanID: plan.ID, Status: "in_stock"}
		}

		if customerID != "" {
			_, activeCustomer, _ := p.Store.GetActivePhoneLineCustomerLink(ctx, pl.ID)
			if activeCustomer == "" {
				_ = p.Store.AssignPhoneLineCustomer(ctx, pl.ID, customerID, header.IssueDate)
				_ = p.Store.AddCustomerProviderLink(ctx, customerID, providerID, header.IssueDate)
				_ = p.Store.ReactivateCustomer(ctx, customerID)
			}
			_ = p.Store.UpdatePhoneLineStatus(ctx, pl.ID, "active")
		} else {
			_ = p.Store.UpdatePhoneLineStatus(ctx, pl.ID, "in_stock")
		}

		_ = p.Store.UpdatePhoneLineCosts(ctx, pl.ID, line.LineTotal, line.LineTotal, invoiceID)
		_ = p.Store.LinkInvoicePhoneLine(ctx, invoiceID, pl.ID)

		if pl.Status == "inactive" || pl.Status == "cancelled" || pl.Status == "suspended" {
			return httputil.BusinessError(notifications.InvoiceImportedLineOrphanDestination)
		}
	}
	return nil
}

func (p *Processor) applyAbsentLines(ctx context.Context, accountID, invoiceID string, numbersInFile map[string]struct{}) error {
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
		_, activeCustomer, _ := p.Store.GetActivePhoneLineCustomerLink(ctx, line.ID)
		if activeCustomer == "" {
			_ = p.Store.UpdatePhoneLineStatus(ctx, line.ID, "in_stock")
		} else {
			_ = p.Store.UpdatePhoneLineStatus(ctx, line.ID, "awaiting_invoice")
		}
		_ = p.Store.UpdatePhoneLineCosts(ctx, line.ID, 0, 0, invoiceID)
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
	panic("missing 010D header")
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
