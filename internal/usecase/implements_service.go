package usecase

import(
	"errors"
	"agri-api/internal/domain"
	"agri-api/internal/repository"
)


// ImplementService conține logica de business pentru gestionarea utilajelor agricole
type ImplementService struct {
	implementRepo repository.ImplementRepository
}

// NewImplementService creează o nouă instanță a serviciului cu repository-ul injectat
func NewImplementService(implementRepo repository.ImplementRepository) *ImplementService {
	return &ImplementService{implementRepo: implementRepo}
}

//toate utilajele agricole
func (implementService *ImplementService) GetImplements() ([]domain.Implement, error) {
	return implementService.implementRepo.GetAll()
}

//utilajul agricol by ID
func (implementService *ImplementService) GetImplementByID(id int64) (*domain.Implement, error) {
	implement, err := implementService.implementRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return implement, nil
}

//creează un utilaj agricol nou 
func (implementService *ImplementService) CreateImplement(implement *domain.Implement) (*domain.Implement, error) {
	if implement.Name == "" || implement.Code == "" || implement.Type == "" || implement.Status == "" {
		return nil, errors.New("name, code, type si status sunt  obligatorii")
	}
	err := implementService.implementRepo.Create(implement)
	if err != nil {
		return nil, err
	}
	//returnează implementul creat si eroare nil
	return implement, nil
}


//actualizează un utilaj agricol existent
func (implementService *ImplementService) UpdateImplement(id int64, implement *domain.Implement) (*domain.Implement, error) {
	if implement.Name == "" || implement.Code == "" || implement.Type == "" || implement.Status == "" {
		return nil, errors.New("name, code, type si status sunt  obligatorii")
	}
	
	err := implementService.implementRepo.Update(id, implement)
	if err != nil {
		return nil, err
	}
	return implement, nil
}

//sterge un utilaj agricol existent
func (implementService *ImplementService) DeleteImplement(id int64) error {
	return implementService.implementRepo.Delete(id)
}	


//activeaza
func (implementService *ImplementService) ActivateImplement(id int64) error {
	return implementService.implementRepo.Activate(id)
}

//dezactiveaza
func (implementService *ImplementService) DeactivateImplement(id int64) error {
	return implementService.implementRepo.Deactivate(id)
}
