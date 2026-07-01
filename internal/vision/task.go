package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chillcat-server/internal/cache"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string     `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Result    any        `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
	CreatedAt int64      `json:"created_at"`
	UpdatedAt int64      `json:"updated_at"`
}

// TaskManager 任务管理器（基于 Redis）
type TaskManager struct {
	rdb *cache.RedisClient
}

// NewTaskManager 创建任务管理器
func NewTaskManager(rdb *cache.RedisClient) *TaskManager {
	return &TaskManager{rdb: rdb}
}

// taskKey 生成 Redis key
func (tm *TaskManager) taskKey(taskID string) string {
	return fmt.Sprintf("vision:task:%s", taskID)
}

// CreateTask 创建任务，初始状态为 pending
func (tm *TaskManager) CreateTask(ctx context.Context, taskID string) error {
	now := time.Now().Unix()
	task := &TaskResult{
		TaskID:    taskID,
		Status:    TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	// 使用 SetEX 设置 30 分钟过期
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}
	return tm.rdb.SetEX(ctx, tm.taskKey(taskID), string(data), 30*time.Minute)
}

// UpdateStatus 更新任务状态
func (tm *TaskManager) UpdateStatus(ctx context.Context, taskID string, status TaskStatus) error {
	task, err := tm.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = status
	task.UpdatedAt = time.Now().Unix()
	return tm.saveTask(ctx, taskID, task)
}

// SetResult 设置任务结果（完成时调用）
func (tm *TaskManager) SetResult(ctx context.Context, taskID string, result any) error {
	task, err := tm.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = TaskStatusCompleted
	task.Result = result
	task.UpdatedAt = time.Now().Unix()
	return tm.saveTask(ctx, taskID, task)
}

// SetError 设置任务错误（失败时调用）
func (tm *TaskManager) SetError(ctx context.Context, taskID string, errMsg string) error {
	task, err := tm.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	task.Status = TaskStatusFailed
	task.Error = errMsg
	task.UpdatedAt = time.Now().Unix()
	return tm.saveTask(ctx, taskID, task)
}

// GetTask 获取任务信息
func (tm *TaskManager) GetTask(ctx context.Context, taskID string) (*TaskResult, error) {
	data, err := tm.rdb.Get(ctx, tm.taskKey(taskID))
	if err != nil {
		return nil, fmt.Errorf("get task %s: %w", taskID, err)
	}
	var task TaskResult
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("unmarshal task: %w", err)
	}
	return &task, nil
}

// saveTask 保存任务到 Redis（保持原有 TTL）
func (tm *TaskManager) saveTask(ctx context.Context, taskID string, task *TaskResult) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}
	// 获取剩余 TTL
	ttl, err := tm.rdb.TTL(ctx, tm.taskKey(taskID))
	if err != nil || ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return tm.rdb.SetEX(ctx, tm.taskKey(taskID), string(data), ttl)
}
