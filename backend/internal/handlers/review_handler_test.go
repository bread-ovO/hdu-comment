package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newReviewHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}, &models.Review{}, &models.ReviewImage{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return db
}

func TestUploadImageRejectsProcessedReview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := newReviewHandlerTestDB(t)
	user := &models.User{
		Email:        "owner@example.com",
		PasswordHash: "hashed",
		DisplayName:  "owner",
		Role:         "user",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	review := &models.Review{
		Title:    "红烧鸡腿饭",
		Address:  "一食堂",
		Rating:   4.5,
		Status:   models.ReviewStatusApproved,
		AuthorID: user.ID,
	}
	if err := db.Create(review).Error; err != nil {
		t.Fatalf("create review failed: %v", err)
	}

	handler := NewReviewHandler(services.NewReviewService(repository.NewReviewRepository(db), nil))

	router := gin.New()
	router.POST("/reviews/:id/images", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.UploadImage(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/reviews/"+review.ID.String()+"/images", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if body["error"] != "images can only be uploaded while review is pending" {
		t.Fatalf("unexpected error response: %+v", body)
	}
}
