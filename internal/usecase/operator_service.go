package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
)

// OperatorService conține logica de business pentru gestionarea operatorilor
type OperatorService struct {
	operatorRepo repository.OperatorRepository
}

func NewOperatorService(operatorRepo repository.OperatorRepository) *OperatorService {
	return &OperatorService{operatorRepo: operatorRepo}
}

// GetOperators returnează lista completă a operatorilor
func (operatorService *OperatorService) GetOperators() ([]domain.Operator, error) {
	return operatorService.operatorRepo.GetAll()
}

// GetOperatorByID returnează un operator după ID sau eroare dacă nu există
func (operatorService *OperatorService) GetOperatorByID(id int64) (*domain.Operator, error) {
	operator, err := operatorService.operatorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return operator, nil
}

// CreateOperator validează și creează un operator nou
func (operatorService *OperatorService) CreateOperator(operator *domain.Operator) (*domain.Operator, error) {
	if operator.Name == "" {
		return nil, errors.New("name is required")
	}
	o := &domain.Operator{
		Name:   operator.Name,
		Phone:  operator.Phone,
		Email:  operator.Email,
		Notes:  operator.Notes,
		Status: domain.OperatorStatusActive,
	}
	if err := operatorService.operatorRepo.Create(o); err != nil {
		return nil, err
	}
	return o, nil
}

// UpdateOperator actualizează câmpurile unui operator existent
func (operatorService *OperatorService) UpdateOperator(id int64, operator *domain.Operator) (*domain.Operator, error) {
	existing, err := operatorService.operatorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("operator not found")
	}
	existing.Name = operator.Name
	existing.Phone = operator.Phone
	existing.Email = operator.Email
	existing.Notes = operator.Notes
	if err := operatorService.operatorRepo.Update(id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (operatorService *OperatorService) DeleteOperator(id int64) error {
	if found, err := operatorService.operatorRepo.GetByID(id); err != nil {
		return err
	} else if found == nil {
		return errors.New("operator not found")
	}

	return operatorService.operatorRepo.Delete(id)
}

// DisableOperator dezactivează un operator (setează statusul la inactive)
func (operatorService *OperatorService) DisableOperator(id int64) (*domain.Operator, error) {
	existing, err := operatorService.operatorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("operator not found")
	}
	if err := operatorService.operatorRepo.UpdateStatusDirect(id, domain.OperatorStatusInactive); err != nil {
		return nil, err
	}
	existing.Status = domain.OperatorStatusInactive
	return existing, nil
}

// EnableOperator reactivează un operator (setează statusul la active)
func (operatorService *OperatorService) EnableOperator(id int64) (*domain.Operator, error) {
	existing, err := operatorService.operatorRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("operator not found")
	}
	if err := operatorService.operatorRepo.UpdateStatusDirect(id, domain.OperatorStatusActive); err != nil {
		return nil, err
	}
	existing.Status = domain.OperatorStatusActive
	return existing, nil
}
