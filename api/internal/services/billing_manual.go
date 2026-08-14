package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (s *Service) ManualBillingPreview(ctx context.Context, customerIDs []string) (*models.BulkBillingPreviewResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.Store.ListManualBillingCandidates(ctx, orgID, customerIDs)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if items == nil {
		items = []models.BulkBillingPreviewItem{}
	}
	eligible := 0
	for _, item := range items {
		if item.Eligible {
			eligible++
		}
	}
	return &models.BulkBillingPreviewResponse{
		Items:         items,
		EligibleCount: eligible,
	}, nil
}

func (s *Service) ManualGenerateBillingDocuments(ctx context.Context, input models.ManualGenerateBillingDocumentsInput) (*models.BulkGenerateBillingDocumentsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	issueDate, err := parseFinancialDate(input.IssueDate)
	if err != nil {
		return nil, err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return nil, err
	}
	templateCode := strings.TrimSpace(input.TemplateCode)
	if templateCode == "" {
		templateCode = "default-billing-invoice"
	}
	layoutCode := strings.TrimSpace(input.LayoutTemplateCode)
	if layoutCode == "" {
		layoutCode = "default-invoice-layout"
	}
	descriptionTemplate := strings.TrimSpace(input.Description)
	if descriptionTemplate == "" {
		descriptionTemplate = "Mensalidade telefonia"
	}

	candidates, err := s.Store.ListManualBillingCandidates(ctx, orgID, input.CustomerIDs)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if len(input.CustomerIDs) > 0 {
		selected := make(map[string]struct{}, len(input.CustomerIDs))
		for _, id := range input.CustomerIDs {
			if id = strings.TrimSpace(id); id != "" {
				selected[id] = struct{}{}
			}
		}
		filtered := candidates[:0]
		for _, c := range candidates {
			if _, ok := selected[c.CustomerID]; ok {
				filtered = append(filtered, c)
			}
		}
		candidates = filtered
	}
	if len(input.BillingGroupIDs) > 0 {
		selected := make(map[string]struct{}, len(input.BillingGroupIDs))
		for _, id := range input.BillingGroupIDs {
			if id = strings.TrimSpace(id); id != "" {
				selected[id] = struct{}{}
			}
		}
		filtered := candidates[:0]
		for _, c := range candidates {
			if _, ok := selected[c.BillingGroupID]; ok {
				filtered = append(filtered, c)
			}
		}
		candidates = filtered
	}

	return s.generateBillingDocumentsForCandidates(ctx, orgID, candidates, nil, issueDate, dueDate, descriptionTemplate, templateCode, layoutCode)
}

func (s *Service) GenerateCustomerBillingDocument(ctx context.Context, customerID string, input models.GenerateCustomerBillingDocumentInput) (*models.GenerateCustomerBillingDocumentResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return nil, httputil.ValidationError(notifications.CustomerNotFound)
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	issueDate, err := parseFinancialDate(input.IssueDate)
	if err != nil {
		return nil, err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return nil, err
	}
	templateCode := strings.TrimSpace(input.TemplateCode)
	if templateCode == "" {
		templateCode = "default-billing-invoice"
	}
	layoutCode := strings.TrimSpace(input.LayoutTemplateCode)
	if layoutCode == "" {
		layoutCode = "default-invoice-layout"
	}
	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = "Mensalidade telefonia"
	}

	candidates, err := s.Store.ListManualBillingCandidates(ctx, orgID, []string{customerID})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if len(candidates) == 0 {
		return nil, httputil.ValidationError(notifications.N("BILLING_NO_LINES", "Cliente sem linhas ou aparelhos ativos vinculados."))
	}
	if input.Amount != nil && *input.Amount > 0 && len(candidates) == 1 {
		candidates[0].MonthlyAmount = *input.Amount
		candidates[0].Eligible = *input.Amount > 0
	}
	bulk, err := s.generateBillingDocumentsForCandidates(ctx, orgID, candidates, nil, issueDate, dueDate, description, templateCode, layoutCode)
	if err != nil {
		return nil, err
	}
	var firstID, firstRec string
	var total float64
	ids := make([]string, 0, bulk.Created)
	for _, item := range bulk.Items {
		if item.Status == "created" && item.DocumentID != nil {
			if firstID == "" {
				firstID = *item.DocumentID
				if item.ReceivableID != nil {
					firstRec = *item.ReceivableID
				}
			}
			ids = append(ids, *item.DocumentID)
			total += item.Amount
		}
	}
	if bulk.Created == 0 {
		msg := "Nenhuma fatura gerada."
		if len(bulk.Items) > 0 && bulk.Items[0].Message != "" {
			msg = bulk.Items[0].Message
		}
		return nil, httputil.ValidationError(notifications.N("BILLING_NO_AMOUNT", msg))
	}
	msg := "Fatura criada com sucesso."
	if bulk.Created > 1 {
		msg = "Faturas criadas (um boleto por linha Normal e um por grupo Titular+dependentes)."
	}
	resp := &models.GenerateCustomerBillingDocumentResponse{
		ID:           firstID,
		ReceivableID: firstRec,
		Amount:       total,
		Message:      msg,
		CreatedCount: bulk.Created,
		DocumentIDs:  ids,
	}
	s.applySicrediFeedback(ctx, firstID, &resp.Message, &resp.SicrediBoletoStatus, &resp.SicrediNossoNumero, &resp.SicrediBoletoError)
	return resp, nil
}

func (s *Service) generateBillingDocumentsForCandidates(
	ctx context.Context,
	orgID string,
	candidates []models.BulkBillingPreviewItem,
	processingMonthID *string,
	issueDate, dueDate time.Time,
	descriptionTemplate, templateCode, layoutCode string,
) (*models.BulkGenerateBillingDocumentsResponse, error) {
	resp := &models.BulkGenerateBillingDocumentsResponse{Items: make([]models.BulkBillingGenerateItemResult, 0, len(candidates))}
	now := time.Now().UTC()
	_ = now

	for _, c := range candidates {
		result := models.BulkBillingGenerateItemResult{
			BillingGroupID:   c.BillingGroupID,
			BillingGroupType: c.BillingGroupType,
			GroupLabel:       c.GroupLabel,
			CustomerID:       c.CustomerID,
			CustomerName:     c.CustomerName,
			Amount:           c.MonthlyAmount,
		}
		if !c.Eligible {
			result.Status = "skipped"
			result.Message = bulkSkipReasonMessage(c.SkipReason)
			resp.Skipped++
			resp.Items = append(resp.Items, result)
			continue
		}

		desc := descriptionTemplate
		if c.GroupLabel != "" {
			desc = descriptionTemplate + " — " + c.GroupLabel
		}
		docID, receivableID, err := s.createBillingDocumentForAmount(ctx, orgID, c.CustomerID, desc, c.MonthlyAmount, processingMonthID, issueDate, dueDate, templateCode, layoutCode, c.PhoneLineID, c.BillingGroupType)
		if err != nil {
			result.Status = "failed"
			if svcErr, ok := err.(*httputil.AppError); ok {
				result.Message = svcErr.Error()
			} else {
				result.Message = err.Error()
			}
			resp.Failed++
			resp.Items = append(resp.Items, result)
			continue
		}

		result.Status = "created"
		result.DocumentID = &docID
		result.ReceivableID = &receivableID
		s.applySicrediFeedback(ctx, docID, &result.Message, &result.SicrediBoletoStatus, &result.SicrediNossoNumero, &result.SicrediBoletoError)
		resp.Created++
		resp.Items = append(resp.Items, result)
	}

	return resp, nil
}

func (s *Service) createBillingDocumentForAmount(
	ctx context.Context,
	orgID, customerID, description string,
	amount float64,
	processingMonthID *string,
	issueDate, dueDate time.Time,
	templateCode, layoutCode string,
	phoneLineID *string,
	groupType string,
) (documentID, receivableID string, err error) {
	receivableID = uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateAccountReceivable(ctx, receivableID, orgID, customerID, description, processingMonthID, issueDate, dueDate, amount, nil, now); err != nil {
		return "", "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	rec, err := s.Store.GetReceivableForBilling(ctx, orgID, receivableID)
	if err != nil || rec == nil {
		return "", "", httputil.InternalError(notifications.N("BILLING_RECEIVABLE_FAILED", "Falha ao carregar conta a receber criada."))
	}
	rec.PhoneLineID = phoneLineID
	if groupType != "" {
		rec.BillingGroupType = &groupType
	}
	docID, err := s.createBillingDocumentFromReceivable(ctx, orgID, rec, templateCode, layoutCode)
	if err != nil {
		return "", "", err
	}
	return docID, receivableID, nil
}
