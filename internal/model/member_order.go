package model

import "time"

// MemberOrder 会员订单模型
type MemberOrder struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64      `gorm:"index;not null" json:"user_id"`
	OrderNo       string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	MemberType    string     `gorm:"type:varchar(32);not null" json:"member_type"`
	OrderStatus   string     `gorm:"type:varchar(32);default:'pending'" json:"order_status"` // pending/success/failed/refunded
	Amount        float64    `gorm:"type:decimal(10,2);default:0" json:"amount"`
	PaymentMethod string     `gorm:"type:varchar(32);default:''" json:"payment_method"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName 表名
func (MemberOrder) TableName() string {
	return "member_orders"
}
