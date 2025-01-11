package controllers

import (
	"api-service/workers" // or wherever you put your workers package
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetJobStatusHandler(c *gin.Context) {
	// The :uuid parameter from the route
	uuid := c.Param("uuid")

	// Fetch status from your worker package
	status, ok := workers.GetJobStatus(uuid)
	if !ok {
		// If not found in your map/DB, return 404
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	// Return the status
	c.JSON(http.StatusOK, gin.H{
		"uuid":   uuid,
		"status": status,
	})
}
