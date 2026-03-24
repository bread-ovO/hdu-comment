package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	requestIDContextKey = "request_id"
	requestIDHeader     = "X-Request-ID"
)

// RequestID ensures each request has a stable identifier for logs and clients.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(requestIDContextKey, requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

// GetRequestID reads the request identifier from gin context.
func GetRequestID(c *gin.Context) string {
	if value, ok := c.Get(requestIDContextKey); ok {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}
	return ""
}
