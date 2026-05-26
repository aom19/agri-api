package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/store"
	"database/sql"
	"errors"
	"log"
)

// AssigmentService conține logica de business pentru gestionarea asignărilor între mașini și operatori
type AssigmentService struct {
	store *store.Store
}

func NewAssigmentService(s *store.Store) *AssigmentService {
	return &AssigmentService{
		store: s,
	}
}

// GetAssigments returnează lista completă a asignmenturilor
func (assigmentService *AssigmentService) GetAssigments() ([]domain.Assigment, error) {
	return assigmentService.store.AssigmentRepo.GetAll()
}

// GetAssigmentByID returnează un asignment după ID sau eroare dacă nu există
func (assigmentService *AssigmentService) GetAssigmentByID(id int64) (*domain.Assigment, error) {
	assigment, err := assigmentService.store.AssigmentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return assigment, nil
}

func (assigmentService *AssigmentService) CreateAssigment(assigment *domain.Assigment) (*domain.Assigment, error) {

	var result *domain.Assigment
	var opErr error

	opErr = assigmentService.store.WithTx(func(tx *sql.Tx) error {
		// Verifică dacă mașina există și este disponibilă
		machine, err := assigmentService.store.MachineRepo.GetByID(assigment.MachineID)
		if err != nil {
			return err
		}
		if machine == nil {
			return errors.New("machine not found")
		}
		// Verifică dacă operatorul există
		operator, err := assigmentService.store.OperatorRepo.GetByID(assigment.OperatorID)
		if err != nil {
			return err
		}
		if operator == nil {
			return errors.New("operator not found")
		}

		//masina  este alocata
		activeByMachine, err := assigmentService.store.AssigmentRepo.GetActiveByMachine(assigment.MachineID)
		if err != nil {
			return err
		}
		if activeByMachine != nil {
			return errors.New("machine is already assigned")
		}

		//operatorul este alocat
		activeByOperator, err := assigmentService.store.AssigmentRepo.GetActiveByOperator(assigment.OperatorID)
		if err != nil {
			return err
		}
		if activeByOperator != nil {
			return errors.New("operator is already assigned")
		}

		assigmentCreated := &domain.Assigment{
			MachineID:  assigment.MachineID,
			OperatorID: assigment.OperatorID,
			StartDate:  assigment.StartDate,
			EndDate:    assigment.EndDate,
			Status:     domain.AssigmentStatusActive,
		}
		if err := assigmentService.store.AssigmentRepo.Create(tx, assigmentCreated); err != nil {
			return err
		}
		if err := assigmentService.store.MachineRepo.UpdateStatus(tx, assigment.MachineID, domain.MachineStatusInUse); err != nil {
			return err
		}
		if err := assigmentService.store.OperatorRepo.UpdateStatus(tx, assigment.OperatorID, domain.OperatorStatusActive); err != nil {
			return err
		}
		result = assigmentCreated
		return nil
	})
	if opErr != nil {
		return nil, opErr
	}
	return result, nil
}

func (assigmentService *AssigmentService) UpdateAssigment(id int64, assigment *domain.Assigment) (*domain.Assigment, error) {
	//verifică dacă asignmentul există
	existing, err := assigmentService.store.AssigmentRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("assigment not found")
	}
	existing.MachineID = assigment.MachineID
	existing.OperatorID = assigment.OperatorID
	existing.StartDate = assigment.StartDate
	existing.EndDate = assigment.EndDate
	existing.Status = assigment.Status

	if err := assigmentService.store.AssigmentRepo.Update(id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (assigmentService *AssigmentService) DeleteAssigment(id int64) error {
	//verifică dacă asignmentul există
	existing, err := assigmentService.store.AssigmentRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("assigment not found")
	}
	isActive, err := assigmentService.store.AssigmentRepo.IsAssigmentActive(id)
	if err != nil {
		log.Default().Printf("Error checking if assigment is active: %v", err)
		return err
	}
	log.Default().Printf("Is assigment active: %v", isActive)
	if isActive {
		return errors.New("cannot delete active assigment")
	}
	return assigmentService.store.AssigmentRepo.Delete(id)
}

func (assigmentService *AssigmentService) CloseAssigment(id int64) (bool, error) {
	assigment, err := assigmentService.store.AssigmentRepo.GetByID(id)
	if err != nil {
		return false, err
	}
	if assigment == nil {
		return false, errors.New("assigment not found")
	}
	if assigment.Status != domain.AssigmentStatusActive {
		return false, errors.New("assigment is not active")
	}

	err = assigmentService.store.WithTx(func(tx *sql.Tx) error {
		assigment.Status = domain.AssigmentStatusClosed
		if err := assigmentService.store.AssigmentRepo.Update(id, assigment); err != nil {
			return err
		}
		if err := assigmentService.store.MachineRepo.UpdateStatus(tx, assigment.MachineID, domain.MachineStatusAvailable); err != nil {
			return err
		}
		return assigmentService.store.OperatorRepo.UpdateStatus(tx, assigment.OperatorID, domain.OperatorStatusInactive)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
