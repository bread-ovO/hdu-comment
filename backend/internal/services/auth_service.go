package services

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/auth"
	"github.com/hdu-dp/backend/internal/common"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/repository"
	"github.com/hdu-dp/backend/internal/utils"
	"gorm.io/gorm"
)

// AuthResult captures token issuance results.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *models.User
}

// AuthService exposes user registration, login, refresh and logout operations.
type AuthService struct {
	users         *repository.UserRepository
	tokens        *auth.JWTManager
	refreshTokens *repository.RefreshTokenRepository
	smsCodes      *repository.SMSCodeRepository
	qqOAuth       *QQOAuthService
	wechatOAuth   *WeChatOAuthService
	refreshTTL    time.Duration
	smsCodeTTL    time.Duration
	smsEnabled    bool
	smsDevMode    bool
	adminEmail    string
}

// AuthServiceOptions groups options for AuthService initialization.
type AuthServiceOptions struct {
	RefreshTTL time.Duration
	SMSCodeTTL time.Duration
	SMSEnabled bool
	SMSDevMode bool
	AdminEmail string
}

// NewAuthService constructs an auth service instance.
func NewAuthService(
	users *repository.UserRepository,
	tokens *auth.JWTManager,
	refreshRepo *repository.RefreshTokenRepository,
	smsCodeRepo *repository.SMSCodeRepository,
	qqOAuth *QQOAuthService,
	wechatOAuth *WeChatOAuthService,
	options AuthServiceOptions,
) *AuthService {
	if options.SMSCodeTTL <= 0 {
		options.SMSCodeTTL = 10 * time.Minute
	}
	return &AuthService{
		users:         users,
		tokens:        tokens,
		refreshTokens: refreshRepo,
		smsCodes:      smsCodeRepo,
		qqOAuth:       qqOAuth,
		wechatOAuth:   wechatOAuth,
		refreshTTL:    options.RefreshTTL,
		smsCodeTTL:    options.SMSCodeTTL,
		smsEnabled:    options.SMSEnabled,
		smsDevMode:    options.SMSDevMode,
		adminEmail:    strings.TrimSpace(strings.ToLower(options.AdminEmail)),
	}
}

// Register creates a new user account and issues token pair.
func (s *AuthService) Register(email, password, displayName string) (*AuthResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)

	if email == "" || password == "" || displayName == "" {
		return nil, errors.New("invalid registration input")
	}

	if _, err := s.users.FindByEmail(email); err == nil {
		return nil, common.ErrEmailAlreadyUsed
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hashed,
		DisplayName:  displayName,
		Role:         "user",
	}

	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return s.issueTokens(user)
}

// Login validates credentials and returns access/refresh tokens.
func (s *AuthService) Login(email, password string) (*AuthResult, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := utils.CheckPassword(user.PasswordHash, password); err != nil {
		return nil, common.ErrInvalidCredentials
	}

	return s.issueTokens(user)
}

// GetQQLoginURL builds QQ oauth authorize URL.
func (s *AuthService) GetQQLoginURL() (string, string, error) {
	if s.qqOAuth == nil || !s.qqOAuth.IsEnabled() {
		return "", "", common.ErrQQServiceUnavailable
	}
	return s.qqOAuth.BuildAuthURL()
}

// LoginWithQQ authenticates/creates account using QQ oauth code.
func (s *AuthService) LoginWithQQ(code, state string) (*AuthResult, error) {
	if s.qqOAuth == nil || !s.qqOAuth.IsEnabled() {
		return nil, common.ErrQQServiceUnavailable
	}

	profile, err := s.qqOAuth.ExchangeCode(code, state)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByQQOpenID(profile.OpenID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		passwordHash, err := s.hashRandomPassword()
		if err != nil {
			return nil, err
		}

		displayName := strings.TrimSpace(profile.Nickname)
		if displayName == "" {
			displayName = "QQ用户" + shortSuffix(profile.OpenID)
		}

		openIDCopy := profile.OpenID
		user = &models.User{
			ID:           uuid.New(),
			Email:        virtualEmail("qq", profile.OpenID),
			QQOpenID:     &openIDCopy,
			PasswordHash: passwordHash,
			DisplayName:  displayName,
			Role:         "user",
		}

		if err := s.users.Create(user); err != nil {
			return nil, err
		}
	}

	return s.issueTokens(user)
}

// LoginWithWeChat authenticates/creates account using WeChat code.
func (s *AuthService) LoginWithWeChat(code string) (*AuthResult, error) {
	if s.wechatOAuth == nil || !s.wechatOAuth.IsEnabled() {
		return nil, common.ErrWeChatServiceUnavailable
	}

	profile, err := s.wechatOAuth.CodeToSession(code)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByWeChatOpenID(profile.OpenID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		passwordHash, err := s.hashRandomPassword()
		if err != nil {
			return nil, err
		}

		displayName := "微信用户" + shortSuffix(profile.OpenID)

		openIDCopy := profile.OpenID
		user = &models.User{
			ID:           uuid.New(),
			Email:        virtualEmail("wechat", profile.OpenID),
			WeChatOpenID: &openIDCopy,
			PasswordHash: passwordHash,
			DisplayName:  displayName,
			Role:         "user",
		}

		if err := s.users.Create(user); err != nil {
			return nil, err
		}
	}

	return s.issueTokens(user)
}

// SendSMSLoginCode creates and stores one-time login code.
func (s *AuthService) SendSMSLoginCode(phone string) (string, error) {
	if !s.smsEnabled || s.smsCodes == nil {
		return "", common.ErrSMSServiceUnavailable
	}
	if !s.smsDevMode {
		return "", common.ErrSMSServiceUnavailable
	}

	normalized, err := normalizeChinaPhone(phone)
	if err != nil {
		return "", err
	}

	code, err := generateSMSNumericCode(6)
	if err != nil {
		return "", err
	}

	codeHash, err := utils.HashPassword(code)
	if err != nil {
		return "", err
	}

	if err := s.smsCodes.DeleteByPhonePurpose(normalized, smsCodePurposeLogin); err != nil {
		return "", err
	}

	record := &models.SMSCode{
		ID:        uuid.New(),
		Phone:     normalized,
		Purpose:   smsCodePurposeLogin,
		CodeHash:  codeHash,
		ExpiresAt: time.Now().Add(s.smsCodeTTL),
	}
	if err := s.smsCodes.Create(record); err != nil {
		return "", err
	}
	_ = s.smsCodes.DeleteExpired(time.Now())

	slog.Info("sms dev code generated",
		slog.String("phone", normalized),
		slog.String("code", code),
	)
	return code, nil
}

// LoginWithSMS verifies code and returns token pair.
func (s *AuthService) LoginWithSMS(phone, code string) (*AuthResult, error) {
	if !s.smsEnabled || s.smsCodes == nil {
		return nil, common.ErrSMSServiceUnavailable
	}

	normalized, err := normalizeChinaPhone(phone)
	if err != nil {
		return nil, err
	}

	code = strings.TrimSpace(code)
	if !numericCodePattern.MatchString(code) {
		return nil, common.ErrInvalidSMSCode
	}

	record, err := s.smsCodes.FindLatestActive(normalized, smsCodePurposeLogin, time.Now())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrInvalidSMSCode
		}
		return nil, err
	}

	if err := utils.CheckPassword(record.CodeHash, code); err != nil {
		return nil, common.ErrInvalidSMSCode
	}

	if err := s.smsCodes.MarkUsed(record.ID); err != nil {
		return nil, err
	}
	_ = s.smsCodes.DeleteExpired(time.Now())

	user, err := s.users.FindByPhone(normalized)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		passwordHash, err := s.hashRandomPassword()
		if err != nil {
			return nil, err
		}

		phoneCopy := normalized
		user = &models.User{
			ID:           uuid.New(),
			Email:        virtualEmail("phone", normalized),
			Phone:        &phoneCopy,
			PasswordHash: passwordHash,
			DisplayName:  "手机用户" + shortSuffix(normalized),
			Role:         "user",
		}
		if err := s.users.Create(user); err != nil {
			return nil, err
		}
	}

	return s.issueTokens(user)
}

// Refresh validates an existing refresh token and rotates it.
func (s *AuthService) Refresh(token string) (*AuthResult, error) {
	tokenID, secret, err := parseRefreshToken(token)
	if err != nil {
		return nil, common.ErrInvalidRefreshToken
	}

	stored, err := s.refreshTokens.FindByID(tokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrInvalidRefreshToken
		}
		return nil, err
	}

	if stored.Revoked || time.Now().After(stored.ExpiresAt) {
		return nil, common.ErrInvalidRefreshToken
	}

	if err := utils.CheckPassword(stored.SecretHash, secret); err != nil {
		return nil, common.ErrInvalidRefreshToken
	}

	user, err := s.users.FindByID(stored.UserID)
	if err != nil {
		return nil, err
	}

	stored.Revoked = true
	if err := s.refreshTokens.Save(stored); err != nil {
		return nil, err
	}
	_ = s.refreshTokens.DeleteExpired(time.Now())

	return s.issueTokens(user)
}

// Logout revokes the provided refresh token without issuing a new one.
func (s *AuthService) Logout(token string) error {
	tokenID, secret, err := parseRefreshToken(token)
	if err != nil {
		return common.ErrInvalidRefreshToken
	}

	stored, err := s.refreshTokens.FindByID(tokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.ErrInvalidRefreshToken
		}
		return err
	}

	if stored.Revoked {
		return nil
	}

	if err := utils.CheckPassword(stored.SecretHash, secret); err != nil {
		return common.ErrInvalidRefreshToken
	}

	stored.Revoked = true
	return s.refreshTokens.Save(stored)
}

func (s *AuthService) issueTokens(user *models.User) (*AuthResult, error) {
	if s.adminEmail != "" && strings.EqualFold(user.Email, s.adminEmail) && user.Role != "admin" {
		user.Role = "admin"
		if err := s.users.Save(user); err != nil {
			return nil, err
		}
	}

	accessToken, err := s.tokens.Generate(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.createRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{AccessToken: accessToken, RefreshToken: refreshToken, User: user}, nil
}

func (s *AuthService) createRefreshToken(userID uuid.UUID) (string, error) {
	tokenID := uuid.New()
	secret, err := randomSecret()
	if err != nil {
		return "", err
	}

	secretHash, err := utils.HashPassword(secret)
	if err != nil {
		return "", err
	}

	refresh := &models.RefreshToken{
		ID:         tokenID,
		UserID:     userID,
		SecretHash: secretHash,
		ExpiresAt:  time.Now().Add(s.refreshTTL),
		Revoked:    false,
	}

	if err := s.refreshTokens.Create(refresh); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", tokenID.String(), secret), nil
}

func randomSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func parseRefreshToken(token string) (uuid.UUID, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return uuid.Nil, "", errors.New("invalid token format")
	}
	tokenID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", err
	}
	if parts[1] == "" {
		return uuid.Nil, "", errors.New("invalid token secret")
	}
	return tokenID, parts[1], nil
}

const smsCodePurposeLogin = "login"

var numericCodePattern = regexp.MustCompile(`^\d{6}$`)
var chinaPhonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func normalizeChinaPhone(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")

	if strings.HasPrefix(phone, "+86") {
		phone = strings.TrimPrefix(phone, "+86")
	}
	if strings.HasPrefix(phone, "86") && len(phone) == 13 {
		phone = strings.TrimPrefix(phone, "86")
	}

	if !chinaPhonePattern.MatchString(phone) {
		return "", common.ErrInvalidPhoneNumber
	}
	return phone, nil
}

func generateSMSNumericCode(digits int) (string, error) {
	if digits <= 0 {
		return "", errors.New("invalid code length")
	}

	max := big.NewInt(1)
	for i := 0; i < digits; i++ {
		max.Mul(max, big.NewInt(10))
	}

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", digits, n.Int64()), nil
}

func (s *AuthService) hashRandomPassword() (string, error) {
	secret, err := randomSecret()
	if err != nil {
		return "", err
	}
	return utils.HashPassword(secret)
}

func virtualEmail(kind, raw string) string {
	sum := sha1.Sum([]byte(kind + ":" + raw))
	return fmt.Sprintf("%s_%s@local.invalid", kind, hex.EncodeToString(sum[:12]))
}

func shortSuffix(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 4 {
		return value
	}
	return value[len(value)-4:]
}
