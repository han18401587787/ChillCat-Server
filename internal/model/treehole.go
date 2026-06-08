package model

import "time"

type TreeHolePost struct {
    ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    UserID      int64     `json:"user_id" gorm:"index"`
    Content     string    `json:"content" gorm:"type:text;not null"`
    Scope       string    `json:"scope" gorm:"size:20;default:'public'"`
    IsAnonymous bool      `json:"is_anonymous" gorm:"default:true"`
    Hugs        int64     `json:"hugs" gorm:"default:0"`
    CreatedAt   time.Time `json:"created_at"`
}
