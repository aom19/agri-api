package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service *usecase.ReportService
	email   *usecase.ReportEmailService
	users   *usecase.UserService
}

func NewReportHandler(service *usecase.ReportService, opts ...func(*ReportHandler)) *ReportHandler {
	h := &ReportHandler{service: service}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func parseReportQuery(c *gin.Context) (usecase.ReportQuery, error) {
	query := usecase.ReportQuery{
		From:          c.Query("from"),
		To:            c.Query("to"),
		FieldID:       c.Query("field_id"),
		OperationType: c.Query("operation_type"),
	}
	if query.OperationType != "" && !domain.OperationType(query.OperationType).IsValid() {
		return query, errors.New("operation_type invalid")
	}

	var err error
	if query.MachineID, err = optionalInt64Query(c.Query("machine_id")); err != nil {
		return query, errors.New("machine_id invalid")
	}
	if query.OperatorID, err = optionalInt64Query(c.Query("operator_id")); err != nil {
		return query, errors.New("operator_id invalid")
	}
	return query, nil
}

func optionalInt64Query(raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func respondReport(c *gin.Context, payload interface{}, err error) {
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidReportPeriod) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payload)
}

// GetSummary returnează indicatorii generali ai raportului, cu comparație față de perioada anterioară.
// @Summary      Raport sumar
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD, implicit ultimele 30 de zile)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD, implicit azi)"
// @Param        field_id query string false "Filtru teren"
// @Param        operation_type query string false "Filtru tip operațiune (soil_preparation, seeding, fertilization, spraying, harvesting, irrigation)"
// @Param        machine_id query int false "Filtru mașină"
// @Param        operator_id query int false "Filtru operator"
// @Success      200 {object} domain.ReportSummary
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/summary [get]
func (h *ReportHandler) GetSummary(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetSummary(query)
	respondReport(c, report, err)
}

// GetOperations returnează raportul operațiunilor pe teren (serie temporală, pe tip, lista detaliată).
// @Summary      Raport operațiuni
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Param        field_id query string false "Filtru teren"
// @Param        operation_type query string false "Filtru tip operațiune (soil_preparation, seeding, fertilization, spraying, harvesting, irrigation)"
// @Param        machine_id query int false "Filtru mașină"
// @Param        operator_id query int false "Filtru operator"
// @Success      200 {object} domain.ReportOperations
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/operations [get]
func (h *ReportHandler) GetOperations(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetOperations(query)
	respondReport(c, report, err)
}

// GetFields returnează raportul terenurilor (suprafețe, lucrări, geometrie pentru hartă).
// @Summary      Raport terenuri
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Param        field_id query string false "Filtru teren"
// @Param        operation_type query string false "Filtru tip operațiune (soil_preparation, seeding, fertilization, spraying, harvesting, irrigation)"
// @Param        machine_id query int false "Filtru mașină"
// @Param        operator_id query int false "Filtru operator"
// @Success      200 {object} domain.ReportFields
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/fields [get]
func (h *ReportHandler) GetFields(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetFields(query)
	respondReport(c, report, err)
}

// GetFleet returnează raportul flotei (mașini și echipamente).
// @Summary      Raport flotă
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Param        field_id query string false "Filtru teren"
// @Param        operation_type query string false "Filtru tip operațiune (soil_preparation, seeding, fertilization, spraying, harvesting, irrigation)"
// @Param        machine_id query int false "Filtru mașină"
// @Param        operator_id query int false "Filtru operator"
// @Success      200 {object} domain.ReportFleet
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/fleet [get]
func (h *ReportHandler) GetFleet(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetFleet(query)
	respondReport(c, report, err)
}

// GetOperators returnează raportul operatorilor (încărcare, finalizări, întârzieri).
// @Summary      Raport operatori
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Param        field_id query string false "Filtru teren"
// @Param        operation_type query string false "Filtru tip operațiune (soil_preparation, seeding, fertilization, spraying, harvesting, irrigation)"
// @Param        machine_id query int false "Filtru mașină"
// @Param        operator_id query int false "Filtru operator"
// @Success      200 {object} domain.ReportOperators
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/operators [get]
func (h *ReportHandler) GetOperators(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetOperators(query)
	respondReport(c, report, err)
}

// GetStocks returnează raportul stocurilor și consumul estimat de resurse.
// @Summary      Raport stocuri
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Param        field_id query string false "Filtru teren"
// @Param        operation_type query string false "Filtru tip operațiune (soil_preparation, seeding, fertilization, spraying, harvesting, irrigation)"
// @Param        machine_id query int false "Filtru mașină"
// @Param        operator_id query int false "Filtru operator"
// @Success      200 {object} domain.ReportStocks
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/stocks [get]
func (h *ReportHandler) GetStocks(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetStocks(query)
	respondReport(c, report, err)
}

// WithReportEmail injectează serviciul de trimitere pe e-mail și serviciul de utilizatori.
func WithReportEmail(email *usecase.ReportEmailService, users *usecase.UserService) func(*ReportHandler) {
	return func(h *ReportHandler) {
		h.email = email
		h.users = users
	}
}

// GetCrops returnează raportul pe culturi și producție pentru un sezon.
// @Summary      Raport culturi
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        season_id query int false "Sezon (implicit cel activ sau cel mai recent)"
// @Param        field_id query string false "Filtru teren"
// @Success      200 {object} domain.ReportCrops
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/crops [get]
func (h *ReportHandler) GetCrops(c *gin.Context) {
	seasonID, err := optionalInt64Query(c.Query("season_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "season_id invalid"})
		return
	}
	report, err := h.service.GetCrops(seasonID, c.Query("field_id"))
	respondReport(c, report, err)
}

// GetWeather returnează istoricul meteo agregat pe zile.
// @Summary      Raport meteo
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Param        field_id query string false "Filtru teren"
// @Success      200 {object} domain.ReportWeather
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/weather [get]
func (h *ReportHandler) GetWeather(c *gin.Context) {
	query, err := parseReportQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	report, err := h.service.GetWeather(query)
	respondReport(c, report, err)
}

// SendSummaryEmail trimite sumarul perioadei pe e-mailul utilizatorului curent.
// @Summary      Trimitere sumar pe e-mail
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from query string false "Data de început (YYYY-MM-DD)"
// @Param        to query string false "Data de sfârșit (YYYY-MM-DD)"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /reports/summary/email [post]
func (h *ReportHandler) SendSummaryEmail(c *gin.Context) {
	if h.email == nil || h.users == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "trimiterea pe e-mail nu este configurată"})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
		return
	}
	user, err := h.users.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "utilizatorul nu există"})
		return
	}
	displayName, _ := h.users.GetUserDisplayName(userID)
	query := usecase.ReportQuery{From: c.Query("from"), To: c.Query("to")}
	if err := h.email.SendSummary(query, user.Email, displayName); err != nil {
		respondReport(c, nil, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Raportul a fost trimis pe " + user.Email})
}
