package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hdu-dp/backend/internal/common"
)

const qqStateTTL = 10 * time.Minute

// QQUserProfile contains required identity fields from QQ OAuth.
type QQUserProfile struct {
	OpenID   string
	Nickname string
}

// QQOAuthService wraps QQ OAuth exchange logic.
type QQOAuthService struct {
	enabled     bool
	appID       string
	appSecret   string
	redirectURI string
	stateSecret string
	httpClient  *http.Client
}

// NewQQOAuthService creates a QQ OAuth service.
func NewQQOAuthService(enabled bool, appID, appSecret, redirectURI, stateSecret string) *QQOAuthService {
	return &QQOAuthService{
		enabled:     enabled,
		appID:       strings.TrimSpace(appID),
		appSecret:   strings.TrimSpace(appSecret),
		redirectURI: strings.TrimSpace(redirectURI),
		stateSecret: stateSecret,
		httpClient:  &http.Client{Timeout: 8 * time.Second},
	}
}

// IsEnabled reports whether QQ OAuth is properly configured.
func (s *QQOAuthService) IsEnabled() bool {
	return s != nil && s.enabled && s.appID != "" && s.appSecret != "" && s.redirectURI != "" && s.stateSecret != ""
}

// BuildAuthURL returns QQ authorize URL and state token.
func (s *QQOAuthService) BuildAuthURL() (string, string, error) {
	if !s.IsEnabled() {
		return "", "", common.ErrQQServiceUnavailable
	}

	state, err := s.generateState()
	if err != nil {
		return "", "", err
	}

	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", s.appID)
	values.Set("redirect_uri", s.redirectURI)
	values.Set("state", state)

	return "https://graph.qq.com/oauth2.0/authorize?" + values.Encode(), state, nil
}

// ExchangeCode exchanges QQ oauth code and returns profile.
func (s *QQOAuthService) ExchangeCode(code, state string) (*QQUserProfile, error) {
	if !s.IsEnabled() {
		return nil, common.ErrQQServiceUnavailable
	}
	if strings.TrimSpace(code) == "" {
		return nil, common.ErrInvalidCredentials
	}
	if err := s.validateState(state); err != nil {
		return nil, common.ErrInvalidQQState
	}

	accessToken, err := s.exchangeToken(code)
	if err != nil {
		return nil, err
	}

	openID, err := s.fetchOpenID(accessToken)
	if err != nil {
		return nil, err
	}

	nickname, err := s.fetchNickname(accessToken, openID)
	if err != nil {
		return nil, err
	}

	return &QQUserProfile{OpenID: openID, Nickname: nickname}, nil
}

func (s *QQOAuthService) exchangeToken(code string) (string, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", s.appID)
	values.Set("client_secret", s.appSecret)
	values.Set("code", code)
	values.Set("redirect_uri", s.redirectURI)

	endpoint := "https://graph.qq.com/oauth2.0/token?" + values.Encode()
	body, err := s.getBody(endpoint)
	if err != nil {
		return "", err
	}

	parsed, err := url.ParseQuery(body)
	if err == nil {
		token := strings.TrimSpace(parsed.Get("access_token"))
		if token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("qq token exchange failed")
}

func (s *QQOAuthService) fetchOpenID(accessToken string) (string, error) {
	endpoint := "https://graph.qq.com/oauth2.0/me?access_token=" + url.QueryEscape(accessToken)
	body, err := s.getBody(endpoint)
	if err != nil {
		return "", err
	}

	rawJSON, err := extractQQCallbackJSON(body)
	if err != nil {
		return "", err
	}

	var payload struct {
		OpenID string `json:"openid"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.OpenID) == "" {
		return "", errors.New("qq openid missing")
	}
	return payload.OpenID, nil
}

func (s *QQOAuthService) fetchNickname(accessToken, openID string) (string, error) {
	values := url.Values{}
	values.Set("access_token", accessToken)
	values.Set("oauth_consumer_key", s.appID)
	values.Set("openid", openID)

	endpoint := "https://graph.qq.com/user/get_user_info?" + values.Encode()
	body, err := s.getBody(endpoint)
	if err != nil {
		return "", err
	}

	var payload struct {
		Ret      int    `json:"ret"`
		Msg      string `json:"msg"`
		Nickname string `json:"nickname"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return "", err
	}
	if payload.Ret != 0 {
		if payload.Msg == "" {
			payload.Msg = "unknown error"
		}
		return "", fmt.Errorf("qq user info failed: %s", payload.Msg)
	}
	return strings.TrimSpace(payload.Nickname), nil
}

func (s *QQOAuthService) generateState() (string, error) {
	nonce, err := randomHex(12)
	if err != nil {
		return "", err
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	payload := nonce + "." + ts
	signature := signState(payload, s.stateSecret)
	return payload + "." + signature, nil
}

func (s *QQOAuthService) validateState(state string) error {
	parts := strings.Split(state, ".")
	if len(parts) != 3 {
		return common.ErrInvalidQQState
	}

	payload := parts[0] + "." + parts[1]
	if signState(payload, s.stateSecret) != parts[2] {
		return common.ErrInvalidQQState
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return common.ErrInvalidQQState
	}

	issuedAt := time.Unix(ts, 0)
	if time.Since(issuedAt) > qqStateTTL || issuedAt.After(time.Now().Add(time.Minute)) {
		return common.ErrInvalidQQState
	}

	return nil
}

func (s *QQOAuthService) getBody(endpoint string) (string, error) {
	resp, err := s.httpClient.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("qq api request failed: status %d", resp.StatusCode)
	}

	return strings.TrimSpace(string(body)), nil
}

func extractQQCallbackJSON(raw string) (string, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return "", errors.New("invalid qq callback payload")
	}
	return raw[start : end+1], nil
}

func signState(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func randomHex(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
