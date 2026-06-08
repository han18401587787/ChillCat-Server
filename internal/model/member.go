package model

import "time"

// MemberInfo 会员信息模型
type MemberInfo struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64      `gorm:"uniqueIndex;not null" json:"user_id"`
	MemberType string     `gorm:"type:varchar(32);default:''" json:"member_type"` // monthly/quarterly/yearly/permanent
	Status     string     `gorm:"type:varchar(32);default:'none'" json:"status"`  // active/expired/cancelled/none
	StartDate  *time.Time `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	AutoRenew  bool       `gorm:"default:false" json:"auto_renew"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TableName 表名
func (MemberInfo) TableName() string {
	return "member_info"
}

// IsValid 是否在有效期内
func (m *MemberInfo) IsValid() bool {
	if m.Status != "active" {
		return false
	}
	if m.EndDate == nil {
		return true // 永久会员
	}
	return m.EndDate.After(time.Now())
}

// RemainingDays 剩余天数（永久会员返回 -1）
func (m *MemberInfo) RemainingDays() int {
	if m.EndDate == nil {
		return -1
	}
	return int(time.Until(*m.EndDate).Hours() / 24)
}
