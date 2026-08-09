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

// Emit creates a notification and delivers it to all admin/manager users
func (s *NotificationService) Emit(notifType domain.NotificationType, title, message, entityType, entityID string) {
	go s.emitAsync(notifType, title, message, entityType, entityID)
}

func (s *NotificationService) emitAsync(notifType domain.NotificationType, title, message, entityType, entityID string) {
	n := &domain.Notification{
		Type:       notifType,
		Title:      title,
		Message:    message,
		EntityType: entityType,
		EntityID:   entityID,
	}

	if err := s.repo.Create(n); err != nil {
		return
	}

	userIDs, err := s.repo.GetManagerAndAdminUserIDs()
	if err != nil {
		return
	}

	for _, uid := range userIDs {
		if err := s.repo.CreateUserNotification(n.ID, uid); err != nil {
			continue
		}
		// push to connected WebSocket subscribers
		s.pushToSubscribers(uid, domain.UserNotification{
			NotificationID: n.ID,
			UserID:         uid,
			Notification:   *n,
			CreatedAt:      n.CreatedAt,
		})
	}
}

// EmitToUser sends notification to a specific user
func (s *NotificationService) EmitToUser(userID int64, notifType domain.NotificationType, title, message, entityType, entityID string) {
	go func() {
		n := &domain.Notification{
			Type:       notifType,
			Title:      title,
			Message:    message,
			EntityType: entityType,
			EntityID:   entityID,
		}
		if err := s.repo.Create(n); err != nil {
			return
		}
		if err := s.repo.CreateUserNotification(n.ID, userID); err != nil {
			return
		}
		s.pushToSubscribers(userID, domain.UserNotification{
			NotificationID: n.ID,
			UserID:         userID,
			Notification:   *n,
			CreatedAt:      n.CreatedAt,
		})
	}()
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
