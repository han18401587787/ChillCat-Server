package repository

import (
	"chillcat-server/internal/model"

	"gorm.io/gorm"
)

type ResonanceRepo struct{ db *gorm.DB }

func NewResonanceRepo(db *gorm.DB) *ResonanceRepo { return &ResonanceRepo{db: db} }

// CreateStory 创建共鸣故事
func (r *ResonanceRepo) CreateStory(story *model.ResonanceStory) error {
	return r.db.Create(story).Error
}

// GetStoryByID 根据 ID 获取共鸣故事
func (r *ResonanceRepo) GetStoryByID(id int64) (*model.ResonanceStory, error) {
	var story model.ResonanceStory
	err := r.db.Where("id = ?", id).First(&story).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

// ListStories 分页获取共鸣故事列表（按最新排序）
func (r *ResonanceRepo) ListStories(page, pageSize int) ([]model.ResonanceStory, int64, error) {
	var items []model.ResonanceStory
	var total int64
	q := r.db.Model(&model.ResonanceStory{})
	q.Count(&total)
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []model.ResonanceStory{}
	}
	return items, total, nil
}

// CreateRecord 创建共鸣记录，同时递增故事的共鸣数
func (r *ResonanceRepo) CreateRecord(record *model.ResonanceRecord) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		return tx.Model(&model.ResonanceStory{}).
			Where("id = ?", record.StoryID).
			UpdateColumn("resonance_count", gorm.Expr("resonance_count + 1")).Error
	})
}

// GetResonators 获取某条故事的共鸣者列表
func (r *ResonanceRepo) GetResonators(storyID int64) ([]model.ResonanceRecord, error) {
	var records []model.ResonanceRecord
	err := r.db.Where("story_id = ?", storyID).Order("id DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []model.ResonanceRecord{}
	}
	return records, nil
}

// CheckUserResonated 检查用户是否已对某条故事表达过共鸣
func (r *ResonanceRepo) CheckUserResonated(storyID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.ResonanceRecord{}).
		Where("story_id = ? AND user_id = ?", storyID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
