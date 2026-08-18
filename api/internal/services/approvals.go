package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

var twoLevelActions = map[string]struct{}{
	"reopen_month": {}, "consolidation": {}, "forced_charge": {},
	"batch_change": {}, "high_discount": {}, "retroactive_cancel": {}, "cycle_impact": {},
}

type approvalExecKey struct{}

func withApprovalExecution(ctx context.Context) context.Context {
	return context.WithValue(ctx, approvalExecKey{}, true)
}

func isApprovalExecution(ctx context.Context) bool {
	v, _ := ctx.Value(approvalExecKey{}).(bool)
	return v
}

func pendingTwoLevelApproval() error {
	return httputil.BusinessError(notifications.N("APPROVAL_PENDING", "Ação enviada para aprovação em dois níveis. Quem solicita não pode aprovar a si mesmo."))
}

func payloadAsMap(v any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	out := map[string]any{}
	if err := json.Unmarshal(b, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func (s *Service) queueTwoLevel(ctx context.Context, action, entityType, entityID, justification string, payload any) error {
	if isApprovalExecution(ctx) {
		return nil
	}
	_, err := s.RequestTwoLevelApproval(ctx, models.CreateApprovalRequestInput{
		ActionType: action, EntityType: entityType, EntityID: entityID,
		Justification: justification, Payload: payloadAsMap(payload),
	})
	if err != nil {
		return err
	}
	return pendingTwoLevelApproval()
}

func (s *Service) RequestTwoLevelApproval(ctx context.Context, input models.CreateApprovalRequestInput) (*models.ApprovalRequestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	action := strings.TrimSpace(input.ActionType)
	if _, ok := twoLevelActions[action]; !ok {
		return nil, httputil.ValidationError(notifications.N("APPROVAL_ACTION_INVALID", "Tipo de aprovação não suportado."))
	}
	if strings.TrimSpace(input.EntityID) == "" {
		return nil, httputil.ValidationError(notifications.N("APPROVAL_ENTITY_REQUIRED", "Informe a entidade da aprovação."))
	}
	if existing, err := s.Store.FindOpenApproval(ctx, orgID, action, input.EntityID); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	} else if existing != nil {
		return toApprovalResponse(*existing), nil
	}
	now := time.Now().UTC()
	before, _ := json.Marshal(input.Before)
	payload, _ := json.Marshal(input.Payload)
	beforeStr := string(before)
	payloadStr := string(payload)
	row := store.ApprovalRequestRow{
		ID: uuid.New().String(), OrganizationID: orgID, ActionType: action,
		EntityType: firstNonEmpty(input.EntityType, action), EntityID: input.EntityID,
		Status: "pending_first", RequesterUserID: user.ID, Justification: optStr(input.Justification),
		PayloadJSON: &payloadStr, BeforeJSON: &beforeStr, CreatedAt: now,
	}
	if err := s.Store.InsertApprovalRequest(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "RequestApproval", "ApprovalRequest", row.ID, nil, map[string]any{
		"action": action, "entity_id": input.EntityID, "requester": user.ID,
	})
	return toApprovalResponse(row), nil
}

func (s *Service) ListApprovalRequests(ctx context.Context, status string) ([]models.ApprovalRequestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.Store.ListApprovalRequests(ctx, orgID, status, 100)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	out := make([]models.ApprovalRequestResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, *toApprovalResponse(r))
	}
	return out, nil
}

func (s *Service) ApproveRequest(ctx context.Context, id string) (*models.ApprovalRequestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	if !auth.CanApproveTwoLevel(ctx) {
		return nil, httputil.ForbiddenError(notifications.N("APPROVAL_ROLE_REQUIRED", "Apenas master/admin/financeiro podem aprovar."))
	}
	row, err := s.Store.GetApprovalRequest(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row == nil {
		return nil, httputil.NotFoundError(notifications.N("APPROVAL_NOT_FOUND", "Pedido de aprovação não encontrado."))
	}
	if row.RequesterUserID == user.ID {
		return nil, httputil.BusinessError(notifications.N("APPROVAL_SELF_FORBIDDEN", "Quem solicita não pode aprovar a si mesmo."))
	}
	now := time.Now().UTC()
	switch row.Status {
	case "pending_first":
		row.Status = "pending_second"
		row.FirstApproverUserID = &user.ID
		row.FirstApprovedAt = &now
		s.auditLog(ctx, "ApproveFirst", "ApprovalRequest", row.ID, map[string]any{"status": "pending_first"}, map[string]any{"status": "pending_second", "approver": user.ID})
	case "pending_second":
		if row.FirstApproverUserID != nil && *row.FirstApproverUserID == user.ID {
			return nil, httputil.BusinessError(notifications.N("APPROVAL_DISTINCT_REQUIRED", "A segunda aprovação deve ser de outro usuário privilegiado."))
		}
		row.Status = "approved"
		row.SecondApproverUserID = &user.ID
		row.SecondApprovedAt = &now
		if err := s.executeApprovedAction(ctx, row); err != nil {
			return nil, err
		}
		row.ExecutedAt = &now
		s.auditLog(ctx, "ApproveSecond", "ApprovalRequest", row.ID,
			map[string]any{"status": "pending_second"},
			map[string]any{"status": "approved", "approver": user.ID, "executed": true})
	default:
		return nil, httputil.BusinessError(notifications.N("APPROVAL_NOT_PENDING", "Este pedido já foi concluído."))
	}
	if err := s.Store.UpdateApprovalRequest(ctx, *row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return toApprovalResponse(*row), nil
}

func (s *Service) RejectApprovalRequest(ctx context.Context, id string, reason string) (*models.ApprovalRequestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	if !auth.CanApproveTwoLevel(ctx) {
		return nil, httputil.ForbiddenError(notifications.N("APPROVAL_ROLE_REQUIRED", "Apenas master/admin/financeiro podem recusar."))
	}
	row, err := s.Store.GetApprovalRequest(ctx, orgID, id)
	if err != nil || row == nil {
		return nil, httputil.NotFoundError(notifications.N("APPROVAL_NOT_FOUND", "Pedido de aprovação não encontrado."))
	}
	if row.RequesterUserID == user.ID {
		return nil, httputil.BusinessError(notifications.N("APPROVAL_SELF_FORBIDDEN", "Quem solicita não pode recusar a si mesmo."))
	}
	now := time.Now().UTC()
	r := strings.TrimSpace(reason)
	row.Status = "rejected"
	row.RejectedBy = &user.ID
	row.RejectedAt = &now
	row.RejectionReason = &r
	if err := s.Store.UpdateApprovalRequest(ctx, *row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "RejectApproval", "ApprovalRequest", row.ID, nil, map[string]any{"reason": r, "rejected_by": user.ID})
	return toApprovalResponse(*row), nil
}

func (s *Service) executeApprovedAction(ctx context.Context, row *store.ApprovalRequestRow) error {
	ctx = withApprovalExecution(ctx)
	switch row.ActionType {
	case "reopen_month":
		return s.executeReopenMonth(ctx, row.EntityID)
	case "consolidation":
		return s.doConsolidateMonth(ctx, row.EntityID)
	case "forced_charge", "high_discount":
		var p struct {
			PhoneLineID  string                                       `json:"phone_line_id"`
			ProcessingID string                                       `json:"processing_id"`
			Item         models.CreateLineBillingCompositionItemInput `json:"item"`
		}
		if err := json.Unmarshal([]byte(deref(row.PayloadJSON)), &p); err != nil {
			return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		_, err := s.CreateLineBillingCompositionItem(ctx, p.PhoneLineID, p.ProcessingID, p.Item)
		return err
	case "batch_change":
		var input models.BulkGenerateBillingDocumentsInput
		if err := json.Unmarshal([]byte(deref(row.PayloadJSON)), &input); err != nil {
			return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		_, err := s.BulkGenerateBillingDocuments(ctx, input)
		return err
	case "cycle_impact":
		var p struct {
			ID    string                         `json:"id"`
			Input models.UpdateBillingCycleInput `json:"input"`
		}
		if err := json.Unmarshal([]byte(deref(row.PayloadJSON)), &p); err != nil {
			return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		return s.UpdateBillingCycle(ctx, p.ID, p.Input)
	case "retroactive_cancel":
		_, err := s.CancelSicrediBoleto(ctx, row.EntityID)
		return err
	default:
		after, _ := json.Marshal(map[string]any{"approved_action": row.ActionType, "entity_id": row.EntityID})
		str := string(after)
		row.AfterJSON = &str
		return nil
	}
}

func (s *Service) executeReopenMonth(ctx context.Context, id string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return err
	}
	m, err := s.Store.GetProcessingMonth(ctx, orgID, id)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if m == nil {
		return httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if err := s.SM().ValidateTransition(statemachine.EntityProcessingMonth, m.Status, "open", []string{"master"}); err != nil {
		return err
	}
	before := map[string]any{"status": m.Status}
	if err := s.Store.ReopenProcessingMonth(ctx, orgID, id); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.SM().RecordTransition(ctx, orgID, statemachine.EntityProcessingMonth, id, m.Status, "open", "reopen_month_two_level", nil, &user.ID, nil)
	s.auditLog(ctx, "Reopen", "ProcessingMonth", id, before, map[string]any{"status": "open", "reopened_by": user.ID})
	return nil
}

func (s *Service) ReopenProcessingMonth(ctx context.Context, id string) (*models.GetProcessingMonthResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.Store.GetProcessingMonth(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if m == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if m.Status != "closed" {
		return nil, httputil.BusinessError(notifications.N("PROCESSING_MONTH_NOT_CLOSED", "Somente meses fechados podem ser reabertos."))
	}
	_, err = s.RequestTwoLevelApproval(ctx, models.CreateApprovalRequestInput{
		ActionType: "reopen_month", EntityType: "processing_month", EntityID: id,
		Justification: "Reabertura do mês de processamento",
		Before:        map[string]any{"status": m.Status},
	})
	if err != nil {
		return nil, err
	}
	return s.GetProcessingMonth(ctx, id)
}

func toApprovalResponse(r store.ApprovalRequestRow) *models.ApprovalRequestResponse {
	return &models.ApprovalRequestResponse{
		ID: r.ID, OrganizationID: r.OrganizationID, ActionType: r.ActionType,
		EntityType: r.EntityType, EntityID: r.EntityID, Status: r.Status,
		RequesterUserID: r.RequesterUserID, FirstApproverUserID: r.FirstApproverUserID,
		SecondApproverUserID: r.SecondApproverUserID, Justification: r.Justification,
		CreatedAt: r.CreatedAt, FirstApprovedAt: r.FirstApprovedAt, SecondApprovedAt: r.SecondApprovedAt,
		RejectedAt: r.RejectedAt, RejectionReason: r.RejectionReason, ExecutedAt: r.ExecutedAt,
	}
}

func optStr(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
