package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderCorrelationID = "X-Correlation-ID"

// CorrelationID propagates or generates unique request tracing ID
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		corID := c.GetHeader(HeaderCorrelationID)
		if corID == "" {
			corID = uuid.New().String()
		}
		c.Set(HeaderCorrelationID, corID)
		c.Writer.Header().Set(HeaderCorrelationID, corID)
		c.Next()
	}
}
