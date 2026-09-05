package domain

import "time"

type ReportFrequency string

const (
	ReportFrequencyDaily   ReportFrequency = "daily"
	ReportFrequencyWeekly  ReportFrequency = "weekly"
	ReportFrequencyMonthly ReportFrequency = "monthly"
)

// ReportSubscription este abonamentul unui utilizator la raportul sumar trimis pe e-mail.
type ReportSubscription struct {
	ID         int64           `json:"id"`
	UserID     int64           `json:"user_id"`
	Frequency  ReportFrequency `json:"frequency"`
	SendHour   int             `json:"send_hour"`
	Weekday    int             `json:"weekday"`
	IsActive   bool            `json:"is_active"`
	LastSentAt *time.Time      `json:"last_sent_at,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// ReportSubscriptionRecipient este abonamentul împreună cu datele de contact ale utilizatorului.
type ReportSubscriptionRecipient struct {
	ReportSubscription
	Email     string
	FirstName string
}

// Tipurile rapoartelor pe culturi.
type ReportCropStat struct {
	CropID             int64    `json:"crop_id"`
	CropName           string   `json:"crop_name"`
	YieldUnit          string   `json:"yield_unit"`
	FieldsCount        int      `json:"fields_count"`
	PlantedAreaHa      float64  `json:"planted_area_ha"`
	ProductionTotal    float64  `json:"production_total"`
	YieldPerHa         *float64 `json:"yield_per_ha,omitempty"`
	ExpectedYieldPerHa *float64 `json:"expected_yield_per_ha,omitempty"`
	EstimatedCost      float64  `json:"estimated_cost"`
	RealCost           float64  `json:"real_cost"`
	CostPerHa          *float64 `json:"cost_per_ha,omitempty"`
}

type ReportFieldCropRow struct {
	FieldCrop
	OperationsCount int      `json:"operations_count"`
	EstimatedCost   float64  `json:"estimated_cost"`
	RealCost        float64  `json:"real_cost"`
	CostPerHa       *float64 `json:"cost_per_ha,omitempty"`
}

type ReportCropsTotals struct {
	FieldsCount     int      `json:"fields_count"`
	PlantedAreaHa   float64  `json:"planted_area_ha"`
	ProductionTotal float64  `json:"production_total"`
	YieldPerHa      *float64 `json:"yield_per_ha,omitempty"`
	EstimatedCost   float64  `json:"estimated_cost"`
	RealCost        float64  `json:"real_cost"`
}

type ReportCrops struct {
	Season  *Season              `json:"season,omitempty"`
	Seasons []Season             `json:"seasons"`
	Totals  ReportCropsTotals    `json:"totals"`
	ByCrop  []ReportCropStat     `json:"by_crop"`
	Items   []ReportFieldCropRow `json:"items"`
}
