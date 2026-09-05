package domain

import "time"

type NotificationType string

const (
	NotifAssetUnavailable NotificationType = "asset_unavailable"
	NotifOperationStarted NotificationType = "operation_started"
	NotifStockLow         NotificationType = "stock_low"
	NotifResourceIssue    NotificationType = "resource_issue"
	NotifOperationOverdue NotificationType = "operation_overdue"
)

type Notification struct {
	ID         int64            `json:"id"`
	Type       NotificationType `json:"type"`
	Title      string           `json:"title"`
	Message    string           `json:"message"`
	EntityType string           `json:"entity_type,omitempty"`
	EntityID   string           `json:"entity_id,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
}

type UserNotification struct {
	ID             int64        `json:"id"`
	NotificationID int64        `json:"notification_id"`
	UserID         int64        `json:"user_id"`
	ReadAt         *time.Time   `json:"read_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	Notification   Notification `json:"notification"`
}
