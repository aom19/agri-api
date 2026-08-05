package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"encoding/json"
	"errors"
	"math"
	"strings"
)

type geoJSONPolygon struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// FieldService contine logica de business pentru terenuri.
type FieldService struct {
	fieldRepo repository.FieldRepository
}

func NewFieldService(fieldRepo repository.FieldRepository) *FieldService {
	return &FieldService{fieldRepo: fieldRepo}
}

func (s *FieldService) GetFields() ([]domain.Field, error) {
	return s.fieldRepo.GetAll()
}

func (s *FieldService) GetFieldByID(id string) (*domain.Field, error) {
	if id == "" {
		return nil, errors.New("field id is required")
	}
	return s.fieldRepo.GetByID(id)
}

func (s *FieldService) CreateField(field *domain.Field) (*domain.Field, error) {
	if err := validateFieldInput(field); err != nil {
		return nil, err
	}
	f := &domain.Field{
		Name:            field.Name,
		CadastralNumber: field.CadastralNumber,
		AreaHa:          field.AreaHa,
		Geometry:        field.Geometry,
	}
	if err := s.fieldRepo.Create(f); err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FieldService) UpdateField(id string, field *domain.Field) (*domain.Field, error) {
	if id == "" {
		return nil, errors.New("field id is required")
	}
	if err := validateFieldInput(field); err != nil {
		return nil, err
	}

	existing, err := s.fieldRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("field not found")
	}

	existing.Name = field.Name
	existing.CadastralNumber = field.CadastralNumber
	existing.AreaHa = field.AreaHa
	existing.Geometry = field.Geometry

	if err := s.fieldRepo.Update(id, existing); err != nil {
		return nil, err
	}

	return s.fieldRepo.GetByID(id)
}

func (s *FieldService) DeleteField(id string) error {
	if id == "" {
		return errors.New("field id is required")
	}
	return s.fieldRepo.Delete(id)
}

func validateFieldInput(field *domain.Field) error {
	if field == nil {
		return errors.New("field payload is required")
	}
	if field.Name == "" {
		return errors.New("name is required")
	}
	if field.AreaHa != nil && *field.AreaHa < 0 {
		return errors.New("area_ha must be greater than or equal to 0")
	}
	if field.CadastralNumber != nil && strings.TrimSpace(*field.CadastralNumber) == "" {
		return errors.New("cadastral_number must not be blank")
	}

	var geometry geoJSONPolygon
	if err := json.Unmarshal(field.Geometry, &geometry); err != nil {
		return errors.New("geometry must be a valid GeoJSON polygon")
	}
	if geometry.Type != "Polygon" {
		return errors.New("geometry.type must be 'Polygon'")
	}
	if len(geometry.Coordinates) == 0 || len(geometry.Coordinates[0]) < 4 {
		return errors.New("geometry must contain at least 4 points in outer ring")
	}

	ring := geometry.Coordinates[0]
	for _, point := range ring {
		if len(point) < 2 {
			return errors.New("each geometry coordinate must have [lng, lat]")
		}
		lng := point[0]
		lat := point[1]
		if lng < -180 || lng > 180 || lat < -90 || lat > 90 {
			return errors.New("geometry coordinates are out of bounds")
		}
	}

	first := ring[0]
	last := ring[len(ring)-1]
	if !almostEqual(first[0], last[0]) || !almostEqual(first[1], last[1]) {
		return errors.New("polygon ring must be closed (first point equals last point)")
	}

	return nil
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
