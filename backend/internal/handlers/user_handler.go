package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/httpx"
	"github.com/hdu-dp/backend/internal/repository"
)

// UserHandler exposes user profile endpoints.
type UserHandler struct {
	users *repository.UserRepository
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(users *repository.UserRepository) *UserHandler {
	return &UserHandler{users: users}
}

// @Summary      获取当前用户信息
// @Description  获取当前已认证用户的详细信息。
// @Tags         用户
// @Produce      json
// @Success      200 {object} object{id=integer,email=string,display_name=string,role=string,created_at=string} "用户信息"
// @Failure      401 {object} object{error=string} "未认证"
// @Failure      404 {object} object{error=string} "用户不存在"
// @Security     ApiKeyAuth
// @Router       /users/me [get]
func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := httpx.MustContextUUID(c, "user_id", "missing user", "invalid user id")
	if !ok {
		return
	}

	user, err := h.users.FindByID(userID)
	if err != nil {
		httpx.Error(c, http.StatusNotFound, "user not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                user.ID,
		"email":             user.Email,
		"phone":             user.Phone,
		"qq_open_id":        user.QQOpenID,
		"display_name":      user.DisplayName,
		"role":              user.Role,
		"email_verified":    user.EmailVerified,
		"email_verified_at": user.EmailVerifiedAt,
		"created_at":        user.CreatedAt,
	})
}
