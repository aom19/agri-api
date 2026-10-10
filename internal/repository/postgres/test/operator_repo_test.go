package postgres_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
)

// T9: operatorul e un singur cont (users) cu un singur profil; lucrarea îi e asignată direct.
func TestOperators_AreUsersWithProfile(t *testing.T) {
	db := requireDB(t)
	repo := postgres.NewOperatorRepo(db)

	// fără e-mail: contul primește o adresă tehnică, ascunsă în API, și nu se poate autentifica
	noEmail := &domain.Operator{FirstName: "Gheorghe", LastName: "Dănilă", Phone: "+37369100003"}
	if err := repo.Create(noEmail); err != nil {
		t.Fatal(err)
	}
	var email, role, hash string
	if err := db.QueryRow(`SELECT u.email, r.code, u.password_hash FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = $1`, noEmail.ID).
		Scan(&email, &role, &hash); err != nil {
		t.Fatal(err)
	}
	if !domain.IsPlaceholderEmail(email) || role != "operator" || hash != "!" {
		t.Errorf("contul operatorului fără e-mail: %s, %s, %s", email, role, hash)
	}
	got, err := repo.GetByID(noEmail.ID)
	if err != nil || got.Name != "Gheorghe Dănilă" || got.Email != "" || got.Phone != "+37369100003" ||
		got.Status != domain.OperatorStatusActive {
		t.Fatalf("operatorul citit: %v, %+v", err, got)
	}

	// adminul completează e-mailul real; un e-mail folosit de alt cont e refuzat
	withEmail := &domain.Operator{FirstName: "Mihai", Email: "mihai@agri.ro"}
	if err := repo.Create(withEmail); err != nil {
		t.Fatal(err)
	}
	if err := repo.Update(noEmail.ID, &domain.Operator{FirstName: "Gheorghe", Email: "mihai@agri.ro"}); !errors.Is(err, repository.ErrEmailTaken) {
		t.Errorf("e-mail duplicat: %v", err)
	}
	if err := repo.Update(noEmail.ID, &domain.Operator{FirstName: "Gheorghe", LastName: "Dănilă", Email: "gheorghe@agri.ro"}); err != nil {
		t.Fatal(err)
	}
	if got, _ = repo.GetByID(noEmail.ID); got.Email != "gheorghe@agri.ro" || got.Phone != "" {
		t.Errorf("după actualizare: %+v", got)
	}

	// dezactivarea e cea a contului
	if err := repo.SetActive(noEmail.ID, false); err != nil {
		t.Fatal(err)
	}
	if got, _ = repo.GetByID(noEmail.ID); got.Status != domain.OperatorStatusInactive {
		t.Errorf("operatorul dezactivat: %+v", got)
	}

	// conturile cu alte roluri nu sunt operatori
	viewerID := insertID(t, db, `INSERT INTO users (email, password_hash, role_id) VALUES ('viewer@agri.ro', '!', (SELECT id FROM roles WHERE code = 'viewer')) RETURNING id`)
	all, err := repo.GetAll()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAll: %v, %+v", err, all)
	}
	if viewer, _ := repo.GetByID(viewerID); viewer != nil {
		t.Errorf("un cont viewer nu e operator: %+v", viewer)
	}

	// lucrarea asignată operatorului e văzută de contul lui, cu numele din profil
	field := insertField(t, db, "Lot operator")
	opID := insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", OperatorID: &withEmail.ID, Status: "planned", PlannedStart: time.Now()})
	ops := postgres.NewFieldOperationRepo(db)
	mine, err := ops.GetByIDForAssignedUser(opID, withEmail.ID)
	if err != nil || mine == nil || mine.OperatorName == nil || strings.TrimSpace(*mine.OperatorName) != "Mihai" {
		t.Fatalf("lucrarea operatorului: %v, %+v", err, mine)
	}
	if other, err := ops.GetByIDForAssignedUser(opID, noEmail.ID); err != nil || other != nil {
		t.Errorf("alt operator nu trebuie să vadă lucrarea: %v, %+v", err, other)
	}
	list, err := ops.GetAll(repository.FieldOperationFilter{AssignedUserID: withEmail.ID})
	if err != nil || len(list) != 1 {
		t.Errorf("lista operatorului: %v, %d", err, len(list))
	}
}
