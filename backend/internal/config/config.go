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
		Port              string
		Mode              string
		ReadHeaderTimeout time.Duration
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		ShutdownTimeout   time.Duration
	}
	Log struct {
		Level  string
		Format string
	}
	Database struct {
		Driver string
		DSN    string
	}
	Auth struct {
		JWTSecret       string
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
		QQ              struct {
			Enabled     bool
			AppID       string
			AppSecret   string
			RedirectURI string
		}
		SMS struct {
			Enabled bool
			CodeTTL time.Duration
			DevMode bool
		}
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
	v.SetDefault("SERVER_READ_HEADER_TIMEOUT", "5s")
	v.SetDefault("SERVER_READ_TIMEOUT", "15s")
	v.SetDefault("SERVER_WRITE_TIMEOUT", "30s")
	v.SetDefault("SERVER_IDLE_TIMEOUT", "60s")
	v.SetDefault("SERVER_SHUTDOWN_TIMEOUT", "10s")

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "text")

	v.SetDefault("DATABASE_DRIVER", "sqlite")
	v.SetDefault("DATABASE_DSN", "file:data/app.db?_fk=1&mode=rwc")

	v.SetDefault("AUTH_ACCESS_TOKEN_TTL", "15m")
	v.SetDefault("AUTH_REFRESH_TOKEN_TTL", "168h")
	v.SetDefault("AUTH_QQ_ENABLED", false)
	v.SetDefault("AUTH_QQ_REDIRECT_URI", "http://localhost:5173/login")
	v.SetDefault("AUTH_SMS_ENABLED", false)
	v.SetDefault("AUTH_SMS_CODE_TTL", "10m")
	v.SetDefault("AUTH_SMS_DEV_MODE", true)

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

	readHeaderTimeout, err := parseDuration(v, "SERVER_READ_HEADER_TIMEOUT")
	if err != nil {
		return nil, fmt.Errorf("invalid READ_HEADER_TIMEOUT: %w", err)
	}

	readTimeout, err := parseDuration(v, "SERVER_READ_TIMEOUT")
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := parseDuration(v, "SERVER_WRITE_TIMEOUT")
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}

	idleTimeout, err := parseDuration(v, "SERVER_IDLE_TIMEOUT")
	if err != nil {
		return nil, fmt.Errorf("invalid IDLE_TIMEOUT: %w", err)
	}

	shutdownTimeout, err := parseDuration(v, "SERVER_SHUTDOWN_TIMEOUT")
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}

	accessTTL, err := parseDuration(v, "AUTH_ACCESS_TOKEN_TTL")
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN ttl: %w", err)
	}

	refreshTTL, err := parseDuration(v, "AUTH_REFRESH_TOKEN_TTL")
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN ttl: %w", err)
	}

	smsCodeTTL, err := parseDuration(v, "AUTH_SMS_CODE_TTL")
	if err != nil {
		return nil, fmt.Errorf("invalid SMS_CODE ttl: %w", err)
	}

	cfg := &Config{}
	cfg.Server.Port = v.GetString("SERVER_PORT")
	cfg.Server.Mode = v.GetString("SERVER_MODE")
	cfg.Server.ReadHeaderTimeout = readHeaderTimeout
	cfg.Server.ReadTimeout = readTimeout
	cfg.Server.WriteTimeout = writeTimeout
	cfg.Server.IdleTimeout = idleTimeout
	cfg.Server.ShutdownTimeout = shutdownTimeout
	cfg.Log.Level = strings.TrimSpace(strings.ToLower(v.GetString("LOG_LEVEL")))
	cfg.Log.Format = strings.TrimSpace(strings.ToLower(v.GetString("LOG_FORMAT")))

	cfg.Database.Driver = v.GetString("DATABASE_DRIVER")
	cfg.Database.DSN = v.GetString("DATABASE_DSN")

	cfg.Auth.JWTSecret = v.GetString("AUTH_JWT_SECRET")
	cfg.Auth.AccessTokenTTL = accessTTL
	cfg.Auth.RefreshTokenTTL = refreshTTL
	cfg.Auth.QQ.Enabled = v.GetBool("AUTH_QQ_ENABLED")
	cfg.Auth.QQ.AppID = strings.TrimSpace(v.GetString("AUTH_QQ_APP_ID"))
	cfg.Auth.QQ.AppSecret = strings.TrimSpace(v.GetString("AUTH_QQ_APP_SECRET"))
	cfg.Auth.QQ.RedirectURI = strings.TrimSpace(v.GetString("AUTH_QQ_REDIRECT_URI"))
	cfg.Auth.SMS.Enabled = v.GetBool("AUTH_SMS_ENABLED")
	cfg.Auth.SMS.CodeTTL = smsCodeTTL
	cfg.Auth.SMS.DevMode = v.GetBool("AUTH_SMS_DEV_MODE")

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

	if cfg.Auth.QQ.Enabled {
		if cfg.Auth.QQ.AppID == "" || cfg.Auth.QQ.AppSecret == "" || cfg.Auth.QQ.RedirectURI == "" {
			return nil, fmt.Errorf("qq login enabled but APP_AUTH_QQ_APP_ID/APP_AUTH_QQ_APP_SECRET/APP_AUTH_QQ_REDIRECT_URI not fully set")
		}
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

func parseDuration(v *viper.Viper, key string) (time.Duration, error) {
	return time.ParseDuration(v.GetString(key))
}
