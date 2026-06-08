package model

import "time"

// FeedItem 内容流条目
type FeedItem struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Subtitle    string    `json:"subtitle" gorm:"size:500"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	ContentType string    `json:"content_type" gorm:"size:50;default:'article'"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	Status      int       `json:"status" gorm:"default:1"` // 1=发布 0=下架
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
