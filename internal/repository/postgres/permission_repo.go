package postgres 

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"

	"database/sql"
)

// PermissionRepo implementeaza repository.PermissionRepository folosind PostgreSQL.
type PermissionRepo struct {
	db *sql.DB
}

func NewPermissionRepo(db *sql.DB) repository.PermissionRepository {
	return &PermissionRepo{db: db}
}

func (repo *PermissionRepo) GetAll() ([]domain.Permission, error) {
	rows, err := repo.db.Query(`SELECT id, name, description FROM permissions ORDER BY id`)
	if err != nil {
		return nil, err
	}
	// Defer closing the rows to ensure resources are released after processing
	defer rows.Close()

	var permissions []domain.Permission
	for rows.Next() {
		var permission domain.Permission
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (repo *PermissionRepo) GetByID(id string) (*domain.Permission, error) {
	var permission domain.Permission
	err := repo.db.QueryRow(
		`SELECT id, name, description FROM permissions WHERE id = $1`, id,
	).Scan(&permission.ID, &permission.Name, &permission.Description)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil if no rows found
	}
	return &permission, err
}

