package model

import "time"

// EncourageChain 鼓励链 — 一人发起鼓励，陌生人接力传递
type EncourageChain struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	InitiatorID   int64      `json:"initiator_id" gorm:"index;not null"`
	Title         string     `json:"title" gorm:"size:200;not null"`
	Description   string     `json:"description" gorm:"type:text"`
	MaxLength     int        `json:"max_length" gorm:"default:20"`
	CurrentLength int        `json:"current_length" gorm:"default:0"`
	Category      string     `json:"category" gorm:"size:20;default:'其他'"`
	Status        string     `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at"`
}

// EncourageLink 鼓励链接力 — 每位参与者的鼓励内容
type EncourageLink struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ChainID     int64     `json:"chain_id" gorm:"index;not null"`
	UserID      int64     `json:"user_id" gorm:"index;not null"`
	Content     string    `json:"content" gorm:"size:140;not null"`
	Position    int       `json:"position" gorm:"not null"`
	IsAnonymous bool      `json:"is_anonymous" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
}
