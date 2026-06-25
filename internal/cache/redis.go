// Package cache Redis 缓存层封装
// 提供 Redis 客户端初始化、基础 KV 操作以及 Pub/Sub 能力
// 连接失败时降级为无缓存模式，不阻断服务启动
package cache

import (
	"chillcat-server/internal/config"
	"chillcat-server/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient Redis 客户端封装
type RedisClient struct {
	rdb  *redis.Client
	ok   bool // 是否可用（连接失败时设为 false，降级运行）
}

// NewRedisClient 创建 Redis 客户端
// cfg: 应用配置
// 返回的客户端在连接失败时 ok=false，不影响服务启动
func NewRedisClient(cfg *config.Config) *RedisClient {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warnf("⚠️ Redis 连接失败（%v），降级为无缓存模式运行", err)
		return &RedisClient{rdb: nil, ok: false}
	}

	logger.Infof("✅ Redis 连接成功 (%s, DB=%d)", cfg.Redis.Addr(), cfg.Redis.DB)
	return &RedisClient{rdb: rdb, ok: true}
}

// IsOK 返回 Redis 是否可用
func (r *RedisClient) IsOK() bool {
	return r.ok && r.rdb != nil
}

// Ping 健康检查
func (r *RedisClient) Ping(ctx context.Context) error {
	if !r.IsOK() {
		return fmt.Errorf("redis 不可用（降级模式）")
	}
	return r.rdb.Ping(ctx).Err()
}

// Get 获取字符串值
// key 不存在时返回 ("", nil)
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if !r.IsOK() {
		return "", fmt.Errorf("redis 不可用（降级模式）")
	}
	val, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Set 设置字符串值（永不过期）
func (r *RedisClient) Set(ctx context.Context, key, value string) error {
	if !r.IsOK() {
		return fmt.Errorf("redis 不可用（降级模式）")
	}
	return r.rdb.Set(ctx, key, value, 0).Err()
}

// SetEX 设置带过期时间的字符串值
func (r *RedisClient) SetEX(ctx context.Context, key, value string, ttl time.Duration) error {
	if !r.IsOK() {
		return fmt.Errorf("redis 不可用（降级模式）")
	}
	return r.rdb.Set(ctx, key, value, ttl).Err()
}

// SetJSON 将对象序列化为 JSON 后存入 Redis
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON 序列化失败: %w", err)
	}
	return r.SetEX(ctx, key, string(data), ttl)
}

// GetJSON 从 Redis 读取 JSON 并反序列化
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return fmt.Errorf("key %s 不存在", key)
	}
	return json.Unmarshal([]byte(val), dest)
}

// Del 删除一个或多个 key
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	if !r.IsOK() {
		return fmt.Errorf("redis 不可用（降级模式）")
	}
	return r.rdb.Del(ctx, keys...).Err()
}

// Exists 检查 key 是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	if !r.IsOK() {
		return false, fmt.Errorf("redis 不可用（降级模式）")
	}
	n, err := r.rdb.Exists(ctx, key).Result()
	return n > 0, err
}

// TTL 获取 key 的剩余过期时间
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	if !r.IsOK() {
		return 0, fmt.Errorf("redis 不可用（降级模式）")
	}
	return r.rdb.TTL(ctx, key).Result()
}

// Publish 发布消息到频道（用于跨实例 WebSocket 消息广播）
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	if !r.IsOK() {
		return fmt.Errorf("redis 不可用（降级模式）")
	}
	var payload string
	switch v := message.(type) {
	case string:
		payload = v
	case []byte:
		payload = string(v)
	default:
		data, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("消息序列化失败: %w", err)
		}
		payload = string(data)
	}
	return r.rdb.Publish(ctx, channel, payload).Err()
}

// Subscribe 订阅频道，返回消息 channel
// channelName: Redis 频道名
// 返回的 <-chan 在连接断开或 ctx 取消时自动关闭
func (r *RedisClient) Subscribe(ctx context.Context, channelName string) (<-chan *redis.Message, error) {
	if !r.IsOK() {
		return nil, fmt.Errorf("redis 不可用（降级模式）")
	}
	pubsub := r.rdb.Subscribe(ctx, channelName)

	// 等待订阅确认
	if _, err := pubsub.Receive(ctx); err != nil {
		return nil, fmt.Errorf("订阅频道 %s 失败: %w", channelName, err)
	}

	ch := pubsub.Channel()
	go func() {
		<-ctx.Done()
		_ = pubsub.Close()
	}()

	return ch, nil
}

// Close 关闭 Redis 连接
func (r *RedisClient) Close() error {
	if r.rdb != nil {
		return r.rdb.Close()
	}
	return nil
}

// Client 获取底层 redis.Client（供高级操作使用，可能为 nil）
func (r *RedisClient) Client() *redis.Client {
	return r.rdb
}
