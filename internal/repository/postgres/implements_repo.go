package postgres

import(
	"agri-api/internal/domain"
	"database/sql"

)


type ImplementRepo struct {
	db *sql.DB
}

// NewImplementRepo se creaza o noua instanta
func NewImplementRepo(db *sql.DB) *ImplementRepo {
	return &ImplementRepo{db: db}
}


// GetAll returneaza toate utilajele a
func (implementRepo *ImplementRepo) GetAll() ([]domain.Implement, error) {
	rows, err := implementRepo.db.Query("SELECT id, name, code, type, brand, model, year, status, notes, working_width, capacity, deleted_at FROM implements WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	//defer pentru a inchide rows dupa ce functia se termina
	defer func() { _ = rows.Close() }()

	var implements []domain.Implement
	//parcurgem toate randurile returnate 
	for rows.Next() {
		var m domain.Implement
		var year sql.NullInt64
		var workingWidth sql.NullFloat64
		var capacity sql.NullFloat64
		var deletedAt sql.NullTime
		if err := rows.Scan(&m.ID, &m.Name, &m.Code, &m.Type, &m.Brand, &m.Model, &year, &m.Status, &m.Notes, &workingWidth, &capacity, &deletedAt); err != nil {
			return nil, err
		}
		if year.Valid {
			y := int(year.Int64)
			m.Year = &y
		}
	
		if workingWidth.Valid {
			w := float64(workingWidth.Float64)
			m.WorkingWidth = &w
		}
		if capacity.Valid {
			c := float64(capacity.Float64)
			m.Capacity = &c
		}
		if deletedAt.Valid {
			d := deletedAt.Time
			m.DeletedAt = &d
		}
		implements = append(implements, m)
	}

	return implements, nil
}

// GetByID returneaza un utilaj agricol dupa ID 

func (implementRepo *ImplementRepo) GetByID(id int64) (*domain.Implement, error) {
	var m domain.Implement
	var year sql.NullInt64
	var workingWidth sql.NullFloat64
	var capacity sql.NullFloat64
	var deletedAt sql.NullTime
	err := implementRepo.db.QueryRow("SELECT id, name, code, type, brand, model, year, status, notes, working_width, capacity, deleted_at FROM implements WHERE id = $1 AND deleted_at IS NULL", id).Scan(&m.ID, &m.Name, &m.Code, &m.Type, &m.Brand, &m.Model, &year, &m.Status, &m.Notes, &workingWidth, &capacity, &deletedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if year.Valid {
		y := int(year.Int64)
		m.Year = &y
	}
	if workingWidth.Valid {
		w := float64(workingWidth.Float64)
		m.WorkingWidth = &w
	}
	if capacity.Valid {
		c := float64(capacity.Float64)
		m.Capacity = &c
	}
	if deletedAt.Valid {
		d := deletedAt.Time
		m.DeletedAt = &d
	}

	return &m, nil
}	

// Create creeaza un utilaj agricol nou
func (implementRepo *ImplementRepo) Create(implement *domain.Implement) error {


	var workingWidth any
	if implement.WorkingWidth != nil {
		workingWidth = *implement.WorkingWidth
	}
	var capacity any
	if implement.Capacity != nil {
		capacity = *implement.Capacity
	}
	return implementRepo.db.QueryRow("INSERT INTO implements (name, code, type, brand, model, year, status, notes, working_width, capacity) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		implement.Name,
		implement.Code,
		implement.Type,
		implement.Brand,
		implement.Model,
		implement.Year,
		implement.Status,
		implement.Notes,
		workingWidth,
		capacity,
	).Scan(&implement.ID)
}	

// Update actualizeaza un utilaj agricol existent
func (implementRepo *ImplementRepo) Update(id int64, implement *domain.Implement) error {
	var workingWidth any
	if implement.WorkingWidth != nil {
		workingWidth = *implement.WorkingWidth
	}
	var capacity any
	if implement.Capacity != nil {
		capacity = *implement.Capacity
	}
	_, err := implementRepo.db.Exec("UPDATE implements SET name = $1, code = $2, type = $3, brand = $4, model = $5, year = $6, status = $7, notes = $8, working_width = $9, capacity = $10, updated_at = NOW() WHERE id = $11 AND deleted_at IS NULL",
		implement.Name,
		implement.Code,
		implement.Type,
		implement.Brand,
		implement.Model,				
		implement.Year,	
		implement.Status,
		implement.Notes,	
		workingWidth,
		capacity,
		id,
	)
	return err
}

// Delete sterge un utilaj agricol existent
func (implementRepo *ImplementRepo) Delete(id int64) error {
	_, err := implementRepo.db.Exec("UPDATE implements SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

// Activate activeaza un utilaj agricol existent
func (implementRepo *ImplementRepo) Activate(id int64) error {
	_, err := implementRepo.db.Exec("UPDATE implements SET status = 'active', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}	

// Deactivate dezactiveaza un utilaj agricol existent
func (implementRepo *ImplementRepo) Deactivate(id int64) error {
	_, err := implementRepo.db.Exec("UPDATE implements SET status = 'inactive', updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

// DB returnează conexiunea la baza de date pentru inițierea tranzacțiilor
func (implementRepo *ImplementRepo) DB() *sql.DB {
	return implementRepo.db
}

