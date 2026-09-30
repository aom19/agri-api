package usecase

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}
}

const openMeteoBody = `{"current":{"time":"2026-09-29T10:00","temperature_2m":21.6,"weather_code":61,"is_day":1,"relative_humidity_2m":70,"precipitation":0.44,"cloud_cover":80,"wind_speed_10m":12.34,"wind_direction_10m":180}}`
const openWeatherBody = `{"name":"Cantemir","weather":[{"id":800,"description":"cer senin","icon":"01n"}],"main":{"temp":15.2,"humidity":50},"wind":{"speed":2.0,"deg":90},"clouds":{"all":0},"rain":{"1h":0.2},"dt":1700000000,"timezone":7200}`

// fakeWeatherClient răspunde diferit în funcție de furnizor și numără apelurile.
func fakeWeatherClient(calls *int, openMeteo, openWeather func() *http.Response) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		*calls++
		if strings.Contains(r.URL.Host, "openweathermap") {
			return openWeather(), nil
		}
		return openMeteo(), nil
	})}
}

func TestWeatherService_OpenMeteo(t *testing.T) {
	calls := 0
	svc := NewWeatherService("")
	svc.client = fakeWeatherClient(&calls, func() *http.Response { return jsonResponse(200, openMeteoBody) }, nil)
	now := time.Now()
	svc.now = func() time.Time { return now }

	weather, err := svc.GetCurrent(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if weather.Location != "Cantemir" || weather.TemperatureC != 22 || weather.Condition != "Ploaie slabă" || weather.Icon != "rain" || weather.Source != "Open-Meteo" {
		t.Errorf("meteo greșit: %+v", weather)
	}
	if *weather.WindSpeedKmh != 12.3 || *weather.PrecipitationMM != 0.4 || *weather.HumidityPercent != 70 {
		t.Errorf("detalii greșite: %+v", weather)
	}

	// cache: același loc, în TTL → fără apel nou
	if _, err := svc.GetCurrent(context.Background()); err != nil || calls != 1 {
		t.Errorf("cache-ul trebuie folosit: %v, apeluri=%d", err, calls)
	}
	// alt loc fără nume → „Teren”, apel nou
	other, err := svc.GetCurrentFor(context.Background(), WeatherLocation{Latitude: 47, Longitude: 28})
	if err != nil || other.Location != "Teren" || calls != 2 {
		t.Errorf("alt loc: %v, %s, apeluri=%d", err, other.Location, calls)
	}
	// după expirarea cache-ului → apel nou
	svc.now = func() time.Time { return now.Add(weatherCacheTTL + time.Second) }
	if _, err := svc.GetCurrent(context.Background()); err != nil || calls != 3 {
		t.Errorf("cache expirat: %v, apeluri=%d", err, calls)
	}

	failures := map[string]func() *http.Response{
		"status 500":   func() *http.Response { return jsonResponse(500, "") },
		"json invalid": func() *http.Response { return jsonResponse(200, "{") },
		"fără timp":    func() *http.Response { return jsonResponse(200, `{"current":{}}`) },
	}
	for name, respond := range failures {
		svc := NewWeatherService("")
		svc.client = fakeWeatherClient(&calls, respond, nil)
		if _, err := svc.GetCurrent(context.Background()); err == nil {
			t.Errorf("%s: mă așteptam la eroare", name)
		}
	}
	svc = NewWeatherService("")
	svc.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, errors.New("offline") })}
	if _, err := svc.GetCurrent(context.Background()); err == nil {
		t.Error("eroarea de rețea trebuie propagată")
	}
}

func TestWeatherService_OpenWeatherMap(t *testing.T) {
	calls := 0
	svc := NewWeatherService("cheie")
	svc.client = fakeWeatherClient(&calls,
		func() *http.Response { return jsonResponse(200, openMeteoBody) },
		func() *http.Response { return jsonResponse(200, openWeatherBody) })

	weather, err := svc.GetCurrent(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if weather.Source != "OpenWeatherMap" || weather.Condition != "Cer senin" || weather.Icon != "moon" || weather.TemperatureC != 15 || *weather.WindSpeedKmh != 7.2 || *weather.PrecipitationMM != 0.2 {
		t.Errorf("OpenWeatherMap: %+v", weather)
	}

	// când OpenWeatherMap eșuează, se folosește Open-Meteo
	fallbacks := []func() *http.Response{
		func() *http.Response { return jsonResponse(500, "") },
		func() *http.Response { return jsonResponse(200, `{"weather":[]}`) },
		func() *http.Response { return jsonResponse(200, "{") },
	}
	for i, respond := range fallbacks {
		svc := NewWeatherService("cheie")
		svc.client = fakeWeatherClient(&calls, func() *http.Response { return jsonResponse(200, openMeteoBody) }, respond)
		if weather, err := svc.GetCurrent(context.Background()); err != nil || weather.Source != "Open-Meteo" {
			t.Errorf("fallback #%d: %v, %s", i, err, weather.Source)
		}
	}
}

func TestWeatherHelpers(t *testing.T) {
	if weatherDescription(0) != "Senin" || weatherDescription(999) != "Condiții meteo" {
		t.Error("weatherDescription")
	}
	icons := map[[2]int]string{
		{0, 0}: "moon", {0, 1}: "sunny", {1, 1}: "partly_cloudy", {3, 1}: "cloudy", {45, 1}: "cloudy",
		{55, 1}: "rain", {73, 1}: "snow", {85, 1}: "snow", {81, 1}: "showers", {95, 1}: "storm", {90, 1}: "partly_cloudy",
	}
	for in, want := range icons {
		if got := weatherIcon(in[0], in[1]); got != want {
			t.Errorf("weatherIcon(%d,%d)=%s, vreau %s", in[0], in[1], got, want)
		}
	}
	if openWeatherDescription("ploaie", 500) != "Ploaie" || openWeatherDescription("", 500) != "Ploaie" || openWeatherDescription("  ", 200) != "Furtună" {
		t.Error("openWeatherDescription")
	}
	fallbacks := map[int]string{250: "Furtună", 350: "Burniță", 550: "Ploaie", 650: "Ninsoare", 750: "Vizibilitate redusă", 800: "Senin", 803: "Înnorat", 100: "Condiții meteo"}
	for code, want := range fallbacks {
		if got := openWeatherFallbackDescription(code); got != want {
			t.Errorf("openWeatherFallbackDescription(%d)=%s", code, got)
		}
	}
	owIcons := map[int]string{250: "storm", 350: "rain", 550: "rain", 650: "snow", 750: "cloudy", 801: "partly_cloudy", 804: "cloudy", 100: "partly_cloudy"}
	for code, want := range owIcons {
		if got := openWeatherIcon(code, "01d"); got != want {
			t.Errorf("openWeatherIcon(%d)=%s", code, got)
		}
	}
	if openWeatherIcon(800, "01d") != "sunny" || openWeatherIcon(800, "01n") != "moon" {
		t.Error("openWeatherIcon zi/noapte")
	}
	if round1(1.26) != 1.3 || *intPtr(3) != 3 || *floatPtr(1.5) != 1.5 {
		t.Error("helpers numerici")
	}
}
