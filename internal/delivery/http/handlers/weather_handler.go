package handlers

import (
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	service *usecase.WeatherService
}

type CurrentWeatherResponse = domain.CurrentWeather

func NewWeatherHandler(service *usecase.WeatherService) *WeatherHandler {
	return &WeatherHandler{service: service}
}

// GetCurrent returneaza vremea curenta pentru Cantemir sau pentru coordonatele cerute.
// @Summary      Vreme curenta
// @Tags         weather
// @Produce      json
// @Param        lat query number false "Latitudine WGS84"
// @Param        lng query number false "Longitudine WGS84"
// @Param        location query string false "Numele locatiei afisat in raspuns"
// @Success      200 {object} CurrentWeatherResponse
// @Failure      400 {object} object{error=string}
// @Failure      502 {object} object{error=string}
// @Router       /weather/current [get]
func (h *WeatherHandler) GetCurrent(c *gin.Context) {
	latQuery := c.Query("lat")
	lngQuery := c.Query("lng")
	locationName := c.DefaultQuery("location", "Cantemir")

	if latQuery == "" && lngQuery == "" {
		weather, err := h.service.GetCurrent(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Serviciul meteo este indisponibil"})
			return
		}

		c.JSON(http.StatusOK, weather)
		return
	}

	if latQuery == "" || lngQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parametrii lat si lng trebuie trimisi impreuna"})
		return
	}

	latitude, err := strconv.ParseFloat(latQuery, 64)
	if err != nil || latitude < -90 || latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parametrul lat este invalid"})
		return
	}

	longitude, err := strconv.ParseFloat(lngQuery, 64)
	if err != nil || longitude < -180 || longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parametrul lng este invalid"})
		return
	}

	weather, err := h.service.GetCurrentFor(c.Request.Context(), usecase.WeatherLocation{
		Latitude:  latitude,
		Longitude: longitude,
		Name:      locationName,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Serviciul meteo este indisponibil"})
		return
	}

	c.JSON(http.StatusOK, weather)
}
