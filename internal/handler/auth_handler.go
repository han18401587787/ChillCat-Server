package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/jwt"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证接口
type AuthHandler struct {
	userService *service.UserService
	jwtSecret   string
	jwtExpire   int
}

// NewAuthHandler 创建认证接口
func NewAuthHandler(userService *service.UserService, jwtSecret string, jwtExpire int) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
		jwtExpire:   jwtExpire,
	}
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	resp, code, err := h.userService.Register(&req)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, resp)
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	resp, code, err := h.userService.Login(&req)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, resp)
}

// RefreshToken 刷新 Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Error(c, response.ErrUnauthorized)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Error(c, response.ErrUnauthorized)
		return
	}

	claims, err := jwt.ParseTokenLenient(parts[1], h.jwtSecret)
	if err != nil {
		logger.Errorf("刷新Token解析失败: %v", err)
		response.Error(c, response.ErrUnauthorized)
		return
	}

	newToken, err := jwt.GenerateToken(claims.UserID, claims.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		logger.Errorf("刷新Token生成失败: %v", err)
		response.Error(c, response.ErrInternal)
		return
	}

	response.Success(c, gin.H{
		"token":      newToken,
		"expires_in": h.jwtExpire * 3600,
	})
}
