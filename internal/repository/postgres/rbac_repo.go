package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type RBACRepo struct {
	db *sql.DB
}

func NewRBACRepo(db *sql.DB) *RBACRepo {
	return &RBACRepo{db: db}
}

// ─── Role ────────────────────────────────────────────────────────────────────

func (r *RBACRepo) GetAll() ([]domain.Role, error) {
	rows, err := r.db.Query(`SELECT id, code, name, description, created_at FROM roles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *RBACRepo) GetByID(id int64) (*domain.Role, error) {
	var role domain.Role
	err := r.db.QueryRow(
		`SELECT id, code, name, description, created_at FROM roles WHERE id = $1`, id,
	).Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &role, err
}

func (r *RBACRepo) GetByCode(code string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.QueryRow(
		`SELECT id, code, name, description, created_at FROM roles WHERE code = $1`, code,
	).Scan(&role.ID, &role.Code, &role.Name, &role.Description, &role.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &role, err
}

func (r *RBACRepo) Create(role *domain.Role) error {
	return r.db.QueryRow(
		`INSERT INTO roles (code, name, description) VALUES ($1, $2, $3) RETURNING id, created_at`,
		role.Code, role.Name, role.Description,
	).Scan(&role.ID, &role.CreatedAt)
}

func (r *RBACRepo) Update(role *domain.Role) error {
	_, err := r.db.Exec(
		`UPDATE roles SET code = $1, name = $2, description = $3 WHERE id = $4`,
		role.Code, role.Name, role.Description, role.ID,
	)
	return err
}

func (r *RBACRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM roles WHERE id = $1`, id)
	return err
}

// GetPermissions returnează toate permisiunile asociate unui rol
func (r *RBACRepo) GetPermissions(roleID int64) ([]domain.Permission, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.description
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
		ORDER BY p.name`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

// SetPermissions înlocuiește complet setul de permisiuni al unui rol (în tranzacție)
func (r *RBACRepo) SetPermissions(roleID int64, permissionIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return err
	}

	for _, pid := range permissionIDs {
		if _, err = tx.Exec(
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, pid,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// ─── Permission ──────────────────────────────────────────────────────────────

func (r *RBACRepo) GetAllPermissions() ([]domain.Permission, error) {
	rows, err := r.db.Query(`SELECT id, name, description FROM permissions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (r *RBACRepo) GetPermissionByID(id int64) (*domain.Permission, error) {
	var p domain.Permission
	err := r.db.QueryRow(
		`SELECT id, name, description FROM permissions WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

// HasPermission este query-ul apelat de RequirePermission middleware la fiecare request
func (r *RBACRepo) HasPermission(roleID int64, permissionName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM role_permissions rp
			JOIN permissions p ON p.id = rp.permission_id
			WHERE rp.role_id = $1 AND p.name = $2
		)`, roleID, permissionName,
	).Scan(&exists)
	return exists, err
}
