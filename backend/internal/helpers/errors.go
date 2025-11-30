package helpers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InternalServerError logs the real error and returns a generic message to the client
func InternalServerError(c *gin.Context, message string, err error) {
	// Log the real error for debugging (server-side only)
	log.Printf("ERROR: %s - %v", message, err)
	
	// Return generic message to client (no internal details)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": message,
	})
}

// BadRequest logs and returns a bad request error
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}

// NotFound returns a not found error
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": resource + " not found",
	})
}

// Forbidden returns a forbidden error
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": message,
	})
}

// Conflict returns a conflict error (e.g., duplicate entry)
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"error": message,
	})
}
