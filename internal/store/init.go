package store

import (
	"agri-api/internal/repository/postgres"
	"database/sql"
)

func NewInitialiedStore(db *sql.DB) *Store {
	rbacRepo := postgres.NewRBACRepo(db)
	return &Store{
		DB:             db,
		MachineRepo:    postgres.NewMachineRepo(db),
		OperatorRepo:   postgres.NewOperatorRepo(db),
		FieldRepo:      postgres.NewFieldRepo(db),
		AssigmentRepo:  postgres.NewAssignmentRepo(db),
		UserRepo:       postgres.NewUserRepo(db),
		RoleRepo:       rbacRepo,
		PermissionRepo: rbacRepo,
	}
}
