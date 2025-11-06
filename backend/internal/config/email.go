package config

import (
	"fmt"
	"os"
)

// EmailConfig holds email service configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	FrontendBaseURL string
}

// LoadEmailConfig loads email configuration from environment variables
func LoadEmailConfig() *EmailConfig {
	return &EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@hdu-food-review.com"),
		FromName:     getEnv("FROM_NAME", "HDU美食点评"),
		FrontendBaseURL: getEnv("FRONTEND_BASE_URL", "http://localhost:5174"),
	}
}

// IsValid checks if email configuration is valid
func (c *EmailConfig) IsValid() bool {
	return c.SMTPHost != "" && c.SMTPUsername != "" && c.SMTPPassword != ""
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}
