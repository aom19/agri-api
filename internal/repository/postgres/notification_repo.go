package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type NotificationRepo struct {
	db *sql.DB
}

func NewNotificationRepo(db *sql.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(n *domain.Notification) error {
	return r.db.QueryRow(`
		INSERT INTO notifications (type, title, message, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		n.Type, n.Title, n.Message, n.EntityType, n.EntityID,
	).Scan(&n.ID, &n.CreatedAt)
}

func (r *NotificationRepo) CreateUserNotification(notificationID, userID int64) error {
	_, err := r.db.Exec(`
		INSERT INTO user_notifications (notification_id, user_id)
		VALUES ($1, $2)`,
		notificationID, userID,
	)
	return err
}

func (r *NotificationRepo) GetByUser(userID int64, onlyUnread bool, limit int) ([]domain.UserNotification, error) {
	query := `
		SELECT un.id, un.notification_id, un.user_id, un.read_at, un.created_at,
		       n.id, n.type, n.title, n.message, n.entity_type, n.entity_id, n.created_at
		FROM user_notifications un
		JOIN notifications n ON n.id = un.notification_id
		WHERE un.user_id = $1`
	if onlyUnread {
		query += " AND un.read_at IS NULL"
	}
	query += " ORDER BY un.created_at DESC LIMIT $2"

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.UserNotification
	for rows.Next() {
		var un domain.UserNotification
		var entityType, entityID sql.NullString
		if err := rows.Scan(
			&un.ID, &un.NotificationID, &un.UserID, &un.ReadAt, &un.CreatedAt,
			&un.Notification.ID, &un.Notification.Type, &un.Notification.Title,
			&un.Notification.Message, &entityType, &entityID, &un.Notification.CreatedAt,
		); err != nil {
			return nil, err
		}
		if entityType.Valid {
			un.Notification.EntityType = entityType.String
		}
		if entityID.Valid {
			un.Notification.EntityID = entityID.String
		}
		result = append(result, un)
	}
	return result, rows.Err()
}

func (r *NotificationRepo) CountUnread(userID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM user_notifications
		WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *NotificationRepo) MarkAsRead(id int64, userID int64) error {
	_, err := r.db.Exec(`
		UPDATE user_notifications SET read_at = NOW()
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
		id, userID,
	)
	return err
}

func (r *NotificationRepo) MarkAllAsRead(userID int64) error {
	_, err := r.db.Exec(`
		UPDATE user_notifications SET read_at = NOW()
		WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	)
	return err
}

func (r *NotificationRepo) GetManagerAndAdminUserIDs() ([]int64, error) {
	rows, err := r.db.Query(`
		SELECT u.id FROM users u
		JOIN roles ro ON ro.id = u.role_id
		WHERE ro.code IN ('admin', 'manager')
		  AND u.deleted_at IS NULL`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
