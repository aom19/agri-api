package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
	"encoding/json"
	"fmt"
)

type AuditRepo struct {
	db *sql.DB
}

func NewAuditRepo(db *sql.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Log(entry *domain.AuditEntry) error {
	var changesJSON []byte
	if entry.Changes != nil {
		var err error
		changesJSON, err = json.Marshal(entry.Changes)
		if err != nil {
			changesJSON = nil
		}
	}

	_, err := r.db.Exec(`
		INSERT INTO audit_log (entity_type, entity_id, action, actor_id, changes)
		VALUES ($1, $2, $3, $4, $5)`,
		entry.EntityType, entry.EntityID, entry.Action, entry.ActorID, changesJSON,
	)
	return err
}

func (r *AuditRepo) GetAll(limit int, entityType, entityID string) ([]domain.AuditEntry, error) {
	query := `
		SELECT
			al.id,
			al.entity_type,
			al.entity_id,
			COALESCE(
				m.name,
				i.name,
				f.name,
				rt.name,
				r.name,
				sr.name,
				NULLIF(BTRIM(CONCAT_WS(' - ', am.name, ao.name)), ''),
				o.name,
				opt.name,
				pl.name,
				NULLIF(BTRIM(CONCAT_WS(' ', eup.first_name, eup.last_name)), ''),
				eu.email,
				NULLIF(BTRIM(CONCAT_WS(' - ', ot.name, fof.name)), '')
			) AS entity_name,
			al.action,
			al.actor_id,
			NULLIF(BTRIM(CONCAT_WS(' ', up.first_name, up.last_name)), '') AS actor_name,
			al.changes,
			al.created_at
		FROM audit_log al
		LEFT JOIN user_profiles up ON up.user_id = al.actor_id
		LEFT JOIN machines m ON al.entity_type = 'machine' AND m.id::text = al.entity_id
		LEFT JOIN implements i ON al.entity_type = 'implement' AND i.id::text = al.entity_id
		LEFT JOIN fields f ON al.entity_type = 'field' AND f.id::text = al.entity_id
		LEFT JOIN resource_types rt ON al.entity_type = 'resource_type' AND rt.id::text = al.entity_id
		LEFT JOIN resources r ON al.entity_type = 'resource' AND r.id::text = al.entity_id
		LEFT JOIN stocks s ON al.entity_type = 'stock' AND s.id::text = al.entity_id
		LEFT JOIN resources sr ON sr.id = s.resource_id
		LEFT JOIN assignments a ON al.entity_type = 'assignment' AND a.id::text = al.entity_id
		LEFT JOIN machines am ON am.id = a.machine_id
		LEFT JOIN operators ao ON ao.id = a.operator_id
		LEFT JOIN operators o ON al.entity_type = 'operator' AND o.id::text = al.entity_id
		LEFT JOIN operation_types opt ON al.entity_type = 'operation_type' AND opt.id::text = al.entity_id
		LEFT JOIN operation_templates pl ON al.entity_type = 'operation_template' AND pl.id::text = al.entity_id
		LEFT JOIN users eu ON al.entity_type = 'user' AND eu.id::text = al.entity_id
		LEFT JOIN user_profiles eup ON eup.user_id = eu.id
		LEFT JOIN field_operations fo ON al.entity_type = 'field_operation' AND fo.id::text = al.entity_id
		LEFT JOIN operation_types ot ON ot.id = fo.operation_type_id
		LEFT JOIN fields fof ON fof.id = fo.field_id
		WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if entityType != "" {
		query += fmt.Sprintf(" AND al.entity_type = $%d", idx)
		args = append(args, entityType)
		idx++
	}
	if entityID != "" {
		query += fmt.Sprintf(" AND al.entity_id = $%d", idx)
		args = append(args, entityID)
		idx++
	}
	query += " ORDER BY al.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d", idx)
	args = append(args, limit)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		var changesRaw []byte
		var entityName sql.NullString
		var actorName sql.NullString
		if err := rows.Scan(&e.ID, &e.EntityType, &e.EntityID, &entityName, &e.Action, &e.ActorID, &actorName, &changesRaw, &e.CreatedAt); err != nil {
			return nil, err
		}
		if entityName.Valid {
			e.EntityName = &entityName.String
		}
		if actorName.Valid {
			e.ActorName = &actorName.String
		}
		if changesRaw != nil {
			_ = json.Unmarshal(changesRaw, &e.Changes)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}
