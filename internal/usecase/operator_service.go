package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
	"strings"
)

var (
	ErrOperatorNotFound = errors.New("operator not found")
	ErrOperatorInvalid  = errors.New("operator invalid")
)

// OperatorService gestionează operatorii: utilizatori cu rolul `operator`, cu profilul lor.
type OperatorService struct {
	operatorRepo repository.OperatorRepository
}

func NewOperatorService(operatorRepo repository.OperatorRepository) *OperatorService {
	return &OperatorService{operatorRepo: operatorRepo}
}

// GetOperators returnează toți operatorii, inclusiv cei dezactivați.
func (operatorService *OperatorService) GetOperators() ([]domain.Operator, error) {
	return operatorService.operatorRepo.GetAll()
}

// GetOperatorByID returnează operatorul (după id-ul contului) sau nil dacă nu există.
func (operatorService *OperatorService) GetOperatorByID(id int64) (*domain.Operator, error) {
	return operatorService.operatorRepo.GetByID(id)
}

// CreateOperator creează contul și profilul operatorului. Contul nu are parolă utilizabilă:
// operatorul primește acces când are un e-mail real și își setează parola.
func (operatorService *OperatorService) CreateOperator(operator *domain.Operator) (*domain.Operator, error) {
	if err := validateOperator(operator); err != nil {
		return nil, err
	}
	if err := operatorService.operatorRepo.Create(operator); err != nil {
		return nil, mapOperatorError(err)
	}
	return operatorService.operatorRepo.GetByID(operator.ID)
}

// UpdateOperator actualizează e-mailul contului și profilul operatorului.
func (operatorService *OperatorService) UpdateOperator(id int64, operator *domain.Operator) (*domain.Operator, error) {
	if err := validateOperator(operator); err != nil {
		return nil, err
	}
	if err := operatorService.ensureExists(id); err != nil {
		return nil, err
	}
	if err := operatorService.operatorRepo.Update(id, operator); err != nil {
		return nil, mapOperatorError(err)
	}
	return operatorService.operatorRepo.GetByID(id)
}

// DeleteOperator dezactivează contul. Lucrările asignate îl păstrează ca operator.
func (operatorService *OperatorService) DeleteOperator(id int64) error {
	_, err := operatorService.setActive(id, false)
	return err
}

// DisableOperator dezactivează contul operatorului.
func (operatorService *OperatorService) DisableOperator(id int64) (*domain.Operator, error) {
	return operatorService.setActive(id, false)
}

// EnableOperator reactivează contul operatorului.
func (operatorService *OperatorService) EnableOperator(id int64) (*domain.Operator, error) {
	return operatorService.setActive(id, true)
}

func (operatorService *OperatorService) setActive(id int64, active bool) (*domain.Operator, error) {
	if err := operatorService.ensureExists(id); err != nil {
		return nil, err
	}
	if err := operatorService.operatorRepo.SetActive(id, active); err != nil {
		return nil, err
	}
	return operatorService.operatorRepo.GetByID(id)
}

func (operatorService *OperatorService) ensureExists(id int64) error {
	existing, err := operatorService.operatorRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrOperatorNotFound
	}
	return nil
}

func validateOperator(operator *domain.Operator) error {
	if operator == nil || strings.TrimSpace(operator.FirstName) == "" {
		return errors.Join(ErrOperatorInvalid, errors.New("prenumele este obligatoriu"))
	}
	if domain.IsPlaceholderEmail(operator.Email) {
		return errors.Join(ErrOperatorInvalid, errors.New("adresa de e-mail nu este validă"))
	}
	return nil
}

func mapOperatorError(err error) error {
	if errors.Is(err, repository.ErrEmailTaken) {
		return errors.Join(ErrOperatorInvalid, errors.New("adresa de e-mail aparține altui cont"))
	}
	return err
}
