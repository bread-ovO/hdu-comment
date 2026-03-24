package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/auth"
	"github.com/hdu-dp/backend/internal/httpx"
	"github.com/hdu-dp/backend/internal/repository"
)

// AuthMiddleware handles JWT extraction and validation.
type AuthMiddleware struct {
	tokens *auth.JWTManager
	users  *repository.UserRepository
}

// NewAuthMiddleware constructs an auth middleware instance.
func NewAuthMiddleware(tokens *auth.JWTManager, users *repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{tokens: tokens, users: users}
}

// RequireAuth ensures a valid JWT is provided.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			c.Abort()
			httpx.Error(c, http.StatusUnauthorized, "missing token")
			return
		}

		claims, err := m.tokens.Parse(token)
		if err != nil {
			c.Abort()
			httpx.Error(c, http.StatusUnauthorized, "invalid token")
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.Abort()
			httpx.Error(c, http.StatusUnauthorized, "invalid user id")
			return
		}

		user, err := m.users.FindByID(userID)
		if err != nil {
			c.Abort()
			httpx.Error(c, http.StatusUnauthorized, "user not found")
			return
		}

		c.Set("user_id", userID)
		c.Set("role", user.Role)
		c.Set("user", user)
		c.Next()
	}
}

// OptionalAuth attaches user context if a valid JWT is provided, otherwise continues.
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.tokens.Parse(token)
		if err != nil {
			c.Next()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.Next()
			return
		}

		user, err := m.users.FindByID(userID)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", user.Role)
		c.Set("user", user)
		c.Next()
	}
}

// RequireRoles checks that the authenticated user is in one of the roles.
func (m *AuthMiddleware) RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, ok := c.Get("role")
		if !ok {
			c.Abort()
			httpx.Error(c, http.StatusForbidden, "missing role")
			return
		}
		roleStr, ok := role.(string)
		if !ok || roleStr == "" {
			c.Abort()
			httpx.Error(c, http.StatusForbidden, "invalid role")
			return
		}
		if _, ok := allowed[roleStr]; !ok {
			c.Abort()
			httpx.Error(c, http.StatusForbidden, fmt.Sprintf("insufficient privileges: got %s, want one of %v", roleStr, roles))
			return
		}
		c.Next()
	}
}

func extractBearer(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
