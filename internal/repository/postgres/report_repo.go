package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"agri-api/internal/domain"

	"github.com/lib/pq"
)

type ReportRepo struct {
	db *sql.DB
}

func NewReportRepo(db *sql.DB) *ReportRepo {
	return &ReportRepo{db: db}
}

// Expresii SQL reutilizate în rapoarte (alias `fo` = field_operations).
const (
	// Data de referință a unei operațiuni: începutul planificat sau, în lipsă, data creării.
	reportOperationDateExpr = "COALESCE(fo.planned_start_at, fo.created_at)"
	// Operațiune întârziată: încă activă, dar cu sfârșitul planificat depășit.
	reportOverdueExpr = "fo.status IN ('planned', 'in_progress') AND fo.planned_end_at IS NOT NULL AND fo.planned_end_at < NOW()"
	// Operațiune finalizată la timp: ultima modificare (finalizarea) nu depășește sfârșitul planificat.
	reportOnTimeExpr = "fo.status = 'completed' AND (fo.planned_end_at IS NULL OR COALESCE(fo.actual_end_at, fo.updated_at) <= fo.planned_end_at)"
	// Costul real = valoarea ieșirilor din stoc legate de operațiune.
	reportRealCostExpr = `COALESCE((
			SELECT SUM(sm.total_cost)
			FROM stock_movements sm
			WHERE sm.field_operation_id = fo.id AND sm.movement_type = 'out'
		), 0)`
	// Durata reală în minute, când există ambii timpi reali.
	reportDurationExpr = `CASE
			WHEN fo.actual_start_at IS NOT NULL AND fo.actual_end_at IS NOT NULL AND fo.actual_end_at > fo.actual_start_at
				THEN FLOOR(EXTRACT(EPOCH FROM (fo.actual_end_at - fo.actual_start_at)) / 60)::bigint
			ELSE NULL
		END`
	// Cost estimat = suprafață planificată × Σ(cantitate pe unitate × preț unitar) din șablonul operațiunii.
	reportEstimatedCostExpr = `COALESCE(fo.area_planned_ha, 0) * COALESCE((
			SELECT SUM(tr.quantity_per_unit * r.price_per_unit)
			FROM template_resources tr
			JOIN resources r ON r.id = tr.resource_id
			WHERE tr.template_id = fo.operation_template_id
		), 0)`
	// Întârziere în minute: față de sfârșitul planificat, la finalizare sau până acum.
	reportDelayMinutesExpr = `CASE
			WHEN fo.status = 'completed' AND fo.planned_end_at IS NOT NULL AND COALESCE(fo.actual_end_at, fo.updated_at) > fo.planned_end_at
				THEN FLOOR(EXTRACT(EPOCH FROM (COALESCE(fo.actual_end_at, fo.updated_at) - fo.planned_end_at)) / 60)::bigint
			WHEN fo.status IN ('planned', 'in_progress') AND fo.planned_end_at IS NOT NULL AND fo.planned_end_at < NOW()
				THEN FLOOR(EXTRACT(EPOCH FROM (NOW() - fo.planned_end_at)) / 60)::bigint
			ELSE 0
		END`
)

// operationsWhere construiește condițiile comune pentru field_operations (alias fo),
// cu parametri numerotați începând de la `start`. Rezultatul nu începe cu AND.
func operationsWhere(filter domain.ReportFilter, start int) (string, []interface{}) {
	var sb strings.Builder
	args := make([]interface{}, 0, 6)
	idx := start

	sb.WriteString("fo.deleted_at IS NULL")
	fmt.Fprintf(&sb, " AND %s >= $%d AND %s < $%d", reportOperationDateExpr, idx, reportOperationDateExpr, idx+1)
	args = append(args, filter.From, filter.To)
	idx += 2

	if filter.FieldID != "" {
		fmt.Fprintf(&sb, " AND fo.field_id = $%d", idx)
		args = append(args, filter.FieldID)
		idx++
	}
	if filter.OperationTypeID > 0 {
		fmt.Fprintf(&sb, " AND fo.operation_type_id = $%d", idx)
		args = append(args, filter.OperationTypeID)
		idx++
	}
	if filter.MachineID > 0 {
		fmt.Fprintf(&sb, " AND fo.machine_id = $%d", idx)
		args = append(args, filter.MachineID)
		idx++
	}
	if filter.OperatorID > 0 {
		fmt.Fprintf(&sb, " AND fo.operator_id = $%d", idx)
		args = append(args, filter.OperatorID)
	}

	return sb.String(), args
}

func (repo *ReportRepo) GetOperationsMetrics(filter domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
	where, args := operationsWhere(filter, 1)
	query := fmt.Sprintf(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE fo.status = 'planned'),
			COUNT(*) FILTER (WHERE fo.status = 'in_progress'),
			COUNT(*) FILTER (WHERE fo.status = 'completed'),
			COUNT(*) FILTER (WHERE fo.status = 'canceled'),
			COUNT(*) FILTER (WHERE %s),
			COUNT(*) FILTER (WHERE %s),
			COALESCE(SUM(fo.area_planned_ha), 0),
			COALESCE(SUM(fo.area_planned_ha) FILTER (WHERE fo.status = 'completed'), 0),
			COALESCE(SUM(%s), 0),
			COUNT(DISTINCT fo.field_id),
			COUNT(DISTINCT fo.machine_id),
			COUNT(DISTINCT fo.operator_id),
			COUNT(*) FILTER (WHERE fo.status = 'completed' AND fo.actual_start_at IS NOT NULL AND fo.actual_end_at IS NOT NULL),
			COALESCE(SUM(%s), 0),
			COALESCE(SUM(fo.area_completed_ha) FILTER (WHERE fo.status = 'completed'), 0),
			COALESCE(SUM(fo.fuel_used_l), 0),
			COALESCE(SUM(fo.machine_hours), 0),
			COALESCE(SUM(%s), 0)
		FROM field_operations fo
		WHERE %s`, reportOverdueExpr, reportOnTimeExpr, reportEstimatedCostExpr, reportDurationExpr, reportRealCostExpr, where)

	metrics := &domain.ReportOperationsMetrics{}
	err := repo.db.QueryRow(query, args...).Scan(
		&metrics.OperationsTotal,
		&metrics.OperationsPlanned,
		&metrics.OperationsInProgress,
		&metrics.OperationsCompleted,
		&metrics.OperationsCanceled,
		&metrics.OverdueOperations,
		&metrics.OnTimeCompleted,
		&metrics.PlannedAreaHa,
		&metrics.CompletedAreaHa,
		&metrics.EstimatedCost,
		&metrics.FieldsWorked,
		&metrics.MachinesUsed,
		&metrics.OperatorsUsed,
		&metrics.CompletedWithActuals,
		&metrics.ActualDurationMinutes,
		&metrics.RealizedAreaHa,
		&metrics.FuelUsedL,
		&metrics.MachineHours,
		&metrics.RealCost,
	)
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (repo *ReportRepo) GetInventorySnapshot() (*domain.ReportInventorySnapshot, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM fields WHERE deleted_at IS NULL),
			(SELECT COALESCE(SUM(area_ha), 0) FROM fields WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL AND asset_status = 'active'),
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL AND asset_status = 'maintenance')
				+ (SELECT COUNT(*) FROM implements WHERE deleted_at IS NULL AND status = 'maintenance'),
			(SELECT COUNT(*) FROM operators WHERE deleted_at IS NULL),
			(SELECT COUNT(*) FROM operators WHERE deleted_at IS NULL AND status = 'active'),
			(SELECT COUNT(*) FROM stocks),
			(SELECT COUNT(*) FROM stocks WHERE minimum_quantity > 0 AND quantity <= minimum_quantity),
			(SELECT COALESCE(SUM(s.quantity * r.price_per_unit), 0) FROM stocks s JOIN resources r ON r.id = s.resource_id)
	`

	snapshot := &domain.ReportInventorySnapshot{}
	err := repo.db.QueryRow(query).Scan(
		&snapshot.TotalFields,
		&snapshot.TotalAreaHa,
		&snapshot.TotalMachines,
		&snapshot.ActiveMachines,
		&snapshot.MaintenanceAssets,
		&snapshot.TotalOperators,
		&snapshot.ActiveOperators,
		&snapshot.TotalStocks,
		&snapshot.LowStocks,
		&snapshot.StockValue,
	)
	if err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (repo *ReportRepo) GetOperationsTimeline(filter domain.ReportFilter, granularity string) ([]domain.ReportTimeBucket, error) {
	where, filterArgs := operationsWhere(filter, 2)
	args := append([]interface{}{granularity}, filterArgs...)
	query := fmt.Sprintf(`
		SELECT
			to_char(date_trunc($1, %s), 'YYYY-MM-DD'),
			COUNT(*) FILTER (WHERE fo.status = 'planned'),
			COUNT(*) FILTER (WHERE fo.status = 'in_progress'),
			COUNT(*) FILTER (WHERE fo.status = 'completed'),
			COUNT(*) FILTER (WHERE fo.status = 'canceled'),
			COALESCE(SUM(fo.area_planned_ha), 0)
		FROM field_operations fo
		WHERE %s
		GROUP BY 1
		ORDER BY 1`, reportOperationDateExpr, where)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := []domain.ReportTimeBucket{}
	for rows.Next() {
		var bucket domain.ReportTimeBucket
		if err := rows.Scan(&bucket.Bucket, &bucket.Planned, &bucket.InProgress, &bucket.Completed, &bucket.Canceled, &bucket.AreaHa); err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}
	return buckets, rows.Err()
}

func (repo *ReportRepo) GetOperationsByType(filter domain.ReportFilter) ([]domain.ReportOperationTypeStat, error) {
	where, args := operationsWhere(filter, 1)
	query := fmt.Sprintf(`
		SELECT
			ot.id,
			ot.name,
			COUNT(*),
			COUNT(*) FILTER (WHERE fo.status = 'completed'),
			COALESCE(SUM(fo.area_planned_ha), 0),
			COALESCE(SUM(%s), 0)
		FROM field_operations fo
		JOIN operation_types ot ON ot.id = fo.operation_type_id
		WHERE %s
		GROUP BY ot.id, ot.name
		ORDER BY COUNT(*) DESC, ot.name`, reportEstimatedCostExpr, where)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []domain.ReportOperationTypeStat{}
	for rows.Next() {
		var stat domain.ReportOperationTypeStat
		if err := rows.Scan(&stat.OperationTypeID, &stat.OperationTypeName, &stat.Total, &stat.Completed, &stat.AreaHa, &stat.EstimatedCost); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

func (repo *ReportRepo) GetOperationRows(filter domain.ReportFilter, limit int) ([]domain.ReportOperationRow, error) {
	where, args := operationsWhere(filter, 1)
	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT
			fo.id,
			fo.field_id,
			f.name,
			ot.name,
			tpl.name,
			m.name,
			i.name,
			o.name,
			fo.planned_start_at,
			fo.planned_end_at,
			fo.area_planned_ha,
			fo.status,
			%s,
			%s,
			fo.actual_start_at,
			fo.actual_end_at,
			%s,
			fo.area_completed_ha,
			fo.fuel_used_l,
			fo.machine_hours,
			%s
		FROM field_operations fo
		JOIN fields f ON f.id = fo.field_id
		JOIN operation_types ot ON ot.id = fo.operation_type_id
		LEFT JOIN operation_templates tpl ON tpl.id = fo.operation_template_id
		LEFT JOIN machines m ON m.id = fo.machine_id
		LEFT JOIN implements i ON i.id = fo.implement_id
		LEFT JOIN operators o ON o.id = fo.operator_id
		WHERE %s
		ORDER BY %s DESC, fo.id DESC
		LIMIT $%d`, reportDelayMinutesExpr, reportEstimatedCostExpr, reportDurationExpr, reportRealCostExpr, where, reportOperationDateExpr, len(args))

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportOperationRow{}
	for rows.Next() {
		var (
			row           domain.ReportOperationRow
			templateName  sql.NullString
			machineName   sql.NullString
			implementName sql.NullString
			operatorName  sql.NullString
			plannedStart  sql.NullTime
			plannedEnd    sql.NullTime
			areaPlanned   sql.NullFloat64
			actualStart   sql.NullTime
			actualEnd     sql.NullTime
			duration      sql.NullInt64
			areaCompleted sql.NullFloat64
			fuelUsed      sql.NullFloat64
			machineHours  sql.NullFloat64
		)
		if err := rows.Scan(
			&row.ID, &row.FieldID, &row.FieldName, &row.OperationTypeName,
			&templateName, &machineName, &implementName, &operatorName,
			&plannedStart, &plannedEnd, &areaPlanned, &row.Status,
			&row.DelayMinutes, &row.EstimatedCost,
			&actualStart, &actualEnd, &duration, &areaCompleted, &fuelUsed, &machineHours, &row.RealCost,
		); err != nil {
			return nil, err
		}
		if actualStart.Valid {
			value := actualStart.Time
			row.ActualStartAt = &value
		}
		if actualEnd.Valid {
			value := actualEnd.Time
			row.ActualEndAt = &value
		}
		if duration.Valid {
			value := duration.Int64
			row.ActualDurationMin = &value
		}
		if areaCompleted.Valid {
			value := areaCompleted.Float64
			row.AreaCompletedHa = &value
		}
		if fuelUsed.Valid {
			value := fuelUsed.Float64
			row.FuelUsedL = &value
		}
		if machineHours.Valid {
			value := machineHours.Float64
			row.MachineHours = &value
		}
		row.TemplateName = nullStringPtr(templateName)
		row.MachineName = nullStringPtr(machineName)
		row.ImplementName = nullStringPtr(implementName)
		row.OperatorName = nullStringPtr(operatorName)
		if plannedStart.Valid {
			value := plannedStart.Time
			row.PlannedStartAt = &value
		}
		if plannedEnd.Valid {
			value := plannedEnd.Time
			row.PlannedEndAt = &value
		}
		if areaPlanned.Valid {
			value := areaPlanned.Float64
			row.AreaPlannedHa = &value
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetFieldRows(filter domain.ReportFilter) ([]domain.ReportFieldRow, error) {
	where, args := operationsWhere(filter, 1)
	fieldClause := ""
	if filter.FieldID != "" {
		args = append(args, filter.FieldID)
		fieldClause = fmt.Sprintf(" AND f.id = $%d", len(args))
	}
	query := fmt.Sprintf(`
		SELECT
			f.id,
			f.name,
			f.cadastral_number,
			f.area_ha,
			f.geometry,
			COUNT(fo.id),
			COUNT(fo.id) FILTER (WHERE fo.status = 'completed'),
			COALESCE(SUM(fo.area_planned_ha), 0),
			COALESCE(SUM(%s), 0),
			COALESCE(SUM(%s), 0),
			COALESCE(SUM(fo.area_completed_ha) FILTER (WHERE fo.status = 'completed'), 0),
			COALESCE(SUM(fo.fuel_used_l), 0),
			MAX(%s)
		FROM fields f
		LEFT JOIN field_operations fo ON fo.field_id = f.id AND %s
		WHERE f.deleted_at IS NULL%s
		GROUP BY f.id, f.name, f.cadastral_number, f.area_ha, f.geometry
		ORDER BY COUNT(fo.id) DESC, f.name`, reportEstimatedCostExpr, reportRealCostExpr, reportOperationDateExpr, where, fieldClause)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportFieldRow{}
	for rows.Next() {
		var (
			row       domain.ReportFieldRow
			cadastral sql.NullString
			area      sql.NullFloat64
			geometry  []byte
			lastAt    sql.NullTime
		)
		if err := rows.Scan(
			&row.ID, &row.Name, &cadastral, &area, &geometry,
			&row.OperationsCount, &row.CompletedCount, &row.PlannedAreaHa, &row.EstimatedCost,
			&row.RealCost, &row.RealizedAreaHa, &row.FuelUsedL, &lastAt,
		); err != nil {
			return nil, err
		}
		row.CadastralNumber = nullStringPtr(cadastral)
		if area.Valid {
			value := area.Float64
			row.AreaHa = &value
		}
		if len(geometry) > 0 && json.Valid(geometry) {
			row.Geometry = json.RawMessage(geometry)
		}
		if lastAt.Valid {
			value := lastAt.Time
			row.LastOperationAt = &value
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetMachineStatusCounts() ([]domain.ReportNamedCount, error) {
	return repo.countBy("SELECT asset_status, COUNT(*) FROM machines WHERE deleted_at IS NULL GROUP BY 1 ORDER BY 1")
}

func (repo *ReportRepo) GetImplementStatusCounts() ([]domain.ReportNamedCount, error) {
	return repo.countBy("SELECT status, COUNT(*) FROM implements WHERE deleted_at IS NULL GROUP BY 1 ORDER BY 1")
}

func (repo *ReportRepo) GetMachinesByType() ([]domain.ReportNamedCount, error) {
	return repo.countBy("SELECT type, COUNT(*) FROM machines WHERE deleted_at IS NULL GROUP BY 1 ORDER BY 2 DESC, 1")
}

func (repo *ReportRepo) GetMachinesByFuel() ([]domain.ReportNamedCount, error) {
	return repo.countBy("SELECT COALESCE(NULLIF(fuel_type, ''), 'unknown'), COUNT(*) FROM machines WHERE deleted_at IS NULL GROUP BY 1 ORDER BY 2 DESC, 1")
}

func (repo *ReportRepo) GetMachinesByYear() ([]domain.ReportNamedCount, error) {
	return repo.countBy("SELECT COALESCE(year::text, 'unknown'), COUNT(*) FROM machines WHERE deleted_at IS NULL GROUP BY year ORDER BY year NULLS LAST")
}

func (repo *ReportRepo) countBy(query string) ([]domain.ReportNamedCount, error) {
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := []domain.ReportNamedCount{}
	for rows.Next() {
		var item domain.ReportNamedCount
		if err := rows.Scan(&item.Key, &item.Count); err != nil {
			return nil, err
		}
		counts = append(counts, item)
	}
	return counts, rows.Err()
}

func (repo *ReportRepo) GetMachineRows(filter domain.ReportFilter) ([]domain.ReportMachineRow, error) {
	where, args := operationsWhere(filter, 1)
	machineClause := ""
	if filter.MachineID > 0 {
		args = append(args, filter.MachineID)
		machineClause = fmt.Sprintf(" AND m.id = $%d", len(args))
	}
	query := fmt.Sprintf(`
		SELECT
			m.id,
			m.name,
			m.code,
			m.type,
			m.asset_status,
			NULLIF(m.fuel_type, ''),
			m.year,
			m.operating_hours,
			(SELECT COUNT(*) FROM field_operations fo WHERE fo.machine_id = m.id AND %[1]s),
			(SELECT COALESCE(SUM(fo.area_planned_ha), 0) FROM field_operations fo WHERE fo.machine_id = m.id AND %[1]s),
			(SELECT COUNT(*) FROM assignments a WHERE a.machine_id = m.id AND a.status = 'active' AND a.deleted_at IS NULL),
			(SELECT COALESCE(SUM(fo.fuel_used_l), 0) FROM field_operations fo WHERE fo.machine_id = m.id AND %[1]s),
			(SELECT COALESCE(SUM(fo.machine_hours), 0) FROM field_operations fo WHERE fo.machine_id = m.id AND %[1]s)
		FROM machines m
		WHERE m.deleted_at IS NULL%[2]s
		ORDER BY 9 DESC, m.name`, where, machineClause)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportMachineRow{}
	for rows.Next() {
		var (
			row      domain.ReportMachineRow
			fuelType sql.NullString
			year     sql.NullInt64
			hours    sql.NullFloat64
		)
		if err := rows.Scan(
			&row.ID, &row.Name, &row.Code, &row.Type, &row.Status, &fuelType, &year, &hours,
			&row.OperationsCount, &row.PlannedAreaHa, &row.ActiveAssignments, &row.FuelUsedL, &row.HoursInPeriod,
		); err != nil {
			return nil, err
		}
		row.FuelType = nullStringPtr(fuelType)
		if year.Valid {
			value := int(year.Int64)
			row.Year = &value
		}
		if hours.Valid {
			value := hours.Float64
			row.OperatingHours = &value
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetImplementRows(filter domain.ReportFilter) ([]domain.ReportImplementRow, error) {
	where, args := operationsWhere(filter, 1)
	query := fmt.Sprintf(`
		SELECT
			i.id,
			i.name,
			i.code,
			i.type,
			i.status,
			i.working_width,
			(SELECT COUNT(*) FROM field_operations fo WHERE fo.implement_id = i.id AND %s)
		FROM implements i
		WHERE i.deleted_at IS NULL
		ORDER BY 7 DESC, i.name`, where)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportImplementRow{}
	for rows.Next() {
		var (
			row   domain.ReportImplementRow
			width sql.NullFloat64
		)
		if err := rows.Scan(&row.ID, &row.Name, &row.Code, &row.Type, &row.Status, &width, &row.OperationsCount); err != nil {
			return nil, err
		}
		if width.Valid {
			value := width.Float64
			row.WorkingWidth = &value
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetOperatorRows(filter domain.ReportFilter) ([]domain.ReportOperatorRow, error) {
	where, args := operationsWhere(filter, 1)
	operatorClause := ""
	if filter.OperatorID > 0 {
		args = append(args, filter.OperatorID)
		operatorClause = fmt.Sprintf(" AND o.id = $%d", len(args))
	}
	query := fmt.Sprintf(`
		SELECT
			o.id,
			o.name,
			o.status,
			o.allowed_machine_types,
			COUNT(fo.id),
			COUNT(fo.id) FILTER (WHERE fo.status = 'completed'),
			COUNT(fo.id) FILTER (WHERE fo.status = 'in_progress'),
			COUNT(fo.id) FILTER (WHERE fo.status = 'planned'),
			COUNT(fo.id) FILTER (WHERE %s),
			COUNT(fo.id) FILTER (WHERE %s),
			COALESCE(SUM(fo.area_planned_ha), 0),
			(SELECT COUNT(*) FROM assignments a WHERE a.operator_id = o.id AND a.status = 'active' AND a.deleted_at IS NULL)
		FROM operators o
		LEFT JOIN field_operations fo ON fo.operator_id = o.id AND %s
		WHERE o.deleted_at IS NULL%s
		GROUP BY o.id, o.name, o.status, o.allowed_machine_types
		ORDER BY COUNT(fo.id) DESC, o.name`, reportOnTimeExpr, reportOverdueExpr, where, operatorClause)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportOperatorRow{}
	for rows.Next() {
		var (
			row          domain.ReportOperatorRow
			allowedTypes pq.StringArray
		)
		if err := rows.Scan(
			&row.ID, &row.Name, &row.Status, &allowedTypes,
			&row.OperationsCount, &row.CompletedCount, &row.InProgressCount, &row.PlannedCount,
			&row.OnTimeCount, &row.OverdueCount, &row.PlannedAreaHa, &row.ActiveAssignments,
		); err != nil {
			return nil, err
		}
		row.AllowedMachineTypes = []string(allowedTypes)
		if row.AllowedMachineTypes == nil {
			row.AllowedMachineTypes = []string{}
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetStockRows() ([]domain.ReportStockRow, error) {
	query := `
		SELECT
			s.id,
			r.name,
			rt.category,
			rt.default_unit,
			s.quantity,
			s.minimum_quantity,
			r.price_per_unit,
			s.quantity * r.price_per_unit,
			(s.minimum_quantity > 0 AND s.quantity <= s.minimum_quantity)
		FROM stocks s
		JOIN resources r ON r.id = s.resource_id
		JOIN resource_types rt ON rt.id = r.resource_type_id
		ORDER BY (s.minimum_quantity > 0 AND s.quantity <= s.minimum_quantity) DESC, r.name`

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportStockRow{}
	for rows.Next() {
		var row domain.ReportStockRow
		if err := rows.Scan(
			&row.ID, &row.ResourceName, &row.Category, &row.Unit,
			&row.Quantity, &row.MinimumQuantity, &row.PricePerUnit, &row.Value, &row.BelowMinimum,
		); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetEstimatedConsumption(filter domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
	where, args := operationsWhere(filter, 1)
	query := fmt.Sprintf(`
		SELECT
			r.id,
			r.name,
			rt.category,
			rt.default_unit,
			COALESCE(SUM(tr.quantity_per_unit * COALESCE(fo.area_planned_ha, 0)), 0),
			COALESCE(SUM(tr.quantity_per_unit * COALESCE(fo.area_planned_ha, 0) * r.price_per_unit), 0),
			(SELECT SUM(s.quantity) FROM stocks s WHERE s.resource_id = r.id)
		FROM field_operations fo
		JOIN template_resources tr ON tr.template_id = fo.operation_template_id
		JOIN resources r ON r.id = tr.resource_id
		JOIN resource_types rt ON rt.id = r.resource_type_id
		WHERE %s AND fo.status <> 'canceled'
		GROUP BY r.id, r.name, rt.category, rt.default_unit
		ORDER BY 6 DESC, r.name`, where)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportResourceConsumption{}
	for rows.Next() {
		var (
			row   domain.ReportResourceConsumption
			stock sql.NullFloat64
		)
		if err := rows.Scan(&row.ResourceID, &row.ResourceName, &row.Category, &row.Unit, &row.Quantity, &row.Cost, &stock); err != nil {
			return nil, err
		}
		if stock.Valid {
			value := stock.Float64
			row.StockQuantity = &value
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid || value.String == "" {
		return nil
	}
	result := value.String
	return &result
}

// GetRealConsumption returnează ieșirile efective din stoc din interval, pe resursă.
// Când există filtre de entitate, sunt luate în calcul doar ieșirile legate de operațiunile filtrate.
func (repo *ReportRepo) GetRealConsumption(filter domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
	args := []interface{}{filter.From, filter.To}
	entityClause := ""
	if filter.FieldID != "" || filter.OperationTypeID > 0 || filter.MachineID > 0 || filter.OperatorID > 0 {
		where, whereArgs := operationsWhere(filter, 3)
		args = append(args, whereArgs...)
		entityClause = fmt.Sprintf(" AND sm.field_operation_id IN (SELECT fo.id FROM field_operations fo WHERE %s)", where)
	}
	query := fmt.Sprintf(`
		SELECT
			r.id,
			r.name,
			rt.category,
			rt.default_unit,
			COALESCE(SUM(-sm.quantity_delta), 0),
			COALESCE(SUM(sm.total_cost), 0),
			(SELECT SUM(s.quantity) FROM stocks s WHERE s.resource_id = r.id)
		FROM stock_movements sm
		JOIN resources r ON r.id = sm.resource_id
		JOIN resource_types rt ON rt.id = r.resource_type_id
		WHERE sm.movement_type = 'out' AND sm.created_at >= $1 AND sm.created_at < $2%s
		GROUP BY r.id, r.name, rt.category, rt.default_unit
		ORDER BY 6 DESC, r.name`, entityClause)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportResourceConsumption{}
	for rows.Next() {
		var (
			row   domain.ReportResourceConsumption
			stock sql.NullFloat64
		)
		if err := rows.Scan(&row.ResourceID, &row.ResourceName, &row.Category, &row.Unit, &row.Quantity, &row.Cost, &stock); err != nil {
			return nil, err
		}
		if stock.Valid {
			value := stock.Float64
			row.StockQuantity = &value
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (repo *ReportRepo) GetMovementTotals(filter domain.ReportFilter) (*domain.ReportMovementTotals, error) {
	totals := &domain.ReportMovementTotals{}
	err := repo.db.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE movement_type = 'in'),
			COALESCE(SUM(total_cost) FILTER (WHERE movement_type = 'in'), 0),
			COUNT(*) FILTER (WHERE movement_type = 'out'),
			COALESCE(SUM(total_cost) FILTER (WHERE movement_type = 'out'), 0)
		FROM stock_movements
		WHERE created_at >= $1 AND created_at < $2`, filter.From, filter.To,
	).Scan(&totals.InCount, &totals.InValue, &totals.OutCount, &totals.OutValue)
	if err != nil {
		return nil, err
	}
	return totals, nil
}

// GetFieldCropRows returnează culturile pe terenuri dintr-un sezon, cu numărul de operațiuni
// și costurile (estimat și real) ale operațiunilor din intervalul sezonului pe fiecare teren.
func (repo *ReportRepo) GetFieldCropRows(seasonID int64, fieldID string) ([]domain.ReportFieldCropRow, error) {
	args := []interface{}{seasonID}
	fieldClause := ""
	if fieldID != "" {
		args = append(args, fieldID)
		fieldClause = fmt.Sprintf(" AND fc.field_id = $%d", len(args))
	}
	// Operațiunile sunt legate explicit de cultura pe teren (field_operations.field_crop_id).
	seasonOps := `FROM field_operations fo
				WHERE fo.field_crop_id = fc.id AND fo.deleted_at IS NULL`
	query := fmt.Sprintf(`
		SELECT
			fc.id, fc.field_id, f.name, f.area_ha,
			fc.season_id, s.name, to_char(s.start_date, 'YYYY-MM-DD'), to_char(s.end_date, 'YYYY-MM-DD'),
			fc.crop_id, c.name, c.yield_unit,
			fc.planted_area_ha,
			to_char(fc.planted_at, 'YYYY-MM-DD'),
			to_char(fc.harvested_at, 'YYYY-MM-DD'),
			fc.production_total, fc.expected_yield_per_ha, fc.notes,
			fc.harvest_recorded_quantity, fc.harvest_recorded_at,
			fc.created_at, fc.updated_at,
			(SELECT COUNT(*) %[1]s),
			(SELECT COALESCE(SUM(%[2]s), 0) %[1]s),
			(SELECT COALESCE(SUM(%[3]s), 0) %[1]s)
		FROM field_crops fc
		JOIN fields f ON f.id = fc.field_id
		JOIN seasons s ON s.id = fc.season_id
		JOIN crops c ON c.id = fc.crop_id
		WHERE fc.season_id = $1 AND f.deleted_at IS NULL%[4]s
		ORDER BY c.name, f.name`, seasonOps, reportEstimatedCostExpr, reportRealCostExpr, fieldClause)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.ReportFieldCropRow{}
	for rows.Next() {
		var (
			row         domain.ReportFieldCropRow
			fieldArea   sql.NullFloat64
			plantedArea sql.NullFloat64
			plantedAt   sql.NullString
			harvestedAt sql.NullString
			production  sql.NullFloat64
			expected    sql.NullFloat64
			recordedQty sql.NullFloat64
			recordedAt  sql.NullTime
		)
		if err := rows.Scan(
			&row.ID, &row.FieldID, &row.FieldName, &fieldArea,
			&row.SeasonID, &row.SeasonName, &row.SeasonStart, &row.SeasonEnd,
			&row.CropID, &row.CropName, &row.YieldUnit,
			&plantedArea, &plantedAt, &harvestedAt, &production, &expected, &row.Notes,
			&recordedQty, &recordedAt, &row.CreatedAt, &row.UpdatedAt,
			&row.OperationsCount, &row.EstimatedCost, &row.RealCost,
		); err != nil {
			return nil, err
		}
		row.HarvestRecordedQty = nullFloatPtr(recordedQty)
		if recordedAt.Valid {
			value := recordedAt.Time
			row.HarvestRecordedAt = &value
		}
		row.FieldAreaHa = nullFloatPtr(fieldArea)
		row.PlantedAreaHa = nullFloatPtr(plantedArea)
		row.PlantedAt = nullStringPtr(plantedAt)
		row.HarvestedAt = nullStringPtr(harvestedAt)
		row.ProductionTotal = nullFloatPtr(production)
		row.ExpectedYieldPerHa = nullFloatPtr(expected)
		row.ComputeYield()
		if row.PlantedAreaHa != nil && *row.PlantedAreaHa > 0 {
			cost := row.RealCost
			if cost <= 0 {
				cost = row.EstimatedCost
			}
			perHa := cost / *row.PlantedAreaHa
			row.CostPerHa = &perHa
		}
		items = append(items, row)
	}
	return items, rows.Err()
}
