package store

import (
	"agri-api/internal/repository/postgres"
	"database/sql"
)

func NewInitialiedStore(db *sql.DB) *Store {

	return &Store{
		DB:            db,
		MachineRepo:   postgres.NewMachineRepo(db),
		OperatorRepo:  postgres.NewOperatorRepo(db),
		AssigmentRepo: postgres.NewAssignmentRepo(db),
		UserRepo:      postgres.NewUserRepo(db),
	}
}
