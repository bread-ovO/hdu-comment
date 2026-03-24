package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hdu-dp/backend/internal/httpx"
	"github.com/hdu-dp/backend/internal/services"
)

// EmailVerificationHandler handles email verification HTTP endpoints
type EmailVerificationHandler struct {
	emailVerificationService *services.EmailVerificationService
}

// NewEmailVerificationHandler creates a new EmailVerificationHandler
func NewEmailVerificationHandler(emailVerificationService *services.EmailVerificationService) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		emailVerificationService: emailVerificationService,
	}
}

// @Summary      发送注册验证码
// @Description  向指定邮箱发送注册验证码
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body body object{email=string} true "邮箱地址"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Router       /auth/send-code [post]
func (h *EmailVerificationHandler) SendRegistrationCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if !httpx.BindJSON(c, &req, "请输入有效的邮箱地址") {
		return
	}

	if err := h.emailVerificationService.SendRegistrationCode(c.Request.Context(), req.Email); err != nil {
		switch {
		case errors.Is(err, services.ErrEmailAlreadyUsed):
			httpx.Error(c, http.StatusConflict, "该邮箱已注册")
		case errors.Is(err, services.ErrEmailServiceNotConfigured):
			httpx.Error(c, http.StatusInternalServerError, "邮件服务未配置，请联系管理员")
		default:
			httpx.Error(c, http.StatusInternalServerError, "发送验证码失败，请稍后重试")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送，请查收邮箱"})
}

// @Summary      发送邮箱验证邮件
// @Description  向当前用户发送邮箱验证邮件
// @Tags         认证
// @Produce      json
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Failure      401 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Security     ApiKeyAuth
// @Router       /auth/send-verification [post]
func (h *EmailVerificationHandler) SendVerificationEmail(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.emailVerificationService.ResendVerificationEmail(c.Request.Context(), userID); err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound):
			httpx.Error(c, http.StatusBadRequest, "用户不存在")
		case errors.Is(err, services.ErrEmailAlreadyVerified):
			httpx.Error(c, http.StatusConflict, "邮箱已验证")
		case errors.Is(err, services.ErrEmailServiceNotConfigured):
			httpx.Error(c, http.StatusInternalServerError, "邮件服务未配置，请联系管理员")
		default:
			httpx.Error(c, http.StatusInternalServerError, "发送失败，请稍后重试")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证邮件已发送，请检查您的邮箱"})
}

// @Summary      验证邮箱
// @Description  使用验证令牌验证用户邮箱
// @Tags         认证
// @Accept       json
// @Produce      json
// @Param        body body object{token=string} true "验证令牌"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /auth/verify-email [post]
func (h *EmailVerificationHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if !httpx.BindJSON(c, &req, "无效的请求参数") {
		return
	}

	if err := h.emailVerificationService.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidVerificationToken):
			httpx.Error(c, http.StatusNotFound, "无效的验证令牌")
		case errors.Is(err, services.ErrVerificationTokenExpired):
			httpx.Error(c, http.StatusBadRequest, "验证令牌已过期或已使用")
		case errors.Is(err, services.ErrUserNotFound):
			httpx.Error(c, http.StatusBadRequest, "用户不存在")
		default:
			httpx.Error(c, http.StatusInternalServerError, "验证失败，请稍后重试")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "邮箱验证成功"})
}

// @Summary      获取邮箱验证状态
// @Description  获取当前用户的邮箱验证状态
// @Tags         认证
// @Produce      json
// @Success      200 {object} object{email_verified=bool}
// @Failure      401 {object} object{error=string}
// @Security     ApiKeyAuth
// @Router       /auth/verification-status [get]
func (h *EmailVerificationHandler) GetVerificationStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	verified, err := h.emailVerificationService.GetVerificationStatus(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			httpx.Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		httpx.Error(c, http.StatusInternalServerError, "获取用户信息失败")
		return
	}

	c.JSON(http.StatusOK, gin.H{"email_verified": verified})
}
