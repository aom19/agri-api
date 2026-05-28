package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck verifica starea serviciului
// @Summary      Health check
// @Tags         system
// @Produce      json
// @Success      200 {object} object{status=string}
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "agri-api",
		"version": "1.0.0",
		"message": "Service is healthy",
	})
}
