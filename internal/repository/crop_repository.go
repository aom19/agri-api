package repository

import (
	"database/sql"

	"agri-api/internal/domain"
)

type CropRepository interface {
	GetSeasons() ([]domain.Season, error)
	GetSeasonByID(id int64) (*domain.Season, error)
	GetActiveOrLatestSeason() (*domain.Season, error)
	CreateSeason(season *domain.Season) error
	UpdateSeason(id int64, season *domain.Season) error
	DeleteSeason(id int64) error

	GetCrops() ([]domain.Crop, error)
	GetCropByID(id int64) (*domain.Crop, error)
	CreateCrop(crop *domain.Crop) error
	UpdateCrop(id int64, crop *domain.Crop) error
	DeleteCrop(id int64) error

	ListFieldCrops(filter domain.FieldCropFilter) ([]domain.FieldCrop, error)
	GetFieldCropByID(id int64) (*domain.FieldCrop, error)
	CreateFieldCrop(item *domain.FieldCrop) error
	UpdateFieldCrop(id int64, item *domain.FieldCrop) error
	DeleteFieldCrop(id int64) error

	// Cultura deschisă pe teren: cel mult una; lucrările noi merg implicit la ea.
	GetOpenFieldCrop(fieldID string) (*domain.FieldCrop, error)
	CloseFieldCrop(id int64, input domain.FieldCropClose) error
	ReopenFieldCrop(id int64) error

	// Recolta.
	LockFieldCropHarvest(tx *sql.Tx, fieldCropID int64) (production *float64, recorded float64, err error)
	EnsureHarvestResource(tx *sql.Tx, crop *domain.Crop) (int64, error)
	MarkHarvestRecorded(tx *sql.Tx, fieldCropID int64, quantity float64) error
}

type WeatherSnapshotRepository interface {
	Insert(snapshot *domain.WeatherSnapshot) (bool, error)
	GetDailyAggregates(filter domain.ReportFilter) ([]domain.WeatherDailyAggregate, error)
	GetLatestPerField() ([]domain.WeatherSnapshot, error)
	CountInPeriod(filter domain.ReportFilter) (int, int, error)
}
