package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/observability"
	"github.com/luxus-connect/telefonia/api/internal/precision"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

const (
	perspectiveLuxusCustomer   = "luxus_customer"
	perspectiveCustomerEndUser = "customer_end_user"
)

func (s *Service) ListLineBillingProcessings(ctx context.Context, phoneLineID string) (*models.ListLineBillingProcessingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if linkID == "" {
		return &models.ListLineBillingProcessingsResponse{Processings: []models.LineBillingProcessingResponse{}}, nil
	}
	rows, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	resp := &models.ListLineBillingProcessingsResponse{LinkID: linkID, Processings: make([]models.LineBillingProcessingResponse, 0, len(rows))}
	for _, row := range rows {
		item, err := s.Store.ToBillingProcessingResponse(ctx, row)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		resp.Processings = append(resp.Processings, item)
	}
	return resp, nil
}

func (s *Service) EnsureBillingProcessingsForLink(ctx context.Context, linkID, customerID string, monthlyAmount *float64) error {
	now := time.Now().UTC()
	existing, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return err
	}
	hasPrimary := false
	hasSecondary := false
	for _, p := range existing {
		if p.Perspective == perspectiveLuxusCustomer {
			hasPrimary = true
		}
		if p.Perspective == perspectiveCustomerEndUser {
			hasSecondary = true
		}
	}
	if !hasPrimary {
		procID := uuid.New().String()
		if err := s.Store.CreateBillingProcessing(ctx, store.BillingProcessingRow{
			ID:                      procID,
			PhoneLineCustomerLinkID: linkID,
			Perspective:             perspectiveLuxusCustomer,
			Active:                  true,
			CreatedAt:               now,
			UpdatedAt:               now,
		}); err != nil {
			return err
		}
		amount := 0.0
		if monthlyAmount != nil && *monthlyAmount > 0 {
			amount = *monthlyAmount
		}
		if amount > 0 {
			if err := s.Store.CreateBillingCompositionItem(ctx, store.BillingCompositionItemRow{
				ID:           uuid.New().String(),
				ProcessingID: procID,
				ItemType:     "service",
				Description:  "Mensalidade",
				Amount:       amount,
				Quantity:     1,
				Active:       true,
				CreatedAt:    now,
				UpdatedAt:    now,
			}); err != nil {
				return err
			}
		}
	}
	isReseller, err := s.Store.CustomerIsReseller(ctx, customerID)
	if err != nil {
		return err
	}
	if isReseller && !hasSecondary {
		label := "Usuário final"
		if err := s.Store.CreateBillingProcessing(ctx, store.BillingProcessingRow{
			ID:                      uuid.New().String(),
			PhoneLineCustomerLinkID: linkID,
			Perspective:             perspectiveCustomerEndUser,
			Label:                   &label,
			Active:                  true,
			CreatedAt:               now,
			UpdatedAt:               now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) EnableEndUserProcessing(ctx context.Context, phoneLineID string) (*models.LineBillingProcessingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil || linkID == "" {
		return nil, httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}
	_, customerID, err := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	isReseller, err := s.Store.CustomerIsReseller(ctx, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !isReseller {
		return nil, httputil.ValidationError(notifications.N("CUSTOMER_NOT_RESELLER", "Cliente não é PJ revendedor. Ative a flag no cadastro do cliente."))
	}
	existing, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, p := range existing {
		if p.Perspective == perspectiveCustomerEndUser {
			resp, err := s.Store.ToBillingProcessingResponse(ctx, p)
			if err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
			return &resp, nil
		}
	}
	now := time.Now().UTC()
	procID := uuid.New().String()
	label := "Usuário final"
	if err := s.Store.CreateBillingProcessing(ctx, store.BillingProcessingRow{
		ID:                      procID,
		PhoneLineCustomerLinkID: linkID,
		Perspective:             perspectiveCustomerEndUser,
		Label:                   &label,
		Active:                  true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Create", "LineBillingProcessing", procID, nil, map[string]any{"perspective": perspectiveCustomerEndUser})
	resp, err := s.Store.ToBillingProcessingResponse(ctx, store.BillingProcessingRow{
		ID: procID, PhoneLineCustomerLinkID: linkID, Perspective: perspectiveCustomerEndUser, Label: &label, Active: true,
	})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &resp, nil
}

func (s *Service) UpdateLineBillingProcessing(ctx context.Context, phoneLineID, processingID string, input models.UpdateLineBillingProcessingInput) (*models.LineBillingProcessingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	before, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || before == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateBillingProcessing(ctx, processingID, input.Label, input.MirrorFromPrimary, input.OrganizationalUnit, input.Department, input.CostCenterLabel, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if input.MirrorFromPrimary != nil && *input.MirrorFromPrimary {
		if err := s.mirrorProcessingFromPrimary(ctx, orgID, before.PhoneLineCustomerLinkID, processingID); err != nil {
			return nil, err
		}
	}
	s.auditLog(ctx, "Update", "LineBillingProcessing", processingID,
		map[string]any{"label": before.Label, "mirror": before.MirrorFromPrimary},
		map[string]any{"label": input.Label, "mirror": input.MirrorFromPrimary})
	after, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || after == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	resp, err := s.Store.ToBillingProcessingResponse(ctx, *after)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &resp, nil
}

func (s *Service) MirrorProcessingFromPrimary(ctx context.Context, phoneLineID, processingID string) (*models.LineBillingProcessingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	target, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || target == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	if target.Perspective != perspectiveCustomerEndUser {
		return nil, httputil.ValidationError(notifications.N("BILLING_MIRROR_SECONDARY_ONLY", "Espelhamento só se aplica ao processamento cliente→usuário final."))
	}
	if err := s.mirrorProcessingFromPrimary(ctx, orgID, target.PhoneLineCustomerLinkID, processingID); err != nil {
		return nil, err
	}
	mirror := true
	now := time.Now().UTC()
	_ = s.Store.UpdateBillingProcessing(ctx, processingID, nil, &mirror, nil, nil, nil, now)
	resp, err := s.Store.ToBillingProcessingResponse(ctx, *target)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &resp, nil
}

func (s *Service) mirrorProcessingFromPrimary(ctx context.Context, orgID, linkID, targetProcessingID string) error {
	processings, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	var primaryID string
	for _, p := range processings {
		if p.Perspective == perspectiveLuxusCustomer {
			primaryID = p.ID
			break
		}
	}
	if primaryID == "" {
		return httputil.BusinessError(notifications.N("BILLING_PRIMARY_MISSING", "Processamento Luxus→Cliente não encontrado."))
	}
	primaryItems, err := s.Store.ListBillingCompositionItems(ctx, primaryID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	now := time.Now().UTC()
	if err := s.Store.DeactivateBillingCompositionItemsForProcessing(ctx, targetProcessingID, now); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, it := range primaryItems {
		copy := it
		copy.ID = uuid.New().String()
		copy.ProcessingID = targetProcessingID
		copy.CreatedAt = now
		copy.UpdatedAt = now
		if err := s.Store.CreateBillingCompositionItem(ctx, copy); err != nil {
			return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	s.auditLog(ctx, "Mirror", "LineBillingProcessing", targetProcessingID, nil, map[string]any{"from": primaryID})
	return nil
}

func (s *Service) CreateLineBillingCompositionItem(ctx context.Context, phoneLineID, processingID string, input models.CreateLineBillingCompositionItemInput) (*models.LineBillingCompositionItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	proc, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || proc == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	if err := validateCompositionItemInput(input.ItemType, input.Description, input.Amount); err != nil {
		return nil, err
	}
	if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, time.Now().UTC()); err != nil {
		return nil, err
	}
	qty := 1.0
	if input.Quantity != nil && *input.Quantity > 0 {
		qty = *input.Quantity
	}
	now := time.Now().UTC()
	id := uuid.New().String()
	proportional := true
	if input.Proportional != nil {
		proportional = *input.Proportional
	}
	row := store.BillingCompositionItemRow{
		ID: id, ProcessingID: processingID, ItemType: strings.ToLower(strings.TrimSpace(input.ItemType)),
		Description: strings.TrimSpace(input.Description), Amount: input.Amount, Quantity: qty,
		InstallmentCount: input.InstallmentCount, InstallmentCurrent: input.InstallmentCurrent,
		Active: true, CreatedAt: now, UpdatedAt: now, Proportional: proportional,
		ServiceType: input.ServiceType, ProviderPlanServiceID: input.ProviderPlanServiceID,
	}
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		t, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		row.StartDate = &t
	}
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		t, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return nil, err
		}
		row.EndDate = &t
	}

	// Regra 2.4: Parcelamento de Aparelhos (calcula término automático se omitido)
	if row.ItemType == "installment" && row.InstallmentCount != nil && *row.InstallmentCount > 0 {
		if row.InstallmentCurrent == nil || *row.InstallmentCurrent <= 0 {
			first := 1
			row.InstallmentCurrent = &first
		}
		if row.StartDate != nil && row.EndDate == nil {
			calcEnd := row.StartDate.AddDate(0, *row.InstallmentCount, 0)
			row.EndDate = &calcEnd
		}
		total := input.Amount
		row.InstallmentTotal = &total
		parcels, splitErr := precision.SplitInstallments(total, *row.InstallmentCount)
		if splitErr != nil {
			return nil, httputil.ValidationError(notifications.N("INSTALLMENT_SPLIT_INVALID", splitErr.Error()))
		}
		idx := 0
		if row.InstallmentCurrent != nil && *row.InstallmentCurrent > 0 {
			idx = *row.InstallmentCurrent - 1
			if idx < 0 {
				idx = 0
			}
			if idx >= len(parcels) {
				idx = len(parcels) - 1
			}
		}
		row.Amount = parcels[idx]
	}

	// Regra 2.3: Validação de Descontos (não pode exceder o valor dos serviços)
	if row.ItemType == "discount" {
		items, _ := s.Store.ListBillingCompositionItems(ctx, processingID)
		var servicesTotal float64
		var existingDiscounts float64
		for _, it := range items {
			if it.ItemType == "service" {
				servicesTotal = precision.SumCents(servicesTotal, it.Amount*it.Quantity)
			} else if it.ItemType == "discount" {
				existingDiscounts = precision.SumCents(existingDiscounts, it.Amount*it.Quantity)
			}
		}
		newDiscountTotal := precision.SumCents(existingDiscounts, input.Amount*qty)
		if servicesTotal > 0 && newDiscountTotal > servicesTotal {
			return nil, httputil.ValidationError(notifications.N("DISCOUNT_EXCEEDS_SERVICES", fmt.Sprintf("O somatório de descontos (R$ %.2f) não pode exceder o total de serviços da linha (R$ %.2f).", newDiscountTotal, servicesTotal)))
		}
	}

	if row.ItemType == "service" && row.ServiceType != nil && strings.TrimSpace(*row.ServiceType) != "" {
		dup, err := s.Store.ActiveCompositionServiceTypeExists(ctx, processingID, strings.ToLower(strings.TrimSpace(*row.ServiceType)), "")
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if dup {
			return nil, httputil.BusinessError(notifications.PhoneLineServiceTypeDuplicated)
		}
	}

	if err := s.gateCompositionTwoLevel(ctx, phoneLineID, processingID, row.ItemType, input.Amount*qty, input); err != nil {
		return nil, err
	}

	if err := s.Store.CreateBillingCompositionItem(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Create", "LineBillingCompositionItem", id, nil, compositionAuditMap(row))
	if proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}
	resp := store.CompositionItemToModel(row)
	return &resp, nil
}

func (s *Service) gateCompositionTwoLevel(ctx context.Context, phoneLineID, processingID, itemType string, amount float64, input models.CreateLineBillingCompositionItemInput) error {
	payload := map[string]any{
		"phone_line_id": phoneLineID,
		"processing_id": processingID,
		"item":          input,
	}
	switch itemType {
	case "extra_charge", "exceedance":
		return s.queueTwoLevel(ctx, "forced_charge", "billing_item", processingID, "Cobrança forçada / avulsa", payload)
	case "discount":
		items, _ := s.Store.ListBillingCompositionItems(ctx, processingID)
		var servicesTotal float64
		for _, it := range items {
			if it.ItemType == "service" {
				servicesTotal = precision.SumCents(servicesTotal, it.Amount*it.Quantity)
			}
		}
		if servicesTotal > 0 && amount >= servicesTotal*0.30 {
			return s.queueTwoLevel(ctx, "high_discount", "billing_item", processingID, "Desconto elevado (≥ 30% dos serviços)", payload)
		}
	}
	return nil
}

func (s *Service) UpdateLineBillingCompositionItem(ctx context.Context, phoneLineID, processingID, itemID string, input models.UpdateLineBillingCompositionItemInput) (*models.LineBillingCompositionItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	before, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || before == nil || before.ProcessingID != processingID {
		return nil, httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de composição não encontrado."))
	}
	if input.Amount != nil && *input.Amount < 0 {
		return nil, httputil.ValidationError(notifications.N("BILLING_ITEM_AMOUNT_INVALID", "Valor não pode ser negativo."))
	}
	now := time.Now().UTC()
	var startDate, endDate *time.Time
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		t, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		startDate = &t
	}
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		t, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return nil, err
		}
		endDate = &t
	}
	if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, now); err != nil {
		return nil, err
	}
	if input.ServiceType != nil && strings.TrimSpace(*input.ServiceType) != "" {
		dup, err := s.Store.ActiveCompositionServiceTypeExists(ctx, processingID, strings.ToLower(strings.TrimSpace(*input.ServiceType)), itemID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if dup {
			return nil, httputil.BusinessError(notifications.PhoneLineServiceTypeDuplicated)
		}
	}
	if err := s.Store.UpdateBillingCompositionItem(ctx, itemID, input.Description, input.Amount, input.Quantity,
		input.InstallmentCount, input.InstallmentCurrent, startDate, endDate, now, input.ServiceType, input.Proportional); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Update", "LineBillingCompositionItem", itemID, compositionAuditMap(*before), input)
	proc, _ := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if proc != nil && proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}
	after, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || after == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de composição não encontrado."))
	}
	resp := store.CompositionItemToModel(*after)
	return &resp, nil
}

func (s *Service) DeleteLineBillingCompositionItem(ctx context.Context, phoneLineID, processingID, itemID string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return err
	}
	before, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || before == nil || before.ProcessingID != processingID {
		return httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de composição não encontrado."))
	}
	items, err := s.Store.ListBillingCompositionItems(ctx, processingID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	serviceCount := 0
	for _, it := range items {
		if it.ItemType == "service" {
			serviceCount++
		}
	}
	if before.ItemType == "service" && serviceCount <= 1 {
		return httputil.ValidationError(notifications.N("BILLING_MIN_ONE_SERVICE", "O processamento deve ter ao menos um serviço ativo."))
	}
	now := time.Now().UTC()
	if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, now); err != nil {
		return err
	}
	if err := s.Store.DeactivateBillingCompositionItem(ctx, itemID, now); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Delete", "LineBillingCompositionItem", itemID, compositionAuditMap(*before), nil)
	proc, _ := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if proc != nil && proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}
	return nil
}

func (s *Service) PayoffInstallmentItem(ctx context.Context, phoneLineID, processingID, itemID string) (*models.LineBillingCompositionItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	item, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || item == nil || item.ProcessingID != processingID {
		return nil, httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de parcelamento não encontrado."))
	}
	if item.ItemType != "installment" {
		return nil, httputil.ValidationError(notifications.N("ITEM_NOT_INSTALLMENT", "Apenas itens de parcelamento podem ser quitados antecipadamente."))
	}
	now := time.Now().UTC()
	if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, now); err != nil {
		return nil, err
	}

	count := 1
	if item.InstallmentCount != nil && *item.InstallmentCount > 0 {
		count = *item.InstallmentCount
	}
	current := 1
	if item.InstallmentCurrent != nil && *item.InstallmentCurrent > 0 {
		current = *item.InstallmentCurrent
	}
	remaining := count - current + 1
	if remaining <= 0 {
		remaining = 1
	}

	total := precision.Round2(item.Amount * float64(count))
	if item.InstallmentTotal != nil && *item.InstallmentTotal > 0 {
		total = *item.InstallmentTotal
	}
	payoffAmount := precision.Round2(item.Amount * float64(remaining))
	if parcels, splitErr := precision.SplitInstallments(total, count); splitErr == nil {
		payoffAmount = 0
		for i := current - 1; i < len(parcels); i++ {
			if i >= 0 {
				payoffAmount = precision.SumCents(payoffAmount, parcels[i])
			}
		}
	}

	// Inativa o parcelamento recorrente
	if err := s.Store.DeactivateBillingCompositionItem(ctx, itemID, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	// Cria o lançamento de quitação avulsa
	payoffID := uuid.New().String()
	payoffDesc := fmt.Sprintf("Quitação antecipada de aparelho (%s) — %d parcela(s) restante(s)", item.Description, remaining)
	payoffRow := store.BillingCompositionItemRow{
		ID:           payoffID,
		ProcessingID: processingID,
		ItemType:     "extra_charge",
		Description:  payoffDesc,
		Amount:       payoffAmount,
		Quantity:     1,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Proportional: false,
	}
	if err := s.Store.CreateBillingCompositionItem(ctx, payoffRow); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	s.auditLog(ctx, "Payoff", "LineBillingCompositionItem", itemID, compositionAuditMap(*item), map[string]any{
		"payoff_item_id":   payoffID,
		"payoff_amount":    payoffAmount,
		"remaining_months": remaining,
	})

	proc, _ := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if proc != nil && proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}

	resp := store.CompositionItemToModel(payoffRow)
	return &resp, nil
}

func (s *Service) ListLineBillingProcessingAudit(ctx context.Context, phoneLineID, processingID string) ([]models.AuditLogResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	if proc, err := s.Store.GetBillingProcessing(ctx, orgID, processingID); err != nil || proc == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	return s.listAuditForProcessing(ctx, processingID)
}

func (s *Service) listAuditForProcessing(ctx context.Context, processingID string) ([]models.AuditLogResponse, error) {
	rows, err := s.Store.ListAuditLogsForEntity(ctx, "LineBillingProcessing", processingID, 100)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	itemRows, _ := s.Store.ListBillingCompositionItems(ctx, processingID)
	ids := make([]string, 0, len(itemRows))
	for _, it := range itemRows {
		ids = append(ids, it.ID)
	}
	for _, id := range ids {
		more, err := s.Store.ListAuditLogsForEntity(ctx, "LineBillingCompositionItem", id, 50)
		if err == nil {
			rows = append(rows, more...)
		}
	}
	out := make([]models.AuditLogResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.AuditLogResponse{
			ID: r.ID, ChangeType: r.ChangeType, EntityName: r.EntityName, KeyValues: r.KeyValues,
			ChangedBy: r.ChangedBy, OldValues: r.OldValues, NewValues: r.NewValues, Timestamp: r.Timestamp,
		})
	}
	return out, nil
}

func (s *Service) syncLinkMonthlyAmountFromPrimary(ctx context.Context, linkID, processingID string) error {
	total, err := s.Store.SumBillingProcessingTotal(ctx, processingID)
	if err != nil {
		return err
	}
	return s.Store.UpdateActivePhoneLineCustomerLinkAmountByLinkID(ctx, linkID, &total)
}

func validateCompositionItemInput(itemType, description string, amount float64) error {
	t := strings.ToLower(strings.TrimSpace(itemType))
	switch t {
	case "service", "discount", "extra_charge", "installment", "exceedance":
	default:
		return httputil.ValidationError(notifications.N("BILLING_ITEM_TYPE_INVALID", "Tipo inválido. Use: service, discount, extra_charge, installment."))
	}
	if strings.TrimSpace(description) == "" {
		return httputil.ValidationError(notifications.N("BILLING_ITEM_DESCRIPTION_REQUIRED", "Descrição obrigatória."))
	}
	if amount < 0 {
		return httputil.ValidationError(notifications.N("BILLING_ITEM_AMOUNT_INVALID", "Valor não pode ser negativo."))
	}
	return nil
}

func (s *Service) auditLog(ctx context.Context, changeType, entityName, key string, oldVal, newVal any) {
	var oldStr, newStr *string
	if oldVal != nil {
		if b, err := json.Marshal(oldVal); err == nil {
			s := string(b)
			oldStr = &s
		}
	}
	if newVal != nil {
		if b, err := json.Marshal(newVal); err == nil {
			s := string(b)
			newStr = &s
		}
	}
	var changedBy *string
	if u, err := userFrom(ctx); err == nil && u != nil {
		changedBy = &u.ID
	}
	corrID := observability.CorrelationIDFromContext(ctx)
	_ = s.Store.InsertAuditLog(ctx, uuid.New().String(), changeType, entityName, key, changedBy, oldStr, newStr, time.Now().UTC(), corrID)
}

func compositionAuditMap(row store.BillingCompositionItemRow) map[string]any {
	return map[string]any{
		"item_type": row.ItemType, "description": row.Description, "amount": row.Amount, "quantity": row.Quantity,
	}
}

func (s *Service) SimulateBillingImpact(ctx context.Context, processingMonthID string) (*models.BillingImpactSimulationResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}

	lines, totalLines, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 1000})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var projectedRevenue float64
	var projectedCost float64
	for _, l := range lines {
		if l.Status == "active" {
			if l.BaseCost != nil {
				projectedCost = precision.SumCents(projectedCost, *l.BaseCost)
			}
			linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
			if err == nil && linkID != "" {
				processings, _ := s.Store.ListBillingProcessingsForLink(ctx, linkID)
				for _, p := range processings {
					if p.Perspective == perspectiveLuxusCustomer {
						tot, _ := s.Store.SumBillingProcessingTotal(ctx, p.ID)
						projectedRevenue = precision.SumCents(projectedRevenue, tot)
						break
					}
				}
			}
		}
	}

	projectedMargin := precision.Round2(projectedRevenue - projectedCost)
	marginPct := 0.0
	if projectedRevenue > 0 {
		marginPct = precision.Round2((projectedMargin / projectedRevenue) * 100.0)
	}

	if invoiceCost, err := s.Store.SumProviderInvoiceTotalsForMonth(ctx, orgID, processingMonthID); err == nil && invoiceCost > 0 {
		projectedCost = invoiceCost
		projectedMargin = precision.Round2(projectedRevenue - projectedCost)
		if projectedRevenue > 0 {
			marginPct = precision.Round2((projectedMargin / projectedRevenue) * 100.0)
		}
	}

	var prevRevenue float64
	prev, _ := s.Store.FindPreviousProcessingMonth(ctx, orgID, month.ProviderID, month.Year, month.Month)
	if prev != nil {
		if rev, err := s.Store.SumCustomerBillingForMonth(ctx, orgID, prev.ID); err == nil {
			prevRevenue = rev
		}
	}

	delta := precision.Round2(projectedRevenue - prevRevenue)
	deltaPct := 0.0
	if prevRevenue > 0 {
		deltaPct = precision.Round2((delta / prevRevenue) * 100.0)
	}

	return &models.BillingImpactSimulationResponse{
		ProcessingMonthID: processingMonthID,
		DisplayName:       month.DisplayName,
		ProjectedRevenue:  projectedRevenue,
		ProjectedCost:     projectedCost,
		ProjectedMargin:   projectedMargin,
		MarginPercentage:  marginPct,
		PreviousRevenue:   prevRevenue,
		RevenueDelta:      delta,
		RevenueDeltaPct:   deltaPct,
		TotalActiveLines:  int(totalLines),
	}, nil
}

func (s *Service) GetProcessingMonthLineReadiness(ctx context.Context, processingMonthID string) (*models.ProcessingMonthLineReadinessResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 1000})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var items []models.LineProcessingReadinessItem
	readyCount := 0
	blockedCount := 0

	for _, l := range lines {
		if l.Status != "active" {
			continue
		}
		var blocking []string
		linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
		var custID, custName *string
		monthlyAmt := 0.0

		if err != nil || linkID == "" {
			blocking = append(blocking, "Linha sem cliente ativo vinculado.")
		} else {
			links, _ := s.Store.ListPhoneLineCustomerLinks(ctx, orgID, l.ID)
			for _, cl := range links {
				if cl.IsActive {
					cID := cl.CustomerID
					cName := cl.CustomerName
					custID = &cID
					custName = &cName
					break
				}
			}

			processings, _ := s.Store.ListBillingProcessingsForLink(ctx, linkID)
			hasService := false
			for _, p := range processings {
				if p.Perspective == perspectiveLuxusCustomer {
					compItems, _ := s.Store.ListBillingCompositionItems(ctx, p.ID)
					for _, ci := range compItems {
						if ci.ItemType == "service" {
							hasService = true
							break
						}
					}
					tot, _ := s.Store.SumBillingProcessingTotal(ctx, p.ID)
					monthlyAmt = tot
					break
				}
			}
			if !hasService {
				blocking = append(blocking, "Nenhum serviço contratado ativo na composição.")
			}
		}

		isReady := len(blocking) == 0
		if isReady {
			readyCount++
		} else {
			blockedCount++
		}

		items = append(items, models.LineProcessingReadinessItem{
			PhoneLineID:   l.ID,
			PhoneNumber:   l.Number,
			CustomerID:    custID,
			CustomerName:  custName,
			MonthlyAmount: monthlyAmt,
			IsReady:       isReady,
			BlockingRules: blocking,
		})
	}

	return &models.ProcessingMonthLineReadinessResponse{
		ProcessingMonthID: processingMonthID,
		TotalLines:        len(items),
		ReadyLines:        readyCount,
		BlockedLines:      blockedCount,
		Items:             items,
	}, nil
}

func (s *Service) GetLineBillingExplanation(ctx context.Context, phoneLineID, processingMonthID string) (*models.BillingExplanationResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}

	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil || linkID == "" {
		return nil, httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}

	var custID, custName string
	links, _ := s.Store.ListPhoneLineCustomerLinks(ctx, orgID, phoneLineID)
	for _, cl := range links {
		if cl.IsActive {
			custID = cl.CustomerID
			custName = cl.CustomerName
			break
		}
	}

	var components []models.BillingExplanationComponent
	var total float64
	var formulaParts []string

	processings, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err == nil {
		for _, p := range processings {
			if p.Perspective == perspectiveLuxusCustomer {
				items, _ := s.Store.ListBillingCompositionItems(ctx, p.ID)
				for _, it := range items {
					itemAmt := precision.Round2(it.Amount * it.Quantity)
					switch it.ItemType {
					case "service":
						components = append(components, models.BillingExplanationComponent{
							Type:        "service",
							Description: it.Description,
							Amount:      itemAmt,
							Details:     "Serviço recorrente de plano",
						})
						total = precision.SumCents(total, itemAmt)
						formulaParts = append(formulaParts, fmt.Sprintf("+ R$ %.2f (%s)", itemAmt, it.Description))
					case "discount":
						components = append(components, models.BillingExplanationComponent{
							Type:        "discount",
							Description: it.Description,
							Amount:      -itemAmt,
							Details:     "Desconto contratual/rateio",
						})
						total = precision.Round2(total - itemAmt)
						formulaParts = append(formulaParts, fmt.Sprintf("- R$ %.2f (%s)", itemAmt, it.Description))
					case "installment":
						cur := 1
						if it.InstallmentCurrent != nil {
							cur = *it.InstallmentCurrent
						}
						cnt := 1
						if it.InstallmentCount != nil {
							cnt = *it.InstallmentCount
						}
						components = append(components, models.BillingExplanationComponent{
							Type:        "device",
							Description: it.Description,
							Amount:      itemAmt,
							Details:     fmt.Sprintf("Parcela %d de %d", cur, cnt),
						})
						total = precision.SumCents(total, itemAmt)
						formulaParts = append(formulaParts, fmt.Sprintf("+ R$ %.2f (Aparelho %d/%d)", itemAmt, cur, cnt))
					case "extra_charge":
						components = append(components, models.BillingExplanationComponent{
							Type:        "exceedance",
							Description: it.Description,
							Amount:      itemAmt,
							Details:     "Cobrança avulsa / excedente",
						})
						total = precision.SumCents(total, itemAmt)
						formulaParts = append(formulaParts, fmt.Sprintf("+ R$ %.2f (%s)", itemAmt, it.Description))
					}
				}
				break
			}
		}
	}

	formulaText := strings.Join(formulaParts, " ") + fmt.Sprintf(" = R$ %.2f", total)
	if len(formulaParts) == 0 {
		formulaText = fmt.Sprintf("Total da linha: R$ %.2f", total)
	}

	return &models.BillingExplanationResponse{
		PhoneLineID:       phoneLineID,
		PhoneNumber:       line.Number,
		CustomerID:        custID,
		CustomerName:      custName,
		ProcessingMonthID: processingMonthID,
		TotalAmount:       total,
		FormulaText:       formulaText,
		Components:        components,
	}, nil
}

func (s *Service) CloseProcessingMonthWithHash(ctx context.Context, processingMonthID string) (*models.CloseMonthWithHashResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if month.Status == "closed" {
		hash := ""
		if month.ConsolidationHash != nil {
			hash = *month.ConsolidationHash
		}
		closedBy := ""
		if month.ClosedBy != nil {
			closedBy = *month.ClosedBy
		}
		closedAt := time.Time{}
		if month.ClosedAt != nil {
			closedAt = *month.ClosedAt
		}
		return &models.CloseMonthWithHashResponse{
			ProcessingMonthID: processingMonthID, Status: "closed", ClosedAt: closedAt,
			ClosedBy: closedBy, ConsolidationHash: hash,
		}, nil
	}
	if _, err := s.RequestTwoLevelApproval(ctx, models.CreateApprovalRequestInput{
		ActionType: "consolidation", EntityType: "processing_month", EntityID: processingMonthID,
		Justification: "Consolidação e fechamento da competência",
	}); err != nil {
		return nil, err
	}
	return &models.CloseMonthWithHashResponse{
		ProcessingMonthID: processingMonthID,
		Status:            "pending_approval",
		ClosedBy:          "",
		ConsolidationHash: "",
	}, nil
}

func (s *Service) doConsolidateMonth(ctx context.Context, processingMonthID string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return err
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, processingMonthID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if err := s.SM().ValidateTransition(statemachine.EntityProcessingMonth, month.Status, "closed", userRoles(ctx)); err != nil {
		return err
	}
	if err := s.Store.CloseProcessingMonth(ctx, orgID, processingMonthID, user.ID, false, nil); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.SM().RecordTransition(ctx, orgID, statemachine.EntityProcessingMonth, processingMonthID, month.Status, "closed", "close_month_hash", nil, &user.ID, nil)
	now := time.Now().UTC()
	hashData := fmt.Sprintf("ORG=%s|MONTH=%s|CLOSED_BY=%s|TIMESTAMP=%d", orgID, processingMonthID, user.ID, now.Unix())
	hashBytes := sha256.Sum256([]byte(hashData))
	consolidationHash := hex.EncodeToString(hashBytes[:])
	_ = s.Store.UpdateProcessingMonthConsolidationHash(ctx, orgID, processingMonthID, consolidationHash)
	s.auditLog(ctx, "ConsolidateClose", "ProcessingMonth", processingMonthID, map[string]any{
		"previous_status": month.Status,
	}, map[string]any{
		"new_status": "closed", "consolidation_hash": consolidationHash, "closed_by": user.ID,
	})
	s.DispatchWebhookEvent(orgID, WebhookEventBillingClosed, map[string]any{
		"processing_month_id": processingMonthID, "consolidation_hash": consolidationHash,
	})
	if expiring, err := s.Store.ListExpiringContracts(ctx, orgID, 30); err == nil && len(expiring) > 0 {
		s.DispatchWebhookEvent(orgID, WebhookEventFidelityExpiring, map[string]any{"count": len(expiring)})
	}
	return nil
}
