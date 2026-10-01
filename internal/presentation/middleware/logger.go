package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger outputs structured HTTP access logs with latency and correlation ID
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		corID := c.GetString(HeaderCorrelationID)

		if raw != "" {
			path = path + "?" + raw
		}

		log.Printf("[HTTP] %3d | %12v | %s | %s %-7s %s | correlation_id=%s",
			statusCode,
			latency,
			clientIP,
			method,
			c.Request.Proto,
			path,
			corID,
		)
	}
}
