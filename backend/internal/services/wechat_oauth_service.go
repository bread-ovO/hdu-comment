package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hdu-dp/backend/internal/common"
)

const weChatInvalidCodeErrCode = 40029

// WeChatUserProfile contains required identity fields from WeChat OAuth.
type WeChatUserProfile struct {
	OpenID     string
	SessionKey string
	UnionID    string
}

// WeChatOAuthService wraps WeChat OAuth exchange logic.
type WeChatOAuthService struct {
	enabled    bool
	appID      string
	appSecret  string
	httpClient *http.Client
}

// NewWeChatOAuthService creates a WeChat OAuth service.
func NewWeChatOAuthService(enabled bool, appID, appSecret string) *WeChatOAuthService {
	return &WeChatOAuthService{
		enabled:    enabled,
		appID:      strings.TrimSpace(appID),
		appSecret:  strings.TrimSpace(appSecret),
		httpClient: &http.Client{Timeout: 8 * time.Second},
	}
}

// IsEnabled reports whether WeChat OAuth is properly configured.
func (s *WeChatOAuthService) IsEnabled() bool {
	return s != nil && s.enabled && s.appID != "" && s.appSecret != ""
}

// CodeToSession exchanges WeChat login code for session info.
func (s *WeChatOAuthService) CodeToSession(code string) (*WeChatUserProfile, error) {
	if !s.IsEnabled() {
		return nil, common.ErrWeChatServiceUnavailable
	}
	if strings.TrimSpace(code) == "" {
		return nil, common.ErrInvalidWeChatCode
	}

	values := url.Values{}
	values.Set("appid", s.appID)
	values.Set("secret", s.appSecret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	endpoint := "https://api.weixin.qq.com/sns/jscode2session?" + values.Encode()
	body, err := s.getBody(endpoint)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Errcode    int    `json:"errcode"`
		Errmsg     string `json:"errmsg"`
		OpenID     string `json:"openid"`
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
	}

	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return nil, fmt.Errorf("wechat session parse failed: %w", err)
	}

	if resp.Errcode != 0 {
		if resp.Errcode == weChatInvalidCodeErrCode {
			return nil, common.ErrInvalidWeChatCode
		}
		return nil, fmt.Errorf("wechat api error: %s", resp.Errmsg)
	}

	if strings.TrimSpace(resp.OpenID) == "" {
		return nil, errors.New("wechat openid missing")
	}

	return &WeChatUserProfile{
		OpenID:     resp.OpenID,
		SessionKey: resp.SessionKey,
		UnionID:    resp.UnionID,
	}, nil
}

func (s *WeChatOAuthService) getBody(endpoint string) (string, error) {
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
		return "", fmt.Errorf("wechat api request failed: status %d", resp.StatusCode)
	}

	return strings.TrimSpace(string(body)), nil
}
