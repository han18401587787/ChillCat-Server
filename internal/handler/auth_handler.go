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

// RefreshTokenRequest 刷新 Token 请求体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新 Token 响应
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// RefreshToken 刷新 Token
// 支持两种方式：
// 1. Body 中传 refresh_token（优先，客户端新逻辑）
// 2. Authorization header 中传旧 access_token（兼容旧客户端）
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var claims *jwt.Claims
	var err error

	// 优先从 body 解析 refresh_token
	var req RefreshTokenRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr == nil && req.RefreshToken != "" {
		// 使用宽容模式解析 refresh token（允许过期 token）
		claims, err = jwt.ParseTokenLenient(req.RefreshToken, h.jwtSecret)
		if err != nil {
			logger.Errorf("RefreshToken body 解析失败: %v", err)
			response.Error(c, response.ErrUnauthorized)
			return
		}
	} else {
		// 兼容旧方式：从 Authorization header 解析
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

		claims, err = jwt.ParseTokenLenient(parts[1], h.jwtSecret)
		if err != nil {
			logger.Errorf("RefreshToken header 解析失败: %v", err)
			response.Error(c, response.ErrUnauthorized)
			return
		}
	}

	// 生成新的 access token
	newAccessToken, err := jwt.GenerateToken(claims.UserID, claims.Username, h.jwtSecret, h.jwtExpire)
	if err != nil {
		logger.Errorf("刷新Token生成 access_token 失败: %v", err)
		response.Error(c, response.ErrInternal)
		return
	}

	// 生成新的 refresh token（有效期更长：7天）
	refreshExpireHour := h.jwtExpire * 3 // 72h * 3 = 9天
	if refreshExpireHour > 168 {
		refreshExpireHour = 168 // 最多 7 天
	}
	newRefreshToken, err := jwt.GenerateRefreshToken(claims.UserID, claims.Username, h.jwtSecret, refreshExpireHour)
	if err != nil {
		logger.Errorf("刷新Token生成 refresh_token 失败: %v", err)
		response.Error(c, response.ErrInternal)
		return
	}

	logger.Infof("Token 刷新成功: user_id=%d", claims.UserID)

	response.Success(c, RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    h.jwtExpire * 3600,
	})
}

// AnonymousLogin 匿名登录
func (h *AuthHandler) AnonymousLogin(c *gin.Context) {
	resp, code, err := h.userService.AnonymousRegister()
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, resp)
}
