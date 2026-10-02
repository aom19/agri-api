package usecase

import (
	"errors"
	"testing"
)

func TestAuditService(t *testing.T) {
	repo := &auditRepoMock{}
	svc := NewAuditService(repo)

	svc.Log("machine", "1", "create", ptr(int64(2)), map[string]interface{}{"name": "T1"})
	waitFor(t, func() bool { return repo.count() == 1 })
	if repo.entries[0].EntityType != "machine" || *repo.entries[0].ActorID != 2 {
		t.Errorf("intrare de audit greșită: %+v", repo.entries[0])
	}

	if err := svc.LogSync("field", "f1", "delete", nil, nil); err != nil || repo.count() != 2 {
		t.Errorf("LogSync: %v", err)
	}
	repo.logErr = errors.New("db down")
	if err := svc.LogSync("field", "f1", "delete", nil, nil); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
}
