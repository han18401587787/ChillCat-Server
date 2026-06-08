package repository

import (
	"chillcat-server/internal/model"

	"gorm.io/gorm"
)

// MemberRepo 会员数据访问
type MemberRepo struct {
	db *gorm.DB
}

// NewMemberRepo 创建会员仓库
func NewMemberRepo(db *gorm.DB) *MemberRepo {
	return &MemberRepo{db: db}
}

// GetByUserID 根据用户 ID 获取会员信息
func (r *MemberRepo) GetByUserID(userID int64) (*model.MemberInfo, error) {
	var info model.MemberInfo
	err := r.db.Where("user_id = ?", userID).First(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Upsert 创建或更新会员信息
func (r *MemberRepo) Upsert(info *model.MemberInfo) error {
	// 使用 ON CONFLICT 实现 upsert
	return r.db.Where("user_id = ?", info.UserID).Assign(info).FirstOrCreate(info).Error
}

// CreateOrder 创建订单
func (r *MemberRepo) CreateOrder(order *model.MemberOrder) error {
	return r.db.Create(order).Error
}

// UpdateOrder 更新订单
func (r *MemberRepo) UpdateOrder(order *model.MemberOrder) error {
	return r.db.Save(order).Error
}

// GetOrderByNo 根据订单号获取订单
func (r *MemberRepo) GetOrderByNo(orderNo string) (*model.MemberOrder, error) {
	var order model.MemberOrder
	err := r.db.Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrdersByUserID 获取用户订单列表
func (r *MemberRepo) GetOrdersByUserID(userID int64, page, pageSize int) ([]model.MemberOrder, int64, error) {
	var orders []model.MemberOrder
	var total int64

	query := r.db.Model(&model.MemberOrder{}).Where("user_id = ?", userID)
	query.Count(&total)

	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&orders).Error

	return orders, total, err
}
