package repository

import (
	"chillcat-server/internal/model"
	"gorm.io/gorm"
)

type MessageRepo struct{ db *gorm.DB }

func NewMessageRepo(db *gorm.DB) *MessageRepo { return &MessageRepo{db: db} }

func (r *MessageRepo) List(userID int64, page, pageSize int) ([]model.Message, int64, error) {
	var items []model.Message
	var total int64
	q := r.db.Model(&model.Message{}).Where("user_id = ?", userID)
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	if items == nil { items = []model.Message{} }
	return items, total, err
}

func (r *MessageRepo) UnreadCount(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Message{}).Where("user_id = ? AND is_read = false", userID).Count(&count).Error
	return count, err
}

func (r *MessageRepo) MarkRead(userID, id int64) error {
	return r.db.Model(&model.Message{}).Where("id = ? AND user_id = ?", id, userID).Update("is_read", true).Error
}
