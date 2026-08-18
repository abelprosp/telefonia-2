package models

import "time"

type InAppNotification struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"` // "billing", "operations", "security", "system"
	Severity    string    `json:"severity"` // "info", "warning", "success", "critical"
	ActionURL   string    `json:"action_url,omitempty"`
	ActionLabel string    `json:"action_label,omitempty"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

type NotificationListResponse struct {
	Items       []InAppNotification `json:"items"`
	UnreadCount int                 `json:"unread_count"`
	TotalCount  int                 `json:"total_count"`
}
