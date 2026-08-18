package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/precision"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) GetOperationalDashboard(ctx context.Context) (*models.OperationalDashboardResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	lines, totalLines, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 2000})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var activeLines, inTransitionLines, canceledLines, orphanLines int32
	var totalBaseCost, projectedMonthlyRevenue float64

	for _, l := range lines {
		switch l.Status {
		case "active":
			activeLines++
			if l.BaseCost != nil {
				totalBaseCost = precision.SumCents(totalBaseCost, *l.BaseCost)
			}
			linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
			if err != nil || linkID == "" {
				orphanLines++
			} else {
				processings, _ := s.Store.ListBillingProcessingsForLink(ctx, linkID)
				for _, p := range processings {
					if p.Perspective == perspectiveLuxusCustomer {
						tot, _ := s.Store.SumBillingProcessingTotal(ctx, p.ID)
						projectedMonthlyRevenue = precision.SumCents(projectedMonthlyRevenue, tot)
						break
					}
				}
			}
		case "in_transition":
			inTransitionLines++
		case "canceled":
			canceledLines++
		}
	}

	custCount, _ := s.Store.CountDashboardCustomers(ctx, orgID)
	projectedMargin := precision.Round2(projectedMonthlyRevenue - totalBaseCost)
	marginPct := 0.0
	if projectedMonthlyRevenue > 0 {
		marginPct = precision.Round2((projectedMargin / projectedMonthlyRevenue) * 100.0)
	}

	// Identificar mês de processamento aberto atual
	var currentMonthStatus *models.OperationalDashboardMonthStatus
	months, _, _ := s.Store.ListProcessingMonths(ctx, orgID, httputil.PageSearch{PageSize: 1})
	if len(months) > 0 {
		m := months[0]
		alerts, _ := s.GetPreClosingAlerts(ctx, m.ID)
		crit := 0
		warn := 0
		if alerts != nil {
			crit = alerts.CriticalCount
			warn = alerts.WarningCount
		}
		currentMonthStatus = &models.OperationalDashboardMonthStatus{
			ProcessingMonthID: m.ID,
			DisplayName:       m.DisplayName,
			Status:            m.Status,
			CriticalAlerts:    crit,
			WarningAlerts:     warn,
		}
	}

	openDiv, _ := s.Store.CountOpenDivergences(ctx, orgID)

	return &models.OperationalDashboardResponse{
		LinesSummary: models.OperationalDashboardLinesSummary{
			TotalLines:        int32(totalLines),
			ActiveLines:       activeLines,
			InTransitionLines: inTransitionLines,
			CanceledLines:     canceledLines,
			OrphanLines:       orphanLines,
		},
		CustomersSummary: models.OperationalDashboardCustomersSummary{
			TotalCustomers:  custCount,
			ActiveCustomers: custCount,
		},
		FinancialSummary: models.OperationalDashboardFinancialSummary{
			ProjectedMonthlyRevenue: projectedMonthlyRevenue,
			TotalBaseCost:           totalBaseCost,
			ProjectedMargin:         projectedMargin,
			MarginPercentage:        marginPct,
		},
		CurrentMonthStatus: currentMonthStatus,
		PendingDivergences: openDiv,
	}, nil
}

func (s *Service) GetPhoneLine360(ctx context.Context, phoneLineID string) (*models.PhoneLine360Response, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}

	// Vínculo ativo
	var activeLink *models.PhoneLineCustomerLinkResponse
	links, err := s.Store.ListPhoneLineCustomerLinks(ctx, orgID, phoneLineID)
	if err == nil {
		for _, l := range links {
			if l.IsActive {
				cl := l
				activeLink = &cl
				break
			}
		}
	}

	// Fidelidade
	fidelity, _ := s.GetLineFidelity(ctx, phoneLineID)
	var penaltyEst *models.FidelityPenaltyEstimateResponse
	if fidelity != nil && fidelity.Status == "active" {
		penaltyEst, _ = s.EstimateLineFidelityPenalty(ctx, phoneLineID, nil, nil)
	}

	// Explicabilidade da cobrança
	explanation, _ := s.GetLineBillingExplanation(ctx, phoneLineID, "")

	// Timeline
	timeline, err := s.GetPhoneLineTimeline(ctx, phoneLineID)
	var recentEvents []models.PhoneLineTimelineEvent
	if timeline != nil {
		recentEvents = timeline.Events
	}

	return &models.PhoneLine360Response{
		Line:               *line,
		ActiveCustomerLink: activeLink,
		ActiveFidelity:     fidelity,
		PenaltyEstimate:    penaltyEst,
		BillingExplanation: explanation,
		RecentTimeline:     recentEvents,
	}, nil
}

func (s *Service) GetCustomer360(ctx context.Context, customerID string) (*models.Customer360Response, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	cust, err := s.Store.GetCustomer(ctx, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if cust == nil {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}

	lines, totalLines, err := s.Store.ListCustomerPhoneLines(ctx, orgID, customerID, httputil.PageSearch{PageSize: 500})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	activeCount := 0
	var totalMonthly float64
	var lineItems []models.ListPhoneLineResponse

	for _, cl := range lines {
		if cl.IsActive {
			activeCount++
			if cl.MonthlyAmount != nil {
				totalMonthly = precision.SumCents(totalMonthly, *cl.MonthlyAmount)
			}
		}
		l, err := s.Store.GetPhoneLine(ctx, orgID, cl.PhoneLineID)
		if err == nil && l != nil {
			lineItems = append(lineItems, l.ListPhoneLineResponse)
		}
	}

	contracts, err := s.Store.ListGeneratedContractsForCustomer(ctx, orgID, customerID)
	if err != nil {
		contracts = []models.GeneratedContractResponse{}
	}

	attachments, _ := s.Store.ListCustomerAttachments(ctx, orgID, customerID)

	return &models.Customer360Response{
		Customer:           *cust,
		TotalLinesCount:    int(totalLines),
		ActiveLinesCount:   activeCount,
		TotalMonthlyAmount: totalMonthly,
		PhoneLines:         lineItems,
		GeneratedContracts: contracts,
		AttachmentsCount:   len(attachments),
	}, nil
}

func (s *Service) ListDivergences(ctx context.Context, processingMonthID string) ([]models.DivergenceItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	page := httputil.PageSearch{PageSize: 200}
	rows, _, err := s.Store.ListBillingDivergences(ctx, orgID, processingMonthID, "", page)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if len(rows) == 0 {
		_ = s.seedDivergencesFromLines(ctx, orgID, processingMonthID)
		rows, _, err = s.Store.ListBillingDivergences(ctx, orgID, processingMonthID, "", page)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	out := make([]models.DivergenceItemResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, divergenceToDTO(r))
	}
	return out, nil
}

func (s *Service) seedDivergencesFromLines(ctx context.Context, orgID, processingMonthID string) error {
	lines, _, err := s.Store.ListPhoneLines(ctx, orgID, nil, httputil.PageSearch{PageSize: 1000})
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, l := range lines {
		if l.Status != "active" {
			continue
		}
		linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, l.ID)
		if err != nil || linkID == "" {
			num := l.Number
			_ = s.Store.InsertBillingDivergence(ctx, store.BillingDivergenceRow{
				ID: uuid.New().String(), OrganizationID: orgID, ProcessingMonthID: optStr(processingMonthID),
				DivergenceType: "LINE_NOT_IN_SYSTEM", Severity: "HIGH", PhoneLineID: &l.ID, PhoneNumber: &num,
				Status: "open", Cause: strPtr("Linha ativa sem cliente vinculado."),
				RecommendedAction: strPtr("Vincular cliente ou retornar ao estoque."),
				FinancialImpact: 0, CreatedAt: now, UpdatedAt: now,
			})
			s.DispatchWebhookEvent(orgID, WebhookEventDivergence, map[string]any{
				"phone_line_id": l.ID, "type": "LINE_NOT_IN_SYSTEM",
			})
		}
	}
	return nil
}

func (s *Service) GetDivergence(ctx context.Context, id string) (*models.DivergenceDetailResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetBillingDivergence(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("DIVERGENCE_NOT_FOUND", "Divergência não encontrada."))
	}
	comments, _ := s.Store.ListDivergenceComments(ctx, id)
	hist, _ := s.Store.ListDivergenceHistory(ctx, id)
	dto := divergenceToDTO(*row)
	detail := models.DivergenceDetailResponse{DivergenceItemResponse: dto}
	for _, c := range comments {
		detail.Comments = append(detail.Comments, models.DivergenceCommentResponse{ID: c.ID, AuthorUserID: c.AuthorUserID, Body: c.Body, CreatedAt: c.CreatedAt})
	}
	for _, h := range hist {
		detail.History = append(detail.History, models.DivergenceHistoryResponse{ID: h.ID, ActorUserID: h.ActorUserID, EventType: h.EventType, FromValue: h.FromValue, ToValue: h.ToValue, Notes: h.Notes, CreatedAt: h.CreatedAt})
	}
	if detail.Comments == nil {
		detail.Comments = []models.DivergenceCommentResponse{}
	}
	if detail.History == nil {
		detail.History = []models.DivergenceHistoryResponse{}
	}
	return &detail, nil
}

func (s *Service) AssignDivergence(ctx context.Context, id, ownerID string) (*models.DivergenceItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetBillingDivergence(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("DIVERGENCE_NOT_FOUND", "Divergência não encontrada."))
	}
	now := time.Now().UTC()
	from := row.OwnerUserID
	row.OwnerUserID = optStr(ownerID)
	row.UpdatedAt = now
	if err := s.Store.UpdateBillingDivergence(ctx, *row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	var fromStr *string
	if from != nil {
		fromStr = from
	}
	_ = s.Store.InsertDivergenceHistory(ctx, store.BillingDivergenceHistoryRow{
		ID: uuid.New().String(), DivergenceID: id, ActorUserID: &user.ID, EventType: "assign", FromValue: fromStr, ToValue: &ownerID, CreatedAt: now,
	})
	dto := divergenceToDTO(*row)
	return &dto, nil
}

func (s *Service) CommentDivergence(ctx context.Context, id, body string) (*models.DivergenceDetailResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetBillingDivergence(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("DIVERGENCE_NOT_FOUND", "Divergência não encontrada."))
	}
	now := time.Now().UTC()
	_ = s.Store.InsertDivergenceComment(ctx, store.BillingDivergenceCommentRow{
		ID: uuid.New().String(), DivergenceID: id, AuthorUserID: &user.ID, Body: strings.TrimSpace(body), CreatedAt: now,
	})
	return s.GetDivergence(ctx, id)
}

func (s *Service) ResolveDivergence(ctx context.Context, id string, input models.ResolveDivergenceInput) (*models.DivergenceItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	row, err := s.Store.GetBillingDivergence(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("DIVERGENCE_NOT_FOUND", "Divergência não encontrada."))
	}
	action := strings.ToLower(strings.TrimSpace(input.Action))
	now := time.Now().UTC()
	from := row.Status
	switch action {
	case "resolve":
		row.Status = "resolved"
		row.ResolvedAt = &now
		row.ResolvedBy = &user.ID
	case "ignore":
		row.Status = "ignored"
		row.ResolvedAt = &now
		row.ResolvedBy = &user.ID
	case "reopen":
		row.Status = "open"
		row.ResolvedAt = nil
		row.ResolvedBy = nil
	default:
		return nil, httputil.ValidationError(notifications.N("INVALID_DIVERGENCE_ACTION", "Action must be resolve, ignore or reopen."))
	}
	notes := strings.TrimSpace(input.Notes)
	row.ResolutionNotes = optStr(notes)
	row.UpdatedAt = now
	if err := s.Store.UpdateBillingDivergence(ctx, *row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.Store.InsertDivergenceHistory(ctx, store.BillingDivergenceHistoryRow{
		ID: uuid.New().String(), DivergenceID: id, ActorUserID: &user.ID, EventType: action, FromValue: &from, ToValue: &row.Status, Notes: optStr(notes), CreatedAt: now,
	})
	s.auditLog(ctx, "ResolveDivergence", "Divergence", id, map[string]any{"status": from}, map[string]any{"status": row.Status, "notes": notes})
	dto := divergenceToDTO(*row)
	return &dto, nil
}

func divergenceToDTO(r store.BillingDivergenceRow) models.DivergenceItemResponse {
	ageHours := int(time.Since(r.CreatedAt).Hours())
	return models.DivergenceItemResponse{
		ID: r.ID, ProcessingMonthID: deref(r.ProcessingMonthID), DivergenceType: r.DivergenceType,
		Severity: r.Severity, PhoneNumber: r.PhoneNumber, Description: deref(r.Cause),
		Status: r.Status, ResolvedBy: r.ResolvedBy, ResolutionNotes: r.ResolutionNotes, CreatedAt: r.CreatedAt,
		OwnerUserID: r.OwnerUserID, FinancialImpact: r.FinancialImpact, Competence: r.Competence,
		OperatorName: r.OperatorName, AccountNumber: r.AccountNumber, CustomerID: r.CustomerID,
		PhoneLineID: r.PhoneLineID, RecommendedAction: r.RecommendedAction, Evidence: r.Evidence, AgeHours: ageHours,
	}
}
