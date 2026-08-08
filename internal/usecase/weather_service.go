package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"agri-api/internal/domain"
)

const (
	defaultWeatherLatitude  = 46.27749
	defaultWeatherLongitude = 28.20052
	openMeteoURL            = "https://api.open-meteo.com/v1/forecast"
	openWeatherMapURL       = "https://api.openweathermap.org/data/2.5/weather"
	weatherCacheTTL         = 10 * time.Minute
)

type WeatherLocation struct {
	Latitude  float64
	Longitude float64
	Name      string
}

type weatherCacheEntry struct {
	weather  domain.CurrentWeather
	cachedAt time.Time
}

type WeatherService struct {
	client            *http.Client
	openWeatherAPIKey string
	now               func() time.Time
	mu                sync.Mutex
	cache             map[string]weatherCacheEntry
}

func NewWeatherService(openWeatherAPIKey string) *WeatherService {
	return &WeatherService{
		client:            &http.Client{Timeout: 4 * time.Second},
		openWeatherAPIKey: openWeatherAPIKey,
		now:               time.Now,
		cache:             make(map[string]weatherCacheEntry),
	}
}

type openMeteoCurrentWeatherResponse struct {
	Current struct {
		Time               string  `json:"time"`
		Temperature2M      float64 `json:"temperature_2m"`
		WeatherCode        int     `json:"weather_code"`
		IsDay              int     `json:"is_day"`
		RelativeHumidity2M int     `json:"relative_humidity_2m"`
		Precipitation      float64 `json:"precipitation"`
		CloudCover         int     `json:"cloud_cover"`
		WindSpeed10M       float64 `json:"wind_speed_10m"`
		WindDirection10M   int     `json:"wind_direction_10m"`
	} `json:"current"`
}

type openWeatherCurrentWeatherResponse struct {
	Name    string `json:"name"`
	Weather []struct {
		ID          int    `json:"id"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Rain     map[string]float64 `json:"rain"`
	Snow     map[string]float64 `json:"snow"`
	Dt       int64              `json:"dt"`
	Timezone int                `json:"timezone"`
}

func (service *WeatherService) GetCurrent(ctx context.Context) (domain.CurrentWeather, error) {
	return service.GetCurrentFor(ctx, WeatherLocation{
		Latitude:  defaultWeatherLatitude,
		Longitude: defaultWeatherLongitude,
		Name:      "Cantemir",
	})
}

func (service *WeatherService) GetCurrentFor(ctx context.Context, location WeatherLocation) (domain.CurrentWeather, error) {
	if location.Name == "" {
		location.Name = "Teren"
	}

	cacheKey := fmt.Sprintf("%.4f:%.4f", location.Latitude, location.Longitude)

	service.mu.Lock()
	defer service.mu.Unlock()

	if entry, ok := service.cache[cacheKey]; ok && service.now().Sub(entry.cachedAt) < weatherCacheTTL {
		return entry.weather, nil
	}

	weather, err := service.fetchCurrent(ctx, location)
	if err != nil {
		return domain.CurrentWeather{}, err
	}

	service.cache[cacheKey] = weatherCacheEntry{weather: weather, cachedAt: service.now()}

	return weather, nil
}

func (service *WeatherService) fetchCurrent(ctx context.Context, location WeatherLocation) (domain.CurrentWeather, error) {
	if service.openWeatherAPIKey != "" {
		weather, err := service.fetchOpenWeatherCurrent(ctx, location)
		if err == nil {
			return weather, nil
		}
	}

	return service.fetchOpenMeteoCurrent(ctx, location)
}

func (service *WeatherService) fetchOpenMeteoCurrent(ctx context.Context, location WeatherLocation) (domain.CurrentWeather, error) {
	requestURL, err := url.Parse(openMeteoURL)
	if err != nil {
		return domain.CurrentWeather{}, err
	}

	query := requestURL.Query()
	query.Set("latitude", fmt.Sprintf("%.5f", location.Latitude))
	query.Set("longitude", fmt.Sprintf("%.5f", location.Longitude))
	query.Set("current", "temperature_2m,relative_humidity_2m,precipitation,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,is_day")
	query.Set("timezone", "Europe/Chisinau")
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return domain.CurrentWeather{}, err
	}

	response, err := service.client.Do(request)
	if err != nil {
		return domain.CurrentWeather{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return domain.CurrentWeather{}, fmt.Errorf("open-meteo returned status %d", response.StatusCode)
	}

	var payload openMeteoCurrentWeatherResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return domain.CurrentWeather{}, err
	}

	timezone, err := time.LoadLocation("Europe/Chisinau")
	if err != nil {
		timezone = time.UTC
	}

	observedAt, err := time.ParseInLocation("2006-01-02T15:04", payload.Current.Time, timezone)
	if err != nil {
		return domain.CurrentWeather{}, errors.New("open-meteo response is missing current weather time")
	}

	return domain.CurrentWeather{
		Location:          location.Name,
		TemperatureC:      int(math.Round(payload.Current.Temperature2M)),
		Condition:         weatherDescription(payload.Current.WeatherCode),
		Icon:              weatherIcon(payload.Current.WeatherCode, payload.Current.IsDay),
		WeatherCode:       payload.Current.WeatherCode,
		ObservedAt:        observedAt,
		Source:            "Open-Meteo",
		HumidityPercent:   intPtr(payload.Current.RelativeHumidity2M),
		WindSpeedKmh:      floatPtr(round1(payload.Current.WindSpeed10M)),
		WindDirectionDeg:  intPtr(payload.Current.WindDirection10M),
		PrecipitationMM:   floatPtr(round1(payload.Current.Precipitation)),
		CloudCoverPercent: intPtr(payload.Current.CloudCover),
	}, nil
}

func (service *WeatherService) fetchOpenWeatherCurrent(ctx context.Context, location WeatherLocation) (domain.CurrentWeather, error) {
	requestURL, err := url.Parse(openWeatherMapURL)
	if err != nil {
		return domain.CurrentWeather{}, err
	}

	query := requestURL.Query()
	query.Set("lat", fmt.Sprintf("%.5f", location.Latitude))
	query.Set("lon", fmt.Sprintf("%.5f", location.Longitude))
	query.Set("units", "metric")
	query.Set("lang", "ro")
	query.Set("appid", service.openWeatherAPIKey)
	requestURL.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return domain.CurrentWeather{}, err
	}

	response, err := service.client.Do(request)
	if err != nil {
		return domain.CurrentWeather{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return domain.CurrentWeather{}, fmt.Errorf("openweathermap returned status %d", response.StatusCode)
	}

	var payload openWeatherCurrentWeatherResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return domain.CurrentWeather{}, err
	}

	if len(payload.Weather) == 0 {
		return domain.CurrentWeather{}, errors.New("openweathermap response is missing weather conditions")
	}

	condition := payload.Weather[0]
	observedAt := time.Unix(payload.Dt, 0).In(time.FixedZone("OpenWeatherMap", payload.Timezone))
	precipitation := payload.Rain["1h"] + payload.Snow["1h"]

	return domain.CurrentWeather{
		Location:          location.Name,
		TemperatureC:      int(math.Round(payload.Main.Temp)),
		Condition:         openWeatherDescription(condition.Description, condition.ID),
		Icon:              openWeatherIcon(condition.ID, condition.Icon),
		WeatherCode:       condition.ID,
		ObservedAt:        observedAt,
		Source:            "OpenWeatherMap",
		HumidityPercent:   intPtr(payload.Main.Humidity),
		WindSpeedKmh:      floatPtr(round1(payload.Wind.Speed * 3.6)),
		WindDirectionDeg:  intPtr(payload.Wind.Deg),
		PrecipitationMM:   floatPtr(round1(precipitation)),
		CloudCoverPercent: intPtr(payload.Clouds.All),
	}, nil
}

func weatherDescription(code int) string {
	descriptions := map[int]string{
		0:  "Senin",
		1:  "Mai mult senin",
		2:  "Parțial noros",
		3:  "Înnorat",
		45: "Ceață",
		48: "Ceață cu chiciură",
		51: "Burniță ușoară",
		53: "Burniță",
		55: "Burniță densă",
		61: "Ploaie slabă",
		63: "Ploaie",
		65: "Ploaie puternică",
		71: "Ninsoare slabă",
		73: "Ninsoare",
		75: "Ninsoare puternică",
		80: "Averse slabe",
		81: "Averse",
		82: "Averse puternice",
		95: "Furtună",
		96: "Furtună cu grindină",
		99: "Furtună cu grindină",
	}

	if description, ok := descriptions[code]; ok {
		return description
	}

	return "Condiții meteo"
}

func weatherIcon(code int, isDay int) string {
	if code == 0 {
		if isDay == 0 {
			return "moon"
		}
		return "sunny"
	}
	if code == 1 || code == 2 {
		return "partly_cloudy"
	}
	if code == 3 || code == 45 || code == 48 {
		return "cloudy"
	}
	if code >= 51 && code <= 67 {
		return "rain"
	}
	if (code >= 71 && code <= 77) || code == 85 || code == 86 {
		return "snow"
	}
	if code >= 80 && code <= 82 {
		return "showers"
	}
	if code >= 95 {
		return "storm"
	}

	return "partly_cloudy"
}

func openWeatherDescription(description string, code int) string {
	if description == "" {
		return openWeatherFallbackDescription(code)
	}

	runes := []rune(strings.TrimSpace(description))
	if len(runes) == 0 {
		return openWeatherFallbackDescription(code)
	}

	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}

func openWeatherFallbackDescription(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "Furtună"
	case code >= 300 && code < 400:
		return "Burniță"
	case code >= 500 && code < 600:
		return "Ploaie"
	case code >= 600 && code < 700:
		return "Ninsoare"
	case code >= 700 && code < 800:
		return "Vizibilitate redusă"
	case code == 800:
		return "Senin"
	case code > 800:
		return "Înnorat"
	default:
		return "Condiții meteo"
	}
}

func openWeatherIcon(code int, icon string) string {
	switch {
	case code >= 200 && code < 300:
		return "storm"
	case code >= 300 && code < 600:
		return "rain"
	case code >= 600 && code < 700:
		return "snow"
	case code >= 700 && code < 800:
		return "cloudy"
	case code == 800:
		if strings.HasSuffix(icon, "n") {
			return "moon"
		}
		return "sunny"
	case code == 801 || code == 802:
		return "partly_cloudy"
	case code > 802:
		return "cloudy"
	default:
		return "partly_cloudy"
	}
}

func intPtr(value int) *int {
	return &value
}

func floatPtr(value float64) *float64 {
	return &value
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}
