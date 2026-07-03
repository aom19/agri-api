package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetAll() ([]domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.email_confirmed, u.role_id, COALESCE(ro.code, ''), COALESCE(ro.name, '')
		FROM users u
		LEFT JOIN roles ro ON ro.id = u.role_id
		WHERE u.deleted_at IS NULL
		ORDER BY u.id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailConfirmed, &u.RoleID, &u.RoleCode, &u.RoleName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.email_confirmed, u.role_id, COALESCE(ro.code, ''), COALESCE(ro.name, '')
		FROM users u
		LEFT JOIN roles ro ON ro.id = u.role_id
		WHERE u.email = $1 AND u.deleted_at IS NULL`
	var u domain.User
	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailConfirmed, &u.RoleID, &u.RoleCode, &u.RoleName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) GetByID(id int64) (*domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.email_confirmed, u.role_id, COALESCE(ro.code, ''), COALESCE(ro.name, '')
		FROM users u
		LEFT JOIN roles ro ON ro.id = u.role_id
		WHERE u.id = $1 AND u.deleted_at IS NULL`
	var u domain.User
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailConfirmed, &u.RoleID, &u.RoleCode, &u.RoleName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

func (r *UserRepo) Create(user *domain.User) error {
	query := `INSERT INTO users (email, password_hash, role_id) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRow(query, user.Email, user.PasswordHash, user.RoleID).Scan(&user.ID)
}

func (r *UserRepo) Update(user *domain.User) error {
	_, err := r.db.Exec(
		`UPDATE users
		 SET email = $1, role_id = $2, email_confirmed = $3, updated_at = NOW()
		 WHERE id = $4 AND deleted_at IS NULL`,
		user.Email,
		user.RoleID,
		user.EmailConfirmed,
		user.ID,
	)
	return err
}

func (r *UserRepo) Delete(id int64) error {
	_, err := r.db.Exec(`UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func (r *UserRepo) MarkEmailConfirmed(id int64) error {
	_, err := r.db.Exec(`UPDATE users SET email_confirmed = TRUE, updated_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *UserRepo) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, passwordHash, id)
	return err
}

func (r *UserRepo) UpdateRole(userID int64, roleID int64) error {
	_, err := r.db.Exec(`UPDATE users SET role_id = $1, updated_at = NOW() WHERE id = $2`, roleID, userID)
	return err
}

func (r *UserRepo) GetProfile(userID int64) (*domain.UserProfile, error) {
	query := `
		SELECT u.id, u.email, u.role_id, COALESCE(ro.code, ''), COALESCE(ro.name, ''),
		       COALESCE(p.first_name,''), COALESCE(p.last_name,''),
		       p.date_of_birth, COALESCE(p.profile_photo,''),
		       COALESCE(p.created_at, NOW()), COALESCE(p.updated_at, NOW())
		FROM users u
		LEFT JOIN roles ro ON ro.id = u.role_id
		LEFT JOIN user_profiles p ON p.user_id = u.id
		WHERE u.id = $1 AND u.deleted_at IS NULL`
	var p domain.UserProfile
	err := r.db.QueryRow(query, userID).Scan(
		&p.UserID, &p.Email, &p.RoleID, &p.RoleCode, &p.RoleName,
		&p.FirstName, &p.LastName,
		&p.DateOfBirth, &p.ProfilePhoto,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &domain.UserProfile{UserID: userID}, nil
	}
	return &p, err
}

func (r *UserRepo) UpsertProfile(p *domain.UserProfile) error {
	query := `
		INSERT INTO user_profiles (user_id, first_name, last_name, date_of_birth, profile_photo, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			first_name    = EXCLUDED.first_name,
			last_name     = EXCLUDED.last_name,
			date_of_birth = EXCLUDED.date_of_birth,
			profile_photo = EXCLUDED.profile_photo,
			updated_at    = NOW()`
	_, err := r.db.Exec(query, p.UserID, p.FirstName, p.LastName, p.DateOfBirth, p.ProfilePhoto)
	return err
}
