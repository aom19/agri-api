package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

func TestPercentageAndMinPositive(t *testing.T) {
	cases := []struct{ value, total, want int }{
		{0, 10, 0}, {5, 0, 0}, {5, 10, 50}, {1, 3, 33}, {2, 3, 67}, {15, 10, 100},
	}
	for _, c := range cases {
		if got := usecase.Percentage(c.value, c.total); got != c.want {
			t.Errorf("percentage(%d,%d)=%d, vreau %d", c.value, c.total, got, c.want)
		}
	}
	minCases := []struct{ a, b, want int }{{0, 5, 5}, {5, 0, 5}, {3, 5, 3}, {7, 5, 5}}
	for _, c := range minCases {
		if got := usecase.MinPositive(c.a, c.b); got != c.want {
			t.Errorf("minPositive(%d,%d)=%d, vreau %d", c.a, c.b, got, c.want)
		}
	}
}

func TestDashboardService_GetCards(t *testing.T) {
	repo := &dashboardRepoMock{getCardStats: func() (*domain.DashboardCardStats, error) {
		return &domain.DashboardCardStats{
			TotalMachines: 10, ActiveMachines: 8, TotalOperators: 4, ActiveOperators: 2,
			ActiveAssignments: 3, NewMachinesLast7Days: 1, NewOperatorsLast7Days: 2,
		}, nil
	}}
	svc := usecase.NewDashboardService(repo, nil)

	cards, err := svc.GetCards()
	if err != nil || len(cards) != 4 {
		t.Fatalf("GetCards: %v, %d carduri", err, len(cards))
	}
	byKey := map[domain.DashboardCardKey]domain.DashboardCard{}
	for _, c := range cards {
		byKey[c.Key] = c
	}
	if byKey[domain.DashboardCardTotalMachines].Value != 10 || byKey[domain.DashboardCardTotalMachines].Progress != 30 {
		t.Errorf("total mașini: %+v", byKey[domain.DashboardCardTotalMachines])
	}
	if byKey[domain.DashboardCardActiveMachines].Progress != 80 {
		t.Errorf("mașini active: %+v", byKey[domain.DashboardCardActiveMachines])
	}
	if byKey[domain.DashboardCardTotalOperators].Progress != 50 || byKey[domain.DashboardCardTotalOperators].Trend != 2 {
		t.Errorf("operatori: %+v", byKey[domain.DashboardCardTotalOperators])
	}
	// capacitatea = min(10 mașini, 4 operatori) = 4 → 3/4 = 75%
	if byKey[domain.DashboardCardActiveAssignments].Progress != 75 {
		t.Errorf("alocări active: %+v", byKey[domain.DashboardCardActiveAssignments])
	}

	repo.getCardStats = func() (*domain.DashboardCardStats, error) { return nil, errors.New("db down") }
	if _, err := svc.GetCards(); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
	if _, err := svc.GetQuickStats(); err != nil {
		t.Errorf("GetQuickStats: %v", err)
	}
}

func TestDashboardService_GetRecentActivity(t *testing.T) {
	svc := usecase.NewDashboardService(&dashboardRepoMock{}, nil)
	items, err := svc.GetRecentActivity(5)
	if err != nil || len(items) != 0 {
		t.Fatalf("fără repo de audit trebuie să întoarcă listă goală: %v", err)
	}

	var gotLimit int
	audit := &auditRepoMock{getAll: func(limit int, _, _ string) ([]domain.AuditEntry, error) {
		gotLimit = limit
		return []domain.AuditEntry{
			{ID: 1, EntityType: "machine", Action: "update", Changes: map[string]interface{}{"status": "active", "old_status": "maintenance"}},
			{ID: 2, EntityType: "field", Action: "create", Changes: map[string]interface{}{"status": 3}},
			{ID: 3, EntityType: "field", Action: "delete"},
		}, nil
	}}
	svc = usecase.NewDashboardService(&dashboardRepoMock{}, audit)

	items, err = svc.GetRecentActivity(0)
	if err != nil || gotLimit != 5 || len(items) != 3 {
		t.Fatalf("limita implicită: %v, limit=%d", err, gotLimit)
	}
	if *items[0].Status != "active" || *items[0].OldStatus != "maintenance" {
		t.Errorf("statusurile nu au fost extrase: %+v", items[0])
	}
	if items[1].Status != nil || items[2].Status != nil {
		t.Error("valorile non-text sau lipsă trebuie să fie nil")
	}
	if _, err = svc.GetRecentActivity(500); err != nil || gotLimit != 50 {
		t.Errorf("limita maximă: %v, limit=%d", err, gotLimit)
	}

	audit.getAll = func(int, string, string) ([]domain.AuditEntry, error) { return nil, errors.New("db down") }
	if _, err := svc.GetRecentActivity(5); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
}
