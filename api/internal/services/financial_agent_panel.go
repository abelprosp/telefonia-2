package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/evolution"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) GetFinancialAgentPanel(ctx context.Context) (*models.FinancialAgentPanelResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := s.Store.GetFinancialAgentSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	statsRow, _ := s.Store.CountFinancialAgentEventStats(ctx, orgID)
	resp := &models.FinancialAgentPanelResponse{
		ToolsReady:       s.FinancialAgentConfigured,
		OrganizationSet:  s.FinancialAgentConfigured,
		Settings:         toAgentSettingsResponse(settings),
		WhatsApp:         models.FinancialAgentWhatsAppStatus{State: "unconfigured", Message: "Configure a Evolution API para conectar o WhatsApp."},
		Stats: models.FinancialAgentPanelStats{
			TotalToday:    statsRow.TotalToday,
			SuccessToday:  statsRow.SuccessToday,
			FailedToday:   statsRow.FailedToday,
			LookupsToday:  statsRow.LookupsToday,
			ReceiptsToday: statsRow.ReceiptsToday,
			PaymentsToday: statsRow.PaymentsToday,
		},
	}
	if s.Sicredi != nil && s.Sicredi.Enabled() {
		resp.SicrediEnabled = true
		if err := s.Sicredi.Ping(ctx); err == nil {
			resp.SicrediConnected = true
		}
	}
	wa, err := s.financialAgentWhatsAppStatus(ctx, settings, false)
	if err == nil {
		resp.WhatsApp = *wa
	} else if settings.EvolutionAPIURL != "" {
		resp.WhatsApp = models.FinancialAgentWhatsAppStatus{
			Configured:   settings.EvolutionAPIKey != "",
			State:        "error",
			InstanceName: settings.EvolutionInstance,
			Message:      err.Error(),
		}
	}
	resp.Message = financialAgentPanelMessage(resp)
	return resp, nil
}

func (s *Service) UpdateFinancialAgentSettings(ctx context.Context, input models.FinancialAgentSettingsInput) (*models.FinancialAgentPanelResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	current, err := s.Store.GetFinancialAgentSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	row := *current
	row.OrganizationID = orgID
	if input.Enabled != nil {
		row.Enabled = *input.Enabled
	}
	if input.EvolutionAPIURL != nil {
		row.EvolutionAPIURL = strings.TrimRight(strings.TrimSpace(*input.EvolutionAPIURL), "/")
	}
	if input.EvolutionAPIKey != nil {
		row.EvolutionAPIKey = strings.TrimSpace(*input.EvolutionAPIKey)
	}
	if input.EvolutionInstance != nil {
		inst := strings.TrimSpace(*input.EvolutionInstance)
		if inst == "" {
			inst = "luxus"
		}
		row.EvolutionInstance = inst
	}
	if input.N8nWebhookURL != nil {
		row.N8nWebhookURL = strings.TrimSpace(*input.N8nWebhookURL)
	}
	if u := auth.UserFromContext(ctx); u != nil {
		row.UpdatedBy = &u.ID
	}
	if err := s.Store.UpsertFinancialAgentSettings(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row.N8nWebhookURL != "" && row.EvolutionAPIURL != "" && row.EvolutionAPIKey != "" {
		cli := evolution.New(row.EvolutionAPIURL, row.EvolutionAPIKey)
		_ = cli.SetWebhook(ctx, row.EvolutionInstance, row.N8nWebhookURL)
	}
	return s.GetFinancialAgentPanel(ctx)
}

func (s *Service) GetFinancialAgentWhatsApp(ctx context.Context) (*models.FinancialAgentWhatsAppStatus, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := s.Store.GetFinancialAgentSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.financialAgentWhatsAppStatus(ctx, settings, false)
}

func (s *Service) ConnectFinancialAgentWhatsApp(ctx context.Context) (*models.FinancialAgentWhatsAppStatus, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := s.Store.GetFinancialAgentSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !settings.Enabled {
		return nil, httputil.ValidationError(notifications.FinancialAgentDisabled)
	}
	if strings.TrimSpace(settings.EvolutionAPIURL) == "" || strings.TrimSpace(settings.EvolutionAPIKey) == "" {
		return nil, httputil.ValidationError(notifications.FinancialAgentWhatsAppNotConfigured)
	}
	cli := evolution.New(settings.EvolutionAPIURL, settings.EvolutionAPIKey)
	conn, err := cli.Connect(ctx, settings.EvolutionInstance)
	if err != nil {
		s.logAgentEvent(ctx, "whatsapp_connect", false, err.Error(), "", nil, nil)
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if settings.N8nWebhookURL != "" {
		_ = cli.SetWebhook(ctx, settings.EvolutionInstance, settings.N8nWebhookURL)
	}
	status := toWhatsAppStatus(settings, conn, "")
	s.logAgentEvent(ctx, "whatsapp_connect", true, whatsappStatusSummary(status), "", nil, nil)
	return status, nil
}

func (s *Service) DisconnectFinancialAgentWhatsApp(ctx context.Context) (*models.FinancialAgentWhatsAppStatus, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	settings, err := s.Store.GetFinancialAgentSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if strings.TrimSpace(settings.EvolutionAPIURL) == "" || strings.TrimSpace(settings.EvolutionAPIKey) == "" {
		return nil, httputil.ValidationError(notifications.FinancialAgentWhatsAppNotConfigured)
	}
	cli := evolution.New(settings.EvolutionAPIURL, settings.EvolutionAPIKey)
	if err := cli.Disconnect(ctx, settings.EvolutionInstance); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.logAgentEvent(ctx, "whatsapp_disconnect", true, "WhatsApp desconectado.", "", nil, nil)
	return s.financialAgentWhatsAppStatus(ctx, settings, false)
}

func (s *Service) ListFinancialAgentEvents(ctx context.Context) ([]models.FinancialAgentEvent, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.Store.ListFinancialAgentEvents(ctx, orgID, 80)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	out := make([]models.FinancialAgentEvent, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.FinancialAgentEvent{
			ID:             r.ID,
			EventType:      r.EventType,
			EventLabel:     agentEventLabel(r.EventType),
			WhatsAppNumber: derefStr(r.WhatsAppNumber),
			CustomerName:   derefStr(r.CustomerName),
			InvoiceNumber:  derefStr(r.InvoiceNumber),
			Success:        r.Success,
			Summary:        r.Summary,
			CreatedAt:      r.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) financialAgentWhatsAppStatus(ctx context.Context, settings *store.FinancialAgentSettingsRow, withQR bool) (*models.FinancialAgentWhatsAppStatus, error) {
	if settings == nil || settings.EvolutionAPIURL == "" || settings.EvolutionAPIKey == "" {
		return &models.FinancialAgentWhatsAppStatus{
			State:   "unconfigured",
			Message: "Configure a Evolution API para conectar o WhatsApp.",
		}, nil
	}
	cli := evolution.New(settings.EvolutionAPIURL, settings.EvolutionAPIKey)
	if withQR {
		conn, err := cli.Connect(ctx, settings.EvolutionInstance)
		if err != nil {
			return nil, err
		}
		return toWhatsAppStatus(settings, conn, ""), nil
	}
	conn, err := cli.ConnectionState(ctx, settings.EvolutionInstance)
	if err != nil {
		return nil, err
	}
	return toWhatsAppStatus(settings, conn, ""), nil
}

func toWhatsAppStatus(settings *store.FinancialAgentSettingsRow, conn *evolution.ConnectionState, fallback string) *models.FinancialAgentWhatsAppStatus {
	state := "close"
	if conn != nil && conn.State != "" {
		state = strings.ToLower(conn.State)
	}
	connected := state == "open" || state == "connected"
	msg := fallback
	if msg == "" {
		switch {
		case connected:
			msg = "WhatsApp conectado."
		case conn != nil && conn.QRCode != "":
			msg = "Escaneie o QR Code no celular para conectar."
		default:
			msg = "WhatsApp desconectado. Clique em conectar para gerar o QR Code."
		}
	}
	out := &models.FinancialAgentWhatsAppStatus{
		Configured:   true,
		State:        state,
		Connected:    connected,
		InstanceName: settings.EvolutionInstance,
		Message:      msg,
	}
	if conn != nil {
		out.QRCode = conn.QRCode
		out.PairingCode = conn.PairingCode
		out.ProfileName = conn.ProfileName
		out.OwnerJID = conn.OwnerJID
		if conn.InstanceName != "" {
			out.InstanceName = conn.InstanceName
		}
	}
	return out
}

func toAgentSettingsResponse(row *store.FinancialAgentSettingsRow) models.FinancialAgentSettingsResponse {
	if row == nil {
		return models.FinancialAgentSettingsResponse{Enabled: true, EvolutionInstance: "luxus"}
	}
	return models.FinancialAgentSettingsResponse{
		Enabled:            row.Enabled,
		EvolutionAPIURL:    row.EvolutionAPIURL,
		EvolutionAPIKeySet: strings.TrimSpace(row.EvolutionAPIKey) != "",
		EvolutionInstance:  row.EvolutionInstance,
		N8nWebhookURL:      row.N8nWebhookURL,
		UpdatedAt:          row.UpdatedAt,
	}
}

func financialAgentPanelMessage(p *models.FinancialAgentPanelResponse) string {
	if !p.ToolsReady {
		return "As ferramentas do agente ainda não estão ativas na API. Defina FINANCIAL_AGENT_API_KEY e FINANCIAL_AGENT_ORG_ID."
	}
	if !p.Settings.EvolutionAPIKeySet || p.Settings.EvolutionAPIURL == "" {
		return "Agente pronto na API. Configure a Evolution para conectar o WhatsApp."
	}
	if p.WhatsApp.Connected {
		return "Agente e WhatsApp conectados. O histórico abaixo mostra o atendimento."
	}
	return "Agente pronto. Conecte o WhatsApp para atender os clientes."
}

func (s *Service) logAgentEvent(ctx context.Context, eventType string, success bool, summary, whatsapp string, customer *models.FinancialAgentCustomer, invoice *models.FinancialAgentInvoice) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return
	}
	row := store.FinancialAgentEventRow{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		EventType:      eventType,
		Success:        success,
		Summary:        strings.TrimSpace(summary),
		CreatedAt:      time.Now().UTC(),
	}
	if whatsapp != "" {
		w := strings.TrimSpace(whatsapp)
		row.WhatsAppNumber = &w
	}
	if customer != nil {
		id := customer.ID
		name := customer.Name
		row.CustomerID = &id
		row.CustomerName = &name
	}
	if invoice != nil {
		id := invoice.ID
		num := invoice.InvoiceNumber
		row.InvoiceID = &id
		row.InvoiceNumber = &num
		if row.CustomerName == nil && invoice.CustomerName != "" {
			n := invoice.CustomerName
			row.CustomerName = &n
		}
	}
	_ = s.Store.InsertFinancialAgentEvent(ctx, row)
}

func agentEventLabel(eventType string) string {
	switch eventType {
	case "lookup":
		return "Consulta de cliente"
	case "invoice_package":
		return "Envio de fatura"
	case "delinquency":
		return "Inadimplência"
	case "verify_receipt":
		return "Comprovante"
	case "confirm_payment":
		return "Baixa de pagamento"
	case "whatsapp_connect":
		return "WhatsApp conectado"
	case "whatsapp_disconnect":
		return "WhatsApp desconectado"
	default:
		return eventType
	}
}

func whatsappStatusSummary(st *models.FinancialAgentWhatsAppStatus) string {
	if st == nil {
		return "Solicitação de conexão WhatsApp."
	}
	if st.Connected {
		return "WhatsApp conectado."
	}
	if st.QRCode != "" {
		return "QR Code gerado para conexão do WhatsApp."
	}
	return st.Message
}
