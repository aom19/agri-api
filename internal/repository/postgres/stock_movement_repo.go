package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

type StockMovementRepo struct {
	db *sql.DB
}

func NewStockMovementRepo(db *sql.DB) *StockMovementRepo {
	return &StockMovementRepo{db: db}
}

func (repo *StockMovementRepo) LockStock(tx *sql.Tx, resourceID int64) (*repository.StockLock, error) {
	lock := &repository.StockLock{}
	err := tx.QueryRow(`
		SELECT r.id, r.name, r.quantity, r.minimum_quantity, r.price_per_unit, rt.category
		FROM resources r
		JOIN resource_types rt ON rt.id = r.resource_type_id
		WHERE r.id = $1
		FOR UPDATE OF r`, resourceID,
	).Scan(&lock.ResourceID, &lock.ResourceName, &lock.Quantity, &lock.Minimum, &lock.PriceUnit, &lock.Category)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return lock, nil
}

func (repo *StockMovementRepo) ApplyMovement(tx *sql.Tx, movement *domain.StockMovement) error {
	if _, err := tx.Exec(`
		UPDATE resources SET quantity = $1, updated_at = NOW() WHERE id = $2`,
		movement.ResultingQuantity, movement.ResourceID,
	); err != nil {
		return err
	}

	return tx.QueryRow(`
		INSERT INTO stock_movements (
			resource_id, field_operation_id, movement_type,
			quantity_delta, resulting_quantity, unit_cost, total_cost, notes, actor_id, field_crop_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`,
		movement.ResourceID,
		movement.FieldOperationID,
		movement.MovementType,
		movement.QuantityDelta,
		movement.ResultingQuantity,
		movement.UnitCost,
		movement.TotalCost,
		movement.Notes,
		movement.ActorID,
		movement.FieldCropID,
	).Scan(&movement.ID, &movement.CreatedAt)
}

func (repo *StockMovementRepo) List(filter domain.StockMovementFilter) ([]domain.StockMovement, error) {
	var where strings.Builder
	where.WriteString("1=1")
	args := []interface{}{}
	add := func(clause string, value interface{}) {
		args = append(args, value)
		fmt.Fprintf(&where, " AND "+clause, len(args))
	}
	if filter.ResourceID > 0 {
		add("sm.resource_id = $%d", filter.ResourceID)
	}
	if filter.FieldOperationID > 0 {
		add("sm.field_operation_id = $%d", filter.FieldOperationID)
	}
	if filter.From != nil {
		add("sm.created_at >= $%d", *filter.From)
	}
	if filter.To != nil {
		add("sm.created_at < $%d", *filter.To)
	}
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT
			sm.id, sm.resource_id, r.name, rt.category, rt.default_unit,
			sm.field_operation_id,
			CASE
				WHEN fo.id IS NOT NULL THEN CONCAT_WS(' - ', `+fieldOperationTypeNameExpr+`, f.name)
				WHEN hc.id IS NOT NULL THEN CONCAT_WS(' - ', 'Recoltă ' || hcr.name, hf.name)
				ELSE NULL
			END,
			sm.field_crop_id,
			sm.movement_type, sm.quantity_delta, sm.resulting_quantity, sm.unit_cost, sm.total_cost,
			sm.notes, sm.actor_id,
			NULLIF(BTRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''),
			sm.created_at
		FROM stock_movements sm
		JOIN resources r ON r.id = sm.resource_id
		JOIN resource_types rt ON rt.id = r.resource_type_id
		LEFT JOIN field_operations fo ON fo.id = sm.field_operation_id
		LEFT JOIN fields f ON f.id = fo.field_id
		LEFT JOIN user_profiles up ON up.user_id = sm.actor_id
		LEFT JOIN field_crops hc ON hc.id = sm.field_crop_id
		LEFT JOIN crops hcr ON hcr.id = hc.crop_id
		LEFT JOIN fields hf ON hf.id = hc.field_id
		WHERE %s
		ORDER BY sm.created_at DESC, sm.id DESC
		LIMIT $%d`, where.String(), len(args))

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.StockMovement{}
	for rows.Next() {
		var (
			item      domain.StockMovement
			opLabel   sql.NullString
			unitCost  sql.NullFloat64
			totalCost sql.NullFloat64
			actorName sql.NullString
		)
		if err := rows.Scan(
			&item.ID, &item.ResourceID, &item.ResourceName, &item.Category, &item.Unit,
			&item.FieldOperationID, &opLabel, &item.FieldCropID,
			&item.MovementType, &item.QuantityDelta, &item.ResultingQuantity, &unitCost, &totalCost,
			&item.Notes, &item.ActorID, &actorName, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.FieldOperationLabel = nullStringPtr(opLabel)
		item.ActorName = nullStringPtr(actorName)
		if unitCost.Valid {
			value := unitCost.Float64
			item.UnitCost = &value
		}
		if totalCost.Valid {
			value := totalCost.Float64
			item.TotalCost = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *StockMovementRepo) ListFuelStocks() ([]domain.FuelStock, error) {
	rows, err := repo.db.Query(`
		SELECT r.id, r.name, rt.default_unit, r.quantity
		FROM resources r
		JOIN resource_types rt ON rt.id = r.resource_type_id
		WHERE rt.category = 'fuel'
		ORDER BY r.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.FuelStock{}
	for rows.Next() {
		var item domain.FuelStock
		if err := rows.Scan(&item.ResourceID, &item.ResourceName, &item.Unit, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
