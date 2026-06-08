package model

import "time"

type Course struct {
    ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    Title       string    `json:"title" gorm:"size:255;not null"`
    Description string    `json:"description" gorm:"size:500"`
    Duration    int       `json:"duration"` // 秒
    Category    string    `json:"category" gorm:"size:50"`
    Tag         string    `json:"tag" gorm:"size:50"`
    IconURL     string    `json:"icon_url" gorm:"size:500"`
    SortOrder   int       `json:"sort_order" gorm:"default:0"`
    CreatedAt   time.Time `json:"created_at"`
}

type UserCourseProgress struct {
    ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    UserID    int64     `json:"user_id" gorm:"index;not null"`
    CourseID  int64     `json:"course_id"`
    Completed bool      `json:"completed" gorm:"default:false"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
