package usecase

import (
	"net/http"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/email"
)

// Acest fișier expune funcțiile și câmpurile interne verificate de testele din usecase/test.
// Testele stau într-un pachet separat și văd doar ce e exportat. Aplicația nu folosește nimic de aici.

const (
	ReportDateLayout = reportDateLayout
	WeatherCacheTTL  = weatherCacheTTL
)

var (
	BuildMovement                  = buildMovement
	MapCropRepoError               = mapCropRepoError
	Percentage                     = percentage
	MinPositive                    = minPositive
	ValidateCompletion             = validateCompletion
	BuildOverdueMessage            = buildOverdueMessage
	FillTimeline                   = fillTimeline
	TruncateTo                     = truncateTo
	IsDigestDue                    = isDigestDue
	IsoWeekday                     = isoWeekday
	DigestPeriod                   = digestPeriod
	FormatRomanianDate             = formatRomanianDate
	PolygonCentroid                = polygonCentroid
	WeatherDescription             = weatherDescription
	WeatherIcon                    = weatherIcon
	OpenWeatherDescription         = openWeatherDescription
	OpenWeatherFallbackDescription = openWeatherFallbackDescription
	OpenWeatherIcon                = openWeatherIcon
	IntPtr                         = intPtr
	FloatPtr                       = floatPtr
	Round1                         = round1
)

func (service *AuthService) ClientOrigin() string { return service.clientOrigin }

func (service *AuthService) JWTService() *auth.JWTService { return service.jwtService }

func (service *AuthService) SetEmailService(emailService *email.EmailService) {
	service.emailService = emailService
}

func (s *NotificationService) PushToSubscribers(userID int64, un domain.UserNotification) {
	s.pushToSubscribers(userID, un)
}

func (s *NotificationService) SubscriberCount(userID int64) int { return len(s.subscribers[userID]) }

func (service *ReportService) ResolvePeriod(query ReportQuery) (domain.ReportFilter, domain.ReportPeriod, error) {
	return service.resolvePeriod(query)
}

func (service *ReportDigestService) Interval() time.Duration { return service.interval }

func (service *ReportDigestService) SetInterval(interval time.Duration) { service.interval = interval }

func (service *ReportDigestService) SetNow(now func() time.Time) { service.now = now }

func (service *WeatherService) SetHTTPClient(client *http.Client) { service.client = client }

func (service *WeatherService) SetNow(now func() time.Time) { service.now = now }

func (service *WeatherService) ResetCache() { service.cache = map[string]weatherCacheEntry{} }

func (m *FieldOperationOverdueMonitor) Interval() time.Duration { return m.interval }

func (m *WeatherSnapshotMonitor) Interval() time.Duration { return m.interval }
