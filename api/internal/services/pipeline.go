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
	"github.com/luxus-connect/telefonia/api/internal/billingcalc"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/observability"
	"github.com/luxus-connect/telefonia/api/internal/precision"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

var pipelineSteps = []struct {
	Key   string
	Label string
}{
	{"import", "Importar faturas"},
	{"validate", "Validar arquivos e chaves"},
	{"simulate", "Simular impacto"},
	{"identify_lines", "Identificar linhas"},
	{"stock_vs_clients", "Estoque vs clientes"},
	{"vigencias", "Vigências"},
	{"composition", "Composição"},
	{"prorata", "Pró-rata"},
	{"dependents", "Dependentes"},
	{"origin_accounts", "Contas de origem"},
	{"pendencies", "Pendências"},
	{"preview", "Prévia"},
	{"audit", "Auditoria"},
	{"release", "Liberação"},
	{"consolidate", "Consolidar"},
}

func (s *Service) RunProcessingMonthPipeline(ctx context.Context, monthID string) (*models.ProcessingMonthRunResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, monthID)
	if err != nil || month == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if month.Status != "open" {
		return nil, httputil.BusinessError(notifications.ProcessingMonthNotOpen)
	}
	start := time.Now()
	version, err := s.Store.NextProcessingMonthRunVersion(ctx, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	run := store.ProcessingMonthRunRow{
		ID: uuid.New().String(), OrganizationID: orgID, ProcessingMonthID: monthID,
		Version: version, Status: "running", TriggeredBy: &user.ID, CreatedAt: time.Now().UTC(),
	}
	if err := s.Store.InsertProcessingMonthRun(ctx, run); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	failed := false
	var failMsg string
	for i, spec := range pipelineSteps {
		step := store.ProcessingMonthRunStepRow{
			ID: uuid.New().String(), RunID: run.ID, StepKey: spec.Key, StepOrder: i + 1, Label: spec.Label, Status: "running",
		}
		_ = s.Store.InsertProcessingMonthRunStep(ctx, step)
		now := time.Now().UTC()
		step.StartedAt = &now
		summary, stepErr := s.runPipelineStep(ctx, orgID, monthID, spec.Key)
		done := time.Now().UTC()
		ms := int(done.Sub(now).Milliseconds())
		step.CompletedAt = &done
		step.DurationMs = &ms
		if stepErr != nil {
			step.Status = "failed"
			msg := stepErr.Error()
			step.Error = &msg
			failed = true
			failMsg = spec.Label + ": " + msg
		} else {
			step.Status = "done"
			if summary != "" {
				step.SummaryJSON = &summary
			}
		}
		_ = s.Store.UpdateProcessingMonthRunStep(ctx, step)
		observability.Observe("processing.pipeline.step."+spec.Key, done.Sub(now), stepErr == nil)
		meta := summary
		if stepErr != nil {
			meta = fmt.Sprintf(`{"error":%q}`, stepErr.Error())
		}
		_ = s.Store.InsertOperationMetric(ctx, uuid.New().String(), orgID, "processing.pipeline.step."+spec.Key, ms, stepErr == nil, optStr(meta))
		if failed {
			break
		}
	}
	completed := time.Now().UTC()
	status := "done"
	if failed {
		status = "failed"
	}
	sum := fmt.Sprintf(`{"failed":%v,"message":%q}`, failed, failMsg)
	_ = s.Store.UpdateProcessingMonthRun(ctx, run.ID, status, &sum, &completed)
	observability.Observe("processing.pipeline", time.Since(start), !failed)
	_ = s.Store.InsertOperationMetric(ctx, uuid.New().String(), orgID, "processing.pipeline", int(time.Since(start).Milliseconds()), !failed, &sum)
	return s.GetProcessingMonthRun(ctx, run.ID)
}

func pipelineSummary(data map[string]any, skipReason string) string {
	if data == nil {
		data = map[string]any{}
	}
	if skipReason != "" {
		data["skip_reason"] = skipReason
	}
	b, err := json.Marshal(data)
	if err != nil {
		return `{"skip_reason":"summary_marshal_failed"}`
	}
	return string(b)
}

func (s *Service) runPipelineStep(ctx context.Context, orgID, monthID, key string) (string, error) {
	month, err := s.Store.GetProcessingMonth(ctx, orgID, monthID)
	if err != nil {
		return "", err
	}
	if month == nil {
		return "", httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	switch key {
	case "import":
		return s.pipelineImport(ctx, orgID, monthID)
	case "validate":
		return s.pipelineValidate(ctx, orgID, monthID)
	case "simulate":
		return s.pipelineSimulate(ctx, monthID)
	case "identify_lines":
		return s.pipelineIdentifyLines(ctx, orgID, monthID)
	case "stock_vs_clients":
		return s.pipelineStockVsClients(ctx, orgID)
	case "vigencias":
		return s.pipelineVigencias(ctx, orgID, month)
	case "composition":
		return s.pipelineComposition(ctx, orgID)
	case "prorata":
		return s.pipelineProrata(ctx, orgID, month)
	case "dependents":
		return s.pipelineDependents(ctx, orgID)
	case "origin_accounts":
		return s.pipelineOriginAccounts(ctx, orgID, monthID)
	case "pendencies":
		return s.pipelinePendencies(ctx, monthID)
	case "preview":
		return s.pipelinePreview(ctx, monthID)
	case "audit":
		return s.pipelineAudit(ctx, orgID, monthID)
	case "release":
		return s.pipelineRelease(ctx, monthID)
	case "consolidate":
		return s.pipelineConsolidate(ctx, orgID, month, monthID)
	default:
		return "", fmt.Errorf("passo de pipeline desconhecido: %s", key)
	}
}

func (s *Service) pipelineImport(ctx context.Context, orgID, monthID string) (string, error) {
	inv, total, err := s.Store.ListProviderInvoices(ctx, orgID, &monthID, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return "", err
	}
	byStatus := map[string]int{}
	for _, i := range inv {
		byStatus[i.Status]++
	}
	skip := ""
	if total == 0 {
		skip = "no_invoices"
	}
	return pipelineSummary(map[string]any{
		"invoices": len(inv), "total": total, "by_status": byStatus,
	}, skip), nil
}

func (s *Service) pipelineValidate(ctx context.Context, orgID, monthID string) (string, error) {
	inv, _, err := s.Store.ListProviderInvoices(ctx, orgID, &monthID, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return "", err
	}
	if len(inv) == 0 {
		return pipelineSummary(map[string]any{"invoices": 0, "invalid": 0}, "no_invoices"), nil
	}
	invalid, missingLines := 0, 0
	for _, i := range inv {
		detail, err := s.Store.GetProviderInvoice(ctx, orgID, i.ID)
		if err != nil || detail == nil {
			invalid++
			continue
		}
		if detail.DueDate.IsZero() || detail.TotalAmount < 0 || detail.ProviderAccountID == "" {
			invalid++
		}
		if len(detail.PhoneLines) == 0 {
			missingLines++
		}
		if detail.Status == "pending" || detail.Status == "failed" {
			invalid++
		}
	}
	return pipelineSummary(map[string]any{
		"invoices": len(inv), "invalid": invalid, "missing_phone_lines": missingLines,
	}, ""), nil
}

func (s *Service) pipelineSimulate(ctx context.Context, monthID string) (string, error) {
	sim, err := s.SimulateBillingImpact(ctx, monthID)
	if err != nil {
		return "", err
	}
	skip := ""
	if sim.TotalActiveLines == 0 {
		skip = "no_active_lines"
	}
	return pipelineSummary(map[string]any{
		"projected_revenue": sim.ProjectedRevenue,
		"projected_cost":    sim.ProjectedCost,
		"projected_margin":  sim.ProjectedMargin,
		"active_lines":      sim.TotalActiveLines,
	}, skip), nil
}

func (s *Service) pipelineIdentifyLines(ctx context.Context, orgID, monthID string) (string, error) {
	inv, _, err := s.Store.ListProviderInvoices(ctx, orgID, &monthID, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return "", err
	}
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	catalog := map[string]string{}
	for _, l := range lines {
		catalog[httputil.NormalizeDigits(l.Number)] = l.ID
	}
	identified, unmatched := 0, 0
	seen := map[string]struct{}{}
	if len(inv) == 0 {
		return pipelineSummary(map[string]any{
			"invoices": 0, "identified": 0, "unmatched": 0, "catalog": len(catalog),
		}, "no_invoices"), nil
	}
	for _, i := range inv {
		detail, err := s.Store.GetProviderInvoice(ctx, orgID, i.ID)
		if err != nil || detail == nil {
			continue
		}
		for _, pl := range detail.PhoneLines {
			digits := httputil.NormalizeDigits(pl.Number)
			if digits == "" {
				continue
			}
			if _, ok := seen[digits]; ok {
				continue
			}
			seen[digits] = struct{}{}
			if _, ok := catalog[digits]; ok {
				identified++
			} else {
				unmatched++
			}
		}
	}
	skip := ""
	if len(seen) == 0 {
		skip = "no_invoice_phone_lines"
	}
	return pipelineSummary(map[string]any{
		"invoices": len(inv), "identified": identified, "unmatched": unmatched, "catalog": len(catalog),
	}, skip), nil
}

func (s *Service) pipelineStockVsClients(ctx context.Context, orgID string) (string, error) {
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	stock, assigned, activeUnlinked := 0, 0, 0
	for _, l := range lines {
		st := strings.ToLower(l.Status)
		if st == "stock" || st == "available" {
			stock++
			continue
		}
		linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
		if err != nil {
			return "", err
		}
		if linkID == "" {
			if st == "active" {
				activeUnlinked++
			}
			stock++
			continue
		}
		assigned++
	}
	skip := ""
	if len(lines) == 0 {
		skip = "no_phone_lines"
	}
	return pipelineSummary(map[string]any{
		"lines": len(lines), "stock_or_unlinked": stock, "assigned": assigned, "active_unlinked": activeUnlinked,
	}, skip), nil
}

func (s *Service) pipelineVigencias(ctx context.Context, orgID string, month *store.ProcessingMonthRow) (string, error) {
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	cycleStart := time.Date(month.Year, time.Month(month.Month), 1, 0, 0, 0, 0, time.UTC)
	active, overlapping, missing := 0, 0, 0
	for _, l := range lines {
		links, err := s.Store.ListPhoneLineCustomerLinks(ctx, orgID, l.ID)
		if err != nil {
			return "", err
		}
		if len(links) == 0 {
			missing++
			continue
		}
		hit := false
		for _, lk := range links {
			start := lk.StartDate
			if billingcalc.ActiveDays(&cycleStart, &start, lk.EndDate) > 0 {
				overlapping++
				hit = true
				break
			}
		}
		if hit {
			active++
		}
	}
	skip := ""
	if len(lines) == 0 {
		skip = "no_phone_lines"
	}
	return pipelineSummary(map[string]any{
		"lines": len(lines), "with_overlap": overlapping, "without_link": missing, "cycle": cycleStart.Format("2006-01"),
	}, skip), nil
}

func (s *Service) pipelineComposition(ctx context.Context, orgID string) (string, error) {
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	processings, items, withService := 0, 0, 0
	var total float64
	for _, l := range lines {
		linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
		if err != nil || linkID == "" {
			continue
		}
		rows, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
		if err != nil {
			return "", err
		}
		processings += len(rows)
		for _, p := range rows {
			comp, err := s.Store.ListBillingCompositionItems(ctx, p.ID)
			if err != nil {
				return "", err
			}
			items += len(comp)
			svc := false
			for _, ci := range comp {
				if ci.ItemType == "service" {
					svc = true
				}
				total = precision.SumCents(total, billingcalc.SignedItemAmount(ci.ItemType, ci.Amount, ci.Quantity))
			}
			if svc {
				withService++
			}
		}
	}
	skip := ""
	if items == 0 {
		skip = "no_composition_items"
	}
	return pipelineSummary(map[string]any{
		"processings": processings, "items": items, "with_service": withService, "total": total,
	}, skip), nil
}

func (s *Service) pipelineProrata(ctx context.Context, orgID string, month *store.ProcessingMonthRow) (string, error) {
	divisor, err := s.Store.GetProrataDivisor(ctx, orgID)
	if err != nil {
		divisor = billingcalc.CycleDays
	}
	divisor = billingcalc.NormalizeDivisor(divisor)
	cycleStart := time.Date(month.Year, time.Month(month.Month), 1, 0, 0, 0, 0, time.UTC)
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	adjusted, full, zeroDays := 0, 0, 0
	var proratedTotal float64
	for _, l := range lines {
		linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
		if err != nil || linkID == "" {
			continue
		}
		rows, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
		if err != nil {
			return "", err
		}
		for _, p := range rows {
			comp, err := s.Store.ListBillingCompositionItems(ctx, p.ID)
			if err != nil {
				return "", err
			}
			for _, ci := range comp {
				if !ci.Active {
					continue
				}
				fullAmt := billingcalc.SignedItemAmount(ci.ItemType, ci.Amount, ci.Quantity)
				pr := billingcalc.ProRataAmountWithDivisor(fullAmt, divisor, &cycleStart, ci.StartDate, ci.EndDate)
				proratedTotal = precision.SumCents(proratedTotal, pr)
				days := billingcalc.ActiveDaysWithDivisor(divisor, &cycleStart, ci.StartDate, ci.EndDate)
				if days == 0 {
					zeroDays++
				} else if days < divisor {
					adjusted++
				} else {
					full++
				}
			}
		}
	}
	skip := ""
	if adjusted+full+zeroDays == 0 {
		skip = "no_composition_items"
	}
	return pipelineSummary(map[string]any{
		"divisor": divisor, "adjusted": adjusted, "full_cycle": full, "zero_days": zeroDays, "prorated_total": proratedTotal,
	}, skip), nil
}

func (s *Service) pipelineDependents(ctx context.Context, orgID string) (string, error) {
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	titular, dependent, normal, orphan := 0, 0, 0, 0
	for _, l := range lines {
		switch strings.ToLower(l.LineClassification) {
		case "titular":
			titular++
			n, err := s.Store.CountDependents(ctx, l.ID)
			if err != nil {
				return "", err
			}
			if n == 0 {
				orphan++
			}
		case "dependent":
			dependent++
			if l.TitularLineID == nil || *l.TitularLineID == "" {
				orphan++
			}
		default:
			normal++
		}
	}
	skip := ""
	if len(lines) == 0 {
		skip = "no_phone_lines"
	}
	return pipelineSummary(map[string]any{
		"titular": titular, "dependent": dependent, "normal": normal, "orphan_links": orphan,
	}, skip), nil
}

func (s *Service) pipelineOriginAccounts(ctx context.Context, orgID, monthID string) (string, error) {
	inv, _, err := s.Store.ListProviderInvoices(ctx, orgID, &monthID, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return "", err
	}
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return "", err
	}
	invoiceAccounts := map[string]struct{}{}
	for _, i := range inv {
		if i.ProviderAccountID != "" {
			invoiceAccounts[i.ProviderAccountID] = struct{}{}
		}
	}
	lineAccounts := map[string]struct{}{}
	missingInvoice := 0
	for _, l := range lines {
		if l.ProviderAccountID == "" {
			continue
		}
		lineAccounts[l.ProviderAccountID] = struct{}{}
		if _, ok := invoiceAccounts[l.ProviderAccountID]; !ok {
			missingInvoice++
		}
	}
	skip := ""
	if len(inv) == 0 {
		skip = "no_invoices"
	}
	return pipelineSummary(map[string]any{
		"invoice_accounts":         len(invoiceAccounts),
		"line_accounts":            len(lineAccounts),
		"accounts_without_invoice": missingInvoice,
		"invoices":                 len(inv),
	}, skip), nil
}

func (s *Service) pipelinePendencies(ctx context.Context, monthID string) (string, error) {
	div, err := s.ListDivergences(ctx, monthID)
	if err != nil {
		return "", err
	}
	alerts, err := s.GetPreClosingAlerts(ctx, monthID)
	if err != nil {
		return "", err
	}
	openDiv := 0
	for _, d := range div {
		if d.Status != "resolved" && d.Status != "ignored" {
			openDiv++
		}
	}
	skip := ""
	if len(div) == 0 && alerts.WarningCount == 0 && alerts.CriticalCount == 0 {
		skip = "no_pendencies"
	}
	return pipelineSummary(map[string]any{
		"divergences": len(div), "open_divergences": openDiv,
		"alerts_warning": alerts.WarningCount, "alerts_critical": alerts.CriticalCount, "can_close": alerts.CanClose,
	}, skip), nil
}

func (s *Service) pipelinePreview(ctx context.Context, monthID string) (string, error) {
	ready, err := s.GetProcessingMonthLineReadiness(ctx, monthID)
	if err != nil {
		return "", err
	}
	sim, err := s.SimulateBillingImpact(ctx, monthID)
	if err != nil {
		return "", err
	}
	skip := ""
	if ready.TotalLines == 0 {
		skip = "no_active_lines"
	}
	return pipelineSummary(map[string]any{
		"total_lines": ready.TotalLines, "ready_lines": ready.ReadyLines, "blocked_lines": ready.BlockedLines,
		"projected_revenue": sim.ProjectedRevenue,
	}, skip), nil
}

func (s *Service) pipelineAudit(ctx context.Context, orgID, monthID string) (string, error) {
	logs, err := s.Store.ListAuditLogsForEntity(ctx, "ProcessingMonth", monthID, 50)
	if err != nil {
		return "", err
	}
	transitions, err := s.Store.ListStateTransitionLogs(ctx, orgID, "processing_month", monthID, 50)
	if err != nil {
		return "", err
	}
	skip := ""
	if len(logs) == 0 && len(transitions) == 0 {
		skip = "no_audit_trail"
	}
	return pipelineSummary(map[string]any{
		"audit_logs": len(logs), "state_transitions": len(transitions),
	}, skip), nil
}

func (s *Service) pipelineRelease(ctx context.Context, monthID string) (string, error) {
	alerts, err := s.GetPreClosingAlerts(ctx, monthID)
	if err != nil {
		return "", err
	}
	ready, err := s.GetProcessingMonthLineReadiness(ctx, monthID)
	if err != nil {
		return "", err
	}
	skip := "awaiting_explicit_release"
	if !alerts.CanClose || ready.BlockedLines > 0 {
		skip = "not_ready_for_release"
	}
	return pipelineSummary(map[string]any{
		"can_close": alerts.CanClose, "blocked_lines": ready.BlockedLines, "ready_lines": ready.ReadyLines,
		"critical_alerts": alerts.CriticalCount,
	}, skip), nil
}

func (s *Service) pipelineConsolidate(ctx context.Context, orgID string, month *store.ProcessingMonthRow, monthID string) (string, error) {
	user, err := userFrom(ctx)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	hashData := fmt.Sprintf("ORG=%s|MONTH=%s|CLOSED_BY=%s|TIMESTAMP=%d|STATUS=%s", orgID, monthID, user.ID, now.Unix(), month.Status)
	sum := sha256.Sum256([]byte(hashData))
	hash := hex.EncodeToString(sum[:])
	ready, err := s.GetProcessingMonthLineReadiness(ctx, monthID)
	if err != nil {
		return "", err
	}
	return pipelineSummary(map[string]any{
		"preview_hash": hash, "month_status": month.Status, "ready_lines": ready.ReadyLines, "blocked_lines": ready.BlockedLines,
	}, "awaiting_explicit_close"), nil
}

func (s *Service) ListProcessingMonthRuns(ctx context.Context, monthID string) ([]models.ProcessingMonthRunResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.Store.ListProcessingMonthRuns(ctx, orgID, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	out := make([]models.ProcessingMonthRunResponse, 0, len(rows))
	for _, r := range rows {
		item, err := s.GetProcessingMonthRun(ctx, r.ID)
		if err != nil {
			continue
		}
		out = append(out, *item)
	}
	return out, nil
}

func (s *Service) GetProcessingMonthRun(ctx context.Context, runID string) (*models.ProcessingMonthRunResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	run, err := s.Store.GetProcessingMonthRun(ctx, orgID, runID)
	if err != nil || run == nil {
		return nil, httputil.NotFoundError(notifications.N("PIPELINE_RUN_NOT_FOUND", "Execução do pipeline não encontrada."))
	}
	steps, err := s.Store.ListProcessingMonthRunSteps(ctx, run.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	resp := &models.ProcessingMonthRunResponse{
		ID: run.ID, ProcessingMonthID: run.ProcessingMonthID, Version: run.Version,
		Status: run.Status, TriggeredBy: run.TriggeredBy, CreatedAt: run.CreatedAt, CompletedAt: run.CompletedAt,
	}
	for _, st := range steps {
		resp.Steps = append(resp.Steps, models.ProcessingMonthRunStepResponse{
			Key: st.StepKey, Order: st.StepOrder, Label: st.Label, Status: st.Status,
			StartedAt: st.StartedAt, CompletedAt: st.CompletedAt, DurationMs: st.DurationMs, Error: st.Error, SummaryJSON: st.SummaryJSON,
		})
	}
	if resp.Steps == nil {
		resp.Steps = []models.ProcessingMonthRunStepResponse{}
	}
	return resp, nil
}
