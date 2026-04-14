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
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newReviewStatsHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Review{},
		&models.ReviewImage{},
		&models.ReviewStats{},
		&models.ReviewReaction{},
		&models.SiteStats{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return db
}

func newReviewStatsHandlerForTest(t *testing.T) (*gorm.DB, *ReviewStatsHandler, *models.Review) {
	t.Helper()

	db := newReviewStatsHandlerTestDB(t)
	user := &models.User{
		Email:        "user@example.com",
		PasswordHash: "hashed",
		DisplayName:  "user",
		Role:         "user",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	review := &models.Review{
		Title:    "小炒肉",
		Address:  "二食堂",
		Rating:   4.0,
		Status:   models.ReviewStatusPending,
		AuthorID: user.ID,
	}
	if err := db.Create(review).Error; err != nil {
		t.Fatalf("create review failed: %v", err)
	}

	reviewRepo := repository.NewReviewRepository(db)
	reviewService := services.NewReviewService(reviewRepo, nil)
	handler := NewReviewStatsHandler(
		services.NewReviewStatsService(
			repository.NewReviewStatsRepository(db),
			repository.NewReviewReactionRepository(db),
			repository.NewSiteStatsRepository(db),
		),
		reviewService,
	)

	return db, handler, review
}

func TestRecordViewRejectsUnknownReviewID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, handler, _ := newReviewStatsHandlerForTest(t)

	router := gin.New()
	router.POST("/reviews/:id/view", handler.RecordView)

	req := httptest.NewRequest(http.MethodPost, "/reviews/"+uuid.New().String()+"/view", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var statsCount int64
	if err := db.Model(&models.ReviewStats{}).Count(&statsCount).Error; err != nil {
		t.Fatalf("count review stats failed: %v", err)
	}
	if statsCount != 0 {
		t.Fatalf("expected no orphan review stats rows, got %d", statsCount)
	}
}

func TestRecordViewRejectsPendingReview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, handler, review := newReviewStatsHandlerForTest(t)

	router := gin.New()
	router.POST("/reviews/:id/view", handler.RecordView)

	req := httptest.NewRequest(http.MethodPost, "/reviews/"+review.ID.String()+"/view", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if body["error"] != "review not accessible" {
		t.Fatalf("unexpected error response: %+v", body)
	}

	var statsCount int64
	if err := db.Model(&models.ReviewStats{}).Count(&statsCount).Error; err != nil {
		t.Fatalf("count review stats failed: %v", err)
	}
	if statsCount != 0 {
		t.Fatalf("expected no stats rows for pending review, got %d", statsCount)
	}
}
