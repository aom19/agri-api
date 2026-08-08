package repository

import "agri-api/internal/domain"

type DashboardRepository interface {
	GetCardStats() (*domain.DashboardCardStats, error)
}
