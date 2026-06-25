package model

import "time"

// ResonanceStory 共鸣故事（用户匿名发布的情绪日记）
type ResonanceStory struct {
	ID               int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID           int64     `json:"user_id" gorm:"index;not null"`
	EmotionCheckinID int64     `json:"emotion_checkin_id" gorm:"index"`
	Content          string    `json:"content" gorm:"type:text;not null"`
	EmotionType      string    `json:"emotion_type" gorm:"size:20;not null"`
	IsAnonymous      bool      `json:"is_anonymous" gorm:"default:true"`
	ResonanceCount   int64     `json:"resonance_count" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ResonanceRecord 共鸣记录（用户对某条故事的共鸣）
type ResonanceRecord struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	StoryID   int64     `json:"story_id" gorm:"index;not null"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	Message   string    `json:"message" gorm:"size:200"`
	CreatedAt time.Time `json:"created_at"`
}
