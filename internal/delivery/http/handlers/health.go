package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "agri-api",
		"version": "1.0.0",
		"message": "Service is healthy",
	})
}
