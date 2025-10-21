package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	idVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user"})
		return
	}

	userID, ok := idVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.users.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
		"role":         user.Role,
		"created_at":   user.CreatedAt,
	})
}
