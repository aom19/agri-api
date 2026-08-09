package repository

import "agri-api/internal/domain"

type NotificationRepository interface {
	Create(notification *domain.Notification) error
	CreateUserNotification(notificationID, userID int64) error
	GetByUser(userID int64, onlyUnread bool, limit int) ([]domain.UserNotification, error)
	CountUnread(userID int64) (int64, error)
	MarkAsRead(id int64, userID int64) error
	MarkAllAsRead(userID int64) error
	GetManagerAndAdminUserIDs() ([]int64, error)
}
