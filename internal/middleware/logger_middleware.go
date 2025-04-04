package middleware

import (
	"time"

	"cred.com/hack25/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// RequestLogger logs HTTP request information
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status
		status := c.Writer.Status()

		// Log request
		logger.WithFields(logger.Fields{
			"status":     status,
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency":    latency.String(),
			"user_agent": c.Request.UserAgent(),
		}).Info("HTTP Request")
	}
}
