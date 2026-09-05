package postgres

import (
	"database/sql"
	"time"

	"agri-api/internal/domain"
)

type ReportSubscriptionRepo struct {
	db *sql.DB
}

func NewReportSubscriptionRepo(db *sql.DB) *ReportSubscriptionRepo {
	return &ReportSubscriptionRepo{db: db}
}

const reportSubscriptionSelect = `
	SELECT id, user_id, frequency, send_hour, weekday, is_active, last_sent_at, created_at, updated_at
	FROM report_subscriptions`

func scanReportSubscription(row interface {
	Scan(dest ...interface{}) error
}) (*domain.ReportSubscription, error) {
	var (
		item     domain.ReportSubscription
		lastSent sql.NullTime
	)
	if err := row.Scan(&item.ID, &item.UserID, &item.Frequency, &item.SendHour, &item.Weekday, &item.IsActive, &lastSent, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	if lastSent.Valid {
		value := lastSent.Time
		item.LastSentAt = &value
	}
	return &item, nil
}

func (repo *ReportSubscriptionRepo) GetByUser(userID int64) (*domain.ReportSubscription, error) {
	item, err := scanReportSubscription(repo.db.QueryRow(reportSubscriptionSelect+" WHERE user_id = $1", userID))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

func (repo *ReportSubscriptionRepo) Upsert(subscription *domain.ReportSubscription) error {
	return repo.db.QueryRow(`
		INSERT INTO report_subscriptions (user_id, frequency, send_hour, weekday, is_active)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET frequency = EXCLUDED.frequency,
		    send_hour = EXCLUDED.send_hour,
		    weekday = EXCLUDED.weekday,
		    is_active = EXCLUDED.is_active,
		    updated_at = NOW()
		RETURNING id, last_sent_at, created_at, updated_at`,
		subscription.UserID, subscription.Frequency, subscription.SendHour, subscription.Weekday, subscription.IsActive,
	).Scan(&subscription.ID, &subscription.LastSentAt, &subscription.CreatedAt, &subscription.UpdatedAt)
}

func (repo *ReportSubscriptionRepo) Deactivate(userID int64) error {
	_, err := repo.db.Exec(`UPDATE report_subscriptions SET is_active = FALSE, updated_at = NOW() WHERE user_id = $1`, userID)
	return err
}

func (repo *ReportSubscriptionRepo) ListActiveRecipients() ([]domain.ReportSubscriptionRecipient, error) {
	rows, err := repo.db.Query(`
		SELECT rs.id, rs.user_id, rs.frequency, rs.send_hour, rs.weekday, rs.is_active, rs.last_sent_at, rs.created_at, rs.updated_at,
		       u.email, COALESCE(up.first_name, '')
		FROM report_subscriptions rs
		JOIN users u ON u.id = rs.user_id AND u.deleted_at IS NULL
		LEFT JOIN user_profiles up ON up.user_id = u.id
		WHERE rs.is_active
		ORDER BY rs.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportSubscriptionRecipient{}
	for rows.Next() {
		var (
			item     domain.ReportSubscriptionRecipient
			lastSent sql.NullTime
		)
		if err := rows.Scan(&item.ID, &item.UserID, &item.Frequency, &item.SendHour, &item.Weekday, &item.IsActive, &lastSent, &item.CreatedAt, &item.UpdatedAt, &item.Email, &item.FirstName); err != nil {
			return nil, err
		}
		if lastSent.Valid {
			value := lastSent.Time
			item.LastSentAt = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *ReportSubscriptionRepo) MarkSent(id int64, at time.Time) error {
	_, err := repo.db.Exec(`UPDATE report_subscriptions SET last_sent_at = $1, updated_at = NOW() WHERE id = $2`, at, id)
	return err
}
