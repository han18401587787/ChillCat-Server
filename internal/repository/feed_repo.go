package repository

import (
	"chillcat-server/internal/model"

	"gorm.io/gorm"
)

// FeedRepo 内容流数据访问
type FeedRepo struct {
	db *gorm.DB
}

// NewFeedRepo 创建内容仓库
func NewFeedRepo(db *gorm.DB) *FeedRepo {
	return &FeedRepo{db: db}
}

// List 分页获取内容列表
func (r *FeedRepo) List(page, pageSize int) ([]model.FeedItem, int64, error) {
	var items []model.FeedItem
	var total int64

	query := r.db.Model(&model.FeedItem{}).Where("status = 1")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("sort_order DESC, id DESC").
		Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	if items == nil {
		items = []model.FeedItem{}
	}
	return items, total, nil
}

// GetByID 根据ID获取内容
func (r *FeedRepo) GetByID(id int64) (*model.FeedItem, error) {
	var item model.FeedItem
	if err := r.db.Where("id = ? AND status = 1", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}


// Search 全文搜索内容
func (r *FeedRepo) Search(query string, page, pageSize int) ([]model.FeedItem, int64, error) {
	var items []model.FeedItem
	var total int64

	q := r.db.Model(&model.FeedItem{}).Where("status = 1").
		Where("title LIKE ? OR subtitle LIKE ?", "%"+query+"%", "%"+query+"%")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Order("sort_order DESC, id DESC").
		Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}

	if items == nil {
		items = []model.FeedItem{}
	}
	return items, total, nil
}
