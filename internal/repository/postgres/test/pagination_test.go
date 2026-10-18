package postgres_test

import (
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
)

// Paginarea în SQL: totalul numără toate rândurile filtrate, pagina le taie în ordinea listei, iar
// căutarea tratează % și _ ca text.
func TestPagination_MachinesAndSearch(t *testing.T) {
	db := requireDB(t)
	for _, name := range []string{"Claas", "Fendt 724", "John Deere", "Fendt 939", "Case 100%", "New_Holland"} {
		insertMachine(t, db, name, "C-"+name)
	}
	repo := postgres.NewMachineRepo(db)

	first, total, err := repo.List(domain.ListFilter{}, &domain.PageRequest{Page: 1, PageSize: 4})
	if err != nil || total != 6 || len(first) != 4 || first[0].Name != "Case 100%" || first[3].Name != "Fendt 939" {
		t.Fatalf("prima pagină: %v, total %d, %+v", err, total, first)
	}
	second, total, err := repo.List(domain.ListFilter{}, &domain.PageRequest{Page: 2, PageSize: 4})
	if err != nil || total != 6 || len(second) != 2 || second[0].Name != "John Deere" {
		t.Errorf("a doua pagină: %v, total %d, %+v", err, total, second)
	}
	empty, total, err := repo.List(domain.ListFilter{}, &domain.PageRequest{Page: 9, PageSize: 4})
	if err != nil || total != 6 || len(empty) != 0 {
		t.Errorf("după ultima pagină: %v, total %d, %d rânduri", err, total, len(empty))
	}

	for query, want := range map[string]int{"fendt": 2, "%": 1, "_": 1, "deere": 1, "nimic": 0} {
		items, total, err := repo.List(domain.ListFilter{Query: query}, &domain.PageRequest{Page: 1, PageSize: 10})
		if err != nil || total != want || len(items) != want {
			t.Errorf("căutare %q: %v, total %d, %d rânduri (aștept %d)", query, err, total, len(items), want)
		}
	}
	all, total, err := repo.List(domain.ListFilter{Query: "fendt"}, nil)
	if err != nil || total != 2 || len(all) != 2 {
		t.Errorf("fără paginare: %v, total %d", err, total)
	}
}

// Terenurile: filtrele de suprafață, ordinea cerută și suprafața totală a tuturor celor filtrate.
func TestPagination_FieldsFiltersAndTotalArea(t *testing.T) {
	db := requireDB(t)
	for name, area := range map[string]float64{"Alfa": 10, "Beta": 50, "Gama": 30, "Delta": 5} {
		if _, err := db.Exec(`UPDATE fields SET area_ha = $2 WHERE id = $1`, insertField(t, db, name), area); err != nil {
			t.Fatal(err)
		}
	}
	repo := postgres.NewFieldRepo(db)
	min := 8.0
	items, total, area, err := repo.List(domain.FieldFilter{MinArea: &min, SortBy: "area", SortDesc: true}, &domain.PageRequest{Page: 1, PageSize: 2})
	if err != nil || total != 3 || area != 90 || len(items) != 2 || items[0].Name != "Beta" || items[1].Name != "Gama" {
		t.Errorf("suprafață ≥ 8, descrescător: %v, total %d, suprafață %.0f, %+v", err, total, area, items)
	}
}

// Alocări: lucrările cu oricare din statusurile cerute, paginate.
func TestPagination_FieldOperationStatuses(t *testing.T) {
	db := requireDB(t)
	field := insertField(t, db, "Tarla")
	start := time.Now().Add(24 * time.Hour)
	for _, status := range []string{"planned", "planned", "in_progress", "completed", "canceled"} {
		insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "seeding", Status: status, PlannedStart: start, AreaPlannedHa: 1})
	}
	repo := postgres.NewFieldOperationRepo(db)
	items, total, err := repo.List(repository.FieldOperationFilter{Statuses: []string{"planned", "in_progress"}}, &domain.PageRequest{Page: 1, PageSize: 2})
	if err != nil || total != 3 || len(items) != 2 {
		t.Errorf("planificate și în lucru: %v, total %d, %d rânduri", err, total, len(items))
	}
}
