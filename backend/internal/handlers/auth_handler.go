package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/common"
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
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Code        string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if h.emailVerificationService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "注册功能暂不可用"})
		return
	}

	var verification *models.EmailVerification
	var err error
	if verification, err = h.emailVerificationService.ValidateRegistrationCode(c.Request.Context(), req.Email, req.Code); err != nil {
		switch {
		case errors.Is(err, services.ErrVerificationCodeRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "请输入验证码"})
		case errors.Is(err, services.ErrInvalidVerificationToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码不正确"})
		case errors.Is(err, services.ErrVerificationTokenExpired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码已过期，请重新获取"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "验证验证码失败"})
		}
		return
	}

	result, err := h.authService.Register(req.Email, req.Password, req.DisplayName)
	if err != nil {
		switch err {
		case common.ErrEmailAlreadyUsed:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	if h.emailVerificationService != nil && verification != nil {
		if verifiedAt, err := h.emailVerificationService.CompleteRegistrationVerification(c.Request.Context(), verification, result.User.ID); err != nil {
			log.Printf("failed to complete registration verification for user %s: %v", result.User.ID, err)
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case common.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	result, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		switch err {
		case common.ErrInvalidRefreshToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		switch err {
		case common.ErrInvalidRefreshToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			"id":               result.User.ID,
			"email":            result.User.Email,
			"display_name":     result.User.DisplayName,
			"role":             result.User.Role,
			"email_verified":   result.User.EmailVerified,
			"email_verified_at": result.User.EmailVerifiedAt,
			"created_at":       result.User.CreatedAt,
		},
	})
}
