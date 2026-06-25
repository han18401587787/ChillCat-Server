package model

import "time"

// HealingPlan 稳情计划（一周治愈计划）
type HealingPlan struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	StartDate string    `json:"start_date" gorm:"size:10;not null"` // 计划开始日期 2006-01-02
	EndDate   string    `json:"end_date" gorm:"size:10;not null"`   // 计划结束日期 2006-01-02
	Status    string    `json:"status" gorm:"size:20;default:'active'"` // active/completed/cancelled
	CreatedAt time.Time `json:"created_at"`
}

// HealingPlanTask 稳情计划任务
type HealingPlanTask struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	PlanID      int64      `json:"plan_id" gorm:"index;not null"`
	DayNumber   int        `json:"day_number" gorm:"not null"`                // 第几天 1-7
	TaskType    string     `json:"task_type" gorm:"size:30;not null"`         // 任务类型: breathing/journal/meditation/music/active
	TaskTitle   string     `json:"task_title" gorm:"size:100;not null"`       // 任务标题
	TaskDesc    string     `json:"task_desc" gorm:"type:text"`                // 任务描述
	IsCompleted bool       `json:"is_completed" gorm:"default:false"`         // 是否完成
	CompletedAt *time.Time `json:"completed_at"`                              // 完成时间
	CreatedAt   time.Time  `json:"created_at"`
}
