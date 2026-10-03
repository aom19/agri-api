package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

func newOperationService() (*usecase.OperationService, *operationTypeRepoMock, *operationTemplateRepoMock) {
	types := &operationTypeRepoMock{
		getByID: func(id int64) (*domain.OperationType, error) {
			if id == 1 {
				return &domain.OperationType{ID: 1, Code: "arat", Name: "Arat"}, nil
			}
			return nil, nil
		},
		create: func(ot *domain.OperationType) error { ot.ID = 2; return nil },
	}
	templates := &operationTemplateRepoMock{
		getByID: func(id int64) (*domain.OperationTemplate, error) {
			if id == 1 {
				return &domain.OperationTemplate{ID: 1, Name: "Arat standard", Unit: "ha", OperationTypeID: 1}, nil
			}
			return nil, nil
		},
		create: func(tpl *domain.OperationTemplate) error { tpl.ID = 1; return nil },
	}
	return usecase.NewOperationService(types, templates), types, templates
}

func TestOperationService_Types(t *testing.T) {
	svc, types, _ := newOperationService()

	if _, err := svc.GetTypeByID(9); !errors.Is(err, usecase.ErrOperationTypeNotFound) {
		t.Errorf("tip inexistent: %v", err)
	}
	if ot, err := svc.GetTypeByID(1); err != nil || ot.Code != "arat" {
		t.Errorf("GetTypeByID: %v", err)
	}
	if _, err := svc.GetAllTypes(); err != nil {
		t.Errorf("GetAllTypes: %v", err)
	}

	if _, err := svc.CreateType(&domain.OperationType{Name: "x"}); !errors.Is(err, usecase.ErrOperationTypeCodeRequired) {
		t.Errorf("cod lipsă: %v", err)
	}
	if _, err := svc.CreateType(&domain.OperationType{Code: "x"}); !errors.Is(err, usecase.ErrOperationTypeNameRequired) {
		t.Errorf("nume lipsă: %v", err)
	}
	created, err := svc.CreateType(&domain.OperationType{Code: "semanat", Name: "Semănat"})
	if err != nil || created.ID != 2 {
		t.Fatalf("CreateType: %v, %+v", err, created)
	}

	if _, err := svc.UpdateType(9, created); !errors.Is(err, usecase.ErrOperationTypeNotFound) {
		t.Errorf("update tip inexistent: %v", err)
	}
	if _, err := svc.UpdateType(1, &domain.OperationType{}); !errors.Is(err, usecase.ErrOperationTypeCodeRequired) {
		t.Errorf("update fără cod: %v", err)
	}
	if _, err := svc.UpdateType(1, &domain.OperationType{Code: "x"}); !errors.Is(err, usecase.ErrOperationTypeNameRequired) {
		t.Errorf("update fără nume: %v", err)
	}
	updated, err := svc.UpdateType(1, &domain.OperationType{Code: "arat2", Name: "Arat 2"})
	if err != nil || updated.ID != 1 || updated.Code != "arat2" {
		t.Fatalf("UpdateType: %v, %+v", err, updated)
	}

	if err := svc.DeleteType(9); !errors.Is(err, usecase.ErrOperationTypeNotFound) {
		t.Errorf("delete tip inexistent: %v", err)
	}
	if err := svc.DeleteType(1); err != nil {
		t.Errorf("DeleteType: %v", err)
	}

	boom := errors.New("db down")
	types.create = func(*domain.OperationType) error { return boom }
	if _, err := svc.CreateType(&domain.OperationType{Code: "a", Name: "b"}); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	types.update = func(int64, *domain.OperationType) error { return boom }
	if _, err := svc.UpdateType(1, &domain.OperationType{Code: "a", Name: "b"}); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	types.getByID = func(int64) (*domain.OperationType, error) { return nil, boom }
	if _, err := svc.GetTypeByID(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteType(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
}

func TestOperationService_Templates(t *testing.T) {
	svc, _, templates := newOperationService()

	if _, err := svc.GetTemplateByID(9); !errors.Is(err, usecase.ErrOperationTemplateNotFound) {
		t.Errorf("template inexistent: %v", err)
	}
	if tpl, err := svc.GetTemplateByID(1); err != nil || tpl.Name != "Arat standard" {
		t.Errorf("GetTemplateByID: %v", err)
	}
	if _, err := svc.GetAllTemplates(); err != nil {
		t.Errorf("GetAllTemplates: %v", err)
	}
	if _, err := svc.GetTemplatesByOperationType(1); err != nil {
		t.Errorf("GetTemplatesByOperationType: %v", err)
	}

	invalid := map[string]struct {
		in  *domain.OperationTemplate
		err error
	}{
		"nume lipsă":     {&domain.OperationTemplate{Unit: "ha", OperationTypeID: 1}, usecase.ErrTemplateNameRequired},
		"unitate lipsă":  {&domain.OperationTemplate{Name: "x", OperationTypeID: 1}, usecase.ErrTemplateUnitRequired},
		"tip invalid":    {&domain.OperationTemplate{Name: "x", Unit: "ha"}, usecase.ErrInvalidOperationTypeID},
		"tip inexistent": {&domain.OperationTemplate{Name: "x", Unit: "ha", OperationTypeID: 9}, usecase.ErrOperationTypeNotFound},
	}
	for name, c := range invalid {
		if _, err := svc.CreateTemplate(c.in); !errors.Is(err, c.err) {
			t.Errorf("create %s: %v", name, err)
		}
	}

	var resources []domain.TemplateResource
	var machineTypes, implementTypes []string
	templates.setResources = func(_ int64, r []domain.TemplateResource) error { resources = r; return nil }
	templates.setMachineTypes = func(_ int64, m []string) error { machineTypes = m; return nil }
	templates.setImplementTypes = func(_ int64, i []string) error { implementTypes = i; return nil }

	created, err := svc.CreateTemplate(&domain.OperationTemplate{
		Name: "Arat adânc", Unit: "ha", OperationTypeID: 1,
		Resources:      []domain.TemplateResource{{ResourceID: 1, QuantityPerUnit: 2}},
		MachineTypes:   []string{"tractor"},
		ImplementTypes: []string{"plow"},
	})
	if err != nil || created == nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if len(resources) != 1 || len(machineTypes) != 1 || len(implementTypes) != 1 {
		t.Error("datele asociate nu au fost salvate")
	}

	if _, err := svc.UpdateTemplate(9, created); !errors.Is(err, usecase.ErrOperationTemplateNotFound) {
		t.Errorf("update template inexistent: %v", err)
	}
	for name, c := range invalid {
		if name == "tip inexistent" {
			continue // update nu verifică existența tipului
		}
		if _, err := svc.UpdateTemplate(1, c.in); !errors.Is(err, c.err) {
			t.Errorf("update %s: %v", name, err)
		}
	}
	resources, machineTypes, implementTypes = nil, nil, nil
	if _, err := svc.UpdateTemplate(1, &domain.OperationTemplate{
		Name: "Arat", Unit: "ha", OperationTypeID: 1,
		Resources: []domain.TemplateResource{}, MachineTypes: []string{"combine"}, ImplementTypes: []string{},
	}); err != nil {
		t.Fatalf("UpdateTemplate: %v", err)
	}
	if resources == nil || len(machineTypes) != 1 || implementTypes == nil {
		t.Error("listele goale (dar nenule) trebuie să golească datele asociate")
	}

	if err := svc.DeleteTemplate(9); !errors.Is(err, usecase.ErrOperationTemplateNotFound) {
		t.Errorf("delete template inexistent: %v", err)
	}
	if err := svc.DeleteTemplate(1); err != nil {
		t.Errorf("DeleteTemplate: %v", err)
	}

	boom := errors.New("db down")
	templates.setResources = func(int64, []domain.TemplateResource) error { return boom }
	if _, err := svc.CreateTemplate(&domain.OperationTemplate{Name: "x", Unit: "ha", OperationTypeID: 1, Resources: []domain.TemplateResource{{}}}); !errors.Is(err, boom) {
		t.Error("eroarea la resurse trebuie propagată")
	}
	templates.create = func(*domain.OperationTemplate) error { return boom }
	if _, err := svc.CreateTemplate(&domain.OperationTemplate{Name: "x", Unit: "ha", OperationTypeID: 1}); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	templates.update = func(int64, *domain.OperationTemplate) error { return boom }
	if _, err := svc.UpdateTemplate(1, &domain.OperationTemplate{Name: "x", Unit: "ha", OperationTypeID: 1}); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
}
