package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

const (
	WebhookEventLineTransition   = "LINE_TRANSITION_ALERT"
	WebhookEventFidelityExpiring = "FIDELITY_EXPIRING_ALERT"
	WebhookEventDivergence       = "DIVERGENCE_DETECTED"
	WebhookEventBillingClosed    = "BILLING_CLOSED"
)

var imeiRegex = regexp.MustCompile(`^[0-9]{15}$`)

var allowedWebhookEvents = map[string]struct{}{
	WebhookEventLineTransition:   {},
	WebhookEventFidelityExpiring: {},
	WebhookEventDivergence:       {},
	WebhookEventBillingClosed:    {},
}

func generateWebhookSecret() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "whsec_" + hex.EncodeToString([]byte(uuid.New().String()))[:24]
	}
	return "whsec_" + hex.EncodeToString(buf)
}

func maskWebhookSecret(secret string) string {
	if len(secret) <= 10 {
		return "whsec_****"
	}
	return secret[:8] + "****" + secret[len(secret)-4:]
}

func webhookToDTO(r store.WebhookSubscriptionRow, revealSecret bool) models.WebhookSubscriptionResponse {
	secret := maskWebhookSecret(r.Secret)
	if revealSecret {
		secret = r.Secret
	}
	return models.WebhookSubscriptionResponse{
		ID:        r.ID,
		URL:       r.URL,
		Events:    r.Events,
		IsActive:  r.IsActive,
		Secret:    secret,
		CreatedAt: r.CreatedAt,
	}
}

func (s *Service) ListWebhooks(ctx context.Context) ([]models.WebhookSubscriptionResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.Store.ListWebhookSubscriptions(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	out := make([]models.WebhookSubscriptionResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, webhookToDTO(r, false))
	}
	return out, nil
}

func (s *Service) CreateWebhook(ctx context.Context, input models.CreateWebhookSubscriptionInput) (*models.WebhookSubscriptionResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	url := strings.TrimSpace(input.URL)
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return nil, httputil.ValidationError(notifications.N("INVALID_WEBHOOK_URL", "URL do webhook deve ser HTTP/HTTPS válida."))
	}
	events := make([]string, 0, len(input.Events))
	for _, e := range input.Events {
		ev := strings.ToUpper(strings.TrimSpace(e))
		if _, ok := allowedWebhookEvents[ev]; !ok {
			return nil, httputil.ValidationError(notifications.N("INVALID_WEBHOOK_EVENT", "Evento inválido. Use LINE_TRANSITION_ALERT, FIDELITY_EXPIRING_ALERT, DIVERGENCE_DETECTED ou BILLING_CLOSED."))
		}
		events = append(events, ev)
	}
	if len(events) == 0 {
		return nil, httputil.ValidationError(notifications.N("WEBHOOK_EVENTS_REQUIRED", "Informe ao menos um evento."))
	}

	now := time.Now().UTC()
	row := store.WebhookSubscriptionRow{
		ID: uuid.New().String(), OrganizationID: orgID, URL: url, Events: events,
		Secret: generateWebhookSecret(), IsActive: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.Store.InsertWebhookSubscription(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "CreateWebhook", "WebhookSubscription", row.ID, nil, map[string]any{
		"url": url, "events": events,
	})
	dto := webhookToDTO(row, true)
	return &dto, nil
}

func (s *Service) DeleteWebhook(ctx context.Context, id string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.DeleteWebhookSubscription(ctx, orgID, id); err != nil {
		return httputil.NotFoundError(notifications.N("WEBHOOK_NOT_FOUND", "Webhook não encontrado."))
	}
	s.auditLog(ctx, "DeleteWebhook", "WebhookSubscription", id, nil, map[string]any{"status": "deleted"})
	return nil
}

func (s *Service) TestWebhook(ctx context.Context, id string) (*models.TestWebhookResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetWebhookSubscription(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("WEBHOOK_NOT_FOUND", "Webhook não encontrado."))
	}
	payload := map[string]any{
		"event": "WEBHOOK_TEST", "occurred_at": time.Now().UTC(), "organization_id": orgID,
		"data": map[string]any{"subscription_id": id},
	}
	status, body, sendErr := postWebhook(ctx, row.URL, row.Secret, payload)
	success := sendErr == nil && status >= 200 && status < 300
	var errMsg *string
	if sendErr != nil {
		m := sendErr.Error()
		errMsg = &m
	}
	_ = s.Store.InsertWebhookDelivery(ctx, uuid.New().String(), row.ID, "WEBHOOK_TEST", &status, success, errMsg, time.Now().UTC())
	return &models.TestWebhookResponse{
		Success:      success,
		StatusCode:   status,
		ResponseBody: truncate(body, 2000),
	}, nil
}

func (s *Service) DispatchWebhookEvent(orgID, event string, data map[string]any) {
	if orgID == "" || event == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		rows, err := s.Store.ListActiveWebhooksForEvent(ctx, orgID, event)
		if err != nil || len(rows) == 0 {
			return
		}
		payload := map[string]any{
			"event": event, "occurred_at": time.Now().UTC(), "organization_id": orgID, "data": data,
		}
		for _, row := range rows {
			status, _, sendErr := postWebhook(ctx, row.URL, row.Secret, payload)
			success := sendErr == nil && status >= 200 && status < 300
			var errMsg *string
			if sendErr != nil {
				m := sendErr.Error()
				errMsg = &m
			}
			_ = s.Store.InsertWebhookDelivery(ctx, uuid.New().String(), row.ID, event, &status, success, errMsg, time.Now().UTC())
		}
	}()
}

func postWebhook(ctx context.Context, url, secret string, payload map[string]any) (int, string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Luxus-Signature", "sha256="+sig)
	if ev, _ := payload["event"].(string); ev != "" {
		req.Header.Set("X-Luxus-Event", ev)
	}

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4000))
	return resp.StatusCode, string(respBody), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func (s *Service) ListInventoryDevices(ctx context.Context) ([]models.InventoryDeviceResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	devs, _, err := s.Store.ListDeviceStockItems(ctx, orgID, nil, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return []models.InventoryDeviceResponse{}, nil
	}

	var items []models.InventoryDeviceResponse
	for _, d := range devs {
		imeiVal := ""
		if d.Imei != nil {
			imeiVal = *d.Imei
		}
		skuVal := d.Sku
		items = append(items, models.InventoryDeviceResponse{
			ID:            d.ID,
			Brand:         d.Brand,
			Model:         d.Model,
			IMEI:          imeiVal,
			SerialNumber:  &skuVal,
			Status:        d.Status,
			PurchaseValue: d.UnitCost,
			CreatedAt:     d.CreatedAt,
		})
	}

	return items, nil
}

func (s *Service) CreateInventoryDevice(ctx context.Context, input models.CreateInventoryDeviceInput) (*models.InventoryDeviceResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	brand := strings.TrimSpace(input.Brand)
	model := strings.TrimSpace(input.Model)
	imei := strings.TrimSpace(input.IMEI)

	if brand == "" || model == "" {
		return nil, httputil.ValidationError(notifications.N("BRAND_MODEL_REQUIRED", "Marca e modelo do dispositivo são obrigatórios."))
	}

	if !imeiRegex.MatchString(imei) {
		return nil, httputil.ValidationError(notifications.N("INVALID_IMEI", "IMEI deve conter exatamente 15 dígitos numéricos."))
	}

	exists, err := s.Store.DeviceStockImeiExists(ctx, orgID, imei, "")
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if exists {
		return nil, httputil.BusinessError(notifications.N("IMEI_DUPLICATE", "Já existe um aparelho com este IMEI."))
	}

	id := uuid.New().String()
	sku := "SKU-" + imei[len(imei)-6:]
	if input.SerialNumber != nil && strings.TrimSpace(*input.SerialNumber) != "" {
		sku = strings.TrimSpace(*input.SerialNumber)
	}

	now := time.Now().UTC()
	row := store.DeviceStockRow{
		ID:             id,
		OrganizationID: orgID,
		Sku:            sku,
		Brand:          brand,
		Model:          model,
		Imei:           &imei,
		UnitCost:       input.PurchaseValue,
		Status:         "in_stock",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.Store.CreateDeviceStockItem(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	s.auditLog(ctx, "CreateDevice", "Device", id, nil, map[string]any{
		"brand": brand,
		"model": model,
		"imei":  imei,
	})

	return &models.InventoryDeviceResponse{
		ID:            id,
		Brand:         brand,
		Model:         model,
		IMEI:          imei,
		SerialNumber:  &sku,
		Status:        "in_stock",
		PurchaseValue: input.PurchaseValue,
		CreatedAt:     now,
	}, nil
}

func (s *Service) UpdateInventoryDevice(ctx context.Context, id string, input models.UpdateInventoryDeviceInput) (*models.InventoryDeviceResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	dev, err := s.Store.GetDeviceStockItem(ctx, orgID, id)
	if err != nil || dev == nil {
		return nil, httputil.NotFoundError(notifications.DeviceStockNotFound)
	}

	brand := dev.Brand
	if input.Brand != nil {
		brand = strings.TrimSpace(*input.Brand)
	}
	model := dev.Model
	if input.Model != nil {
		model = strings.TrimSpace(*input.Model)
	}
	status := dev.Status
	if input.Status != nil {
		status = strings.TrimSpace(*input.Status)
	}

	now := time.Now().UTC()
	row := store.DeviceStockRow{
		ID:              dev.ID,
		OrganizationID:  orgID,
		Sku:             dev.Sku,
		Brand:           brand,
		Model:           model,
		Imei:            dev.Imei,
		Color:           dev.Color,
		StorageCapacity: dev.StorageCapacity,
		UnitCost:        input.PurchaseValue,
		SalePrice:       dev.SalePrice,
		Status:          status,
		Notes:           dev.Notes,
		UpdatedAt:       now,
	}

	if err := s.Store.UpdateDeviceStockItem(ctx, orgID, id, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	imeiVal := ""
	if dev.Imei != nil {
		imeiVal = *dev.Imei
	}

	return &models.InventoryDeviceResponse{
		ID:            dev.ID,
		Brand:         brand,
		Model:         model,
		IMEI:          imeiVal,
		SerialNumber:  &dev.Sku,
		Status:        status,
		PurchaseValue: input.PurchaseValue,
		CreatedAt:     dev.CreatedAt,
	}, nil
}

func (s *Service) ExportOrganizationData(ctx context.Context) (*models.OrganizationDataExportResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	if !auth.IsMaster(ctx) {
		return nil, httputil.BusinessError(notifications.N("MASTER_REQUIRED", "Apenas administradores Master podem exportar o dump completo da organização."))
	}

	custs, _, _ := s.Store.ListCustomers(ctx, orgID, nil, nil, httputil.PageSearch{PageSize: 5000})
	lines, _, _ := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 5000})
	invoices, _, _ := s.Store.ListProviderInvoices(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})

	type customerSnap struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		CpfCnpj string `json:"cpf_cnpj"`
		Active  bool   `json:"active"`
	}
	type lineSnap struct {
		ID     string `json:"id"`
		Number string `json:"number"`
		Status string `json:"status"`
	}
	custOut := make([]customerSnap, 0, len(custs))
	for _, c := range custs {
		custOut = append(custOut, customerSnap{ID: c.ID, Name: c.Name, CpfCnpj: c.CpfCnpj, Active: c.Active})
	}
	lineOut := make([]lineSnap, 0, len(lines))
	for _, l := range lines {
		lineOut = append(lineOut, lineSnap{ID: l.ID, Number: l.Number, Status: l.Status})
	}

	now := time.Now().UTC()
	payload := map[string]any{
		"organization_id": orgID,
		"exported_at":     now,
		"exported_by":     user.ID,
		"customers":       custOut,
		"phone_lines":     lineOut,
		"invoice_count":   len(invoices),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	sum := sha256.Sum256(raw)
	checksum := hex.EncodeToString(sum[:])
	summary := map[string]int{
		"total_lines":     len(lines),
		"total_customers": len(custs),
		"total_invoices":  len(invoices),
	}
	summaryJSON, _ := json.Marshal(summary)
	_ = s.Store.InsertOrganizationDataExport(ctx, uuid.New().String(), orgID, user.ID, checksum, len(raw), string(summaryJSON), now)

	s.auditLog(ctx, "OrganizationBackupExport", "Organization", orgID, nil, map[string]any{
		"exported_by": user.ID, "checksum": checksum, "payload_bytes": len(raw),
	})

	return &models.OrganizationDataExportResponse{
		OrganizationID: orgID,
		ExportedAt:     now,
		ExportedBy:     user.Email,
		ChecksumSHA256: checksum,
		Summary:        summary,
	}, nil
}
