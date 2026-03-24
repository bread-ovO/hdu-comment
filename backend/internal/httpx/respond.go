package httpx

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const requestIDContextKey = "request_id"

// ErrorResponse is the standard API error payload.
type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// Error writes a standardized error response.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{
		Error:     message,
		RequestID: requestID(c),
	})
}

// BindJSON binds request JSON and returns false after writing a 400 response on failure.
func BindJSON(c *gin.Context, dst any, message string) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		if message == "" {
			message = "invalid payload"
		}
		Error(c, http.StatusBadRequest, message)
		return false
	}
	return true
}

// ParamUUID parses a UUID path parameter and writes a 400 response on failure.
func ParamUUID(c *gin.Context, name, message string) (uuid.UUID, bool) {
	value, err := uuid.Parse(c.Param(name))
	if err != nil {
		Error(c, http.StatusBadRequest, message)
		return uuid.Nil, false
	}
	return value, true
}

// MustContextUUID reads a UUID from gin context and writes 401 on failure.
func MustContextUUID(c *gin.Context, key, missingMessage, invalidMessage string) (uuid.UUID, bool) {
	value, exists := c.Get(key)
	if !exists {
		Error(c, http.StatusUnauthorized, missingMessage)
		return uuid.Nil, false
	}

	id, ok := value.(uuid.UUID)
	if !ok {
		Error(c, http.StatusUnauthorized, invalidMessage)
		return uuid.Nil, false
	}

	return id, true
}

// QueryInt reads an integer query parameter and clamps it to a range.
func QueryInt(c *gin.Context, key string, defaultValue, minValue, maxValue int) int {
	value, exists := c.GetQuery(key)
	if !exists || value == "" {
		return defaultValue
	}

	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil {
		return defaultValue
	}

	if parsed < minValue {
		return minValue
	}
	if maxValue > 0 && parsed > maxValue {
		return maxValue
	}
	return parsed
}

// IsNotFound reports whether err means a missing record.
func IsNotFound(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrRecordNotFound)
}

func requestID(c *gin.Context) string {
	value, ok := c.Get(requestIDContextKey)
	if !ok {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}

	return id
}
