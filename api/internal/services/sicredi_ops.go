package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/sicredi"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) GetSicrediBoletoPDF(ctx context.Context, documentID string) ([]byte, string, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, "", err
	}
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return nil, "", httputil.BusinessError(notifications.SicrediNotConfigured)
	}
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil {
		return nil, "", err
	}
	if doc.SicrediLinhaDigitavel == nil || strings.TrimSpace(*doc.SicrediLinhaDigitavel) == "" {
		return nil, "", httputil.ValidationError(notifications.N("SICREDI_BOLETO_NOT_ISSUED", "Esta fatura ainda não possui boleto Sicredi."))
	}
	pdf, err := s.Sicredi.GetBoletoPDF(ctx, strings.TrimSpace(*doc.SicrediLinhaDigitavel))
	if err != nil {
		return nil, "", httputil.BusinessError(notifications.N("SICREDI_PDF_FAILED", err.Error()))
	}
	filename := "boleto-" + doc.InvoiceNumber + ".pdf"
	_ = orgID
	return pdf, filename, nil
}

func (s *Service) CancelSicrediBoleto(ctx context.Context, documentID string) (*models.IssueSicrediBoletoResponse, error) {
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
	if doc.SicrediNossoNumero == nil || strings.TrimSpace(*doc.SicrediNossoNumero) == "" {
		return nil, httputil.ValidationError(notifications.N("SICREDI_BOLETO_NOT_ISSUED", "Esta fatura ainda não possui boleto Sicredi."))
	}
	nossoNumero := strings.TrimSpace(*doc.SicrediNossoNumero)
	if err := s.Sicredi.CancelBoleto(ctx, nossoNumero); err != nil {
		return nil, httputil.BusinessError(notifications.N("SICREDI_CANCEL_FAILED", err.Error()))
	}
	now := time.Now().UTC()
	_ = s.Store.UpdateCustomerBillingDocumentSicrediStatus(ctx, orgID, documentID, "cancelled", now)
	return &models.IssueSicrediBoletoResponse{
		Success: true,
		Message: "Boleto baixado no Sicredi.",
	}, nil
}

type AlterSicrediBoletoDueDateInput struct {
	DueDate string `json:"due_date"`
}

func (s *Service) AlterSicrediBoletoDueDate(ctx context.Context, documentID string, input AlterSicrediBoletoDueDateInput) (*models.IssueSicrediBoletoResponse, error) {
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
	if doc.SicrediNossoNumero == nil || strings.TrimSpace(*doc.SicrediNossoNumero) == "" {
		return nil, httputil.ValidationError(notifications.N("SICREDI_BOLETO_NOT_ISSUED", "Esta fatura ainda não possui boleto Sicredi."))
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return nil, err
	}
	nossoNumero := strings.TrimSpace(*doc.SicrediNossoNumero)
	if err := s.Sicredi.AlterBoletoDueDate(ctx, nossoNumero, dueDate); err != nil {
		return nil, httputil.BusinessError(notifications.N("SICREDI_ALTER_FAILED", err.Error()))
	}
	now := time.Now().UTC()
	_ = s.Store.UpdateCustomerBillingDocumentSicrediStatus(ctx, orgID, documentID, "issued", now)
	return &models.IssueSicrediBoletoResponse{
		Success: true,
		Message: "Vencimento do boleto alterado no Sicredi.",
	}, nil
}

func (s *Service) RegisterSicrediWebhook(ctx context.Context, input *models.RegisterSicrediWebhookInput) (*models.IssueSicrediBoletoResponse, error) {
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return nil, httputil.BusinessError(notifications.SicrediNotConfigured)
	}
	cfg := s.Sicredi.Config()
	publicURL := strings.TrimSpace(cfg.PublicAPIURL)
	if input != nil && strings.TrimSpace(input.PublicAPIURL) != "" {
		publicURL = strings.TrimSpace(input.PublicAPIURL)
	}
	if publicURL == "" {
		return nil, httputil.ValidationError(notifications.N("SICREDI_WEBHOOK_URL_REQUIRED", "Configure SICREDI_PUBLIC_API_URL com a URL pública da API (ex.: ngrok)."))
	}
	if strings.Contains(publicURL, "localhost") || strings.Contains(publicURL, "127.0.0.1") {
		return nil, httputil.ValidationError(notifications.N("SICREDI_WEBHOOK_URL_LOCAL", "O Sicredi exige URL pública HTTPS para o webhook. Use ngrok/cloudflare tunnel e informe em SICREDI_PUBLIC_API_URL."))
	}
	webhookURL := strings.TrimRight(publicURL, "/") + "/v1/webhooks/sicredi"
	if err := s.Sicredi.RegisterWebhookContract(ctx, sicredi.WebhookContractInput{
		URL:     webhookURL,
		Token:   cfg.WebhookToken,
		Eventos: []string{"LIQUIDACAO", "BAIXA", "REGISTRO", "ALTERACAO"},
	}); err != nil {
		return nil, httputil.BusinessError(notifications.N("SICREDI_WEBHOOK_REGISTER_FAILED", err.Error()))
	}
	return &models.IssueSicrediBoletoResponse{
		Success: true,
		Message: "Contrato de webhook registrado no Sicredi: " + webhookURL,
	}, nil
}

func (s *Service) TestSicrediConnection(ctx context.Context) (*models.SicrediTestConnectionResponse, error) {
	if _, err := orgFrom(ctx); err != nil {
		return nil, err
	}
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return nil, httputil.BusinessError(notifications.SicrediNotConfigured)
	}
	cfg := s.Sicredi.Config()
	if err := s.Sicredi.Ping(ctx); err != nil {
		env := "produção"
		if cfg.Sandbox {
			env = "sandbox"
		}
		return &models.SicrediTestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Falha na conexão Sicredi (%s): %s", env, err.Error()),
			Sandbox: cfg.Sandbox,
		}, nil
	}
	env := "produção"
	if cfg.Sandbox {
		env = "sandbox"
	}
	return &models.SicrediTestConnectionResponse{
		Success: true,
		Message: fmt.Sprintf("Conexão Sicredi OK (%s). OAuth autenticado com sucesso.", env),
		Sandbox: cfg.Sandbox,
	}, nil
}

func (s *Service) GetSicrediStatus(ctx context.Context) (*models.SicrediIntegrationStatusResponse, error) {
	_, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	enabled := s.Sicredi != nil && s.Sicredi.Enabled()
	resp := &models.SicrediIntegrationStatusResponse{Enabled: enabled}
	if s.Sicredi != nil {
		cfg := s.Sicredi.Config()
		resp.Sandbox = cfg.Sandbox
		resp.Cooperativa = cfg.Cooperativa
		resp.Posto = cfg.Posto
		resp.CodigoBeneficiario = cfg.CodigoBeneficiario
		resp.PublicAPIURL = cfg.PublicAPIURL
		resp.WebhookConfigured = cfg.WebhookToken != "" && cfg.PublicAPIURL != "" &&
			!strings.Contains(cfg.PublicAPIURL, "localhost") && !strings.Contains(cfg.PublicAPIURL, "127.0.0.1")
		if cfg.PublicAPIURL != "" {
			resp.WebhookURL = strings.TrimRight(cfg.PublicAPIURL, "/") + "/v1/webhooks/sicredi"
		}
		if err := s.Sicredi.Ping(ctx); err != nil {
			resp.ConnectionError = err.Error()
		} else {
			resp.Connected = true
			if contracts, err := s.Sicredi.ListWebhookContracts(ctx); err == nil && len(contracts) > 0 {
				resp.WebhookRegistered = true
				if contracts[0].URL != "" {
					resp.WebhookURL = contracts[0].URL
				}
			}
		}
	}
	return resp, nil
}

func (s *Service) HandleSicrediWebhook(ctx context.Context, r *http.Request) error {
	if s.Sicredi == nil || !s.Sicredi.Enabled() {
		return httputil.BusinessError(notifications.SicrediNotConfigured)
	}
	cfg := s.Sicredi.Config()
	if token := strings.TrimSpace(cfg.WebhookToken); token != "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if auth != "Bearer "+token && auth != token {
			return httputil.BusinessError(notifications.N("SICREDI_WEBHOOK_UNAUTHORIZED", "Token de webhook inválido."))
		}
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return httputil.ValidationError(notifications.N("SICREDI_WEBHOOK_INVALID", "Payload inválido."))
	}

	eventType := firstWebhookString(payload, "tipoEvento", "evento", "tipo", "eventType")
	nossoNumero := firstWebhookString(payload, "nossoNumero", "nosso_numero")
	seuNumero := firstWebhookString(payload, "seuNumero", "seu_numero")
	idEmpresa := firstWebhookString(payload, "idTituloEmpresa", "id_titulo_empresa", "idEmpresa")

	eventID := uuid.New().String()
	now := time.Now().UTC()
	row := store.SicrediWebhookEventRow{
		ID:              eventID,
		EventType:       eventType,
		Payload:         json.RawMessage(raw),
		Processed:       false,
		CreatedAt:       now,
	}
	if nossoNumero != "" {
		row.NossoNumero = &nossoNumero
	}
	if seuNumero != "" {
		row.SeuNumero = &seuNumero
	}
	if idEmpresa != "" {
		row.IdTituloEmpresa = &idEmpresa
	}

	if err := s.Store.InsertSicrediWebhookEvent(ctx, row); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	processErr := s.processSicrediWebhookEvent(ctx, eventType, nossoNumero, idEmpresa, payload)
	var errMsg *string
	if processErr != nil {
		msg := processErr.Error()
		errMsg = &msg
	}
	_ = s.Store.MarkSicrediWebhookEventProcessed(ctx, eventID, errMsg, time.Now().UTC())
	return nil
}

func (s *Service) processSicrediWebhookEvent(ctx context.Context, eventType, nossoNumero, idEmpresa string, payload map[string]any) error {
	upper := strings.ToUpper(eventType)
	if strings.Contains(upper, "LIQUID") || strings.Contains(upper, "PAG") {
		return s.processSicrediLiquidationWebhook(ctx, nossoNumero, idEmpresa, payload)
	}
	if strings.Contains(upper, "BAIXA") || strings.Contains(upper, "CANCEL") {
		return s.processSicrediBaixaWebhook(ctx, nossoNumero, idEmpresa)
	}
	return nil
}

func (s *Service) processSicrediLiquidationWebhook(ctx context.Context, nossoNumero, idEmpresa string, payload map[string]any) error {
	doc, err := s.findWebhookBillingDocument(ctx, nossoNumero, idEmpresa)
	if err != nil || doc == nil {
		return err
	}
	paidAt := time.Now().UTC()
	if raw := firstWebhookString(payload, "dataPagamento", "data_pagamento", "dataLiquidacao"); raw != "" {
		if t, err := time.Parse("2006-01-02", raw); err == nil {
			paidAt = t
		} else if t, err := time.Parse("02/01/2006", raw); err == nil {
			paidAt = t
		}
	}
	amount := doc.Amount
	if v, ok := payload["valorLiquidado"].(float64); ok && v > 0 {
		amount = v
	}
	item := sicredi.LiquidadoItem{
		NossoNumero:    doc.SicrediNossoNumero,
		DataPagamento:  paidAt,
		ValorLiquidado: amount,
		TipoLiquidacao: "WEBHOOK",
	}
	return s.applySicrediPayment(ctx, *doc, item)
}

func (s *Service) processSicrediBaixaWebhook(ctx context.Context, nossoNumero, idEmpresa string) error {
	doc, err := s.findWebhookBillingDocument(ctx, nossoNumero, idEmpresa)
	if err != nil || doc == nil {
		return err
	}
	now := time.Now().UTC()
	return s.Store.UpdateCustomerBillingDocumentSicrediStatus(ctx, doc.OrganizationID, doc.ID, "cancelled", now)
}

func (s *Service) findWebhookBillingDocument(ctx context.Context, nossoNumero, idEmpresa string) (*store.UnpaidSicrediBillingDocument, error) {
	if idEmpresa != "" {
		if doc, err := s.Store.GetBillingDocumentByIDGlobal(ctx, idEmpresa); err == nil && doc != nil {
			return doc, nil
		}
	}
	if nossoNumero != "" {
		return s.Store.GetBillingDocumentByNossoNumeroGlobal(ctx, nossoNumero)
	}
	return nil, nil
}

func firstWebhookString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
