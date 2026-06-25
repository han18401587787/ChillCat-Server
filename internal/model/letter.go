package model

import "time"

// ThankYouLetter 感谢信模型
type ThankYouLetter struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	SenderID    int64     `json:"sender_id" gorm:"index;not null"`
	ReceiverID  int64     `json:"receiver_id" gorm:"index;not null"`
	Content     string    `json:"content" gorm:"type:text;not null"`
	IsAnonymous bool      `json:"is_anonymous" gorm:"default:false"`
	IsPublic    bool      `json:"is_public" gorm:"default:true"`
	Status      string    `json:"status" gorm:"size:20;default:'sent'"` // sent/blocked
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 表名
func (ThankYouLetter) TableName() string {
	return "thankyou_letters"
}
