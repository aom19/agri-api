package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

func TestImplementService(t *testing.T) {
	repo := &implementRepoMock{
		getAll:  func() ([]domain.Implement, error) { return []domain.Implement{{ID: 1}}, nil },
		getByID: func(id int64) (*domain.Implement, error) { return &domain.Implement{ID: id}, nil },
	}
	svc := usecase.NewImplementService(repo)

	valid := &domain.Implement{Name: "Plug", Code: "PL-1", Type: domain.ImplementTypePlow, Status: domain.AssetStatusActive}
	if _, err := svc.CreateImplement(&domain.Implement{Name: "fara cod"}); err == nil {
		t.Error("câmpurile obligatorii lipsă trebuie să dea eroare")
	}
	if created, err := svc.CreateImplement(valid); err != nil || created != valid {
		t.Errorf("CreateImplement: %v", err)
	}

	if _, err := svc.UpdateImplement(1, &domain.Implement{}); err == nil {
		t.Error("actualizarea fără câmpuri obligatorii trebuie să dea eroare")
	}
	if updated, err := svc.UpdateImplement(1, valid); err != nil || updated != valid {
		t.Errorf("UpdateImplement: %v", err)
	}

	if all, err := svc.GetImplements(); err != nil || len(all) != 1 {
		t.Errorf("GetImplements: %v", err)
	}
	if one, err := svc.GetImplementByID(4); err != nil || one.ID != 4 {
		t.Errorf("GetImplementByID: %v", err)
	}
	if err := svc.DeleteImplement(1); err != nil {
		t.Errorf("DeleteImplement: %v", err)
	}
	if err := svc.ActivateImplement(1); err != nil {
		t.Errorf("ActivateImplement: %v", err)
	}
	if err := svc.DeactivateImplement(1); err != nil {
		t.Errorf("DeactivateImplement: %v", err)
	}

	boom := errors.New("db down")
	repo.create = func(*domain.Implement) error { return boom }
	repo.update = func(int64, *domain.Implement) error { return boom }
	repo.getByID = func(int64) (*domain.Implement, error) { return nil, boom }
	if _, err := svc.CreateImplement(valid); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	if _, err := svc.UpdateImplement(1, valid); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	if _, err := svc.GetImplementByID(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
}
