package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"gorm.io/gorm"
)

var (
	// ErrEmailServiceNotConfigured indicates the SMTP service is not ready.
	ErrEmailServiceNotConfigured = errors.New("email service not configured")
	// ErrInvalidVerificationToken indicates the token cannot be found.
	ErrInvalidVerificationToken = errors.New("invalid verification token")
	// ErrVerificationTokenExpired indicates the token has been used or expired.
	ErrVerificationTokenExpired = errors.New("verification token has expired or been used")
	// ErrEmailAlreadyVerified indicates the user has already verified their email.
	ErrEmailAlreadyVerified = errors.New("email already verified")
	// ErrUserNotFound indicates the requested user cannot be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailAlreadyUsed indicates the email has already been registered.
	ErrEmailAlreadyUsed = errors.New("email already used")
	// ErrVerificationCodeRequired indicates that verification code is missing.
	ErrVerificationCodeRequired = errors.New("verification code required")
)

// EmailVerificationService handles email verification business logic.
type EmailVerificationService struct {
	verificationRepo   *repository.EmailVerificationRepository
	userRepo           *repository.UserRepository
	emailService       *EmailService
	verificationBaseURL string
}

// NewEmailVerificationService creates a new EmailVerificationService.
func NewEmailVerificationService(
	emailVerificationRepo *repository.EmailVerificationRepository,
	userRepo *repository.UserRepository,
	emailService *EmailService,
	verificationBaseURL string,
) *EmailVerificationService {
	baseURL := strings.TrimRight(verificationBaseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:5174"
	}

	return &EmailVerificationService{
		verificationRepo:   emailVerificationRepo,
		userRepo:           userRepo,
		emailService:       emailService,
		verificationBaseURL: baseURL,
	}
}

// GenerateVerificationToken generates a secure random token.
func (s *EmailVerificationService) GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateNumericCode(digits int) (string, error) {
	if digits <= 0 {
		return "", fmt.Errorf("invalid code length")
	}

	max := big.NewInt(1)
	for i := 0; i < digits; i++ {
		max.Mul(max, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n.Int64()), nil
}

// SendVerificationEmail sends a verification email to the user.
func (s *EmailVerificationService) SendVerificationEmail(ctx context.Context, userID uuid.UUID, email string) error {
	if s.emailService == nil || !s.emailService.IsConfigured() {
		return ErrEmailServiceNotConfigured
	}

	token, err := s.GenerateVerificationToken()
	if err != nil {
		return fmt.Errorf("generate verification token: %w", err)
	}

	userIDCopy := userID
	verification := &models.EmailVerification{
		UserID:    &userIDCopy,
		Email:     email,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.verificationRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("create verification record: %w", err)
	}

	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.verificationBaseURL, token)
	subject := "请验证您的邮箱地址"
	body := fmt.Sprintf(`
		<h1>邮箱验证</h1>
		<p>您好！感谢您注册我们的服务。</p>
		<p>请点击下面的链接验证您的邮箱地址：</p>
		<p><a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px;">验证邮箱</a></p>
		<p>如果链接无法点击，请复制以下地址到浏览器：</p>
		<p>%s</p>
		<p>此链接将在24小时后过期。</p>
		<p>如果您没有注册我们的服务，请忽略此邮件。</p>
	`, verificationURL, verificationURL)

	if err := s.emailService.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

// VerifyEmail verifies an email using the provided token.
func (s *EmailVerificationService) VerifyEmail(ctx context.Context, token string) error {
	verification, err := s.verificationRepo.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvalidVerificationToken
		}
		return fmt.Errorf("get verification token: %w", err)
	}

	if !verification.IsValid() {
		return ErrVerificationTokenExpired
	}

	if verification.UserID == nil {
		return ErrInvalidVerificationToken
	}

	user, err := s.userRepo.FindByID(*verification.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("find user: %w", err)
	}

	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now

	if err := s.userRepo.Save(user); err != nil {
		return fmt.Errorf("update user verification status: %w", err)
	}

	if err := s.verificationRepo.MarkAsUsed(ctx, token); err != nil {
		return fmt.Errorf("mark token as used: %w", err)
	}

	return nil
}

// SendRegistrationCode sends a verification code to an email before registration.
func (s *EmailVerificationService) SendRegistrationCode(ctx context.Context, email string) error {
	if s.emailService == nil || !s.emailService.IsConfigured() {
		return ErrEmailServiceNotConfigured
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if _, err := s.userRepo.FindByEmail(email); err == nil {
		return ErrEmailAlreadyUsed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("check email: %w", err)
	}

	if err := s.verificationRepo.DeleteByEmail(ctx, email); err != nil {
		return fmt.Errorf("delete existing tokens: %w", err)
	}

	code, err := generateNumericCode(6)
	if err != nil {
		return fmt.Errorf("generate verification code: %w", err)
	}

	verification := &models.EmailVerification{
		Email:     email,
		Token:     code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.verificationRepo.Create(ctx, verification); err != nil {
		return fmt.Errorf("create verification record: %w", err)
	}

	subject := "您的注册验证码"
	body := fmt.Sprintf(`
		<h1>注册验证码</h1>
		<p>您好！您正在注册我们的服务。</p>
		<p>请使用以下验证码完成注册：</p>
		<p style="font-size: 24px; font-weight: bold;">%s</p>
		<p>验证码有效期为10分钟。请勿泄露给他人。</p>
		<p>如果您未发起此请求，请忽略本邮件。</p>
	`, code)

	if err := s.emailService.SendEmail(ctx, email, subject, body); err != nil {
		return fmt.Errorf("send verification email: %w", err)
	}

	return nil
}

// ValidateRegistrationCode validates a registration verification code without consuming it.
func (s *EmailVerificationService) ValidateRegistrationCode(ctx context.Context, email, code string) (*models.EmailVerification, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	code = strings.TrimSpace(code)

	if email == "" || code == "" {
		return nil, ErrVerificationCodeRequired
	}

	verification, err := s.verificationRepo.GetByEmailAndToken(ctx, email, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidVerificationToken
		}
		return nil, fmt.Errorf("get verification token: %w", err)
	}

	if !verification.IsValid() {
		return nil, ErrVerificationTokenExpired
	}

	if verification.UserID != nil {
		return nil, ErrInvalidVerificationToken
	}

	return verification, nil
}

// CompleteRegistrationVerification links a verification token to the user and marks the user as verified.
func (s *EmailVerificationService) CompleteRegistrationVerification(ctx context.Context, verification *models.EmailVerification, userID uuid.UUID) (*time.Time, error) {
	if verification == nil {
		return nil, fmt.Errorf("verification record is required")
	}

	now := time.Now()

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	user.EmailVerified = true
	user.EmailVerifiedAt = &now

	if err := s.userRepo.Save(user); err != nil {
		return nil, fmt.Errorf("update user verification status: %w", err)
	}

	if err := s.verificationRepo.LinkTokenToUser(ctx, verification.ID, userID); err != nil {
		_ = s.verificationRepo.MarkAsUsed(ctx, verification.Token)
		return nil, fmt.Errorf("link verification token: %w", err)
	}

	verification.UserID = &userID
	verification.Used = true

	return &now, nil
}

// ResendVerificationEmail resends a verification email.
func (s *EmailVerificationService) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("find user: %w", err)
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	if err := s.verificationRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("delete existing tokens: %w", err)
	}

	return s.SendVerificationEmail(ctx, user.ID, user.Email)
}

// GetVerificationStatus returns the email verification status for the user.
func (s *EmailVerificationService) GetVerificationStatus(ctx context.Context, userID uuid.UUID) (bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrUserNotFound
		}
		return false, fmt.Errorf("find user: %w", err)
	}
	return user.EmailVerified, nil
}

// CleanExpiredTokens removes expired verification tokens.
func (s *EmailVerificationService) CleanExpiredTokens(ctx context.Context) error {
	return s.verificationRepo.DeleteExpired(ctx)
}
