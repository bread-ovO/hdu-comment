package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
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
	refreshTTL    time.Duration
	adminEmail    string
}

// NewAuthService constructs an auth service instance.
func NewAuthService(users *repository.UserRepository, tokens *auth.JWTManager, refreshRepo *repository.RefreshTokenRepository, refreshTTL time.Duration, adminEmail string) *AuthService {
	adminEmail = strings.TrimSpace(strings.ToLower(adminEmail))
	return &AuthService{users: users, tokens: tokens, refreshTokens: refreshRepo, refreshTTL: refreshTTL, adminEmail: adminEmail}
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
