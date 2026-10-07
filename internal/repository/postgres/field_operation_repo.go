package postgres

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type FieldOperationRepo struct {
	db *sql.DB
}

func NewFieldOperationRepo(db *sql.DB) *FieldOperationRepo {
	return &FieldOperationRepo{db: db}
}

// fuelUsedExpr este combustibilul consumat de o operațiune (alias `fo`): ieșirile din stoc pe
// resurse de combustibil legate de ea. Este singura sursă pentru combustibil, folosită și de
// rapoarte, deci cifrele raportate sunt aceleași cu cele scăzute din stoc. NULL când nu există.
const fuelUsedExpr = `(
		SELECT SUM(-fsm.quantity_delta)
		FROM stock_movements fsm
		JOIN resources fr ON fr.id = fsm.resource_id
		JOIN resource_types frt ON frt.id = fr.resource_type_id
		WHERE fsm.field_operation_id = fo.id AND fsm.movement_type = 'out' AND frt.category = 'fuel'
	)`

const fieldOperationBaseSelect = `
	SELECT
		fo.id,
		fo.field_id, f.name, f.geometry,
		fo.operation_type_id, ot.code, ot.name,
		fo.operation_template_id, t.name,
		fo.machine_id, m.name, m.asset_status,
		fo.implement_id, imp.name, imp.status,
		fo.operator_id, ` + operatorNameExpr + `,
		fo.planned_start_at,
		fo.planned_end_at,
		fo.area_planned_ha,
		fo.notes,
		fo.status,
		fo.field_crop_id, cr.name, se.name,
		fo.actual_start_at,
		fo.actual_end_at,
		fo.area_completed_ha,
		` + fuelUsedExpr + `,
		fo.machine_hours,
		fo.completion_notes,
		fo.created_at,
		fo.updated_at
	FROM field_operations fo
	JOIN fields f              ON f.id  = fo.field_id
	JOIN operation_types ot    ON ot.id = fo.operation_type_id
	LEFT JOIN operation_templates t ON t.id  = fo.operation_template_id
	LEFT JOIN machines m       ON m.id  = fo.machine_id
	LEFT JOIN implements imp   ON imp.id = fo.implement_id
	` + operatorJoin + `
	LEFT JOIN field_crops fc   ON fc.id = fo.field_crop_id
	LEFT JOIN crops cr         ON cr.id = fc.crop_id
	LEFT JOIN seasons se       ON se.id = fc.season_id
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
		&r.FieldCropID, &r.CropName, &r.SeasonName,
		&r.ActualStartAt,
		&r.ActualEndAt,
		&r.AreaCompletedHa,
		&r.FuelUsedL,
		&r.MachineHours,
		&r.CompletionNotes,
		&r.CreatedAt,
		&r.UpdatedAt,
	); err != nil {
		return nil, err
	}
	r.FieldGeometry = fieldGeometry
	if r.ActualStartAt != nil && r.ActualEndAt != nil && r.ActualEndAt.After(*r.ActualStartAt) {
		minutes := int64(r.ActualEndAt.Sub(*r.ActualStartAt).Minutes())
		r.ActualDurationMin = &minutes
	}
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
		query += fmt.Sprintf(" AND fo.operator_id = $%d", i)
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
	query := fieldOperationBaseSelect + " AND fo.id = $1 AND fo.operator_id = $2"
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
			notes, status, field_crop_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11, COALESCE($12, (
				SELECT fc.id
				FROM field_crops fc
				JOIN seasons s ON s.id = fc.season_id
				WHERE fc.field_id = $1
				  AND COALESCE($7, NOW()) >= s.start_date
				  AND COALESCE($7, NOW()) < s.end_date + INTERVAL '1 day'
				ORDER BY s.is_active DESC, s.start_date DESC
				LIMIT 1
			)))
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
		op.FieldCropID,
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
			status                = $11::text,
			field_crop_id         = COALESCE($13, (
				SELECT fc.id
				FROM field_crops fc
				JOIN seasons s ON s.id = fc.season_id
				WHERE fc.field_id = $1
				  AND COALESCE($7, NOW()) >= s.start_date
				  AND COALESCE($7, NOW()) < s.end_date + INTERVAL '1 day'
				ORDER BY s.is_active DESC, s.start_date DESC
				LIMIT 1
			)),
			actual_start_at       = CASE
				WHEN $11::text IN ('in_progress', 'completed') AND actual_start_at IS NULL THEN COALESCE(planned_start_at, NOW())
				ELSE actual_start_at
			END,
			actual_end_at         = CASE
				WHEN $11::text = 'completed' AND actual_end_at IS NULL THEN NOW()
				WHEN $11::text <> 'completed' THEN NULL
				ELSE actual_end_at
			END,
			overdue_notified_at   = CASE
				WHEN planned_end_at IS DISTINCT FROM $8 THEN NULL
				ELSE overdue_notified_at
			END,
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
		op.FieldCropID,
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

// GetOverdueInProgress returnează operațiunile în lucru care au depășit sfârșitul planificat
// și nu au fost încă notificate. Operatorul este chiar utilizatorul care primește notificarea.
func (repo *FieldOperationRepo) GetOverdueInProgress(now time.Time) ([]dto.OverdueFieldOperation, error) {
	query := `
		SELECT
			fo.id,
			f.name,
			ot.name,
			fo.operator_id,
			` + operatorNameExpr + `,
			CASE WHEN ou.deleted_at IS NULL THEN ou.id END,
			fo.planned_start_at,
			fo.planned_end_at
		FROM field_operations fo
		JOIN fields f           ON f.id  = fo.field_id
		JOIN operation_types ot ON ot.id = fo.operation_type_id
		` + operatorJoin + `
		WHERE fo.deleted_at IS NULL
		  AND fo.status = $1
		  AND fo.planned_end_at IS NOT NULL
		  AND fo.planned_end_at < $2
		  AND fo.overdue_notified_at IS NULL
		ORDER BY fo.planned_end_at ASC, fo.id ASC
	`
	rows, err := repo.db.Query(query, domain.FieldOperationStatusInProgress, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]dto.OverdueFieldOperation, 0)
	for rows.Next() {
		var item dto.OverdueFieldOperation
		if err := rows.Scan(
			&item.ID,
			&item.FieldName,
			&item.OperationTypeName,
			&item.OperatorID,
			&item.OperatorName,
			&item.OperatorUserID,
			&item.PlannedStartAt,
			&item.PlannedEndAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// MarkOverdueNotified reține momentul în care s-a trimis notificarea de depășire.
func (repo *FieldOperationRepo) MarkOverdueNotified(id int64, at time.Time) error {
	_, err := repo.db.Exec(`
		UPDATE field_operations SET
			overdue_notified_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`, at, id)
	return err
}

// MarkStarted trece operațiunea în lucru și păstrează momentul primei porniri.
func (repo *FieldOperationRepo) MarkStarted(id int64, at time.Time) error {
	query := `
		UPDATE field_operations SET
			status          = $1,
			actual_start_at = COALESCE(actual_start_at, $2),
			updated_at      = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := repo.db.Exec(query, domain.FieldOperationStatusInProgress, at, id)
	return err
}

// Complete înregistrează datele reale ale finalizării. Dacă operațiunea nu a fost pornită
// explicit, momentul pornirii devine începutul planificat (sau momentul finalizării).
func (repo *FieldOperationRepo) Complete(tx *sql.Tx, id int64, completion domain.FieldOperationCompletion, at time.Time) error {
	query := `
		UPDATE field_operations SET
			status            = $1,
			actual_start_at   = COALESCE(actual_start_at, planned_start_at, $2),
			actual_end_at     = $2,
			area_completed_ha = COALESCE($3, area_completed_ha, area_planned_ha),
			machine_hours     = COALESCE($4, machine_hours),
			completion_notes  = $5,
			updated_at        = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`
	_, err := tx.Exec(query,
		domain.FieldOperationStatusCompleted,
		at,
		completion.AreaCompletedHa,
		completion.MachineHours,
		completion.Notes,
		id,
	)
	return err
}

func (repo *FieldOperationRepo) GetAssetCompatibility(templateID int64, machineID, implementID *int64) (*domain.AssetCompatibility, error) {
	var c domain.AssetCompatibility
	err := repo.db.QueryRow(`
		SELECT
			ARRAY(SELECT machine_type FROM template_machine_types WHERE template_id = $1 ORDER BY machine_type),
			ARRAY(SELECT implement_type FROM template_implement_types WHERE template_id = $1 ORDER BY implement_type),
			COALESCE((SELECT type FROM machines WHERE id = $2), ''),
			COALESCE((SELECT type FROM implements WHERE id = $3), '')`,
		templateID, machineID, implementID,
	).Scan(pq.Array(&c.TemplateMachineTypes), pq.Array(&c.TemplateImplementTypes), &c.MachineType, &c.ImplementType)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (repo *FieldOperationRepo) GetTemplateResources(templateID int64) ([]domain.TemplateResourceUsage, error) {
	rows, err := repo.db.Query(`
		SELECT tr.resource_id, r.name, rt.category, rt.default_unit, tr.quantity_per_unit, r.price_per_unit
		FROM template_resources tr
		JOIN resources r ON r.id = tr.resource_id
		JOIN resource_types rt ON rt.id = r.resource_type_id
		WHERE tr.template_id = $1
		ORDER BY r.name`, templateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.TemplateResourceUsage{}
	for rows.Next() {
		var item domain.TemplateResourceUsage
		if err := rows.Scan(&item.ResourceID, &item.ResourceName, &item.Category, &item.Unit, &item.QuantityPerUnit, &item.PricePerUnit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *FieldOperationRepo) AddMachineHours(tx *sql.Tx, machineID int64, hours float64) error {
	_, err := tx.Exec(`
		UPDATE machines
		SET operating_hours = COALESCE(operating_hours, 0) + $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL`, hours, machineID)
	return err
}
