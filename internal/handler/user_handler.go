package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户接口
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户接口
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	user, code, err := h.userService.GetProfile(userID)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"status":   user.Status,
	})
}
