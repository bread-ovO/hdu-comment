package repository

import (
	"context"
	"testing"

	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
