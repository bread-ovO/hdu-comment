package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/common"
	"github.com/hdu-dp/backend/internal/httpx"
	"github.com/hdu-dp/backend/internal/models"
	"github.com/hdu-dp/backend/internal/services"
)

// AuthHandler exposes HTTP endpoints for authentication flows.
type AuthHandler struct {
	authService              *services.AuthService
	emailVerificationService *services.EmailVerificationService
}

// NewAuthHandler constructs an AuthHandler.
func NewAuthHandler(authService *services.AuthService, emailVerificationService *services.EmailVerificationService) *AuthHandler {
	return &AuthHandler{
		authService:              authService,
		emailVerificationService: emailVerificationService,
	}
}

// @Summary      用户注册
// @Description  接收用户邮箱、密码和昵称进行注册，成功后返回认证信息。
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body body object{email=string,password=string,display_name=string,code=string} true "注册信息"
// @Success      201  {object} object{access_token=string,refresh_token=string,user=object{id=integer,email=string,display_name=string,role=string,created_at=string,email_verified=bool}} "注册成功"
// @Failure      400  {object} object{error=string} "请求参数错误"
// @Failure      409  {object} object{error=string} "邮箱已被占用"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		DisplayName string `json:"display_name" binding:"required,max=64"`
		Code        string `json:"code" binding:"required,len=6"`
	}

	if !httpx.BindJSON(c, &req, "请输入完整且有效的注册信息") {
		return
	}

	if h.emailVerificationService == nil {
		httpx.Error(c, http.StatusServiceUnavailable, "注册功能暂不可用")
		return
	}

	var verification *models.EmailVerification
	var err error
	if verification, err = h.emailVerificationService.ValidateRegistrationCode(c.Request.Context(), req.Email, req.Code); err != nil {
		switch {
		case errors.Is(err, services.ErrVerificationCodeRequired):
			httpx.Error(c, http.StatusBadRequest, "请输入验证码")
		case errors.Is(err, services.ErrInvalidVerificationToken):
			httpx.Error(c, http.StatusBadRequest, "验证码不正确")
		case errors.Is(err, services.ErrVerificationTokenExpired):
			httpx.Error(c, http.StatusBadRequest, "验证码已过期，请重新获取")
		default:
			httpx.Error(c, http.StatusInternalServerError, "验证验证码失败")
		}
		return
	}

	result, err := h.authService.Register(req.Email, req.Password, req.DisplayName)
	if err != nil {
		switch err {
		case common.ErrEmailAlreadyUsed:
			httpx.Error(c, http.StatusConflict, err.Error())
		default:
			httpx.Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	if h.emailVerificationService != nil && verification != nil {
		if verifiedAt, err := h.emailVerificationService.CompleteRegistrationVerification(c.Request.Context(), verification, result.User.ID); err != nil {
			slog.Warn("failed to complete registration verification",
				slog.String("user_id", result.User.ID.String()),
				slog.Any("error", err),
			)
		} else {
			result.User.EmailVerified = true
			result.User.EmailVerifiedAt = verifiedAt
		}
	}

	respondAuthSuccess(c, http.StatusCreated, result)
}

// @Summary      用户登录
// @Description  接收用户邮箱和密码进行登录，成功后返回认证信息。
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body body object{email=string,password=string} true "登录信息"
// @Success      200  {object} object{access_token=string,refresh_token=string,user=object{id=integer,email=string,display_name=string,role=string,created_at=string,email_verified=bool}} "登录成功"
// @Failure      400  {object} object{error=string} "请求参数错误"
// @Failure      401  {object} object{error=string} "邮箱或密码错误"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if !httpx.BindJSON(c, &req, "请输入有效的邮箱和密码") {
		return
	}

	result, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case common.ErrInvalidCredentials:
			httpx.Error(c, http.StatusUnauthorized, err.Error())
		default:
			httpx.Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// QQAuthURL returns a QQ oauth authorize URL.
func (h *AuthHandler) QQAuthURL(c *gin.Context) {
	url, state, err := h.authService.GetQQLoginURL()
	if err != nil {
		switch err {
		case common.ErrQQServiceUnavailable:
			httpx.Error(c, http.StatusServiceUnavailable, "QQ 登录暂不可用")
		default:
			httpx.Error(c, http.StatusInternalServerError, "生成 QQ 登录链接失败")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":   url,
		"state": state,
	})
}

// QQLogin handles login with QQ oauth code.
func (h *AuthHandler) QQLogin(c *gin.Context) {
	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}

	if !httpx.BindJSON(c, &req, "缺少 QQ 登录参数") {
		return
	}

	result, err := h.authService.LoginWithQQ(req.Code, req.State)
	if err != nil {
		switch err {
		case common.ErrQQServiceUnavailable:
			httpx.Error(c, http.StatusServiceUnavailable, "QQ 登录暂不可用")
		case common.ErrInvalidQQState:
			httpx.Error(c, http.StatusBadRequest, "QQ 登录态已失效，请重试")
		default:
			httpx.Error(c, http.StatusUnauthorized, "QQ 登录失败")
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// WeChatLogin handles login with WeChat code.
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if !httpx.BindJSON(c, &req, "缺少微信登录参数") {
		return
	}

	result, err := h.authService.LoginWithWeChat(req.Code)
	if err != nil {
		slog.Error("wechat login failed", slog.Any("error", err))

		switch err {
		case common.ErrWeChatServiceUnavailable:
			httpx.Error(c, http.StatusServiceUnavailable, "微信登录暂不可用")
		case common.ErrInvalidWeChatCode:
			httpx.Error(c, http.StatusBadRequest, "微信登录凭证无效")
		default:
			httpx.Error(c, http.StatusUnauthorized, "微信登录失败")
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// SendSMSCode sends login verification code to a phone number.
func (h *AuthHandler) SendSMSCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}
	if !httpx.BindJSON(c, &req, "请输入手机号") {
		return
	}

	code, err := h.authService.SendSMSLoginCode(req.Phone)
	if err != nil {
		switch err {
		case common.ErrInvalidPhoneNumber:
			httpx.Error(c, http.StatusBadRequest, "手机号格式不正确")
		case common.ErrSMSServiceUnavailable:
			httpx.Error(c, http.StatusServiceUnavailable, "短信登录暂不可用")
		default:
			httpx.Error(c, http.StatusInternalServerError, "发送短信验证码失败")
		}
		return
	}

	resp := gin.H{"message": "验证码已发送"}
	if code != "" {
		resp["debug_code"] = code
	}
	c.JSON(http.StatusOK, resp)
}

// SMSLogin logs in user by phone and sms code.
func (h *AuthHandler) SMSLogin(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}
	if !httpx.BindJSON(c, &req, "请输入手机号和验证码") {
		return
	}

	result, err := h.authService.LoginWithSMS(req.Phone, req.Code)
	if err != nil {
		switch err {
		case common.ErrInvalidPhoneNumber:
			httpx.Error(c, http.StatusBadRequest, "手机号格式不正确")
		case common.ErrInvalidSMSCode:
			httpx.Error(c, http.StatusUnauthorized, "验证码错误或已过期")
		case common.ErrSMSServiceUnavailable:
			httpx.Error(c, http.StatusServiceUnavailable, "短信登录暂不可用")
		default:
			httpx.Error(c, http.StatusInternalServerError, "短信登录失败")
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// @Summary      刷新令牌
// @Description  使用有效的刷新令牌获取新的访问令牌和刷新令牌。
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body body object{refresh_token=string} true "刷新令牌"
// @Success      200  {object} object{access_token=string,refresh_token=string,user=object{id=integer,email=string,display_name=string,role=string,created_at=string,email_verified=bool}} "刷新成功"
// @Failure      400  {object} object{error=string} "请求参数错误"
// @Failure      401  {object} object{error=string} "无效的刷新令牌"
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if !httpx.BindJSON(c, &req, "refresh_token 不能为空") {
		return
	}

	result, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		switch err {
		case common.ErrInvalidRefreshToken:
			httpx.Error(c, http.StatusUnauthorized, err.Error())
		default:
			httpx.Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	respondAuthSuccess(c, http.StatusOK, result)
}

// @Summary      用户登出
// @Description  接收刷新令牌并使其失效。
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body body object{refresh_token=string} true "刷新令牌"
// @Success      204 "登出成功"
// @Failure      400  {object} object{error=string} "请求参数错误"
// @Failure      401  {object} object{error=string} "无效的刷新令牌"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if !httpx.BindJSON(c, &req, "refresh_token 不能为空") {
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		switch err {
		case common.ErrInvalidRefreshToken:
			httpx.Error(c, http.StatusUnauthorized, err.Error())
		default:
			httpx.Error(c, http.StatusBadRequest, err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func respondAuthSuccess(c *gin.Context, status int, result *services.AuthResult) {
	c.JSON(status, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"user": gin.H{
			"id":                result.User.ID,
			"email":             result.User.Email,
			"phone":             result.User.Phone,
			"qq_open_id":        result.User.QQOpenID,
			"wechat_open_id":    result.User.WeChatOpenID,
			"display_name":      result.User.DisplayName,
			"role":              result.User.Role,
			"email_verified":    result.User.EmailVerified,
			"email_verified_at": result.User.EmailVerifiedAt,
			"created_at":        result.User.CreatedAt,
		},
	})
}
