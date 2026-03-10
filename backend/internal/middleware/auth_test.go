package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/auth"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type optionalAuthResponse struct {
	HasUserID bool `json:"has_user_id"`
	HasRole   bool `json:"has_role"`
}

func newAuthMiddlewareForTest(t *testing.T) (*AuthMiddleware, *models.User, *auth.JWTManager) {
	t.Helper()

	dbName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", dbName, time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hashed",
		DisplayName:  "test-user",
		Role:         "user",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	users := repository.NewUserRepository(db)
	tokens := auth.NewJWTManager("test-secret", time.Hour)
	return NewAuthMiddleware(tokens, users), user, tokens
}

func TestOptionalAuthAllowsAnonymousWhenTokenInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw, _, _ := newAuthMiddlewareForTest(t)

	router := gin.New()
	router.GET("/detail", mw.OptionalAuth(), func(c *gin.Context) {
		_, hasUserID := c.Get("user_id")
		_, hasRole := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"has_user_id": hasUserID, "has_role": hasRole})
	})

	req := httptest.NewRequest(http.MethodGet, "/detail", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for invalid optional auth token, got %d", rec.Code)
	}

	var body optionalAuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.HasUserID || body.HasRole {
		t.Fatalf("expected invalid token to be treated as anonymous, got %+v", body)
	}
}

func TestOptionalAuthAttachesContextWhenTokenValid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mw, user, tokens := newAuthMiddlewareForTest(t)

	token, err := tokens.Generate(user)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := gin.New()
	router.GET("/detail", mw.OptionalAuth(), func(c *gin.Context) {
		_, hasUserID := c.Get("user_id")
		_, hasRole := c.Get("role")
		c.JSON(http.StatusOK, gin.H{"has_user_id": hasUserID, "has_role": hasRole})
	})

	req := httptest.NewRequest(http.MethodGet, "/detail", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 for valid token, got %d", rec.Code)
	}

	var body optionalAuthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !body.HasUserID || !body.HasRole {
		t.Fatalf("expected valid token to attach user context, got %+v", body)
	}
}
