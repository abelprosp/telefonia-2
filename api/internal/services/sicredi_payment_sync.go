package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) SyncSicrediPayments(ctx context.Context, daysBack int) (*models.SyncSicrediPaymentsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return nil, httputil.BusinessError(notifications.SicrediNotConfigured)
	}
	if daysBack <= 0 {
		daysBack = 7
	}
	if daysBack > 30 {
		daysBack = 30
	}

	resp := &models.SyncSicrediPaymentsResponse{Items: []models.SyncSicrediPaymentItemResult{}}
	liquidados := make(map[string]sicredi.LiquidadoItem)

	now := time.Now().UTC()
	for i := 0; i < daysBack; i++ {
		day := now.AddDate(0, 0, -i)
		page := 0
		for {
			batch, err := s.Sicredi.ListLiquidadosDia(ctx, day, page)
			if err != nil {
				return nil, httputil.BusinessError(notifications.N("SICREDI_SYNC_FAILED", err.Error()))
			}
			for _, item := range batch.Items {
				key := normalizeNossoNumero(item.NossoNumero)
				if key != "" {
					liquidados[key] = item
				}
			}
			if !batch.HasNext {
				break
			}
			page++
		}
	}

	pending, err := s.Store.ListUnpaidSicrediBillingDocuments(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	for _, doc := range pending {
		resp.Checked++
		result := models.SyncSicrediPaymentItemResult{
			DocumentID:    doc.ID,
			InvoiceNumber: doc.InvoiceNumber,
			CustomerName:  doc.CustomerName,
			Status:        "pending",
		}

		paidItem, found := liquidados[normalizeNossoNumero(doc.SicrediNossoNumero)]
		if !found {
			detail, err := s.Sicredi.GetBoleto(ctx, doc.SicrediNossoNumero)
			if err == nil && detail != nil && sicredi.IsSituacaoLiquidada(detail.Situacao) {
				amount := detail.ValorLiquidado
				if amount <= 0 {
					amount = doc.Amount
				}
				paidAt := now
				if detail.DataPagamento != nil {
					paidAt = detail.DataPagamento.UTC()
				}
				paidItem = sicredi.LiquidadoItem{
					NossoNumero:    doc.SicrediNossoNumero,
					DataPagamento:  paidAt,
					ValorLiquidado: amount,
					TipoLiquidacao: detail.Situacao,
				}
				found = true
			}
		}

		if !found {
			resp.Items = append(resp.Items, result)
			continue
		}

		if err := s.applySicrediPayment(ctx, doc, paidItem); err != nil {
			result.Status = "failed"
			result.Message = err.Error()
		} else {
			result.Status = "paid"
			result.Amount = paidItem.ValorLiquidado
			paidAt := paidItem.DataPagamento.UTC()
			result.PaidAt = &paidAt
			resp.Paid++
		}
		resp.Items = append(resp.Items, result)
	}

	return resp, nil
}

func (s *Service) SyncSicrediPaymentForDocument(ctx context.Context, documentID string) (*models.SyncSicrediPaymentsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return nil, httputil.BusinessError(notifications.SicrediNotConfigured)
	}
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	if doc.SicrediPaidAt != nil {
		paidAt := doc.SicrediPaidAt.UTC()
		return &models.SyncSicrediPaymentsResponse{
			Checked: 1,
			Paid:    1,
			Items: []models.SyncSicrediPaymentItemResult{{
				DocumentID:    doc.ID,
				InvoiceNumber: doc.InvoiceNumber,
				CustomerName:  doc.CustomerName,
				Status:        "paid",
				PaidAt:        &paidAt,
				Amount:        doc.Amount,
			}},
		}, nil
	}
	if doc.SicrediNossoNumero == nil || strings.TrimSpace(*doc.SicrediNossoNumero) == "" {
		return nil, httputil.ValidationError(notifications.N("SICREDI_BOLETO_NOT_ISSUED", "Esta fatura ainda não possui boleto Sicredi."))
	}

	nossoNumero := strings.TrimSpace(*doc.SicrediNossoNumero)
	detail, err := s.Sicredi.GetBoleto(ctx, nossoNumero)
	if err != nil {
		return nil, httputil.BusinessError(notifications.N("SICREDI_SYNC_FAILED", err.Error()))
	}
	if detail == nil || !sicredi.IsSituacaoLiquidada(detail.Situacao) {
		return &models.SyncSicrediPaymentsResponse{
			Checked: 1,
			Items: []models.SyncSicrediPaymentItemResult{{
				DocumentID:    doc.ID,
				InvoiceNumber: doc.InvoiceNumber,
				CustomerName:  doc.CustomerName,
				Status:        "pending",
				Message:       "Boleto ainda não liquidado no Sicredi.",
			}},
		}, nil
	}

	amount := detail.ValorLiquidado
	if amount <= 0 {
		amount = doc.Amount
	}
	paidAt := time.Now().UTC()
	if detail.DataPagamento != nil {
		paidAt = detail.DataPagamento.UTC()
	}
	row := store.UnpaidSicrediBillingDocument{
		ID:                   doc.ID,
		OrganizationID:       orgID,
		InvoiceNumber:        doc.InvoiceNumber,
		CustomerName:         doc.CustomerName,
		AccountsReceivableID: doc.AccountsReceivableID,
		Amount:               doc.Amount,
		SicrediNossoNumero:   nossoNumero,
	}
	item := sicredi.LiquidadoItem{
		NossoNumero:    nossoNumero,
		DataPagamento:  paidAt,
		ValorLiquidado: amount,
		TipoLiquidacao: detail.Situacao,
	}
	if err := s.applySicrediPayment(ctx, row, item); err != nil {
		return nil, httputil.BusinessError(notifications.N("SICREDI_PAYMENT_APPLY_FAILED", err.Error()))
	}
	paidAtCopy := paidAt
	return &models.SyncSicrediPaymentsResponse{
		Checked: 1,
		Paid:    1,
		Items: []models.SyncSicrediPaymentItemResult{{
			DocumentID:    doc.ID,
			InvoiceNumber: doc.InvoiceNumber,
			CustomerName:  doc.CustomerName,
			Status:        "paid",
			PaidAt:        &paidAtCopy,
			Amount:        amount,
		}},
	}, nil
}

func (s *Service) applySicrediPayment(ctx context.Context, doc store.UnpaidSicrediBillingDocument, paid sicredi.LiquidadoItem) error {
	orgID := doc.OrganizationID
	paidAt := paid.DataPagamento
	if paidAt.IsZero() {
		paidAt = time.Now().UTC()
	}
	amount := paid.ValorLiquidado
	if amount <= 0 {
		amount = doc.Amount
	}

	if doc.AccountsReceivableID != nil && strings.TrimSpace(*doc.AccountsReceivableID) != "" {
		receivableID := strings.TrimSpace(*doc.AccountsReceivableID)
		ref := "Sicredi " + paid.NossoNumero
		notes := "Liquidação Sicredi (" + paid.TipoLiquidacao + ")"
		if err := s.Store.RegisterReceivablePaymentAuto(ctx, uuid.New().String(), orgID, receivableID, amount, paidAt, ref, notes, time.Now().UTC()); err != nil {
			return err
		}
	}

	if err := s.Store.MarkSicrediBoletoPaid(ctx, orgID, doc.ID, paidAt); err != nil && !isPgNoRows(err) {
		return err
	}
	return nil
}

func (s *Service) RunSicrediPaymentSyncAllOrgs(ctx context.Context, daysBack int) {
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return
	}
	pending, err := s.Store.ListUnpaidSicrediBillingDocuments(ctx, "")
	if err != nil || len(pending) == 0 {
		return
	}
	orgIDs := make(map[string]struct{})
	for _, doc := range pending {
		orgIDs[doc.OrganizationID] = struct{}{}
	}
	for orgID := range orgIDs {
		bgCtx := auth.WithOrganization(ctx, &auth.Organization{ID: orgID})
		_, _ = s.SyncSicrediPayments(bgCtx, daysBack)
	}
}

func normalizeNossoNumero(v string) string {
	return strings.TrimSpace(v)
}
