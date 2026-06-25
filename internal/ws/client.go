// Package ws WebSocket 客户端连接管理
// 每个连接对应一个在线的已认证用户
package ws

import (
	"chillcat-server/pkg/logger"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 写入超时
	writeWait = 10 * time.Second

	// 读取 pong 的超时
	pongWait = 60 * time.Second

	// 发送 ping 的间隔（必须小于 pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// Client 表示单个 WebSocket 客户端连接
type Client struct {
	UserID int64  // 绑定的用户 ID
	hub    *Hub   // 所属 Hub

	conn *websocket.Conn // WebSocket 连接
	send chan []byte      // 发送消息的缓冲 channel
}

// NewClient 创建一个新的 WebSocket 客户端
func NewClient(userID int64, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		UserID: userID,
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 64),
	}
}

// readPump 从 WebSocket 连接中读取消息
// 每个连接启动一个 goroutine，确保所有读取操作在同一个 goroutine 中进行
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Warnf("WebSocket 读异常 (user_id=%d): %v", c.UserID, err)
			}
			break
		}

		// 尝试解析客户端发送的消息（目前主要处理 ping 心跳）
		var msg map[string]interface{}
		if err := json.Unmarshal(raw, &msg); err != nil {
			logger.Debugf("WebSocket 无法解析消息 (user_id=%d): %s", c.UserID, string(raw))
			continue
		}

		// 客户端心跳处理
		if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
			select {
			case c.send <- []byte(`{"type":"pong"}`):
			default:
			}
		}
	}
}

// writePump 向 WebSocket 连接写入消息
// 每个连接启动一个 goroutine，确保所有写入操作在同一个 goroutine 中进行
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// Hub 关闭了 send channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Warnf("WebSocket 写异常 (user_id=%d): %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SendMessage 向客户端发送一条 JSON 消息
func (c *Client) SendMessage(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		// 发送缓冲区满，丢弃消息
		logger.Warnf("WebSocket 发送缓冲区满 (user_id=%d)，丢弃消息", c.UserID)
		return nil
	}
}
