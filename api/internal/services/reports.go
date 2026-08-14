package services

import (
	"context"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (s *Service) GetMovementReports(ctx context.Context, processingMonthID string) (*models.MovementReportsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	monthID := strings.TrimSpace(processingMonthID)
	if monthID == "" {
		return nil, httputil.ValidationError(notifications.ProcessingMonthNotFound)
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	resp, err := s.Store.ListMovementReports(ctx, orgID, month)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return resp, nil
}
