package usecase

import (
	"errors"
	"testing"
	"time"

	"agri-api/internal/domain"
)

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condiția nu s-a îndeplinit la timp")
}

func TestNotificationService_EmitToAudience(t *testing.T) {
	repo := &notificationRepoMock{adminIDs: []int64{1, 2}}
	svc := NewNotificationService(repo)

	if err := svc.EmitToAudience(domain.NotifStockLow, "Stoc", "mesaj", "stock", "1", 2, 3, 0); err != nil {
		t.Fatalf("EmitToAudience: %v", err)
	}
	if repo.createdCount() != 1 {
		t.Errorf("trebuie creată o singură notificare, am %d", repo.createdCount())
	}
	for _, uid := range []int64{1, 2, 3} {
		if repo.deliveries(uid) != 1 {
			t.Errorf("utilizatorul %d trebuie să primească exact o notificare, are %d", uid, repo.deliveries(uid))
		}
	}
	if repo.deliveries(0) != 0 {
		t.Error("id-urile invalide trebuie ignorate")
	}

	repo.userNotifErr = func(userID int64) error {
		if userID == 2 {
			return errors.New("db down")
		}
		return nil
	}
	if err := svc.EmitToAudience(domain.NotifStockLow, "Stoc", "mesaj", "stock", "1"); err != nil {
		t.Errorf("eroarea per utilizator nu trebuie să oprească livrarea: %v", err)
	}
	if repo.deliveries(1) != 2 || repo.deliveries(2) != 1 {
		t.Error("livrarea trebuie să continue pentru ceilalți utilizatori")
	}

	repo.adminErr = errors.New("db down")
	if err := svc.EmitToAudience(domain.NotifStockLow, "Stoc", "mesaj", "stock", "1"); err == nil {
		t.Error("eroarea la determinarea audienței trebuie propagată")
	}
	repo.adminErr = nil
	repo.createErr = errors.New("db down")
	if err := svc.EmitToAudience(domain.NotifStockLow, "Stoc", "mesaj", "stock", "1"); err == nil {
		t.Error("eroarea la creare trebuie propagată")
	}
}

func TestNotificationService_AsyncEmit(t *testing.T) {
	repo := &notificationRepoMock{adminIDs: []int64{1}}
	svc := NewNotificationService(repo)

	svc.Emit(domain.NotifResourceIssue, "t", "m", "e", "1")
	waitFor(t, func() bool { return repo.deliveries(1) == 1 })

	svc.EmitToUser(5, domain.NotifOperationStarted, "t", "m", "e", "1")
	waitFor(t, func() bool { return repo.deliveries(5) == 1 })
}

func TestNotificationService_Subscribers(t *testing.T) {
	repo := &notificationRepoMock{adminIDs: []int64{1}}
	svc := NewNotificationService(repo)

	sub := svc.Subscribe(1)
	other := svc.Subscribe(2)
	if err := svc.EmitToAudience(domain.NotifOperationCompleted, "Gata", "m", "field_operation", "1"); err != nil {
		t.Fatal(err)
	}
	select {
	case n := <-sub.Ch:
		if n.Notification.Title != "Gata" || n.UserID != 1 {
			t.Errorf("notificare greșită: %+v", n)
		}
	case <-time.After(time.Second):
		t.Fatal("abonatul nu a primit notificarea")
	}
	select {
	case <-other.Ch:
		t.Error("utilizatorul 2 nu face parte din audiență")
	default:
	}

	// buffer plin: livrarea nu trebuie să blocheze
	for i := 0; i < 20; i++ {
		svc.pushToSubscribers(1, domain.UserNotification{})
	}

	svc.Unsubscribe(sub)
	if _, open := <-sub.Ch; open {
		// canalul are încă mesaje din buffer; golește-l până se închide
		for range sub.Ch {
		}
	}
	svc.Unsubscribe(other)
	if len(svc.subscribers[1]) != 0 || len(svc.subscribers[2]) != 0 {
		t.Error("abonații trebuie eliminați la dezabonare")
	}
	// dezabonare repetată nu trebuie să dea panic
	svc.Unsubscribe(&NotificationSubscriber{UserID: 1, Ch: make(chan domain.UserNotification)})
}

func TestNotificationService_Queries(t *testing.T) {
	var gotLimit int
	repo := &notificationRepoMock{
		getByUser:   func(_ int64, _ bool, limit int) ([]domain.UserNotification, error) { gotLimit = limit; return nil, nil },
		countUnread: func(int64) (int64, error) { return 3, nil },
	}
	svc := NewNotificationService(repo)

	if _, err := svc.GetByUser(1, true, 0); err != nil || gotLimit != 50 {
		t.Errorf("limita implicită: %v, %d", err, gotLimit)
	}
	if _, err := svc.GetByUser(1, false, 10); err != nil || gotLimit != 10 {
		t.Errorf("limita explicită: %v, %d", err, gotLimit)
	}
	if n, err := svc.CountUnread(1); err != nil || n != 3 {
		t.Errorf("CountUnread: %v, %d", err, n)
	}
	if err := svc.MarkAsRead(1, 1); err != nil {
		t.Errorf("MarkAsRead: %v", err)
	}
	if err := svc.MarkAllAsRead(1); err != nil {
		t.Errorf("MarkAllAsRead: %v", err)
	}
}
