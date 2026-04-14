package services

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newReviewStatsServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&models.ReviewStats{}, &models.ReviewReaction{}, &models.SiteStats{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return db
}

func TestToggleReactionCreatesStatsRowOnFirstReaction(t *testing.T) {
	db := newReviewStatsServiceTestDB(t)
	service := NewReviewStatsService(
		repository.NewReviewStatsRepository(db),
		repository.NewReviewReactionRepository(db),
		repository.NewSiteStatsRepository(db),
	)

	reviewID := uuid.New()
	userID := uuid.New()

	if err := service.ToggleReaction(t.Context(), reviewID, userID, models.ReactionTypeLike); err != nil {
		t.Fatalf("toggle reaction failed: %v", err)
	}

	var stats models.ReviewStats
	if err := db.Where("review_id = ?", reviewID).First(&stats).Error; err != nil {
		t.Fatalf("load review stats failed: %v", err)
	}
	if stats.Likes != 1 || stats.Dislikes != 0 {
		t.Fatalf("unexpected stats after first reaction: %+v", stats)
	}
}
