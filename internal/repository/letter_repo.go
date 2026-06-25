package repository

import (
	"chillcat-server/internal/model"
	"time"

	"gorm.io/gorm"
)

// LetterRepo 感谢信仓库
type LetterRepo struct{ db *gorm.DB }

// NewLetterRepo 创建感谢信仓库
func NewLetterRepo(db *gorm.DB) *LetterRepo { return &LetterRepo{db: db} }

// Create 创建感谢信
func (r *LetterRepo) Create(letter *model.ThankYouLetter) error {
	return r.db.Create(letter).Error
}

// CountTodayBySender 统计发送者今日发信数量
func (r *LetterRepo) CountTodayBySender(senderID int64) (int64, error) {
	today := time.Now().Format("2006-01-02")
	todayStart := today + " 00:00:00"
	todayEnd := today + " 23:59:59"
	var count int64
	err := r.db.Model(&model.ThankYouLetter{}).
		Where("sender_id = ? AND created_at >= ? AND created_at <= ?", senderID, todayStart, todayEnd).
		Count(&count).Error
	return count, err
}

// ListBySender 查询发送者发出的信件
func (r *LetterRepo) ListBySender(senderID int64, page, pageSize int) ([]model.ThankYouLetter, int64, error) {
	var items []model.ThankYouLetter
	var total int64
	q := r.db.Model(&model.ThankYouLetter{}).Where("sender_id = ?", senderID)
	q.Count(&total)
	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []model.ThankYouLetter{}
	}
	return items, total, nil
}

// ListByReceiver 查询接收者收到的信件
func (r *LetterRepo) ListByReceiver(receiverID int64, page, pageSize int) ([]model.ThankYouLetter, int64, error) {
	var items []model.ThankYouLetter
	var total int64
	q := r.db.Model(&model.ThankYouLetter{}).Where("receiver_id = ?", receiverID)
	q.Count(&total)
	offset := (page - 1) * pageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []model.ThankYouLetter{}
	}
	return items, total, nil
}

// GetByID 根据 ID 获取信件
func (r *LetterRepo) GetByID(id int64) (*model.ThankYouLetter, error) {
	var letter model.ThankYouLetter
	if err := r.db.First(&letter, id).Error; err != nil {
		return nil, err
	}
	return &letter, nil
}
