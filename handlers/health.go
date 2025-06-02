package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheck handles the health check endpoint
func HealthCheck(c *gin.Context) {
	// Obtém informações da aplicação do contexto
	appName := c.GetString("app_name")
	appVersion := c.GetString("app_version")

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"app": gin.H{
			"name":    appName,
			"version": appVersion,
		},
	})
} 