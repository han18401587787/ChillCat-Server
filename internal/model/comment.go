package model
import "time"
type CourseComment struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`
	CourseID int64 `json:"course_id" gorm:"index;not null"`
	UserID int64 `json:"user_id" gorm:"index;not null"`
	Content string `json:"content" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at"`
}
