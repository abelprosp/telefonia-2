package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/invoicelayout"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) IssueSicrediBoleto(ctx context.Context, documentID string) (*models.IssueSicrediBoletoResponse, error) {
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
	if doc.SicrediNossoNumero != nil && strings.TrimSpace(*doc.SicrediNossoNumero) != "" {
		return nil, httputil.BusinessError(notifications.SicrediBoletoAlreadyIssued)
	}
	result, err := s.issueSicrediBoletoForDocument(ctx, orgID, doc)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) tryAttachSicrediBoleto(ctx context.Context, orgID, documentID string) (issued bool, errMsg string) {
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return false, ""
	}
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil || doc == nil {
		return false, ""
	}
	if doc.SicrediNossoNumero != nil && strings.TrimSpace(*doc.SicrediNossoNumero) != "" {
		return true, ""
	}
	_, err = s.issueSicrediBoletoForDocument(ctx, orgID, doc)
	if err != nil {
		if ae, ok := err.(*httputil.AppError); ok {
			return false, ae.Error()
		}
		return false, err.Error()
	}
	return true, ""
}

func (s *Service) applySicrediFeedback(ctx context.Context, documentID string, message *string, status **string, nosso **string, boletoErr **string) {
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return
	}
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil || doc == nil {
		return
	}
	if status != nil {
		*status = doc.SicrediBoletoStatus
	}
	if nosso != nil {
		*nosso = doc.SicrediNossoNumero
	}
	if boletoErr != nil {
		*boletoErr = doc.SicrediBoletoError
	}
	if message == nil {
		return
	}
	if doc.SicrediNossoNumero != nil && strings.TrimSpace(*doc.SicrediNossoNumero) != "" {
		*message = "Fatura e boleto Sicredi gerados com sucesso."
		return
	}
	if doc.SicrediBoletoStatus != nil && *doc.SicrediBoletoStatus == "failed" {
		msg := "Fatura criada; boleto Sicredi falhou."
		if doc.SicrediBoletoError != nil && strings.TrimSpace(*doc.SicrediBoletoError) != "" {
			msg += " " + strings.TrimSpace(*doc.SicrediBoletoError)
		}
		*message = msg
	}
}

func (s *Service) issueSicrediBoletoForDocument(ctx context.Context, orgID string, doc *models.GetCustomerBillingDocumentResponse) (*models.IssueSicrediBoletoResponse, error) {
	recID := ""
	if doc.AccountsReceivableID != nil {
		recID = *doc.AccountsReceivableID
	}
	var rec *store.ReceivableForBilling
	if recID != "" {
		rec, _ = s.Store.GetReceivableForBilling(ctx, orgID, recID)
	}
	if rec == nil {
		docNum, _ := s.Store.GetCustomerPrimaryDocument(ctx, doc.CustomerID)
		rec = &store.ReceivableForBilling{
			ID:               recID,
			CustomerID:       doc.CustomerID,
			CustomerName:     doc.CustomerName,
			CustomerDocument: docNum,
			Description:      doc.EmailSubject,
			IssueDate:        doc.IssueDate,
			DueDate:          doc.DueDate,
			Amount:           doc.Amount,
			BillingEmail:     doc.RecipientEmail,
		}
	}

	pagador, err := s.buildSicrediPagador(ctx, doc.CustomerID, rec)
	if err != nil {
		s.saveSicrediBoletoError(ctx, orgID, doc.ID, doc.EmailBodyHtml, err.Error())
		return nil, httputil.ValidationError(notifications.N("SICREDI_PAGADOR_INVALID", err.Error()))
	}

	input := sicredi.CreateBoletoInput{
		SeuNumero:       doc.InvoiceNumber,
		IdTituloEmpresa: doc.ID,
		DataVencimento:  doc.DueDate,
		Valor:           doc.Amount,
		Mensagens: []string{
			"Fatura " + doc.InvoiceNumber,
			rec.Description,
		},
		Pagador: pagador,
	}

	boleto, err := s.Sicredi.CreateHybridBoleto(ctx, input)
	now := time.Now().UTC()
	if err != nil {
		s.saveSicrediBoletoError(ctx, orgID, doc.ID, doc.EmailBodyHtml, err.Error())
		return nil, httputil.BusinessError(notifications.N("SICREDI_BOLETO_FAILED", err.Error()))
	}

	payment := invoicelayout.PaymentData{
		LinhaDigitavel:   boleto.LinhaDigitavel,
		CodigoBarras:     boleto.CodigoBarras,
		PixCopyPaste:     boleto.PixQrCode,
		PixQrCodeDataURL: invoicelayout.PixQRCodeDataURL(boleto.PixQrCode),
		NossoNumero:      boleto.NossoNumero,
	}
	body := invoicelayout.AppendPaymentSection(doc.EmailBodyHtml, payment, invoicelayout.DefaultConfig().Theme)

	if err := s.Store.UpdateCustomerBillingDocumentSicredi(ctx, orgID, doc.ID,
		boleto.NossoNumero, boleto.LinhaDigitavel, boleto.CodigoBarras, boleto.PixQrCode, boleto.PixTxID,
		"issued", nil, body, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	nn := boleto.NossoNumero
	ld := boleto.LinhaDigitavel
	return &models.IssueSicrediBoletoResponse{
		Success:               true,
		Message:               "Boleto Sicredi gerado com sucesso.",
		SicrediNossoNumero:    &nn,
		SicrediLinhaDigitavel: &ld,
	}, nil
}

func (s *Service) saveSicrediBoletoError(ctx context.Context, orgID, documentID, body, errMsg string) {
	now := time.Now().UTC()
	msg := errMsg
	_ = s.Store.UpdateCustomerBillingDocumentSicredi(ctx, orgID, documentID,
		"", "", "", "", "", "failed", &msg, body, now)
}

func (s *Service) buildSicrediPagador(ctx context.Context, customerID string, rec *store.ReceivableForBilling) (sicredi.Pagador, error) {
	doc := httputil.NormalizeDigits(rec.CustomerDocument)
	if doc == "" {
		return sicredi.Pagador{}, fmt.Errorf("cliente sem CPF/CNPJ cadastrado")
	}
	tipo := "PESSOA_FISICA"
	if len(doc) == 14 {
		tipo = "PESSOA_JURIDICA"
	} else if len(doc) != 11 {
		return sicredi.Pagador{}, fmt.Errorf("CPF/CNPJ do cliente inválido")
	}

	addr, _ := s.Store.GetCustomerAddressParts(ctx, customerID)
	pagador := sicredi.Pagador{
		TipoPessoa: tipo,
		Documento:  doc,
		Nome:       rec.CustomerName,
		Email:      rec.BillingEmail,
	}
	if addr != nil {
		pagador.Endereco = strings.TrimSpace(addr.Street + ", " + addr.Number)
		pagador.Cidade = addr.City
		pagador.UF = addr.State
		pagador.CEP = addr.ZipCode
	}
	return pagador, nil
}
