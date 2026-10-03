package store_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"agri-api/internal/store"
)

func TestStore_WithTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st := store.NewStore(db)

	mock.ExpectBegin()
	mock.ExpectCommit()
	if err := st.WithTx(func(tx *sql.Tx) error { return nil }); err != nil {
		t.Errorf("commit: %v", err)
	}

	boom := errors.New("logic error")
	mock.ExpectBegin()
	mock.ExpectRollback()
	if err := st.WithTx(func(tx *sql.Tx) error { return boom }); !errors.Is(err, boom) {
		t.Errorf("rollback: %v", err)
	}

	mock.ExpectBegin().WillReturnError(boom)
	if err := st.WithTx(func(tx *sql.Tx) error { return nil }); !errors.Is(err, boom) {
		t.Errorf("begin eșuat: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(boom)
	if err := st.WithTx(func(tx *sql.Tx) error { return nil }); !errors.Is(err, boom) {
		t.Errorf("commit eșuat: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestNewInitialiedStore(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	st := store.NewInitialiedStore(db)
	if st.DB != db || st.MachineRepo == nil || st.OperatorRepo == nil || st.FieldRepo == nil ||
		st.UserRepo == nil || st.RoleRepo == nil || st.PermissionRepo == nil {
		t.Error("toate repository-urile trebuie inițializate")
	}
}
