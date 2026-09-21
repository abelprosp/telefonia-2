package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"github.com/xuri/excelize/v2"
)

func (s *Service) ListDomainAuditEvents(ctx context.Context, entityType, entityID, action, actor, phoneLineID, customerID string, from, to *time.Time, page httputil.PageSearch) ([]models.DomainAuditEventResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if !auth.HasRole(ctx, auth.RoleMaster) && !auth.HasRole(ctx, auth.RoleAdmin) && !auth.HasRole(ctx, auth.RoleFinancial) {
		return nil, 0, httputil.ForbiddenError(notifications.N("AUDIT_FORBIDDEN", "Sem permissão para consultar o histórico de auditoria."))
	}
	rows, total, err := s.Store.ListDomainAuditEvents(ctx, orgID, store.DomainAuditFilter{
		EntityType:  entityType,
		EntityID:    entityID,
		Action:      action,
		ActorUserID: actor,
		From:        from,
		To:          to,
		PhoneLineID: phoneLineID,
		CustomerID:  customerID,
	}, page)
	if err != nil {
		return nil, 0, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	items := make([]models.DomainAuditEventResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, models.DomainAuditEventResponse{
			ID: r.ID, EntityType: r.EntityType, EntityID: r.EntityID, Action: r.Action,
			ActorUserID: r.ActorUserID, ActorKind: r.ActorKind, Source: r.Source,
			CorrelationID: r.CorrelationID, BeforeJSON: r.BeforeJSON, AfterJSON: r.AfterJSON,
			MetadataJSON: r.MetadataJSON, CreatedAt: r.CreatedAt,
		})
	}
	return items, total, nil
}

func (s *Service) ListLineConsumptionReport(ctx context.Context, f store.ConsumptionReportFilter, page httputil.PageSearch) (*models.LineConsumptionReportResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	items, total, sum, err := s.Store.ListLineConsumptionReport(ctx, orgID, f, page)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &models.LineConsumptionReportResponse{Items: items, TotalCount: total, TotalAmount: sum}, nil
}

func (s *Service) ExportLineConsumptionReport(ctx context.Context, f store.ConsumptionReportFilter, format string) ([]byte, string, error) {
	page := httputil.PageSearch{PageIndex: 0, PageSize: 100}
	var all []models.LineConsumptionReportItem
	for {
		resp, err := s.ListLineConsumptionReport(ctx, f, page)
		if err != nil {
			return nil, "", err
		}
		all = append(all, resp.Items...)
		if int64(len(all)) >= resp.TotalCount || len(resp.Items) == 0 {
			break
		}
		page.PageIndex++
		if page.PageIndex > 1000 {
			break
		}
	}
	switch strings.ToLower(format) {
	case "txt":
		return consumptionTXT(all), "text/plain; charset=utf-8", nil
	case "csv":
		return consumptionCSV(all), "text/csv; charset=utf-8", nil
	case "xlsx":
		b, err := consumptionXLSX(all)
		if err != nil {
			return nil, "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		return b, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
	case "pdf":
		return consumptionPDF(all), "application/pdf", nil
	default:
		return nil, "", httputil.ValidationError(notifications.N("EXPORT_FORMAT_INVALID", "Formato deve ser pdf, txt ou xlsx."))
	}
}

func consumptionXLSX(items []models.LineConsumptionReportItem) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	sheet := f.GetSheetName(0)
	headers := []string{
		"phone_line_id", "phone_number", "customer", "provider", "competence", "status",
		"base_cost", "cost_with_consumption", "invoice_id", "invoice_total",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
	for r, it := range items {
		cust := ""
		if it.CustomerName != nil {
			cust = *it.CustomerName
		}
		row := []any{
			it.PhoneLineID, it.PhoneNumber, cust, it.ProviderName,
			fmt.Sprintf("%02d/%04d", it.Month, it.Year), it.Status,
			fmtFloat(it.BaseCost), fmtFloat(it.CostWithConsumption),
			it.InvoiceID, it.InvoiceTotal,
		}
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func consumptionCSV(items []models.LineConsumptionReportItem) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"phone_line_id", "phone_number", "customer", "provider", "competence", "status",
		"base_cost", "cost_with_consumption", "invoice_id", "invoice_total",
	})
	for _, it := range items {
		cust := ""
		if it.CustomerName != nil {
			cust = *it.CustomerName
		}
		_ = w.Write([]string{
			it.PhoneLineID, it.PhoneNumber, cust, it.ProviderName,
			fmt.Sprintf("%02d/%04d", it.Month, it.Year), it.Status,
			fmtFloat(it.BaseCost), fmtFloat(it.CostWithConsumption),
			it.InvoiceID, strconv.FormatFloat(it.InvoiceTotal, 'f', 2, 64),
		})
	}
	w.Flush()
	return buf.Bytes()
}

func consumptionTXT(items []models.LineConsumptionReportItem) []byte {
	var b strings.Builder
	b.WriteString("Relatório de linhas e consumo\n")
	b.WriteString(strings.Repeat("-", 72) + "\n")
	for _, it := range items {
		cust := "—"
		if it.CustomerName != nil {
			cust = *it.CustomerName
		}
		b.WriteString(fmt.Sprintf("%s | %s | %s | %02d/%04d | base=%s consumo=%s\n",
			it.PhoneNumber, cust, it.ProviderName, it.Month, it.Year,
			fmtFloat(it.BaseCost), fmtFloat(it.CostWithConsumption)))
	}
	return []byte(b.String())
}

func consumptionPDF(items []models.LineConsumptionReportItem) []byte {
	// Minimal single-page PDF with plain text content (no invented metrics).
	body := string(consumptionTXT(items))
	if len(body) > 3500 {
		body = body[:3500] + "\n...(truncatedado)"
	}
	escaped := strings.ReplaceAll(body, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `(`, `\(`)
	escaped = strings.ReplaceAll(escaped, `)`, `\)`)
	stream := "BT /F1 9 Tf 40 800 Td 12 TL (" + strings.ReplaceAll(escaped, "\n", ") Tj T* (") + ") Tj ET"
	objects := []string{
		"1 0 obj<< /Type /Catalog /Pages 2 0 R >>endobj\n",
		"2 0 obj<< /Type /Pages /Kids [3 0 R] /Count 1 >>endobj\n",
		"3 0 obj<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>endobj\n",
		fmt.Sprintf("4 0 obj<< /Length %d >>stream\n%s\nendstream\nendobj\n", len(stream), stream),
		"5 0 obj<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>endobj\n",
	}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for _, obj := range objects {
		offsets = append(offsets, out.Len())
		out.WriteString(obj)
	}
	xref := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n", len(offsets)))
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(offsets); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(offsets), xref))
	return out.Bytes()
}

func fmtFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func (s *Service) ListMissingInOperator(ctx context.Context, source, providerID, jobID string, isPartial bool) ([]models.ReconciliationFindingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if source == "" {
		source = "axon"
	}
	items, err := s.Store.ListLinesMissingFromExternalSnapshot(ctx, orgID, source, providerID, jobID, isPartial)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func (s *Service) ListCancelledExternallyActiveInternally(ctx context.Context, source, jobID string) ([]models.ReconciliationFindingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if source == "" {
		source = "axon"
	}
	items, err := s.Store.ListCancelledExternallyActiveInternally(ctx, orgID, source, jobID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func (s *Service) RegisterExternalLineImportJob(ctx context.Context, input models.ExternalLineImportJobInput) (*models.ExternalLineImportJobResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	source := strings.TrimSpace(strings.ToLower(input.Source))
	if source == "" {
		return nil, httputil.ValidationError(notifications.N("EXTERNAL_SOURCE_REQUIRED", "Informe a origem (ex.: axon)."))
	}
	layout := strings.TrimSpace(input.LayoutCode)
	if layout == "" {
		return nil, httputil.ValidationError(notifications.N(
			"EXTERNAL_LAYOUT_REQUIRED",
			"Informe o layout_code correspondente ao arquivo real da origem. O parser não inventa colunas.",
		))
	}
	if strings.TrimSpace(input.StorageBucket) == "" || strings.TrimSpace(input.StorageObjectKey) == "" {
		return nil, httputil.ValidationError(notifications.ImportStorageObjectKeyRequired)
	}
	id := uuid.New().String()
	var actor *string
	if u := auth.UserFromContext(ctx); u != nil && u.ID != "" {
		actor = &u.ID
	}
	now := time.Now().UTC()
	msg := "Job registrado. O parser definitivo exige o layout real da Axon (layout_code=" + layout + "). Nenhum dado foi inventado."
	if err := s.Store.InsertExternalLineImportJob(ctx, id, orgID, source, input.ProviderID, input.StorageBucket, input.StorageObjectKey,
		input.FileName, layout, input.ReferencePeriod, input.IsPartialSource, msg, actor, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.recordDomainAudit(ctx, orgID, "external_line_import", id, "register", map[string]any{
		"source": source, "layout_code": layout, "partial": input.IsPartialSource,
	}, nil)
	return &models.ExternalLineImportJobResponse{
		ID: id, Source: source, Status: "awaiting_layout", LayoutCode: layout,
		IsPartialSource: input.IsPartialSource, ErrorMessage: &msg, CreatedAt: now,
	}, nil
}

type ApplyCancelledExternallyInput struct {
	PhoneLineIDs []string `json:"phone_line_ids"`
	Source       string   `json:"source"`
	ImportJobID  *string  `json:"import_job_id,omitempty"`
	Confirm      bool     `json:"confirm"`
}

type ApplyCancelledExternallyResponse struct {
	Applied int      `json:"applied"`
	Skipped int      `json:"skipped"`
	Errors  []string `json:"errors,omitempty"`
}

func (s *Service) ApplyCancelledExternallyActive(ctx context.Context, input ApplyCancelledExternallyInput) (*ApplyCancelledExternallyResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if !input.Confirm {
		return nil, httputil.ValidationError(notifications.N(
			"RECON_CONFIRM_REQUIRED",
			"Confirmação explícita obrigatória para cancelar linhas com base na conciliação externa.",
		))
	}
	if len(input.PhoneLineIDs) == 0 {
		return nil, httputil.ValidationError(notifications.N("RECON_LINES_REQUIRED", "Informe ao menos uma linha."))
	}
	if len(input.PhoneLineIDs) > 200 {
		return nil, httputil.ValidationError(notifications.N("RECON_LINES_LIMIT", "Máximo de 200 linhas por lote."))
	}

	out := &ApplyCancelledExternallyResponse{}
	jobID := ""
	if input.ImportJobID != nil {
		jobID = *input.ImportJobID
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = "axon"
	}
	// Re-validate against current findings so stale selections are not cancelled blindly.
	findings, err := s.Store.ListCancelledExternallyActiveInternally(ctx, orgID, source, jobID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	allowed := map[string]struct{}{}
	for _, f := range findings {
		if f.PhoneLineID != nil && !f.IsInconclusive {
			allowed[*f.PhoneLineID] = struct{}{}
		}
	}

	for _, id := range input.PhoneLineIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			out.Skipped++
			continue
		}
		if _, ok := allowed[id]; !ok {
			out.Skipped++
			out.Errors = append(out.Errors, id+": não consta como divergência confirmada na base atual")
			continue
		}
		line, err := s.GetPhoneLine(ctx, id)
		if err != nil {
			out.Skipped++
			out.Errors = append(out.Errors, id+": "+err.Error())
			continue
		}
		if line.Status == "cancelled" {
			out.Skipped++
			continue
		}
		if err := s.Store.UpdatePhoneLineStatus(ctx, id, "cancelled"); err != nil {
			out.Skipped++
			out.Errors = append(out.Errors, id+": "+err.Error())
			continue
		}
		_ = s.recordDomainAudit(ctx, orgID, "phone_line", id, "reconcile_cancel_external", map[string]any{
			"status": "cancelled", "source": source, "import_job_id": jobID,
		}, map[string]any{"status": line.Status})
		out.Applied++
	}
	return out, nil
}
