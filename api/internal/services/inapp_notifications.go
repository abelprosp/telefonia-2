package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

var (
	readNotificationsMu sync.RWMutex
	readNotifications   = make(map[string]bool) // key: userID+notificationID
)

func (s *Service) ListInAppNotifications(ctx context.Context, userID string) (*models.NotificationListResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	var notifications []models.InAppNotification

	// 1. Obter status operacional
	dashboard, _ := s.GetOperationalDashboard(ctx)

	// 2. Verificar Divergências Pendentes
	monthID := ""
	if dashboard != nil && dashboard.CurrentMonthStatus != nil {
		monthID = dashboard.CurrentMonthStatus.ProcessingMonthID
	}
	divergences, err := s.ListDivergences(ctx, monthID)
	if err == nil && len(divergences) > 0 {
		pendingCount := 0
		for _, d := range divergences {
			if d.Status == "open" || d.Status == "PENDING" {
				pendingCount++
			}
		}
		if pendingCount > 0 {
			notifID := fmt.Sprintf("div-%s-%d", orgID, pendingCount)
			notifications = append(notifications, models.InAppNotification{
				ID:          notifID,
				Title:       "Divergências no Faturamento",
				Description: fmt.Sprintf("Existem %d divergência(s) pendente(s) identificadas na conciliação.", pendingCount),
				Category:    "billing",
				Severity:    "warning",
				ActionURL:   "/divergences",
				ActionLabel: "Ver divergências",
				CreatedAt:   time.Now().Add(-15 * time.Minute),
			})
		}
	}

	// 3. Verificar Aprovações Pendentes (SoD)
	approvals, err := s.ListApprovalRequests(ctx, "PENDING")
	if err == nil && len(approvals) > 0 {
		notifID := fmt.Sprintf("appr-%s-%d", orgID, len(approvals))
		notifications = append(notifications, models.InAppNotification{
			ID:          notifID,
			Title:       "Aprovações Administrativas",
			Description: fmt.Sprintf("Você tem %d ação(ões) aguardando aprovação de 2º nível.", len(approvals)),
			Category:    "security",
			Severity:    "info",
			ActionURL:   "/settings",
			ActionLabel: "Revisar aprovações",
			CreatedAt:   time.Now().Add(-1 * time.Hour),
		})
	}

	// 4. Notificação do Ciclo Vigente
	if dashboard != nil && dashboard.CurrentMonthStatus != nil {
		notifID := fmt.Sprintf("month-%s-%s", orgID, dashboard.CurrentMonthStatus.ProcessingMonthID)
		notifications = append(notifications, models.InAppNotification{
			ID:          notifID,
			Title:       fmt.Sprintf("Ciclo %s Aberto", dashboard.CurrentMonthStatus.DisplayName),
			Description: fmt.Sprintf("O mês de processamento %s está com status %s.", dashboard.CurrentMonthStatus.DisplayName, dashboard.CurrentMonthStatus.Status),
			Category:    "billing",
			Severity:    "info",
			ActionURL:   "/invoices",
			ActionLabel: "Acessar faturas",
			CreatedAt:   time.Now().Add(-4 * time.Hour),
		})
	}

	// 5. Notificação de Boas-Vindas e Segurança
	notifID := "sec-mfa-recommendation"
	notifications = append(notifications, models.InAppNotification{
		ID:          notifID,
		Title:       "Segurança da Conta",
		Description: "A autenticação em dois fatores (MFA/OTP) está disponível no Keycloak.",
		Category:    "system",
		Severity:    "success",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
	})

	// Preencher status de leitura
	readNotificationsMu.RLock()
	defer readNotificationsMu.RUnlock()

	unreadCount := 0
	for i := range notifications {
		key := fmt.Sprintf("%s:%s", userID, notifications[i].ID)
		if readNotifications[key] {
			notifications[i].IsRead = true
		} else {
			notifications[i].IsRead = false
			unreadCount++
		}
	}

	return &models.NotificationListResponse{
		Items:       notifications,
		UnreadCount: unreadCount,
		TotalCount:  len(notifications),
	}, nil
}

func (s *Service) MarkNotificationAsRead(ctx context.Context, userID, notifID string) error {
	readNotificationsMu.Lock()
	defer readNotificationsMu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, notifID)
	readNotifications[key] = true
	return nil
}

func (s *Service) MarkAllNotificationsAsRead(ctx context.Context, userID string) error {
	res, err := s.ListInAppNotifications(ctx, userID)
	if err != nil {
		return err
	}

	readNotificationsMu.Lock()
	defer readNotificationsMu.Unlock()

	for _, item := range res.Items {
		key := fmt.Sprintf("%s:%s", userID, item.ID)
		readNotifications[key] = true
	}
	return nil
}
