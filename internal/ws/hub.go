// Package ws WebSocket 连接管理中心
// Hub 维护所有活跃的 WebSocket 客户端连接，处理注册/注销/广播
// 同时订阅 Redis Pub/Sub 频道，实现跨实例消息广播
package ws

import (
	"chillcat-server/internal/cache"
	"chillcat-server/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Message WebSocket 推送消息格式
type Message struct {
	Type string      `json:"type"` // 消息类型：resonance.new / treehole.reply / encourage.joined
	Data interface{} `json:"data"` // 消息负载
}

// Hub 维护所有活跃的 WebSocket 连接
type Hub struct {
	// 已注册的客户端映射：user_id → Client
	// 同一个用户只能有一个活跃 WebSocket 连接
	clients    map[int64]*Client
	clientsMu  sync.RWMutex

	broadcast  chan Message   // 广播消息 channel
	register   chan *Client   // 客户端注册 channel
	unregister chan *Client   // 客户端注销 channel

	rdb *cache.RedisClient // Redis 客户端，用于 Pub/Sub
}

// NewHub 创建一个新的 Hub
func NewHub(rdb *cache.RedisClient) *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rdb:        rdb,
	}
}

// Register 注册一个客户端到 Hub（供 handler 调用）
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Run 启动 Hub 的主事件循环
// 在单独的 goroutine 中调用
func (h *Hub) Run(ctx context.Context) {
	logger.Info("🔌 WebSocket Hub 已启动")

	// 订阅 Redis Pub/Sub 频道（如果 Redis 可用）
	if h.rdb != nil && h.rdb.IsOK() {
		go h.subscribeRedis(ctx)
	}

	for {
		select {
		case client := <-h.register:
			h.clientsMu.Lock()
			// 如果同一用户已有旧连接，先关闭旧连接
			if oldClient, ok := h.clients[client.UserID]; ok {
				close(oldClient.send)
				oldClient.conn.Close()
			}
			h.clients[client.UserID] = client
			count := len(h.clients)
			h.clientsMu.Unlock()

			// 启动客户端读写协程
			go client.readPump()
			go client.writePump()

			logger.Infof("WebSocket 客户端注册 (user_id=%d)，当前在线: %d", client.UserID, count)

		case client := <-h.unregister:
			h.clientsMu.Lock()
			if existingClient, ok := h.clients[client.UserID]; ok {
				if existingClient == client {
					delete(h.clients, client.UserID)
					close(client.send)
				}
			}
			count := len(h.clients)
			h.clientsMu.Unlock()
			logger.Infof("WebSocket 客户端注销 (user_id=%d)，当前在线: %d", client.UserID, count)

		case message := <-h.broadcast:
			h.broadcastToAll(message)

		case <-ctx.Done():
			logger.Info("WebSocket Hub 正在关闭...")
			h.clientsMu.Lock()
			for userID, client := range h.clients {
				close(client.send)
				client.conn.Close()
				delete(h.clients, userID)
			}
			h.clientsMu.Unlock()
			logger.Info("WebSocket Hub 已关闭")
			return
		}
	}
}

// broadcastToAll 向所有在线客户端广播消息
func (h *Hub) broadcastToAll(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("WebSocket 消息序列化失败: %v", err)
		return
	}

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for userID, client := range h.clients {
		select {
		case client.send <- data:
		default:
			logger.Warnf("WebSocket 发送失败 (user_id=%d)，关闭连接", userID)
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}
}

// BroadcastToUser 向指定用户推送消息
func (h *Hub) BroadcastToUser(userID int64, msg Message) error {
	h.clientsMu.RLock()
	client, ok := h.clients[userID]
	h.clientsMu.RUnlock()

	if !ok {
		return fmt.Errorf("用户 %d 不在线", userID)
	}

	return client.SendMessage(msg)
}

// OnlineCount 返回当前在线客户端数量
func (h *Hub) OnlineCount() int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	return len(h.clients)
}

// subscribeRedis 订阅 Redis Pub/Sub 频道
// 收到消息后广播给所有本地客户端
func (h *Hub) subscribeRedis(ctx context.Context) {
	// 持续重连订阅（处理 Redis 断连情况）
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ch, err := h.rdb.Subscribe(ctx, "chillcat:events")
		if err != nil {
			logger.Warnf("Redis 订阅失败，5秒后重试: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}

		logger.Info("✅ Redis Pub/Sub 订阅成功 (chillcat:events)")

		for msg := range ch {
			var wsMsg Message
			if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
				logger.Warnf("Redis 消息解析失败: %v", err)
				continue
			}

			// 广播给所有本地客户端
			h.broadcastToAll(wsMsg)
		}

		// channel 关闭，尝试重连
		logger.Warn("Redis Pub/Sub 频道关闭，5秒后重连...")
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
