package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var (
	ErrFieldOperationNotFound              = errors.New("operațiunea pe teren nu a fost găsită")
	ErrFieldOperationInvalidStatus         = errors.New("statusul operațiunii nu este valid")
	ErrFieldOperationFieldRequired         = errors.New("terenul este obligatoriu")
	ErrFieldOperationTypeRequired          = errors.New("tipul operațiunii este obligatoriu")
	ErrFieldOperationTypeInvalid           = errors.New("tipul operațiunii nu este valid")
	ErrFieldOperationTemplateNotFound      = errors.New("template-ul operațiunii nu a fost găsit")
	ErrFieldOperationTypeMismatch          = errors.New("tipul operațiunii nu corespunde template-ului")
	ErrFieldOperationBadDates              = errors.New("sfârșitul planificat nu poate fi înaintea începutului planificat")
	ErrFieldOperationResourcesUnavailable  = errors.New("mașina și echipamentul trebuie să fie active înainte de pornire")
	ErrFieldOperationCannotStart           = errors.New("lucrarea nu poate fi pornită din statusul curent")
	ErrFieldOperationMachineIncompatible   = errors.New("tipul mașinii nu este acceptat de template-ul operațiunii")
	ErrFieldOperationImplementIncompatible = errors.New("tipul echipamentului nu este acceptat de template-ul operațiunii")
	ErrFieldOperationArchivedReference     = errors.New("nu poți alege elemente șterse")
	// ErrFieldOperationResourceBusy: operatorul, mașina sau echipamentul sunt deja pe o lucrare în lucru.
	ErrFieldOperationResourceBusy = errors.New("resursele lucrării sunt deja pe o lucrare în lucru")
)

type FieldOperationService struct {
	repo repository.FieldOperationRepository
}

func NewFieldOperationService(repo repository.FieldOperationRepository) *FieldOperationService {
	return &FieldOperationService{repo: repo}
}

func (s *FieldOperationService) GetAll(filter repository.FieldOperationFilter) ([]dto.FieldOperationResponse, error) {
	return s.repo.GetAll(filter)
}

func (s *FieldOperationService) GetByID(id int64) (*dto.FieldOperationResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrFieldOperationNotFound
	}
	return item, nil
}

func (s *FieldOperationService) GetByIDForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error) {
	item, err := s.repo.GetByIDForAssignedUser(id, userID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrFieldOperationNotFound
	}
	return item, nil
}

func (s *FieldOperationService) Create(input *domain.FieldOperation) (*dto.FieldOperationResponse, error) {
	if err := s.validate(input); err != nil {
		return nil, err
	}
	if err := s.validateActiveReferences(input, nil); err != nil {
		return nil, err
	}
	if input.Status == "" {
		input.Status = domain.FieldOperationStatusPlanned
	}
	if input.Status == domain.FieldOperationStatusInProgress {
		if err := s.ensureResourcesFree(0, input, nil); err != nil {
			return nil, err
		}
	}
	input.Notes = strings.TrimSpace(input.Notes)

	if err := s.repo.Create(input); err != nil {
		return nil, err
	}
	return s.repo.GetByID(input.ID)
}

func (s *FieldOperationService) Update(id int64, input *domain.FieldOperation) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := s.validate(input); err != nil {
		return nil, err
	}
	if err := s.validateActiveReferences(input, existing); err != nil {
		return nil, err
	}
	if input.Status == "" {
		input.Status = domain.FieldOperationStatus(existing.Status)
	}
	// trecerea în lucru din formular are aceeași regulă ca pornirea
	if input.Status == domain.FieldOperationStatusInProgress && existing.Status != string(domain.FieldOperationStatusInProgress) {
		if err := s.ensureResourcesFree(id, input, nil); err != nil {
			return nil, err
		}
	}
	input.Notes = strings.TrimSpace(input.Notes)

	if err := s.repo.Update(id, input); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *FieldOperationService) Start(id int64) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := validateStartReadiness(existing); err != nil {
		return nil, err
	}
	if existing.Status == string(domain.FieldOperationStatusInProgress) {
		return existing, nil
	}
	if err := s.ensureResourcesFree(id, operationResources(existing), existing); err != nil {
		return nil, err
	}
	if err := s.repo.MarkStarted(id, time.Now()); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *FieldOperationService) StartForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByIDForAssignedUser(id, userID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := validateStartReadiness(existing); err != nil {
		return nil, err
	}
	if existing.Status == string(domain.FieldOperationStatusInProgress) {
		return existing, nil
	}
	if err := s.ensureResourcesFree(id, operationResources(existing), existing); err != nil {
		return nil, err
	}
	if err := s.repo.MarkStarted(id, time.Now()); err != nil {
		return nil, err
	}
	return s.repo.GetByIDForAssignedUser(id, userID)
}

// FindConflicts întoarce lucrările planificate sau în lucru care folosesc operatorul, mașina sau
// echipamentul dat într-un interval suprapus. E doar un avertisment pentru formular: salvarea nu
// e blocată. Fără dată de început sau fără nicio resursă, nu există conflicte.
func (s *FieldOperationService) FindConflicts(query repository.FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error) {
	if query.StartAt.IsZero() || (query.OperatorID == nil && query.MachineID == nil && query.ImplementID == nil) {
		return []dto.FieldOperationConflict{}, nil
	}
	if query.EndAt != nil && query.EndAt.Before(query.StartAt) {
		return nil, ErrFieldOperationBadDates
	}
	return s.repo.FindConflicts(query)
}

// ensureResourcesFree refuză trecerea în lucru cât timp operatorul, mașina sau echipamentul sunt
// pe altă lucrare în lucru: nimeni nu poate fi în două locuri deodată. names dă numele resurselor
// în mesaj (nil = fără nume).
func (s *FieldOperationService) ensureResourcesFree(id int64, op *domain.FieldOperation, names *dto.FieldOperationResponse) error {
	if op.OperatorID == nil && op.MachineID == nil && op.ImplementID == nil {
		return nil
	}
	busy, err := s.repo.InProgressConflicts(repository.FieldOperationConflictQuery{
		ExcludeID:   id,
		OperatorID:  op.OperatorID,
		MachineID:   op.MachineID,
		ImplementID: op.ImplementID,
	})
	if err != nil {
		return err
	}
	if len(busy) == 0 {
		return nil
	}
	return &resourceBusyError{message: busyMessage(busy[0], names)}
}

// operationResources sunt resursele unei lucrări salvate.
func operationResources(op *dto.FieldOperationResponse) *domain.FieldOperation {
	return &domain.FieldOperation{OperatorID: op.OperatorID, MachineID: op.MachineID, ImplementID: op.ImplementID}
}

type resourceBusyError struct{ message string }

func (e *resourceBusyError) Error() string        { return e.message }
func (e *resourceBusyError) Is(target error) bool { return target == ErrFieldOperationResourceBusy }

// busyMessage: „Operatorul Ion Pop și mașina JD 6R sunt deja pe lucrarea #12 (Arat - Tarla Mare),
// aflată în lucru. Finalizeaz-o întâi.”
func busyMessage(conflict dto.FieldOperationConflict, names *dto.FieldOperationResponse) string {
	labels := make([]string, 0, len(conflict.Resources))
	for _, resource := range conflict.Resources {
		labels = append(labels, resourceLabel(resource, names))
	}
	who := strings.Join(labels, " și ")
	verb := "este"
	if len(labels) > 1 {
		verb = "sunt"
	}
	if who == "" {
		who, verb = "Resursele lucrării", "sunt"
	} else {
		who = strings.ToUpper(who[:1]) + who[1:]
	}
	return fmt.Sprintf("%s %s deja pe lucrarea #%d (%s - %s), aflată în lucru. Finalizeaz-o întâi.",
		who, verb, conflict.OperationID, conflict.OperationTypeName, conflict.FieldName)
}

func resourceLabel(resource string, names *dto.FieldOperationResponse) string {
	var label string
	var name *string
	switch resource {
	case dto.ConflictResourceOperator:
		label = "operatorul"
		if names != nil {
			name = names.OperatorName
		}
	case dto.ConflictResourceMachine:
		label = "mașina"
		if names != nil {
			name = names.MachineName
		}
	case dto.ConflictResourceImplement:
		label = "echipamentul"
		if names != nil {
			name = names.ImplementName
		}
	default:
		return resource
	}
	if name != nil && strings.TrimSpace(*name) != "" {
		return label + " " + strings.TrimSpace(*name)
	}
	return label
}

func (s *FieldOperationService) Delete(id int64) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrFieldOperationNotFound
	}
	return s.repo.Delete(id)
}

func (s *FieldOperationService) validate(input *domain.FieldOperation) error {
	if input == nil {
		return errors.New("datele operațiunii lipsesc")
	}
	if strings.TrimSpace(input.FieldID) == "" {
		return ErrFieldOperationFieldRequired
	}
	if err := s.resolveOperationType(input); err != nil {
		return err
	}
	if input.Status != "" && !isValidFieldOperationStatus(input.Status) {
		return ErrFieldOperationInvalidStatus
	}
	if input.AreaPlannedHa != nil && *input.AreaPlannedHa < 0 {
		return errors.New("suprafața planificată nu poate fi negativă")
	}
	if input.PlannedStartAt != nil && input.PlannedEndAt != nil &&
		input.PlannedEndAt.Before(*input.PlannedStartAt) {
		return ErrFieldOperationBadDates
	}
	return s.validateAssetCompatibility(input)
}

// resolveOperationType aplică regula tipului: cu template, operațiunea moștenește tipul lui și nu
// reține unul propriu (un tip trimis trebuie să fie același); fără template, tipul e obligatoriu.
func (s *FieldOperationService) resolveOperationType(input *domain.FieldOperation) error {
	if input.OperationTemplateID == nil {
		if input.OperationType == "" {
			return ErrFieldOperationTypeRequired
		}
		if !input.OperationType.IsValid() {
			return ErrFieldOperationTypeInvalid
		}
		return nil
	}
	templateType, err := s.repo.GetTemplateOperationType(*input.OperationTemplateID)
	if err != nil {
		return err
	}
	if templateType == "" {
		return ErrFieldOperationTemplateNotFound
	}
	if input.OperationType != "" && input.OperationType != templateType {
		return ErrFieldOperationTypeMismatch
	}
	input.OperationType = ""
	return nil
}

// validateAssetCompatibility verifică mașina și echipamentul față de tipurile acceptate de
// template. Un template fără tipuri, sau o operațiune fără template, acceptă orice.
func (s *FieldOperationService) validateAssetCompatibility(input *domain.FieldOperation) error {
	if input.OperationTemplateID == nil || (input.MachineID == nil && input.ImplementID == nil) {
		return nil
	}
	c, err := s.repo.GetAssetCompatibility(*input.OperationTemplateID, input.MachineID, input.ImplementID)
	if err != nil {
		return err
	}
	if !typeAllowed(c.TemplateMachineTypes, c.MachineType) {
		return ErrFieldOperationMachineIncompatible
	}
	if !typeAllowed(c.TemplateImplementTypes, c.ImplementType) {
		return ErrFieldOperationImplementIncompatible
	}
	return nil
}

// validateActiveReferences refuză un teren, template, mașină, echipament sau operator șters
// (arhivat). La editare se verifică doar ce se schimbă: o lucrare veche rămâne editabilă și după
// ce mașina ei a fost scoasă din flotă.
func (s *FieldOperationService) validateActiveReferences(input *domain.FieldOperation, existing *dto.FieldOperationResponse) error {
	refs := domain.FieldOperation{
		FieldID:             input.FieldID,
		OperationTemplateID: input.OperationTemplateID,
		MachineID:           input.MachineID,
		ImplementID:         input.ImplementID,
		OperatorID:          input.OperatorID,
	}
	if existing != nil {
		if refs.FieldID == existing.FieldID {
			refs.FieldID = ""
		}
		refs.OperationTemplateID = changedID(refs.OperationTemplateID, existing.OperationTemplateID)
		refs.MachineID = changedID(refs.MachineID, existing.MachineID)
		refs.ImplementID = changedID(refs.ImplementID, existing.ImplementID)
		refs.OperatorID = changedID(refs.OperatorID, existing.OperatorID)
	}
	archived, err := s.repo.ArchivedReferences(&refs)
	if err != nil {
		return err
	}
	if len(archived) > 0 {
		return fmt.Errorf("%w: %s", ErrFieldOperationArchivedReference, strings.Join(archived, ", "))
	}
	return nil
}

// changedID întoarce id-ul nou doar dacă diferă de cel existent.
func changedID(next, current *int64) *int64 {
	if next != nil && current != nil && *next == *current {
		return nil
	}
	return next
}

func typeAllowed(allowed []string, assetType string) bool {
	return assetType == "" || len(allowed) == 0 || slices.Contains(allowed, assetType)
}

func validateStartReadiness(operation *dto.FieldOperationResponse) error {
	if operation.Status == string(domain.FieldOperationStatusInProgress) {
		return nil
	}
	if operation.Status != string(domain.FieldOperationStatusPlanned) {
		return ErrFieldOperationCannotStart
	}
	if operation.MachineID != nil && !isActiveAssetStatus(operation.MachineStatus) {
		return ErrFieldOperationResourcesUnavailable
	}
	if operation.ImplementID != nil && !isActiveAssetStatus(operation.ImplementStatus) {
		return ErrFieldOperationResourcesUnavailable
	}
	return nil
}

func isActiveAssetStatus(status *string) bool {
	return status != nil && *status == string(domain.AssetStatusActive)
}

func isValidFieldOperationStatus(status domain.FieldOperationStatus) bool {
	switch status {
	case domain.FieldOperationStatusPlanned,
		domain.FieldOperationStatusInProgress,
		domain.FieldOperationStatusCompleted,
		domain.FieldOperationStatusCanceled:
		return true
	default:
		return false
	}
}
