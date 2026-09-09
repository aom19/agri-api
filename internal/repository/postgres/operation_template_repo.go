package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type OperationTemplateRepo struct {
	db *sql.DB
}

func NewOperationTemplateRepo(db *sql.DB) *OperationTemplateRepo {
	return &OperationTemplateRepo{db: db}
}

func (r *OperationTemplateRepo) GetAll() ([]domain.OperationTemplate, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.operation_type_id, t.name, COALESCE(t.description, ''), t.unit, t.crop_id, c.name,
		        t.created_at, t.updated_at,
		        ot.id, ot.code, ot.name, COALESCE(ot.description, ''), ot.created_at, ot.updated_at
		 FROM operation_templates t
		 JOIN operation_types ot ON ot.id = t.operation_type_id
		 LEFT JOIN crops c ON c.id = t.crop_id
		 WHERE t.deleted_at IS NULL
		 ORDER BY t.name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	byID := make(map[int64]*domain.OperationTemplate)
	var result []domain.OperationTemplate
	for rows.Next() {
		var t domain.OperationTemplate
		var ot domain.OperationType
		if err := rows.Scan(
			&t.ID, &t.OperationTypeID, &t.Name, &t.Description, &t.Unit, &t.CropID, &t.CropName,
			&t.CreatedAt, &t.UpdatedAt,
			&ot.ID, &ot.Code, &ot.Name, &ot.Description, &ot.CreatedAt, &ot.UpdatedAt,
		); err != nil {
			return nil, err
		}
		t.OperationType = &ot
		t.Resources = []domain.TemplateResource{}
		t.MachineTypes = []string{}
		t.ImplementTypes = []string{}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range result {
		byID[result[i].ID] = &result[i]
	}
	if len(result) == 0 {
		return result, nil
	}

	if err := r.attachResources(byID); err != nil {
		return nil, err
	}
	if err := r.attachMachineTypes(byID); err != nil {
		return nil, err
	}
	if err := r.attachImplementTypes(byID); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *OperationTemplateRepo) attachResources(byID map[int64]*domain.OperationTemplate) error {
	rows, err := r.db.Query(
		`SELECT tr.id, tr.template_id, tr.resource_id, tr.quantity_per_unit, COALESCE(tr.notes, ''),
		        tr.created_at, tr.updated_at,
		        res.id, res.name, res.resource_type_id, res.price_per_unit, COALESCE(res.notes, ''),
		        res.created_at, res.updated_at
		 FROM template_resources tr
		 JOIN resources res ON res.id = tr.resource_id
		 JOIN operation_templates t ON t.id = tr.template_id
		 WHERE t.deleted_at IS NULL
		 ORDER BY res.name`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tr domain.TemplateResource
		var res domain.Resource
		if err := rows.Scan(
			&tr.ID, &tr.TemplateID, &tr.ResourceID, &tr.QuantityPerUnit, &tr.Notes,
			&tr.CreatedAt, &tr.UpdatedAt,
			&res.ID, &res.Name, &res.ResourceTypeID, &res.PricePerUnit, &res.Notes,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return err
		}
		tr.Resource = &res
		if tpl, ok := byID[tr.TemplateID]; ok {
			tpl.Resources = append(tpl.Resources, tr)
		}
	}
	return rows.Err()
}

func (r *OperationTemplateRepo) attachMachineTypes(byID map[int64]*domain.OperationTemplate) error {
	rows, err := r.db.Query(
		`SELECT tmt.template_id, tmt.machine_type
		 FROM template_machine_types tmt
		 JOIN operation_templates t ON t.id = tmt.template_id
		 WHERE t.deleted_at IS NULL
		 ORDER BY tmt.machine_type`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var templateID int64
		var mt string
		if err := rows.Scan(&templateID, &mt); err != nil {
			return err
		}
		if tpl, ok := byID[templateID]; ok {
			tpl.MachineTypes = append(tpl.MachineTypes, mt)
		}
	}
	return rows.Err()
}

func (r *OperationTemplateRepo) attachImplementTypes(byID map[int64]*domain.OperationTemplate) error {
	rows, err := r.db.Query(
		`SELECT tit.template_id, tit.implement_type
		 FROM template_implement_types tit
		 JOIN operation_templates t ON t.id = tit.template_id
		 WHERE t.deleted_at IS NULL
		 ORDER BY tit.implement_type`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var templateID int64
		var it string
		if err := rows.Scan(&templateID, &it); err != nil {
			return err
		}
		if tpl, ok := byID[templateID]; ok {
			tpl.ImplementTypes = append(tpl.ImplementTypes, it)
		}
	}
	return rows.Err()
}

func (r *OperationTemplateRepo) GetByID(id int64) (*domain.OperationTemplate, error) {
	var t domain.OperationTemplate
	var ot domain.OperationType

	err := r.db.QueryRow(
		`SELECT t.id, t.operation_type_id, t.name, COALESCE(t.description, ''), t.unit, t.crop_id, c.name,
		        t.created_at, t.updated_at,
		        ot.id, ot.code, ot.name, COALESCE(ot.description, ''), ot.created_at, ot.updated_at
		 FROM operation_templates t
		 JOIN operation_types ot ON ot.id = t.operation_type_id
		 LEFT JOIN crops c ON c.id = t.crop_id
		 WHERE t.id = $1 AND t.deleted_at IS NULL`, id,
	).Scan(
		&t.ID, &t.OperationTypeID, &t.Name, &t.Description, &t.Unit, &t.CropID, &t.CropName,
		&t.CreatedAt, &t.UpdatedAt,
		&ot.ID, &ot.Code, &ot.Name, &ot.Description, &ot.CreatedAt, &ot.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.OperationType = &ot

	// Load resources
	resources, err := r.loadResources(id)
	if err != nil {
		return nil, err
	}
	t.Resources = resources

	// Load machine types
	machineTypes, err := r.loadMachineTypes(id)
	if err != nil {
		return nil, err
	}
	t.MachineTypes = machineTypes

	// Load implement types
	implementTypes, err := r.loadImplementTypes(id)
	if err != nil {
		return nil, err
	}
	t.ImplementTypes = implementTypes

	return &t, nil
}

func (r *OperationTemplateRepo) GetByOperationType(operationTypeID int64) ([]domain.OperationTemplate, error) {
	rows, err := r.db.Query(
		`SELECT id, operation_type_id, name, COALESCE(description, ''), unit, crop_id, created_at, updated_at
		 FROM operation_templates
		 WHERE operation_type_id = $1 AND deleted_at IS NULL
		 ORDER BY name`, operationTypeID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.OperationTemplate
	for rows.Next() {
		var t domain.OperationTemplate
		if err := rows.Scan(&t.ID, &t.OperationTypeID, &t.Name, &t.Description, &t.Unit, &t.CropID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *OperationTemplateRepo) Create(t *domain.OperationTemplate) error {
	return r.db.QueryRow(
		`INSERT INTO operation_templates (operation_type_id, name, description, unit, crop_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		t.OperationTypeID, t.Name, t.Description, t.Unit, t.CropID,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *OperationTemplateRepo) Update(id int64, t *domain.OperationTemplate) error {
	return r.db.QueryRow(
		`UPDATE operation_templates
		 SET operation_type_id=$1, name=$2, description=$3, unit=$4, crop_id=$6, updated_at=NOW()
		 WHERE id=$5 AND deleted_at IS NULL
		 RETURNING updated_at`,
		t.OperationTypeID, t.Name, t.Description, t.Unit, id, t.CropID,
	).Scan(&t.UpdatedAt)
}

func (r *OperationTemplateRepo) Delete(id int64) error {
	_, err := r.db.Exec(
		`UPDATE operation_templates SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *OperationTemplateRepo) SetResources(templateID int64, resources []domain.TemplateResource) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM template_resources WHERE template_id = $1`, templateID); err != nil {
		return err
	}

	for i := range resources {
		err := tx.QueryRow(
			`INSERT INTO template_resources (template_id, resource_id, quantity_per_unit, notes)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, created_at, updated_at`,
			templateID, resources[i].ResourceID, resources[i].QuantityPerUnit, resources[i].Notes,
		).Scan(&resources[i].ID, &resources[i].CreatedAt, &resources[i].UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *OperationTemplateRepo) SetMachineTypes(templateID int64, machineTypes []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM template_machine_types WHERE template_id = $1`, templateID); err != nil {
		return err
	}

	if len(machineTypes) > 0 {
		stmt, err := tx.Prepare(
			`INSERT INTO template_machine_types (template_id, machine_type) VALUES ($1, $2)`)
		if err != nil {
			return err
		}
		defer func() { _ = stmt.Close() }()

		for _, mt := range machineTypes {
			if _, err := stmt.Exec(templateID, mt); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *OperationTemplateRepo) SetImplementTypes(templateID int64, implementTypes []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM template_implement_types WHERE template_id = $1`, templateID); err != nil {
		return err
	}

	if len(implementTypes) > 0 {
		stmt, err := tx.Prepare(
			`INSERT INTO template_implement_types (template_id, implement_type) VALUES ($1, $2)`)
		if err != nil {
			return err
		}
		defer func() { _ = stmt.Close() }()

		for _, it := range implementTypes {
			if _, err := stmt.Exec(templateID, it); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *OperationTemplateRepo) loadResources(templateID int64) ([]domain.TemplateResource, error) {
	rows, err := r.db.Query(
		`SELECT tr.id, tr.template_id, tr.resource_id, tr.quantity_per_unit, COALESCE(tr.notes, ''),
		        tr.created_at, tr.updated_at,
		        res.id, res.name, res.resource_type_id, res.price_per_unit, COALESCE(res.notes, ''),
		        res.created_at, res.updated_at
		 FROM template_resources tr
		 JOIN resources res ON res.id = tr.resource_id
		 WHERE tr.template_id = $1
		 ORDER BY res.name`, templateID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.TemplateResource
	for rows.Next() {
		var tr domain.TemplateResource
		var res domain.Resource
		if err := rows.Scan(
			&tr.ID, &tr.TemplateID, &tr.ResourceID, &tr.QuantityPerUnit, &tr.Notes,
			&tr.CreatedAt, &tr.UpdatedAt,
			&res.ID, &res.Name, &res.ResourceTypeID, &res.PricePerUnit, &res.Notes,
			&res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tr.Resource = &res
		result = append(result, tr)
	}
	return result, rows.Err()
}

func (r *OperationTemplateRepo) loadMachineTypes(templateID int64) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT machine_type FROM template_machine_types WHERE template_id = $1 ORDER BY machine_type`,
		templateID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []string
	for rows.Next() {
		var mt string
		if err := rows.Scan(&mt); err != nil {
			return nil, err
		}
		result = append(result, mt)
	}
	return result, rows.Err()
}

func (r *OperationTemplateRepo) loadImplementTypes(templateID int64) ([]string, error) {
	rows, err := r.db.Query(
		`SELECT implement_type FROM template_implement_types WHERE template_id = $1 ORDER BY implement_type`,
		templateID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []string
	for rows.Next() {
		var it string
		if err := rows.Scan(&it); err != nil {
			return nil, err
		}
		result = append(result, it)
	}
	return result, rows.Err()
}

// Ensure interface compliance
var _ interface {
	GetAll() ([]domain.OperationTemplate, error)
} = (*OperationTemplateRepo)(nil)
