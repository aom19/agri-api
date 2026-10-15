package postgres

import (
	"encoding/json"
	"fmt"
	"strings"

	"agri-api/internal/dto"
	"agri-api/internal/repository"
)

// Suprapunerile de resurse între lucrări. O lucrare planificată sau în lucru ține ocupați
// operatorul, mașina și echipamentul ei pe intervalul de mai jos; două lucrări care împart o
// resursă pe intervale suprapuse sunt în conflict. Regula e într-un singur loc, folosit de lista
// de lucrări (chip-ul din tabel), de verificarea din formular și de pornire.

// busyStartExpr este începutul intervalului ocupat: pornirea reală sau, înainte de ea, cea planificată.
func busyStartExpr(alias string) string {
	return fmt.Sprintf("COALESCE(%[1]s.actual_start_at, %[1]s.planned_start_at)", alias)
}

// busyEndExpr este sfârșitul intervalului ocupat: cel planificat (o oră, dacă lipsește). O lucrare
// în lucru ține resursele cel puțin până acum, chiar dacă și-a depășit sfârșitul planificat.
func busyEndExpr(alias string) string {
	planned := fmt.Sprintf("COALESCE(%[1]s.planned_end_at, %[2]s + INTERVAL '1 hour')", alias, busyStartExpr(alias))
	return fmt.Sprintf("CASE WHEN %[1]s.status = 'in_progress' THEN GREATEST(%[2]s, NOW()) ELSE %[2]s END", alias, planned)
}

// conflictTarget sunt expresiile SQL ale lucrării pentru care se caută conflicte: coloane ale unei
// lucrări salvate sau parametri din formular.
type conflictTarget struct {
	id, operatorID, machineID, implementID, start, end string
}

// otherOperationTypeNameExpr este numele tipului pentru cealaltă lucrare (alias o2).
var otherOperationTypeNameExpr = operationTypeNameExpr(strings.ReplaceAll(fieldOperationTypeExpr, "fo.", "o2."))

// conflictsJSONExpr întoarce, ca tablou JSON, celelalte lucrări care împart o resursă cu ținta.
// Cu onlyInProgress, doar cele în lucru, indiferent de interval (ele blochează pornirea).
func conflictsJSONExpr(t conflictTarget, onlyInProgress bool) string {
	scope := "o2.status IN ('planned', 'in_progress')" +
		" AND " + busyStartExpr("o2") + " < " + t.end +
		" AND " + t.start + " < " + busyEndExpr("o2")
	if onlyInProgress {
		scope = "o2.status = 'in_progress'"
	}
	return `COALESCE((
		SELECT json_agg(json_build_object(
			'operation_id', o2.id,
			'field_name', f2.name,
			'operation_type_name', ` + otherOperationTypeNameExpr + `,
			'status', o2.status,
			'start_at', ` + busyStartExpr("o2") + `,
			'end_at', ` + busyEndExpr("o2") + `,
			'resources', to_json(array_remove(ARRAY[
				CASE WHEN o2.operator_id = ` + t.operatorID + ` THEN 'operator' END,
				CASE WHEN o2.machine_id = ` + t.machineID + ` THEN 'machine' END,
				CASE WHEN o2.implement_id = ` + t.implementID + ` THEN 'implement' END
			], NULL))
		) ORDER BY ` + busyStartExpr("o2") + `, o2.id)
		FROM field_operations o2
		JOIN fields f2 ON f2.id = o2.field_id
		WHERE o2.id <> ` + t.id + `
		  AND o2.deleted_at IS NULL
		  AND (o2.operator_id = ` + t.operatorID + ` OR o2.machine_id = ` + t.machineID + ` OR o2.implement_id = ` + t.implementID + `)
		  AND ` + scope + `
	), '[]'::json)`
}

// savedOperationConflictsExpr sunt conflictele unei lucrări din listă (alias fo); lucrările
// finalizate, anulate sau fără dată nu au.
var savedOperationConflictsExpr = `CASE WHEN fo.status IN ('planned', 'in_progress') AND ` + busyStartExpr("fo") + ` IS NOT NULL THEN ` +
	conflictsJSONExpr(conflictTarget{
		id:          "fo.id",
		operatorID:  "fo.operator_id",
		machineID:   "fo.machine_id",
		implementID: "fo.implement_id",
		start:       busyStartExpr("fo"),
		end:         busyEndExpr("fo"),
	}, false) + ` ELSE '[]'::json END`

// queryConflictTarget: lucrarea vine din parametri ($1 id exclus, $2–$4 resursele, $5–$6 intervalul).
var queryConflictTarget = conflictTarget{
	id:          "$1::bigint",
	operatorID:  "$2::bigint",
	machineID:   "$3::bigint",
	implementID: "$4::bigint",
	start:       "$5::timestamptz",
	end:         "COALESCE($6::timestamptz, $5::timestamptz + INTERVAL '1 hour')",
}

func decodeConflicts(raw []byte) ([]dto.FieldOperationConflict, error) {
	conflicts := make([]dto.FieldOperationConflict, 0)
	if len(raw) == 0 {
		return conflicts, nil
	}
	if err := json.Unmarshal(raw, &conflicts); err != nil {
		return nil, fmt.Errorf("conflicte lucrare: %w", err)
	}
	return conflicts, nil
}

func (repo *FieldOperationRepo) queryConflicts(onlyInProgress bool, q repository.FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error) {
	args := []interface{}{q.ExcludeID, q.OperatorID, q.MachineID, q.ImplementID}
	if !onlyInProgress {
		// intervalul ($5, $6) contează doar pentru suprapunere; la pornire nu apare în interogare
		args = append(args, q.StartAt, q.EndAt)
	}
	var raw []byte
	err := repo.db.QueryRow("SELECT "+conflictsJSONExpr(queryConflictTarget, onlyInProgress), args...).Scan(&raw)
	if err != nil {
		return nil, err
	}
	return decodeConflicts(raw)
}

func (repo *FieldOperationRepo) FindConflicts(q repository.FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error) {
	return repo.queryConflicts(false, q)
}

func (repo *FieldOperationRepo) InProgressConflicts(q repository.FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error) {
	return repo.queryConflicts(true, q)
}
