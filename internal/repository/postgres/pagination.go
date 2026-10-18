package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"agri-api/internal/domain"
)

// listPage execută o listă: query (fără ORDER BY) cu ordinea dată. Cu page, numără întâi toate
// rândurile care corespund filtrelor și citește doar pagina (LIMIT/OFFSET); fără page, citește
// tot, iar totalul e numărul de rânduri citite.
func listPage[T any](
	db *sql.DB,
	query string,
	args []interface{},
	orderBy string,
	page *domain.PageRequest,
	scan func(*sql.Rows) (T, error),
) ([]T, int, error) {
	total := -1
	if page != nil {
		if err := db.QueryRow("SELECT COUNT(*) FROM ("+query+") AS counted", args...).Scan(&total); err != nil {
			return nil, 0, err
		}
	}
	full := query + " ORDER BY " + orderBy
	if page != nil {
		full += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(append([]interface{}{}, args...), page.PageSize, page.Offset())
	}
	rows, err := db.Query(full, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]T, 0)
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if total < 0 {
		total = len(items)
	}
	return items, total, nil
}

// textSearchClause întoarce „ AND (col1 ILIKE $n OR col2 ILIKE $n …)” pentru textul căutat (gol =
// fără condiție). Caracterele speciale din LIKE sunt căutate ca atare.
func textSearchClause(text string, args *[]interface{}, columns ...string) string {
	text = strings.TrimSpace(text)
	if text == "" || len(columns) == 0 {
		return ""
	}
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(text)
	*args = append(*args, "%"+escaped+"%")
	placeholder := fmt.Sprintf("$%d", len(*args))
	conditions := make([]string, len(columns))
	for i, column := range columns {
		conditions[i] = fmt.Sprintf("COALESCE(%s, '') ILIKE %s", column, placeholder)
	}
	return " AND (" + strings.Join(conditions, " OR ") + ")"
}
