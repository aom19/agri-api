package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

const soil = domain.OperationTypeSoilPreparation

func newOperationService() (*usecase.OperationService, *operationTemplateRepoMock) {
	templates := &operationTemplateRepoMock{
		getByID: func(id int64) (*domain.OperationTemplate, error) {
			if id == 1 {
				return &domain.OperationTemplate{ID: 1, Name: "Arat standard", Unit: "ha", OperationType: soil}, nil
			}
			return nil, nil
		},
		create: func(tpl *domain.OperationTemplate) error { tpl.ID = 1; return nil },
	}
	return usecase.NewOperationService(templates), templates
}

func TestOperationService_Templates(t *testing.T) {
	svc, templates := newOperationService()

	if _, err := svc.GetTemplateByID(9); !errors.Is(err, usecase.ErrOperationTemplateNotFound) {
		t.Errorf("template inexistent: %v", err)
	}
	if tpl, err := svc.GetTemplateByID(1); err != nil || tpl.Name != "Arat standard" {
		t.Errorf("GetTemplateByID: %v", err)
	}
	if _, err := svc.GetAllTemplates(); err != nil {
		t.Errorf("GetAllTemplates: %v", err)
	}

	invalid := map[string]struct {
		in  *domain.OperationTemplate
		err error
	}{
		"nume lipsă":             {&domain.OperationTemplate{Unit: "ha", OperationType: soil}, usecase.ErrTemplateNameRequired},
		"unitate lipsă":          {&domain.OperationTemplate{Name: "x", OperationType: soil}, usecase.ErrTemplateUnitRequired},
		"tip lipsă":              {&domain.OperationTemplate{Name: "x", Unit: "ha"}, usecase.ErrInvalidOperationType},
		"tip în afara enum-ului": {&domain.OperationTemplate{Name: "x", Unit: "ha", OperationType: "plowing"}, usecase.ErrInvalidOperationType},
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

	var savedType domain.OperationType
	templates.create = func(tpl *domain.OperationTemplate) error { tpl.ID = 1; savedType = tpl.OperationType; return nil }
	created, err := svc.CreateTemplate(&domain.OperationTemplate{
		Name: "Arat adânc", Unit: "ha", OperationType: soil,
		Resources:      []domain.TemplateResource{{ResourceID: 1, QuantityPerUnit: 2}},
		MachineTypes:   []string{"tractor"},
		ImplementTypes: []string{"plow"},
	})
	if err != nil || created == nil {
		t.Fatalf("CreateTemplate: %v", err)
	}
	if savedType != soil {
		t.Errorf("tipul salvat: %q", savedType)
	}
	if len(resources) != 1 || len(machineTypes) != 1 || len(implementTypes) != 1 {
		t.Error("datele asociate nu au fost salvate")
	}

	if _, err := svc.UpdateTemplate(9, created); !errors.Is(err, usecase.ErrOperationTemplateNotFound) {
		t.Errorf("update template inexistent: %v", err)
	}
	for name, c := range invalid {
		if _, err := svc.UpdateTemplate(1, c.in); !errors.Is(err, c.err) {
			t.Errorf("update %s: %v", name, err)
		}
	}
	resources, machineTypes, implementTypes = nil, nil, nil
	if _, err := svc.UpdateTemplate(1, &domain.OperationTemplate{
		Name: "Arat", Unit: "ha", OperationType: soil,
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
	if _, err := svc.CreateTemplate(&domain.OperationTemplate{Name: "x", Unit: "ha", OperationType: soil, Resources: []domain.TemplateResource{{}}}); !errors.Is(err, boom) {
		t.Error("eroarea la resurse trebuie propagată")
	}
	templates.create = func(*domain.OperationTemplate) error { return boom }
	if _, err := svc.CreateTemplate(&domain.OperationTemplate{Name: "x", Unit: "ha", OperationType: soil}); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	templates.update = func(int64, *domain.OperationTemplate) error { return boom }
	if _, err := svc.UpdateTemplate(1, &domain.OperationTemplate{Name: "x", Unit: "ha", OperationType: soil}); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
}
