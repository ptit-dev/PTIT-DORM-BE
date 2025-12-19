package handlers

import (
	"Backend_Dorm_PTIT/models"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

// Health godoc
// @Summary Health check
// @Description Check if the service is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthResponse
// @Router /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:  "ok",
		Message: "Service is running version 1.9.0",
		Time:    time.Now().Format(time.RFC3339),
	})
}