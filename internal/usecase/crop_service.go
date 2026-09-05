package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

var (
	ErrCropNotFound  = errors.New("înregistrarea nu există")
	ErrCropInvalid   = errors.New("date invalide")
	ErrCropDuplicate = errors.New("există deja o înregistrare cu aceleași date")
)

type CropService struct {
	repo   repository.CropRepository
	fields repository.FieldRepository
}

func NewCropService(repo repository.CropRepository, fields repository.FieldRepository) *CropService {
	return &CropService{repo: repo, fields: fields}
}

func mapCropRepoError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sql.ErrNoRows):
		return ErrCropNotFound
	case errors.Is(err, repository.ErrDuplicateEntry):
		return ErrCropDuplicate
	default:
		return err
	}
}

func parseISODate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.UTC)
}

// ─── Seasons ─────────────────────────────────────────────────────────────────

func (service *CropService) GetSeasons() ([]domain.Season, error) { return service.repo.GetSeasons() }

func (service *CropService) validateSeason(season *domain.Season) error {
	season.Name = strings.TrimSpace(season.Name)
	if season.Name == "" {
		return fmt.Errorf("%w: numele sezonului este obligatoriu", ErrCropInvalid)
	}
	start, err := parseISODate(season.StartDate)
	if err != nil {
		return fmt.Errorf("%w: data de început trebuie să fie YYYY-MM-DD", ErrCropInvalid)
	}
	end, err := parseISODate(season.EndDate)
	if err != nil {
		return fmt.Errorf("%w: data de sfârșit trebuie să fie YYYY-MM-DD", ErrCropInvalid)
	}
	if end.Before(start) {
		return fmt.Errorf("%w: sfârșitul sezonului este înaintea începutului", ErrCropInvalid)
	}
	season.StartDate = start.Format("2006-01-02")
	season.EndDate = end.Format("2006-01-02")
	return nil
}

func (service *CropService) CreateSeason(season *domain.Season) (*domain.Season, error) {
	if err := service.validateSeason(season); err != nil {
		return nil, err
	}
	if err := service.repo.CreateSeason(season); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetSeasonByID(season.ID)
}

func (service *CropService) UpdateSeason(id int64, season *domain.Season) (*domain.Season, error) {
	if err := service.validateSeason(season); err != nil {
		return nil, err
	}
	if err := service.repo.UpdateSeason(id, season); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetSeasonByID(id)
}

func (service *CropService) DeleteSeason(id int64) error {
	return mapCropRepoError(service.repo.DeleteSeason(id))
}

// ─── Crops ───────────────────────────────────────────────────────────────────

func (service *CropService) GetCrops() ([]domain.Crop, error) { return service.repo.GetCrops() }

func validateCrop(crop *domain.Crop) error {
	crop.Name = strings.TrimSpace(crop.Name)
	if crop.Name == "" {
		return fmt.Errorf("%w: numele culturii este obligatoriu", ErrCropInvalid)
	}
	if crop.YieldUnit = strings.TrimSpace(crop.YieldUnit); crop.YieldUnit == "" {
		crop.YieldUnit = "t"
	}
	if crop.Code != nil {
		code := strings.TrimSpace(*crop.Code)
		if code == "" {
			crop.Code = nil
		} else {
			crop.Code = &code
		}
	}
	return nil
}

func (service *CropService) CreateCrop(crop *domain.Crop) (*domain.Crop, error) {
	if err := validateCrop(crop); err != nil {
		return nil, err
	}
	if err := service.repo.CreateCrop(crop); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetCropByID(crop.ID)
}

func (service *CropService) UpdateCrop(id int64, crop *domain.Crop) (*domain.Crop, error) {
	if err := validateCrop(crop); err != nil {
		return nil, err
	}
	if err := service.repo.UpdateCrop(id, crop); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetCropByID(id)
}

func (service *CropService) DeleteCrop(id int64) error {
	return mapCropRepoError(service.repo.DeleteCrop(id))
}

// ─── Field crops ─────────────────────────────────────────────────────────────

func (service *CropService) ListFieldCrops(filter domain.FieldCropFilter) ([]domain.FieldCrop, error) {
	return service.repo.ListFieldCrops(filter)
}

func (service *CropService) validateFieldCrop(item *domain.FieldCrop) error {
	if strings.TrimSpace(item.FieldID) == "" || item.SeasonID <= 0 || item.CropID <= 0 {
		return fmt.Errorf("%w: terenul, sezonul și cultura sunt obligatorii", ErrCropInvalid)
	}
	field, err := service.fields.GetByID(item.FieldID)
	if err != nil {
		return err
	}
	if field == nil {
		return fmt.Errorf("%w: terenul nu există", ErrCropInvalid)
	}
	if season, err := service.repo.GetSeasonByID(item.SeasonID); err != nil {
		return err
	} else if season == nil {
		return fmt.Errorf("%w: sezonul nu există", ErrCropInvalid)
	}
	if crop, err := service.repo.GetCropByID(item.CropID); err != nil {
		return err
	} else if crop == nil {
		return fmt.Errorf("%w: cultura nu există", ErrCropInvalid)
	}
	if item.PlantedAreaHa != nil {
		if *item.PlantedAreaHa <= 0 {
			return fmt.Errorf("%w: suprafața cultivată trebuie să fie pozitivă", ErrCropInvalid)
		}
		if field.AreaHa != nil && *item.PlantedAreaHa > *field.AreaHa*1.0001 {
			return fmt.Errorf("%w: suprafața cultivată (%.2f ha) depășește suprafața terenului (%.2f ha)", ErrCropInvalid, *item.PlantedAreaHa, *field.AreaHa)
		}
	} else if field.AreaHa != nil {
		area := *field.AreaHa
		item.PlantedAreaHa = &area
	}
	if item.ProductionTotal != nil && *item.ProductionTotal < 0 {
		return fmt.Errorf("%w: producția nu poate fi negativă", ErrCropInvalid)
	}
	if item.ExpectedYieldPerHa != nil && *item.ExpectedYieldPerHa < 0 {
		return fmt.Errorf("%w: randamentul estimat nu poate fi negativ", ErrCropInvalid)
	}
	for _, value := range []*string{item.PlantedAt, item.HarvestedAt} {
		if value == nil || strings.TrimSpace(*value) == "" {
			continue
		}
		if _, err := parseISODate(*value); err != nil {
			return fmt.Errorf("%w: datele trebuie să fie YYYY-MM-DD", ErrCropInvalid)
		}
	}
	return nil
}

func (service *CropService) CreateFieldCrop(item *domain.FieldCrop) (*domain.FieldCrop, error) {
	if err := service.validateFieldCrop(item); err != nil {
		return nil, err
	}
	if err := service.repo.CreateFieldCrop(item); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetFieldCropByID(item.ID)
}

func (service *CropService) UpdateFieldCrop(id int64, item *domain.FieldCrop) (*domain.FieldCrop, error) {
	if err := service.validateFieldCrop(item); err != nil {
		return nil, err
	}
	if err := service.repo.UpdateFieldCrop(id, item); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetFieldCropByID(id)
}

func (service *CropService) DeleteFieldCrop(id int64) error {
	return mapCropRepoError(service.repo.DeleteFieldCrop(id))
}
