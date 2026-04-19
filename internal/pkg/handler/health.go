package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/steellink/backend/pkg/response"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
}

// Health returns the health status of the service
func Health(c *gin.Context) {
	response.Success(c, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
		GoVersion: runtime.Version(),
	})
}

// Ready returns the readiness status of the service
func Ready(c *gin.Context) {
	// Check database connection
	// Check redis connection
	// For now, just return ok
	response.Success(c, gin.H{
		"status": "ready",
	})
}

// Live returns the liveness status of the service
func Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
