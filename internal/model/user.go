package model

import "time"

// User 用户模型
type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string    `gorm:"type:varchar(64);default:''" json:"nickname"`
	Avatar    string    `gorm:"type:varchar(512);default:''" json:"avatar"`
	Status    int       `gorm:"type:smallint;default:1" json:"status"` // 1:正常 0:禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}
