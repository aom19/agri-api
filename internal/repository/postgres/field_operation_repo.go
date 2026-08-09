package postgres

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"database/sql"
	"fmt"
	"strings"
)

type FieldOperationRepo struct {
	db *sql.DB
}

func NewFieldOperationRepo(db *sql.DB) *FieldOperationRepo {
	return &FieldOperationRepo{db: db}
}

const fieldOperationBaseSelect = `
	SELECT
		fo.id,
		fo.field_id, f.name, f.geometry,
		fo.operation_type_id, ot.code, ot.name,
		fo.operation_template_id, t.name,
		fo.machine_id, m.name, m.asset_status,
		fo.implement_id, imp.name, imp.status,
		fo.operator_id, op.name,
		fo.planned_start_at,
		fo.planned_end_at,
		fo.area_planned_ha,
		fo.notes,
		fo.status,
		fo.check_machine_status,
		fo.check_implement_status,
		fo.check_field_area,
		fo.check_notes_confirmed,
		fo.checklist_updated_at,
		fo.created_at,
		fo.updated_at
	FROM field_operations fo
	JOIN fields f              ON f.id  = fo.field_id
	JOIN operation_types ot    ON ot.id = fo.operation_type_id
	LEFT JOIN operation_templates t ON t.id  = fo.operation_template_id
	LEFT JOIN machines m       ON m.id  = fo.machine_id
	LEFT JOIN implements imp   ON imp.id = fo.implement_id
	LEFT JOIN operators op     ON op.id = fo.operator_id
	WHERE fo.deleted_at IS NULL
`

func scanFieldOperationRow(row interface {
	Scan(dest ...interface{}) error
}) (*dto.FieldOperationResponse, error) {
	var r dto.FieldOperationResponse
	var fieldGeometry []byte
	if err := row.Scan(
		&r.ID,
		&r.FieldID, &r.FieldName, &fieldGeometry,
		&r.OperationTypeID, &r.OperationTypeCode, &r.OperationTypeName,
		&r.OperationTemplateID, &r.OperationTemplate,
		&r.MachineID, &r.MachineName, &r.MachineStatus,
		&r.ImplementID, &r.ImplementName, &r.ImplementStatus,
		&r.OperatorID, &r.OperatorName,
		&r.PlannedStartAt,
		&r.PlannedEndAt,
		&r.AreaPlannedHa,
		&r.Notes,
		&r.Status,
		&r.Checklist.MachineStatus,
		&r.Checklist.ImplementStatus,
		&r.Checklist.FieldArea,
		&r.Checklist.NotesConfirmed,
		&r.Checklist.UpdatedAt,
		&r.CreatedAt,
		&r.UpdatedAt,
	); err != nil {
		return nil, err
	}
	r.FieldGeometry = fieldGeometry
	return &r, nil
}

func (repo *FieldOperationRepo) GetAll(filter repository.FieldOperationFilter) ([]dto.FieldOperationResponse, error) {
	query := fieldOperationBaseSelect
	var args []interface{}
	i := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND fo.status = $%d", i)
		args = append(args, filter.Status)
		i++
	}
	if strings.TrimSpace(filter.FieldID) != "" {
		query += fmt.Sprintf(" AND fo.field_id = $%d", i)
		args = append(args, filter.FieldID)
		i++
	}
	if filter.OperationTypeID != "" {
		query += fmt.Sprintf(" AND fo.operation_type_id = $%d", i)
		args = append(args, filter.OperationTypeID)
		i++
	}
	if filter.MachineID != "" {
		query += fmt.Sprintf(" AND fo.machine_id = $%d", i)
		args = append(args, filter.MachineID)
		i++
	}
	if filter.OperatorID != "" {
		query += fmt.Sprintf(" AND fo.operator_id = $%d", i)
		args = append(args, filter.OperatorID)
		i++
	}
	if filter.AssignedUserID > 0 {
		query += fmt.Sprintf(`
			AND EXISTS (
				SELECT 1
				FROM users u
				JOIN operators assigned_op ON (
					assigned_op.user_id = u.id
					OR (
						assigned_op.user_id IS NULL
						AND assigned_op.email IS NOT NULL
						AND LOWER(assigned_op.email) = LOWER(u.email)
					)
				)
				WHERE u.id = $%d
				  AND assigned_op.deleted_at IS NULL
				  AND assigned_op.id = fo.operator_id
			)`, i)
		args = append(args, filter.AssignedUserID)
		i++
	}

	query += " ORDER BY COALESCE(fo.planned_start_at, fo.created_at) DESC, fo.id DESC"

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]dto.FieldOperationResponse, 0)
	for rows.Next() {
		item, err := scanFieldOperationRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, rows.Err()
}

func (repo *FieldOperationRepo) GetByID(id int64) (*dto.FieldOperationResponse, error) {
	query := fieldOperationBaseSelect + " AND fo.id = $1"
	row := repo.db.QueryRow(query, id)
	item, err := scanFieldOperationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (repo *FieldOperationRepo) GetByIDForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error) {
	query := fieldOperationBaseSelect + `
		AND fo.id = $1
		AND EXISTS (
			SELECT 1
			FROM users u
			JOIN operators assigned_op ON (
				assigned_op.user_id = u.id
				OR (
					assigned_op.user_id IS NULL
					AND assigned_op.email IS NOT NULL
					AND LOWER(assigned_op.email) = LOWER(u.email)
				)
			)
			WHERE u.id = $2
			  AND assigned_op.deleted_at IS NULL
			  AND assigned_op.id = fo.operator_id
		)
	`
	row := repo.db.QueryRow(query, id, userID)
	item, err := scanFieldOperationRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (repo *FieldOperationRepo) Create(op *domain.FieldOperation) error {
	query := `
		INSERT INTO field_operations (
			field_id, operation_type_id, operation_template_id,
			machine_id, implement_id, operator_id,
			planned_start_at, planned_end_at, area_planned_ha,
			notes, status
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, created_at, updated_at
	`
	return repo.db.QueryRow(query,
		op.FieldID,
		op.OperationTypeID,
		op.OperationTemplateID,
		op.MachineID,
		op.ImplementID,
		op.OperatorID,
		op.PlannedStartAt,
		op.PlannedEndAt,
		op.AreaPlannedHa,
		op.Notes,
		op.Status,
	).Scan(&op.ID, &op.CreatedAt, &op.UpdatedAt)
}

func (repo *FieldOperationRepo) Update(id int64, op *domain.FieldOperation) error {
	query := `
		UPDATE field_operations SET
			field_id              = $1,
			operation_type_id     = $2,
			operation_template_id = $3,
			machine_id            = $4,
			implement_id          = $5,
			operator_id           = $6,
			planned_start_at      = $7,
			planned_end_at        = $8,
			area_planned_ha       = $9,
			notes                 = $10,
			status                = $11,
			updated_at            = NOW()
		WHERE id = $12 AND deleted_at IS NULL
	`
	_, err := repo.db.Exec(query,
		op.FieldID,
		op.OperationTypeID,
		op.OperationTemplateID,
		op.MachineID,
		op.ImplementID,
		op.OperatorID,
		op.PlannedStartAt,
		op.PlannedEndAt,
		op.AreaPlannedHa,
		op.Notes,
		op.Status,
		id,
	)
	return err
}

func (repo *FieldOperationRepo) UpdateChecklist(id int64, checklist domain.FieldOperationChecklist) error {
	query := `
		UPDATE field_operations SET
			check_machine_status   = $1,
			check_implement_status = $2,
			check_field_area       = $3,
			check_notes_confirmed  = $4,
			checklist_updated_at   = NOW(),
			updated_at             = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`
	_, err := repo.db.Exec(
		query,
		checklist.MachineStatus,
		checklist.ImplementStatus,
		checklist.FieldArea,
		checklist.NotesConfirmed,
		id,
	)
	return err
}

func (repo *FieldOperationRepo) UpdateStatus(id int64, status domain.FieldOperationStatus) error {
	query := `
		UPDATE field_operations SET
			status     = $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := repo.db.Exec(query, status, id)
	return err
}

func (repo *FieldOperationRepo) Delete(id int64) error {
	_, err := repo.db.Exec(
		`UPDATE field_operations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
