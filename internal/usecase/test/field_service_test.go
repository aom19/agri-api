package usecase_test

import (
	"encoding/json"
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

const validPolygon = `{"type":"Polygon","coordinates":[[[28.1,46.2],[28.2,46.2],[28.2,46.3],[28.1,46.2]]]}`

func TestFieldService_Validation(t *testing.T) {
	svc := usecase.NewFieldService(&fieldRepoMock{})

	cases := map[string]*domain.Field{
		"nil":                   nil,
		"nume lipsă":            {Geometry: json.RawMessage(validPolygon)},
		"suprafață negativă":    {Name: "A", AreaHa: ptr(-1.0), Geometry: json.RawMessage(validPolygon)},
		"cadastral gol":         {Name: "A", CadastralNumber: ptr("  "), Geometry: json.RawMessage(validPolygon)},
		"geometrie invalidă":    {Name: "A", Geometry: json.RawMessage(`nu-e-json`)},
		"tip greșit":            {Name: "A", Geometry: json.RawMessage(`{"type":"Point","coordinates":[]}`)},
		"prea puține puncte":    {Name: "A", Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[[[1,1],[2,2],[1,1]]]}`)},
		"coordonată incompletă": {Name: "A", Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[[[1],[2,2],[3,3],[1]]]}`)},
		"coordonate în afara":   {Name: "A", Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[[[200,1],[2,2],[3,3],[200,1]]]}`)},
		"inel neînchis":         {Name: "A", Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[[[1,1],[2,2],[3,3],[4,4]]]}`)},
	}
	for name, in := range cases {
		if _, err := svc.CreateField(in); err == nil {
			t.Errorf("%s: mă așteptam la eroare de validare", name)
		}
	}
}

func TestFieldService_CRUD(t *testing.T) {
	existing := &domain.Field{ID: "f1", Name: "Vechi", Geometry: json.RawMessage(validPolygon)}
	repo := &fieldRepoMock{
		create: func(f *domain.Field) error { f.ID = "f2"; return nil },
		getByID: func(id string) (*domain.Field, error) {
			if id == "f1" {
				return existing, nil
			}
			return nil, nil
		},
		getAll: func() ([]domain.Field, error) { return []domain.Field{*existing}, nil },
	}
	svc := usecase.NewFieldService(repo)

	created, err := svc.CreateField(&domain.Field{Name: "Lot", AreaHa: ptr(12.5), Geometry: json.RawMessage(validPolygon)})
	if err != nil || created.ID != "f2" || *created.AreaHa != 12.5 {
		t.Fatalf("CreateField: %v, %+v", err, created)
	}

	if _, err := svc.GetFieldByID(""); err == nil {
		t.Error("id gol trebuie să dea eroare")
	}
	if f, err := svc.GetFieldByID("f1"); err != nil || f.Name != "Vechi" {
		t.Errorf("GetFieldByID: %v", err)
	}
	if all, err := svc.GetFields(); err != nil || len(all) != 1 {
		t.Errorf("GetFields: %v", err)
	}

	input := &domain.Field{Name: "Nou", Geometry: json.RawMessage(validPolygon)}
	if _, err := svc.UpdateField("", input); err == nil {
		t.Error("update cu id gol trebuie să dea eroare")
	}
	if _, err := svc.UpdateField("f1", &domain.Field{}); err == nil {
		t.Error("update cu payload invalid trebuie să dea eroare")
	}
	if _, err := svc.UpdateField("missing", input); err == nil {
		t.Error("update pe teren inexistent trebuie să dea eroare")
	}
	updated, err := svc.UpdateField("f1", input)
	if err != nil || updated.Name != "Nou" {
		t.Fatalf("UpdateField: %v, %+v", err, updated)
	}

	if err := svc.DeleteField(""); err == nil {
		t.Error("delete cu id gol trebuie să dea eroare")
	}
	if err := svc.DeleteField("f1"); err != nil {
		t.Errorf("DeleteField: %v", err)
	}

	boom := errors.New("db down")
	repo.create = func(*domain.Field) error { return boom }
	if _, err := svc.CreateField(input); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	repo.update = func(string, *domain.Field) error { return boom }
	if _, err := svc.UpdateField("f1", input); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	repo.getByID = func(string) (*domain.Field, error) { return nil, boom }
	if _, err := svc.UpdateField("f1", input); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
}
