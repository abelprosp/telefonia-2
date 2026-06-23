package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

var validSaleLineItemTypes = map[string]struct{}{
	"phone_line": {},
	"device":     {},
	"other":      {},
}

func normalizeSaleStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func normalizeLineItemType(t string) string {
	return strings.ToLower(strings.TrimSpace(t))
}

func (s *Service) ListContractTemplates(ctx context.Context, activeOnly bool, page httputil.PageSearch) ([]models.ListContractTemplateResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListContractTemplates(ctx, orgID, activeOnly, page)
}

func (s *Service) GetContractTemplate(ctx context.Context, id string) (*models.GetContractTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	t, err := s.Store.GetContractTemplate(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if t == nil {
		return nil, httputil.NotFoundError(notifications.ContractTemplateNotFound)
	}
	return t, nil
}

func (s *Service) CreateContractTemplate(ctx context.Context, input models.CreateContractTemplateInput) (*models.GetContractTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	code := strings.ToLower(strings.TrimSpace(input.Code))
	body := strings.TrimSpace(input.BodyTemplate)
	if name == "" {
		return nil, httputil.ValidationError(notifications.ContractTemplateNameRequired)
	}
	if code == "" {
		return nil, httputil.ValidationError(notifications.ContractTemplateCodeRequired)
	}
	if body == "" {
		return nil, httputil.ValidationError(notifications.ContractTemplateBodyRequired)
	}
	exists, err := s.Store.ContractTemplateCodeExists(ctx, orgID, code, nil)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if exists {
		return nil, httputil.BusinessError(notifications.ContractTemplateCodeDuplicated)
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateContractTemplate(ctx, id, orgID, name, code, body, active, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetContractTemplate(ctx, id)
}

func (s *Service) UpdateContractTemplate(ctx context.Context, id string, input models.UpdateContractTemplateInput) (*models.GetContractTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	current, err := s.Store.GetContractTemplate(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if current == nil {
		return nil, httputil.NotFoundError(notifications.ContractTemplateNotFound)
	}
	var name, code, body *string
	if input.Name != nil {
		n := strings.TrimSpace(*input.Name)
		if n == "" {
			return nil, httputil.ValidationError(notifications.ContractTemplateNameRequired)
		}
		name = &n
	}
	if input.Code != nil {
		c := strings.ToLower(strings.TrimSpace(*input.Code))
		if c == "" {
			return nil, httputil.ValidationError(notifications.ContractTemplateCodeRequired)
		}
		exists, err := s.Store.ContractTemplateCodeExists(ctx, orgID, c, &id)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if exists {
			return nil, httputil.BusinessError(notifications.ContractTemplateCodeDuplicated)
		}
		code = &c
	}
	if input.BodyTemplate != nil {
		b := strings.TrimSpace(*input.BodyTemplate)
		if b == "" {
			return nil, httputil.ValidationError(notifications.ContractTemplateBodyRequired)
		}
		body = &b
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateContractTemplate(ctx, orgID, id, name, code, body, input.Active, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetContractTemplate(ctx, id)
}

func (s *Service) ListSales(ctx context.Context, status *string, page httputil.PageSearch) ([]models.ListSaleResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		n := normalizeSaleStatus(*status)
		status = &n
	}
	return s.Store.ListSales(ctx, orgID, status, nil, page)
}

func (s *Service) GetSale(ctx context.Context, id string) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	sale, err := s.Store.GetSaleInOrg(ctx, orgID, id, nil)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if sale == nil {
		return nil, httputil.NotFoundError(notifications.SaleNotFound)
	}
	return sale, nil
}

func (s *Service) CreateSale(ctx context.Context, input models.CreateSaleInput, partnerScoped bool) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	customerID := strings.TrimSpace(input.CustomerID)
	if customerID == "" {
		return nil, httputil.ValidationError(notifications.CustomerNotFound)
	}
	if partnerScoped {
		ok, err := s.Store.CustomerOwnedBySalesperson(ctx, orgID, customerID, user.ID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if !ok {
			return nil, httputil.BusinessError(notifications.PartnerCustomerAccessDenied)
		}
	} else {
		c, err := s.Store.GetCustomerInOrg(ctx, orgID, customerID, nil)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if c == nil {
			return nil, httputil.NotFoundError(notifications.CustomerNotFound)
		}
	}
	if input.ContractTemplateID != nil && *input.ContractTemplateID != "" {
		t, err := s.Store.GetContractTemplate(ctx, orgID, *input.ContractTemplateID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if t == nil || !t.Active {
			return nil, httputil.NotFoundError(notifications.ContractTemplateNotFound)
		}
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	saleNumber, err := s.Store.NextSaleNumber(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if err := s.Store.CreateSale(ctx, id, orgID, customerID, user.ID, saleNumber, input.ContractTemplateID, input.Notes, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	for i, item := range input.Items {
		if err := s.addSaleLineItemInternal(ctx, id, item, int32(i)); err != nil {
			return nil, err
		}
	}
	if len(input.Items) > 0 {
		if err := s.Store.RecalculateSaleTotal(ctx, id, now); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	return s.GetSale(ctx, id)
}

func (s *Service) UpdateSale(ctx context.Context, id string, input models.UpdateSaleInput, partnerScoped bool) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	sale, err := s.getDraftSale(ctx, orgID, id, partnerScoped)
	if err != nil {
		return nil, err
	}
	if input.CustomerID != nil {
		customerID := strings.TrimSpace(*input.CustomerID)
		if partnerScoped {
			user, err := userFrom(ctx)
			if err != nil {
				return nil, err
			}
			ok, err := s.Store.CustomerOwnedBySalesperson(ctx, orgID, customerID, user.ID)
			if err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
			if !ok {
				return nil, httputil.BusinessError(notifications.PartnerCustomerAccessDenied)
			}
		}
		input.CustomerID = &customerID
	}
	if input.ContractTemplateID != nil && *input.ContractTemplateID != "" {
		t, err := s.Store.GetContractTemplate(ctx, orgID, *input.ContractTemplateID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if t == nil || !t.Active {
			return nil, httputil.NotFoundError(notifications.ContractTemplateNotFound)
		}
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateSaleDraft(ctx, orgID, sale.ID, input.CustomerID, input.ContractTemplateID, input.Notes, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if partnerScoped {
		user, err := userFrom(ctx)
		if err != nil {
			return nil, err
		}
		return s.partnerGetSale(ctx, orgID, id, user.ID)
	}
	return s.GetSale(ctx, id)
}

func (s *Service) AddSaleLineItem(ctx context.Context, saleID string, input models.AddSaleLineItemInput, partnerScoped bool) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	sale, err := s.getDraftSale(ctx, orgID, saleID, partnerScoped)
	if err != nil {
		return nil, err
	}
	count, err := s.Store.SaleItemCount(ctx, sale.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if err := s.addSaleLineItemInternal(ctx, sale.ID, models.CreateSaleLineItemInput(input), int32(count)); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := s.Store.RecalculateSaleTotal(ctx, sale.ID, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if partnerScoped {
		user, err := userFrom(ctx)
		if err != nil {
			return nil, err
		}
		return s.partnerGetSale(ctx, orgID, saleID, user.ID)
	}
	return s.GetSale(ctx, saleID)
}

func (s *Service) DeleteSaleLineItem(ctx context.Context, saleID, itemID string, partnerScoped bool) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	sale, err := s.getDraftSale(ctx, orgID, saleID, partnerScoped)
	if err != nil {
		return nil, err
	}
	if err := s.Store.DeleteSaleLineItem(ctx, sale.ID, itemID); err != nil {
		return nil, httputil.NotFoundError(notifications.SaleLineItemInvalid)
	}
	now := time.Now().UTC()
	if err := s.Store.RecalculateSaleTotal(ctx, sale.ID, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if partnerScoped {
		user, err := userFrom(ctx)
		if err != nil {
			return nil, err
		}
		return s.partnerGetSale(ctx, orgID, saleID, user.ID)
	}
	return s.GetSale(ctx, saleID)
}

func (s *Service) ConfirmSale(ctx context.Context, id string, partnerScoped bool) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	sale, err := s.getDraftSale(ctx, orgID, id, partnerScoped)
	if err != nil {
		return nil, err
	}
	count, err := s.Store.SaleItemCount(ctx, sale.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if count == 0 {
		return nil, httputil.BusinessError(notifications.SaleItemsRequired)
	}
	now := time.Now().UTC()
	soldAt := now
	if err := s.Store.ConfirmSale(ctx, orgID, sale.ID, soldAt, now); err != nil {
		return nil, httputil.BusinessError(notifications.SaleStatusInvalid)
	}
	if sale.ContractTemplateID != nil && *sale.ContractTemplateID != "" {
		if err := s.generateContractForSale(ctx, orgID, sale.ID, *sale.ContractTemplateID, userFromSafe(ctx)); err != nil {
			_ = s.Store.SaveFailedGeneratedContract(ctx, uuid.New().String(), orgID, sale.ID, *sale.ContractTemplateID, now)
		}
	}
	if partnerScoped {
		user, err := userFrom(ctx)
		if err != nil {
			return nil, err
		}
		return s.partnerGetSale(ctx, orgID, id, user.ID)
	}
	return s.GetSale(ctx, id)
}

func (s *Service) CancelSale(ctx context.Context, id string, partnerScoped bool) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	var salespersonFilter *string
	if partnerScoped {
		user, err := partnerUserFrom(ctx)
		if err != nil {
			return nil, err
		}
		salespersonFilter = &user.ID
	}
	sale, err := s.Store.GetSaleInOrg(ctx, orgID, id, salespersonFilter)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if sale == nil {
		return nil, httputil.NotFoundError(notifications.SaleNotFound)
	}
	if sale.Status == "cancelled" {
		return sale, nil
	}
	now := time.Now().UTC()
	if err := s.Store.CancelSale(ctx, orgID, id, now); err != nil {
		return nil, httputil.BusinessError(notifications.SaleStatusInvalid)
	}
	if partnerScoped {
		user, _ := userFrom(ctx)
		return s.partnerGetSale(ctx, orgID, id, user.ID)
	}
	return s.GetSale(ctx, id)
}

func (s *Service) PartnerListCommercialSales(ctx context.Context, status *string, page httputil.PageSearch) ([]models.ListSaleResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		n := normalizeSaleStatus(*status)
		status = &n
	}
	return s.Store.ListSales(ctx, orgID, status, &user.ID, page)
}

func (s *Service) PartnerGetCommercialSale(ctx context.Context, id string) (*models.GetSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, err
	}
	return s.partnerGetSale(ctx, orgID, id, user.ID)
}

func (s *Service) PartnerListContractTemplates(ctx context.Context, page httputil.PageSearch) ([]models.ListContractTemplateResponse, int64, error) {
	if _, err := partnerUserFrom(ctx); err != nil {
		return nil, 0, err
	}
	return s.ListContractTemplates(ctx, true, page)
}

func (s *Service) partnerGetSale(ctx context.Context, orgID, id, salespersonUserID string) (*models.GetSaleResponse, error) {
	sale, err := s.Store.GetSaleInOrg(ctx, orgID, id, &salespersonUserID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if sale == nil {
		return nil, httputil.NotFoundError(notifications.SaleNotFound)
	}
	return sale, nil
}

func (s *Service) getDraftSale(ctx context.Context, orgID, id string, partnerScoped bool) (*models.GetSaleResponse, error) {
	var salespersonFilter *string
	if partnerScoped {
		user, err := partnerUserFrom(ctx)
		if err != nil {
			return nil, err
		}
		salespersonFilter = &user.ID
	}
	sale, err := s.Store.GetSaleInOrg(ctx, orgID, id, salespersonFilter)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if sale == nil {
		return nil, httputil.NotFoundError(notifications.SaleNotFound)
	}
	if sale.Status != "draft" {
		return nil, httputil.BusinessError(notifications.SaleStatusInvalid)
	}
	return sale, nil
}

func (s *Service) addSaleLineItemInternal(ctx context.Context, saleID string, input models.CreateSaleLineItemInput, sortOrder int32) error {
	lineType := normalizeLineItemType(input.LineItemType)
	if _, ok := validSaleLineItemTypes[lineType]; !ok {
		return httputil.ValidationError(notifications.SaleLineItemInvalid)
	}
	desc := strings.TrimSpace(input.Description)
	if desc == "" {
		return httputil.ValidationError(notifications.SaleLineItemInvalid)
	}
	qty := input.Quantity
	if qty <= 0 {
		qty = 1
	}
	if input.UnitPrice < 0 {
		return httputil.ValidationError(notifications.SaleLineItemInvalid)
	}
	total := qty * input.UnitPrice
	id := uuid.New().String()
	if err := s.Store.AddSaleLineItem(ctx, id, saleID, lineType, desc, qty, input.UnitPrice, total, input.PhoneLineID, input.DeviceSku, sortOrder); err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return httputil.ValidationError(notifications.SaleLineItemInvalid)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func userFromSafe(ctx context.Context) *auth.User {
	u, _ := userFrom(ctx)
	return u
}

func (s *Service) generateContractForSale(ctx context.Context, orgID, saleID, templateID string, salesperson *auth.User) error {
	template, err := s.Store.GetContractTemplate(ctx, orgID, templateID)
	if err != nil || template == nil {
		return fmt.Errorf("template not found")
	}
	sale, err := s.Store.GetSaleInOrg(ctx, orgID, saleID, nil)
	if err != nil || sale == nil {
		return fmt.Errorf("sale not found")
	}
	customer, err := s.Store.GetCustomerContractData(ctx, orgID, sale.CustomerID)
	if err != nil || customer == nil {
		return fmt.Errorf("customer not found")
	}

	rendered := renderContractTemplate(template.BodyTemplate, customer, sale, salesperson)
	now := time.Now().UTC()
	return s.Store.SaveGeneratedContract(ctx, uuid.New().String(), orgID, saleID, templateID, rendered, now)
}

func renderContractTemplate(body string, customer *store.CustomerContractData, sale *models.GetSaleResponse, salesperson *auth.User) string {
	replacements := map[string]string{
		"{{customer.name}}":       customer.Name,
		"{{customer.legal_name}}": derefString(customer.LegalName, customer.Name),
		"{{customer.document}}":   customer.Document,
		"{{customer.type}}":       customer.Type,
		"{{customer.address.full}}": formatFullAddress(customer),
		"{{customer.address.street}}": customer.Street,
		"{{customer.address.number}}": customer.Number,
		"{{customer.address.neighborhood}}": customer.Neighborhood,
		"{{customer.address.city}}": customer.City,
		"{{customer.address.state}}": customer.State,
		"{{customer.address.zip_code}}": customer.ZipCode,
		"{{customer.address.country}}": customer.Country,
		"{{sale.sale_number}}":    sale.SaleNumber,
		"{{sale.total_amount}}":   formatMoneyBR(sale.TotalAmount),
		"{{sale.sold_at}}":        formatDateBR(sale.SoldAt),
		"{{sale.items_table}}":    buildItemsTableHTML(sale.Items),
		"{{salesperson.name}}":    salespersonDisplayName(salesperson),
	}
	out := body
	for k, v := range replacements {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}

func derefString(v *string, fallback string) string {
	if v != nil && strings.TrimSpace(*v) != "" {
		return *v
	}
	return fallback
}

func formatFullAddress(c *store.CustomerContractData) string {
	parts := []string{}
	if c.Street != "" {
		line := c.Street
		if c.Number != "" {
			line += ", " + c.Number
		}
		parts = append(parts, line)
	}
	if c.Neighborhood != "" {
		parts = append(parts, c.Neighborhood)
	}
	if c.Complement != nil && *c.Complement != "" {
		parts = append(parts, *c.Complement)
	}
	cityState := strings.TrimSpace(c.City)
	if c.State != "" {
		if cityState != "" {
			cityState += " - " + c.State
		} else {
			cityState = c.State
		}
	}
	if cityState != "" {
		parts = append(parts, cityState)
	}
	if c.ZipCode != "" {
		parts = append(parts, "CEP "+c.ZipCode)
	}
	if len(parts) == 0 {
		return "—"
	}
	return strings.Join(parts, ", ")
}

func formatMoneyBR(value float64) string {
	s := fmt.Sprintf("%.2f", value)
	s = strings.ReplaceAll(s, ".", "#")
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, "#", ",")
	return "R$ " + s
}

func formatDateBR(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Format("02/01/2006")
}

func buildItemsTableHTML(items []models.SaleLineItemResponse) string {
	if len(items) == 0 {
		return "<p>Nenhum item.</p>"
	}
	var b strings.Builder
	b.WriteString(`<table border="1" cellpadding="6" cellspacing="0" style="border-collapse:collapse;width:100%">`)
	b.WriteString(`<thead><tr><th>Descrição</th><th>Tipo</th><th>Qtd</th><th>Unitário</th><th>Total</th></tr></thead><tbody>`)
	for _, item := range items {
		b.WriteString("<tr>")
		b.WriteString("<td>" + htmlEscape(item.Description) + "</td>")
		b.WriteString("<td>" + htmlEscape(item.LineItemType) + "</td>")
		b.WriteString("<td>" + fmt.Sprintf("%.2f", item.Quantity) + "</td>")
		b.WriteString("<td>" + formatMoneyBR(item.UnitPrice) + "</td>")
		b.WriteString("<td>" + formatMoneyBR(item.TotalPrice) + "</td>")
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>")
	return b.String()
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}

func salespersonDisplayName(u *auth.User) string {
	if u == nil {
		return "—"
	}
	if strings.TrimSpace(u.Name) != "" {
		return u.Name
	}
	if strings.TrimSpace(u.Username) != "" {
		return u.Username
	}
	if utf8.RuneCountInString(u.ID) > 8 {
		return u.ID[:8] + "…"
	}
	return u.ID
}
