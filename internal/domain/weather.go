package domain

import "time"

type CurrentWeather struct {
	Location          string    `json:"location"`
	TemperatureC      int       `json:"temperature_c"`
	Condition         string    `json:"condition"`
	Icon              string    `json:"icon"`
	WeatherCode       int       `json:"weather_code"`
	ObservedAt        time.Time `json:"observed_at"`
	Source            string    `json:"source"`
	HumidityPercent   *int      `json:"humidity_percent,omitempty"`
	WindSpeedKmh      *float64  `json:"wind_speed_kmh,omitempty"`
	WindDirectionDeg  *int      `json:"wind_direction_deg,omitempty"`
	PrecipitationMM   *float64  `json:"precipitation_mm,omitempty"`
	CloudCoverPercent *int      `json:"cloud_cover_percent,omitempty"`
}
