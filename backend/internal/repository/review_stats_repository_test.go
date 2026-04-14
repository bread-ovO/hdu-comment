package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"), time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&models.SiteStats{}); err != nil {
		t.Fatalf("auto migrate site stats failed: %v", err)
	}

	return db
}

func TestSiteStatsRepository_IncrementTotalViews(t *testing.T) {
	db := newTestDB(t)
	repo := NewSiteStatsRepository(db)
	ctx := context.Background()

	if err := repo.IncrementTotalViews(ctx); err != nil {
		t.Fatalf("first increment failed: %v", err)
	}

	if err := repo.IncrementTotalViews(ctx); err != nil {
		t.Fatalf("second increment failed: %v", err)
	}

	totalViews, err := repo.GetTotalViews(ctx)
	if err != nil {
		t.Fatalf("get total views failed: %v", err)
	}

	if totalViews != 2 {
		t.Fatalf("expected total views 2, got %d", totalViews)
	}
}

func TestReviewRepositoryDeleteRemovesStatsAndReactions(t *testing.T) {
	db := newTestDB(t)
	if err := db.AutoMigrate(&models.User{}, &models.Review{}, &models.ReviewImage{}, &models.ReviewStats{}, &models.ReviewReaction{}); err != nil {
		t.Fatalf("auto migrate additional models failed: %v", err)
	}

	user := &models.User{
		Email:        "cleanup@example.com",
		PasswordHash: "hashed",
		DisplayName:  "cleanup",
		Role:         "user",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	review := &models.Review{
		Title:    "牛肉面",
		Address:  "三食堂",
		Rating:   4.2,
		Status:   models.ReviewStatusApproved,
		AuthorID: user.ID,
	}
	if err := db.Create(review).Error; err != nil {
		t.Fatalf("create review failed: %v", err)
	}

	image := &models.ReviewImage{
		ID:         uuid.New(),
		ReviewID:   review.ID,
		StorageKey: "reviews/test.png",
		URL:        "/uploads/test.png",
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image failed: %v", err)
	}

	stats := &models.ReviewStats{
		ReviewID: review.ID,
		Views:    3,
		Likes:    2,
	}
	if err := db.Create(stats).Error; err != nil {
		t.Fatalf("create stats failed: %v", err)
	}

	reaction := &models.ReviewReaction{
		ReviewID: review.ID,
		UserID:   user.ID,
		Type:     models.ReactionTypeLike,
	}
	if err := db.Create(reaction).Error; err != nil {
		t.Fatalf("create reaction failed: %v", err)
	}

	repo := NewReviewRepository(db)
	if err := repo.Delete(review.ID); err != nil {
		t.Fatalf("delete review failed: %v", err)
	}

	var reviewCount int64
	if err := db.Model(&models.Review{}).Where("id = ?", review.ID).Count(&reviewCount).Error; err != nil {
		t.Fatalf("count reviews failed: %v", err)
	}
	if reviewCount != 0 {
		t.Fatalf("expected review to be deleted, got %d rows", reviewCount)
	}

	var imageCount int64
	if err := db.Model(&models.ReviewImage{}).Where("review_id = ?", review.ID).Count(&imageCount).Error; err != nil {
		t.Fatalf("count images failed: %v", err)
	}
	if imageCount != 0 {
		t.Fatalf("expected images to be deleted, got %d rows", imageCount)
	}

	var statsCount int64
	if err := db.Model(&models.ReviewStats{}).Where("review_id = ?", review.ID).Count(&statsCount).Error; err != nil {
		t.Fatalf("count stats failed: %v", err)
	}
	if statsCount != 0 {
		t.Fatalf("expected stats to be deleted, got %d rows", statsCount)
	}

	var reactionCount int64
	if err := db.Model(&models.ReviewReaction{}).Where("review_id = ?", review.ID).Count(&reactionCount).Error; err != nil {
		t.Fatalf("count reactions failed: %v", err)
	}
	if reactionCount != 0 {
		t.Fatalf("expected reactions to be deleted, got %d rows", reactionCount)
	}
}
