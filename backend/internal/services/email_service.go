package services

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/hdu-dp/backend/internal/config"
)

// EmailService handles email sending functionality
type EmailService struct {
	config *config.EmailConfig
}

// NewEmailService creates a new EmailService
func NewEmailService(config *config.EmailConfig) *EmailService {
	return &EmailService{config: config}
}

// SendEmail sends an email to the specified recipient
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	if !s.config.IsValid() {
		return fmt.Errorf("email service not properly configured")
	}

	from := s.config.FromEmail
	password := s.config.SMTPPassword

	// 构建邮件内容
	msg := fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += body

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	auth := smtp.PlainAuth("", s.config.SMTPUsername, password, s.config.SMTPHost)

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

// IsConfigured checks if email service is properly configured
func (s *EmailService) IsConfigured() bool {
	if s == nil || s.config == nil {
		return false
	}
	return s.config.IsValid()
}
