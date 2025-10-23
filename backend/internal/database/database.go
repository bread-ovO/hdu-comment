package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hdu-dp/backend/internal/config"
	"github.com/hdu-dp/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Init connects to the database and performs migrations.
func Init(cfg *config.Config) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	switch cfg.Database.Driver {
	case "sqlite":
		if err = prepareSQLite(cfg.Database.DSN); err != nil {
			return nil, err
		}
		db, err = gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&models.User{}, &models.Review{}, &models.ReviewImage{}, &models.RefreshToken{}, &models.ReviewStats{}, &models.ReviewReaction{}, &models.SiteStats{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	if err = seedAdmin(db, cfg); err != nil {
		return nil, err
	}

	if cfg.Storage.Provider == "local" {
		if err = ensureDir(cfg.Storage.UploadDir); err != nil {
			return nil, fmt.Errorf("ensure upload dir: %w", err)
		}
	}

	log.Printf("database initialised with driver %s", cfg.Database.Driver)
	return db, nil
}

func ensureDir(path string) error {
	if path == "" || path == "." {
		return nil
	}
	if err := os.MkdirAll(path, 0o755); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func prepareSQLite(dsn string) error {
	path := dsn
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	path = strings.TrimPrefix(path, "file:")
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return ensureDir(dir)
}
