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
	// ErrFieldHasOpenCrop: terenul are deja o cultură deschisă; închide-o înainte să deschizi alta.
	ErrFieldHasOpenCrop = fmt.Errorf("%w: terenul are deja o cultură deschisă", ErrCropDuplicate)
)

type CropService struct {
	repo      repository.CropRepository
	fields    repository.FieldRepository
	db        *sql.DB
	movements repository.StockMovementRepository
}

func NewCropService(
	repo repository.CropRepository,
	fields repository.FieldRepository,
	db *sql.DB,
	movements repository.StockMovementRepository,
) *CropService {
	return &CropService{repo: repo, fields: fields, db: db, movements: movements}
}

var ErrHarvestNotRecordable = errors.New("recolta nu poate fi înregistrată")

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
	season.StartDate = start.Format("2006-01-02")
	// sfârșitul e opțional: un sezon în curs nu are încă dată de încheiere
	if season.EndDate == nil || strings.TrimSpace(*season.EndDate) == "" {
		season.EndDate = nil
		return nil
	}
	end, err := parseISODate(*season.EndDate)
	if err != nil {
		return fmt.Errorf("%w: data de sfârșit trebuie să fie YYYY-MM-DD", ErrCropInvalid)
	}
	if end.Before(start) {
		return fmt.Errorf("%w: sfârșitul sezonului este înaintea începutului", ErrCropInvalid)
	}
	formatted := end.Format("2006-01-02")
	season.EndDate = &formatted
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
	// Sezonul e doar o etichetă de raportare: modificarea lui nu atinge nicio lucrare.
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
	if item.Status == "" {
		item.Status = domain.FieldCropStatusOpen
	}
	if !isValidFieldCropStatus(item.Status) {
		return fmt.Errorf("%w: starea trebuie să fie open, closed sau compromised", ErrCropInvalid)
	}
	if item.Status != domain.FieldCropStatusOpen && (item.ClosedAt == nil || strings.TrimSpace(*item.ClosedAt) == "") {
		return fmt.Errorf("%w: o cultură închisă are nevoie de data închiderii", ErrCropInvalid)
	}
	if item.Status == domain.FieldCropStatusOpen {
		item.ClosedAt = nil
	}
	if item.ProductionTotal != nil && *item.ProductionTotal < 0 {
		return fmt.Errorf("%w: producția nu poate fi negativă", ErrCropInvalid)
	}
	if item.ExpectedYieldPerHa != nil && *item.ExpectedYieldPerHa < 0 {
		return fmt.Errorf("%w: randamentul estimat nu poate fi negativ", ErrCropInvalid)
	}
	for _, value := range []*string{item.PlantedAt, item.HarvestedAt, item.ClosedAt} {
		if value == nil || strings.TrimSpace(*value) == "" {
			continue
		}
		if _, err := parseISODate(*value); err != nil {
			return fmt.Errorf("%w: datele trebuie să fie YYYY-MM-DD", ErrCropInvalid)
		}
	}
	return nil
}

func isValidFieldCropStatus(status string) bool {
	switch status {
	case domain.FieldCropStatusOpen, domain.FieldCropStatusClosed, domain.FieldCropStatusCompromised:
		return true
	}
	return false
}

// ensureNoOtherOpenCrop aplică regula „pe un teren e cel mult o cultură deschisă”, cu un mesaj
// care spune ce cultură trebuie închisă. Indexul unic parțial rămâne plasa de siguranță.
func (service *CropService) ensureNoOtherOpenCrop(fieldID string, exceptID int64) error {
	open, err := service.repo.GetOpenFieldCrop(fieldID)
	if err != nil {
		return err
	}
	if open == nil || open.ID == exceptID {
		return nil
	}
	return fmt.Errorf("%w: %s (%s). Închide-o întâi, apoi deschide cultura nouă",
		ErrFieldHasOpenCrop, open.CropName, open.SeasonName)
}

func (service *CropService) CreateFieldCrop(item *domain.FieldCrop) (*domain.FieldCrop, error) {
	if err := service.validateFieldCrop(item); err != nil {
		return nil, err
	}
	if item.Status == domain.FieldCropStatusOpen {
		if err := service.ensureNoOtherOpenCrop(item.FieldID, 0); err != nil {
			return nil, err
		}
	}
	if err := service.repo.CreateFieldCrop(item); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetFieldCropByID(item.ID)
}

// UpdateFieldCrop modifică datele culturii; starea se schimbă doar prin CloseFieldCrop / ReopenFieldCrop.
func (service *CropService) UpdateFieldCrop(id int64, item *domain.FieldCrop) (*domain.FieldCrop, error) {
	existing, err := service.repo.GetFieldCropByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCropNotFound
	}
	item.Status, item.ClosedAt = existing.Status, existing.ClosedAt
	if err := service.validateFieldCrop(item); err != nil {
		return nil, err
	}
	if item.Status == domain.FieldCropStatusOpen && item.FieldID != existing.FieldID {
		if err := service.ensureNoOtherOpenCrop(item.FieldID, id); err != nil {
			return nil, err
		}
	}
	if err := service.repo.UpdateFieldCrop(id, item); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetFieldCropByID(id)
}

// CloseFieldCrop închide cultura (recoltată sau compromisă). Lucrările deja salvate rămân legate
// de ea; doar lucrările noi de pe teren nu mai merg la această cultură.
func (service *CropService) CloseFieldCrop(id int64, input domain.FieldCropClose) (*domain.FieldCrop, error) {
	existing, err := service.repo.GetFieldCropByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCropNotFound
	}
	if input.Status != domain.FieldCropStatusClosed && input.Status != domain.FieldCropStatusCompromised {
		return nil, fmt.Errorf("%w: starea finală trebuie să fie closed sau compromised", ErrCropInvalid)
	}
	closedAt, err := parseISODate(input.ClosedAt)
	if err != nil {
		return nil, fmt.Errorf("%w: data închiderii trebuie să fie YYYY-MM-DD", ErrCropInvalid)
	}
	input.ClosedAt = closedAt.Format("2006-01-02")
	if input.ProductionTotal != nil && *input.ProductionTotal < 0 {
		return nil, fmt.Errorf("%w: producția nu poate fi negativă", ErrCropInvalid)
	}
	if input.Status == domain.FieldCropStatusCompromised {
		zero := 0.0
		input.ProductionTotal = &zero
		if input.Notes == nil || strings.TrimSpace(*input.Notes) == "" {
			return nil, fmt.Errorf("%w: la o cultură compromisă, observațiile sunt obligatorii (ce s-a întâmplat)", ErrCropInvalid)
		}
	}
	if err := service.repo.CloseFieldCrop(id, input); err != nil {
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetFieldCropByID(id)
}

// ReopenFieldCrop redeschide o cultură închisă, dacă terenul nu are deja alta deschisă.
func (service *CropService) ReopenFieldCrop(id int64) (*domain.FieldCrop, error) {
	existing, err := service.repo.GetFieldCropByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrCropNotFound
	}
	if existing.Status == domain.FieldCropStatusOpen {
		return existing, nil
	}
	if err := service.ensureNoOtherOpenCrop(existing.FieldID, id); err != nil {
		return nil, err
	}
	if err := service.repo.ReopenFieldCrop(id); err != nil {
		if errors.Is(err, repository.ErrDuplicateEntry) {
			return nil, ErrFieldHasOpenCrop
		}
		return nil, mapCropRepoError(err)
	}
	return service.repo.GetFieldCropByID(id)
}

func (service *CropService) DeleteFieldCrop(id int64) error {
	return mapCropRepoError(service.repo.DeleteFieldCrop(id))
}

// HarvestResult conține cultura pe teren actualizată și mișcarea de stoc generată (dacă a fost nevoie).
type HarvestResult struct {
	FieldCrop *domain.FieldCrop
	Movement  *domain.StockMovement
}

// RecordHarvest înregistrează producția obținută ca intrare în stoc, pe resursa de recoltă a
// culturii. La o nouă apelare, doar diferența față de cantitatea deja înregistrată este mutată.
func (service *CropService) RecordHarvest(fieldCropID int64, actorID *int64) (*HarvestResult, error) {
	item, err := service.repo.GetFieldCropByID(fieldCropID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrCropNotFound
	}
	if item.ProductionTotal == nil || *item.ProductionTotal <= 0 {
		return nil, fmt.Errorf("%w: introdu mai întâi producția obținută", ErrHarvestNotRecordable)
	}
	crop, err := service.repo.GetCropByID(item.CropID)
	if err != nil {
		return nil, err
	}
	if crop == nil {
		return nil, ErrCropNotFound
	}

	recorded := 0.0
	if item.HarvestRecordedQty != nil {
		recorded = *item.HarvestRecordedQty
	}
	delta := *item.ProductionTotal - recorded
	if delta == 0 {
		return &HarvestResult{FieldCrop: item}, nil
	}

	tx, err := service.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	// Producția și cantitatea înregistrată se recitesc sub blocare: cererile simultane pentru
	// aceeași cultură pe teren (ex. dublu-click) sunt serializate și nu mută recolta de două ori.
	production, recorded, err := service.repo.LockFieldCropHarvest(tx, item.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCropNotFound
	}
	if err != nil {
		return nil, err
	}
	if production == nil || *production <= 0 {
		return nil, fmt.Errorf("%w: introdu mai întâi producția obținută", ErrHarvestNotRecordable)
	}
	delta = *production - recorded
	if delta == 0 {
		// o cerere simultană a înregistrat deja aceeași producție
		_ = tx.Rollback()
		updated, err := service.repo.GetFieldCropByID(item.ID)
		if err != nil {
			return nil, err
		}
		return &HarvestResult{FieldCrop: updated}, nil
	}

	resourceID, err := service.repo.EnsureHarvestResource(tx, crop)
	if err != nil {
		return nil, err
	}
	lock, err := service.movements.LockStock(tx, resourceID)
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, fmt.Errorf("%w: resursa de recoltă nu a fost găsită", ErrHarvestNotRecordable)
	}

	movementType := domain.StockMovementIn
	quantity := delta
	if delta < 0 {
		movementType = domain.StockMovementOut
		quantity = -delta
	}
	// O recoltă repetată (lucernă: mai multe coase) e tot „Recoltă”; „Corecție” doar când scade.
	notes := fmt.Sprintf("Recoltă %s - %s (%s)", crop.Name, item.FieldName, item.SeasonName)
	if delta < 0 {
		notes = fmt.Sprintf("Corecție recoltă %s - %s (%s)", crop.Name, item.FieldName, item.SeasonName)
	}
	movement, err := buildMovement(lock, movementType, quantity, nil, nil, notes, actorID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHarvestNotRecordable, err)
	}
	movement.FieldCropID = &item.ID
	if err := service.movements.ApplyMovement(tx, movement); err != nil {
		return nil, err
	}
	if err := service.repo.MarkHarvestRecorded(tx, item.ID, *production); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	updated, err := service.repo.GetFieldCropByID(item.ID)
	if err != nil {
		return nil, err
	}
	return &HarvestResult{FieldCrop: updated, Movement: movement}, nil
}
