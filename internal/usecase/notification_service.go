package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"sync"
)

// Subscriber receives notifications in real-time (WebSocket prep)
type NotificationSubscriber struct {
	UserID int64
	Ch     chan domain.UserNotification
}

type NotificationService struct {
	repo repository.NotificationRepository

	mu          sync.RWMutex
	subscribers map[int64][]*NotificationSubscriber
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo:        repo,
		subscribers: make(map[int64][]*NotificationSubscriber),
	}
}

// Emit creates a notification and delivers it (asynchronously) to all admin/manager users
func (s *NotificationService) Emit(notifType domain.NotificationType, title, message, entityType, entityID string) {
	go func() {
		_ = s.EmitToAudience(notifType, title, message, entityType, entityID)
	}()
}

// EmitToAudience creates a single notification and delivers it synchronously to all
// admin/manager users plus the given extra users (e.g. the assigned operator), without duplicates.
// It returns an error only if the notification itself could not be created or the audience
// could not be resolved.
func (s *NotificationService) EmitToAudience(notifType domain.NotificationType, title, message, entityType, entityID string, extraUserIDs ...int64) error {
	userIDs, err := s.repo.GetManagerAndAdminUserIDs()
	if err != nil {
		return err
	}
	return s.deliver(notifType, title, message, entityType, entityID, append(userIDs, extraUserIDs...))
}

// EmitToUser sends notification to a specific user
func (s *NotificationService) EmitToUser(userID int64, notifType domain.NotificationType, title, message, entityType, entityID string) {
	go func() {
		_ = s.deliver(notifType, title, message, entityType, entityID, []int64{userID})
	}()
}

// deliver persists one notification and fans it out to the given users (deduplicated),
// pushing it to connected WebSocket subscribers as well.
func (s *NotificationService) deliver(notifType domain.NotificationType, title, message, entityType, entityID string, userIDs []int64) error {
	n := &domain.Notification{
		Type:       notifType,
		Title:      title,
		Message:    message,
		EntityType: entityType,
		EntityID:   entityID,
	}
	if err := s.repo.Create(n); err != nil {
		return err
	}

	seen := make(map[int64]struct{}, len(userIDs))
	for _, uid := range userIDs {
		if uid <= 0 {
			continue
		}
		if _, duplicate := seen[uid]; duplicate {
			continue
		}
		seen[uid] = struct{}{}

		if err := s.repo.CreateUserNotification(n.ID, uid); err != nil {
			continue
		}
		s.pushToSubscribers(uid, domain.UserNotification{
			NotificationID: n.ID,
			UserID:         uid,
			Notification:   *n,
			CreatedAt:      n.CreatedAt,
		})
	}
	return nil
}

func (s *NotificationService) GetByUser(userID int64, onlyUnread bool, limit int) ([]domain.UserNotification, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetByUser(userID, onlyUnread, limit)
}

func (s *NotificationService) CountUnread(userID int64) (int64, error) {
	return s.repo.CountUnread(userID)
}

func (s *NotificationService) MarkAsRead(id int64, userID int64) error {
	return s.repo.MarkAsRead(id, userID)
}

func (s *NotificationService) MarkAllAsRead(userID int64) error {
	return s.repo.MarkAllAsRead(userID)
}

// Subscribe registers a channel for real-time delivery (WebSocket handler will use this)
func (s *NotificationService) Subscribe(userID int64) *NotificationSubscriber {
	sub := &NotificationSubscriber{
		UserID: userID,
		Ch:     make(chan domain.UserNotification, 16),
	}
	s.mu.Lock()
	s.subscribers[userID] = append(s.subscribers[userID], sub)
	s.mu.Unlock()
	return sub
}

// Unsubscribe removes a subscriber
func (s *NotificationService) Unsubscribe(sub *NotificationSubscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := s.subscribers[sub.UserID]
	for i, existing := range subs {
		if existing == sub {
			s.subscribers[sub.UserID] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	close(sub.Ch)
}

func (s *NotificationService) pushToSubscribers(userID int64, un domain.UserNotification) {
	s.mu.RLock()
	subs := s.subscribers[userID]
	s.mu.RUnlock()

	for _, sub := range subs {
		select {
		case sub.Ch <- un:
		default:
			// subscriber buffer full, skip
		}
	}
}
