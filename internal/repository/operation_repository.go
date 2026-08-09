package repository

import "agri-api/internal/domain"

type OperationTypeRepository interface {
	GetAll() ([]domain.OperationType, error)
	GetByID(id int64) (*domain.OperationType, error)
	Create(ot *domain.OperationType) error
	Update(id int64, ot *domain.OperationType) error
	Delete(id int64) error
}

type OperationTemplateRepository interface {
	GetAll() ([]domain.OperationTemplate, error)
	GetByID(id int64) (*domain.OperationTemplate, error)
	GetByOperationType(operationTypeID int64) ([]domain.OperationTemplate, error)
	Create(t *domain.OperationTemplate) error
	Update(id int64, t *domain.OperationTemplate) error
	Delete(id int64) error
	SetResources(templateID int64, resources []domain.TemplateResource) error
	SetMachineTypes(templateID int64, machineTypes []string) error
	SetImplementTypes(templateID int64, implementTypes []string) error
}
