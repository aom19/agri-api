package postgres_test

import (
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
)

// Filtrarea resurselor se face în SQL: după categoria tipului și după text (nume, tip, unitate),
// iar numărătoarea pe categorii acoperă toate rândurile, nu doar lista filtrată.
func TestResourceListFilters(t *testing.T) {
	db := requireDB(t)
	insertStock(t, db, "Motorină (test)", 100, 0, 7) // tip 'fuel', unitate l
	insertStock(t, db, "Benzină", 50, 0, 8)          // tot fuel
	seedType := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ('Semințe', 'seed', 'kg') RETURNING id`)
	for _, name := range []string{"Sămânță grâu", "Sămânță porumb"} {
		if _, err := db.Exec(`INSERT INTO resources (name, resource_type_id, price_per_unit) VALUES ($1, $2, 4.75)`, name, seedType); err != nil {
			t.Fatal(err)
		}
	}

	resources := postgres.NewResourceRepo(db)

	all, err := resources.List(domain.ResourceFilter{})
	if err != nil || len(all) != 4 {
		t.Fatalf("fără filtru: %v, %d", err, len(all))
	}
	seeds, err := resources.List(domain.ResourceFilter{Category: "seed"})
	if err != nil || len(seeds) != 2 {
		t.Errorf("categoria seed: %v, %d", err, len(seeds))
	}
	// textul caută în nume, tip și unitate, fără să țină cont de majuscule
	for query, want := range map[string]int{"GRÂU": 1, "semințe": 2, "kg": 2, "nimic": 0} {
		got, err := resources.List(domain.ResourceFilter{Query: query})
		if err != nil || len(got) != want {
			t.Errorf("text %q: %v, %d (aștept %d)", query, err, len(got), want)
		}
	}
	both, err := resources.List(domain.ResourceFilter{Category: "fuel", Query: "benz"})
	if err != nil || len(both) != 1 || both[0].Name != "Benzină" || both[0].Quantity != 50 {
		t.Errorf("categorie + text: %v, %+v", err, both)
	}
	counts, err := resources.CountByCategory()
	if err != nil || counts.Total != 4 || len(counts.ByCategory) != 2 {
		t.Fatalf("numărătoarea resurselor: %v, %+v", err, counts)
	}
	byCategory := map[string]int{}
	for _, c := range counts.ByCategory {
		byCategory[c.Category] = c.Count
	}
	if byCategory["fuel"] != 2 || byCategory["seed"] != 2 {
		t.Errorf("repartiția pe categorii: %v", byCategory)
	}
}
