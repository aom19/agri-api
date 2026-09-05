package domain

import "time"

// WeatherSnapshot este o observație meteo salvată periodic pentru un teren.
type WeatherSnapshot struct {
	ID                int64     `json:"id"`
	FieldID           *string   `json:"field_id,omitempty"`
	FieldName         *string   `json:"field_name,omitempty"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	ObservedAt        time.Time `json:"observed_at"`
	TemperatureC      float64   `json:"temperature_c"`
	Condition         string    `json:"condition"`
	WeatherCode       *int      `json:"weather_code,omitempty"`
	HumidityPercent   *int      `json:"humidity_percent,omitempty"`
	WindSpeedKmh      *float64  `json:"wind_speed_kmh,omitempty"`
	WindDirectionDeg  *int      `json:"wind_direction_deg,omitempty"`
	PrecipitationMM   *float64  `json:"precipitation_mm,omitempty"`
	CloudCoverPercent *int      `json:"cloud_cover_percent,omitempty"`
	Source            string    `json:"source"`
}

// WeatherDailyAggregate este sinteza zilnică a observațiilor.
type WeatherDailyAggregate struct {
	Day             string   `json:"day"`
	AvgTemperatureC float64  `json:"avg_temperature_c"`
	MinTemperatureC float64  `json:"min_temperature_c"`
	MaxTemperatureC float64  `json:"max_temperature_c"`
	PrecipitationMM float64  `json:"precipitation_mm"`
	AvgHumidity     *float64 `json:"avg_humidity_percent,omitempty"`
	AvgWindKmh      *float64 `json:"avg_wind_speed_kmh,omitempty"`
	Samples         int      `json:"samples"`
}

type ReportWeatherSummary struct {
	AvgTemperatureC      float64 `json:"avg_temperature_c"`
	MinTemperatureC      float64 `json:"min_temperature_c"`
	MaxTemperatureC      float64 `json:"max_temperature_c"`
	TotalPrecipitationMM float64 `json:"total_precipitation_mm"`
	RainyDays            int     `json:"rainy_days"`
	Samples              int     `json:"samples"`
	FieldsCovered        int     `json:"fields_covered"`
}

type ReportWeather struct {
	Period  ReportPeriod            `json:"period"`
	Summary ReportWeatherSummary    `json:"summary"`
	Series  []WeatherDailyAggregate `json:"series"`
	Latest  []WeatherSnapshot       `json:"latest"`
}
