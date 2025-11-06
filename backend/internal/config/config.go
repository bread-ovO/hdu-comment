package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents application level configuration values.
type Config struct {
	Server struct {
		Port string
		Mode string
	}
	Database struct {
		Driver string
		DSN    string
	}
	Auth struct {
		JWTSecret       string
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
	}
	Storage struct {
		Provider      string
		UploadDir     string
		PublicBaseURL string
		S3            struct {
			Endpoint  string
			Bucket    string
			Region    string
			AccessKey string
			SecretKey string
			UseSSL    bool
			BaseURL   string
		}
	}
	Admin struct {
		Email    string
		Password string
	}
	CORS struct {
		AllowOrigins []string
	}
}

// Load reads configuration from environment variables with sane defaults.
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("SERVER_MODE", "release")

	v.SetDefault("DATABASE_DRIVER", "sqlite")
	v.SetDefault("DATABASE_DSN", "file:data/app.db?_fk=1&mode=rwc")

	v.SetDefault("AUTH_ACCESS_TOKEN_TTL", "15m")
	v.SetDefault("AUTH_REFRESH_TOKEN_TTL", "168h")

	v.SetDefault("STORAGE_PROVIDER", "local")
	v.SetDefault("STORAGE_UPLOAD_DIR", "uploads")
	v.SetDefault("STORAGE_PUBLIC_BASE_URL", "/api/v1/uploads")

	v.SetDefault("STORAGE_S3_ENDPOINT", "")
	v.SetDefault("STORAGE_S3_BUCKET", "")
	v.SetDefault("STORAGE_S3_REGION", "")
	v.SetDefault("STORAGE_S3_ACCESS_KEY", "")
	v.SetDefault("STORAGE_S3_SECRET_KEY", "")
	v.SetDefault("STORAGE_S3_USE_SSL", true)
	v.SetDefault("STORAGE_S3_BASE_URL", "")
	v.SetDefault("CORS_ALLOW_ORIGINS", "http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173,http://127.0.0.1:5174,https://hddp.blueloaf.top")

	accessTTL, err := time.ParseDuration(v.GetString("AUTH_ACCESS_TOKEN_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN ttl: %w", err)
	}

	refreshTTL, err := time.ParseDuration(v.GetString("AUTH_REFRESH_TOKEN_TTL"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN ttl: %w", err)
	}

	cfg := &Config{}
	cfg.Server.Port = v.GetString("SERVER_PORT")
	cfg.Server.Mode = v.GetString("SERVER_MODE")

	cfg.Database.Driver = v.GetString("DATABASE_DRIVER")
	cfg.Database.DSN = v.GetString("DATABASE_DSN")

	cfg.Auth.JWTSecret = v.GetString("AUTH_JWT_SECRET")
	cfg.Auth.AccessTokenTTL = accessTTL
	cfg.Auth.RefreshTokenTTL = refreshTTL

	cfg.Storage.Provider = v.GetString("STORAGE_PROVIDER")
	cfg.Storage.UploadDir = v.GetString("STORAGE_UPLOAD_DIR")
	cfg.Storage.PublicBaseURL = v.GetString("STORAGE_PUBLIC_BASE_URL")
	cfg.Storage.S3.Endpoint = v.GetString("STORAGE_S3_ENDPOINT")
	cfg.Storage.S3.Bucket = v.GetString("STORAGE_S3_BUCKET")
	cfg.Storage.S3.Region = v.GetString("STORAGE_S3_REGION")
	cfg.Storage.S3.AccessKey = v.GetString("STORAGE_S3_ACCESS_KEY")
	cfg.Storage.S3.SecretKey = v.GetString("STORAGE_S3_SECRET_KEY")
	cfg.Storage.S3.UseSSL = v.GetBool("STORAGE_S3_USE_SSL")
	cfg.Storage.S3.BaseURL = v.GetString("STORAGE_S3_BASE_URL")

	cfg.Admin.Email = strings.TrimSpace(strings.ToLower(v.GetString("ADMIN_EMAIL")))
	cfg.Admin.Password = strings.TrimSpace(v.GetString("ADMIN_PASSWORD"))
	cfg.CORS.AllowOrigins = splitAndClean(v.GetString("CORS_ALLOW_ORIGINS"))

	if cfg.Auth.JWTSecret == "" {
		return nil, fmt.Errorf("missing auth jwt secret: set APP_AUTH_JWT_SECRET")
	}

	return cfg, nil
}

func splitAndClean(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
