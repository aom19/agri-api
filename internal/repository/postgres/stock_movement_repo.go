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

const stockLockSelect = `
	SELECT s.id, s.resource_id, s.quantity, s.minimum_quantity, r.price_per_unit
	FROM stocks s
	JOIN resources r ON r.id = s.resource_id
	WHERE %s
	FOR UPDATE OF s`

func scanStockLock(row *sql.Row) (*repository.StockLock, error) {
	lock := &repository.StockLock{}
	if err := row.Scan(&lock.StockID, &lock.ResourceID, &lock.Quantity, &lock.Minimum, &lock.PriceUnit); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return lock, nil
}

func (repo *StockMovementRepo) LockStockByID(tx *sql.Tx, stockID int64) (*repository.StockLock, error) {
	return scanStockLock(tx.QueryRow(fmt.Sprintf(stockLockSelect, "s.id = $1"), stockID))
}

func (repo *StockMovementRepo) LockStockByResource(tx *sql.Tx, resourceID int64) (*repository.StockLock, error) {
	return scanStockLock(tx.QueryRow(fmt.Sprintf(stockLockSelect, "s.resource_id = $1"), resourceID))
}

func (repo *StockMovementRepo) ApplyMovement(tx *sql.Tx, movement *domain.StockMovement) error {
	if _, err := tx.Exec(`
		UPDATE stocks SET quantity = $1, updated_at = NOW() WHERE id = $2`,
		movement.ResultingQuantity, movement.StockID,
	); err != nil {
		return err
	}

	return tx.QueryRow(`
		INSERT INTO stock_movements (
			stock_id, resource_id, field_operation_id, movement_type,
			quantity_delta, resulting_quantity, unit_cost, total_cost, notes, actor_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`,
		movement.StockID,
		movement.ResourceID,
		movement.FieldOperationID,
		movement.MovementType,
		movement.QuantityDelta,
		movement.ResultingQuantity,
		movement.UnitCost,
		movement.TotalCost,
		movement.Notes,
		movement.ActorID,
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
	if filter.StockID > 0 {
		add("sm.stock_id = $%d", filter.StockID)
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
			sm.id, sm.stock_id, sm.resource_id, r.name, rt.category, rt.default_unit,
			sm.field_operation_id,
			CASE WHEN fo.id IS NULL THEN NULL ELSE CONCAT_WS(' - ', ot.name, f.name) END,
			sm.movement_type, sm.quantity_delta, sm.resulting_quantity, sm.unit_cost, sm.total_cost,
			sm.notes, sm.actor_id,
			NULLIF(BTRIM(CONCAT_WS(' ', up.first_name, up.last_name)), ''),
			sm.created_at
		FROM stock_movements sm
		JOIN resources r ON r.id = sm.resource_id
		JOIN resource_types rt ON rt.id = r.resource_type_id
		LEFT JOIN field_operations fo ON fo.id = sm.field_operation_id
		LEFT JOIN operation_types ot ON ot.id = fo.operation_type_id
		LEFT JOIN fields f ON f.id = fo.field_id
		LEFT JOIN user_profiles up ON up.user_id = sm.actor_id
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
			&item.ID, &item.StockID, &item.ResourceID, &item.ResourceName, &item.Category, &item.Unit,
			&item.FieldOperationID, &opLabel,
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
