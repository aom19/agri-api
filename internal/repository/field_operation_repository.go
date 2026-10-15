package repository

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"database/sql"
	"time"
)

type FieldOperationFilter struct {
	Status         string
	FieldID        string
	OperationType  string
	MachineID      string
	OperatorID     string
	AssignedUserID int64
}

// FieldOperationConflictQuery descrie o lucrare (salvată sau doar în formular) pentru care se
// caută suprapuneri: resursele ei și intervalul planificat. ExcludeID e lucrarea însăși, la editare.
type FieldOperationConflictQuery struct {
	ExcludeID   int64
	OperatorID  *int64
	MachineID   *int64
	ImplementID *int64
	StartAt     time.Time
	EndAt       *time.Time
}

type FieldOperationRepository interface {
	GetAll(filter FieldOperationFilter) ([]dto.FieldOperationResponse, error)
	GetByID(id int64) (*dto.FieldOperationResponse, error)
	GetByIDForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error)
	Create(op *domain.FieldOperation) error
	Update(id int64, op *domain.FieldOperation) error
	UpdateStatus(id int64, status domain.FieldOperationStatus) error
	// MarkStarted trece operațiunea în lucru și reține momentul real al pornirii (prima pornire).
	MarkStarted(id int64, at time.Time) error
	// Complete finalizează operațiunea cu datele reale, în tranzacția dată.
	Complete(tx *sql.Tx, id int64, completion domain.FieldOperationCompletion, at time.Time) error
	// GetAssetCompatibility returnează tipurile de mașini și echipamente acceptate de template
	// și tipurile mașinii și echipamentului date (nil = neales).
	GetAssetCompatibility(templateID int64, machineID, implementID *int64) (*domain.AssetCompatibility, error)
	// ArchivedReferences întoarce, dintre terenul, template-ul, mașina, echipamentul și operatorul
	// operațiunii, pe cele șterse (arhivate); câmpurile goale nu se verifică.
	ArchivedReferences(op *domain.FieldOperation) ([]string, error)
	// GetTemplateOperationType returnează tipul template-ului (gol = template inexistent).
	GetTemplateOperationType(templateID int64) (domain.OperationType, error)
	// GetTemplateResources returnează normele de consum ale șablonului (resurse + prețuri).
	GetTemplateResources(templateID int64) ([]domain.TemplateResourceUsage, error)
	// AddMachineHours adaugă ore de funcționare la mașină, în tranzacția dată.
	AddMachineHours(tx *sql.Tx, machineID int64, hours float64) error
	Delete(id int64) error
	// FindConflicts întoarce lucrările planificate sau în lucru care folosesc operatorul, mașina
	// sau echipamentul dat într-un interval suprapus.
	FindConflicts(query FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error)
	// InProgressConflicts întoarce celelalte lucrări în lucru care folosesc operatorul, mașina
	// sau echipamentul dat, indiferent de interval (intervalul din query e ignorat): acestea
	// blochează pornirea.
	InProgressConflicts(query FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error)
	// GetOverdueInProgress returnează operațiunile în lucru al căror sfârșit planificat
	// este înainte de `now` și pentru care nu s-a trimis încă notificarea de depășire.
	GetOverdueInProgress(now time.Time) ([]dto.OverdueFieldOperation, error)
	// MarkOverdueNotified marchează că notificarea de depășire a fost trimisă.
	MarkOverdueNotified(id int64, at time.Time) error
}
