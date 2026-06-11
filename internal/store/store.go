package store

import (
	"agri-api/internal/repository"
	"database/sql"
)

type Store struct {
	DB *sql.DB

	MachineRepo    repository.MachineRepository
	OperatorRepo   repository.OperatorRepository
	FieldRepo      repository.FieldRepository
	AssigmentRepo  repository.AssigmentRepository
	UserRepo       repository.UserRepository
	RoleRepo       repository.RoleRepository
	PermissionRepo repository.PermissionRepository
}

// NewStore creează o nouă instanță a Store-ului și inițializează repository-urile
func NewStore(db *sql.DB) *Store {
	return &Store{
		DB: db,
	}
}

// WithTx execută o funcție care primește un *sql.Tx, asigurând commit sau rollback în funcție de rezultat
func (s *Store) WithTx(fn func(tx *sql.Tx) error) error {

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
