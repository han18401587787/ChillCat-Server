// Package handler WebSocket 升级处理
// GET /api/v1/ws — 将 HTTP 连接升级为 WebSocket，支持 JWT 认证
package handler

import (
	"chillcat-server/internal/ws"
	"chillcat-server/pkg/jwt"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	gorillaWS "github.com/gorilla/websocket"
)

var upgrader = gorillaWS.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有来源（内部使用，安全由 JWT 保障）
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSHandler WebSocket 升级处理器
type WSHandler struct {
	hub       *ws.Hub
	jwtSecret string
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub *ws.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret}
}

// Upgrade 处理 WebSocket 升级请求
// GET /api/v1/ws?token=<jwt_token>
// 在升级过程中验证 JWT，无需 Auth 中间件
func (h *WSHandler) Upgrade(c *gin.Context) {
	// 1. 从 query 参数提取并验证 JWT token
	token := c.Query("token")
	if token == "" {
		response.Error(c, response.ErrUnauthorized)
		return
	}

	claims, err := jwt.ParseToken(token, h.jwtSecret)
	if err != nil {
		logger.Warnf("WebSocket JWT 验证失败: %v", err)
		response.Error(c, response.ErrUnauthorized)
		return
	}

	userID := claims.UserID

	// 2. HTTP → WebSocket 协议升级
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Warnf("WebSocket 升级失败 (user_id=%d): %v", userID, err)
		// 此时 HTTP 响应头已发送，无法再返回错误 JSON
		return
	}

	// 3. 创建客户端并注册到 Hub（Hub 会自动启动读写协程）
	client := ws.NewClient(userID, h.hub, conn)
	h.hub.Register(client)

	logger.Infof("WebSocket 连接已建立 (user_id=%d)", userID)
}
