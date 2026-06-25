package repository

import (
	"chillcat-server/internal/model"

	"gorm.io/gorm"
)

type EncourageRepo struct{ db *gorm.DB }

func NewEncourageRepo(db *gorm.DB) *EncourageRepo { return &EncourageRepo{db: db} }

// CreateChain 创建鼓励链
func (r *EncourageRepo) CreateChain(chain *model.EncourageChain) error {
	return r.db.Create(chain).Error
}

// GetChainByID 根据 ID 获取鼓励链
func (r *EncourageRepo) GetChainByID(id int64) (*model.EncourageChain, error) {
	var chain model.EncourageChain
	err := r.db.Where("id = ?", id).First(&chain).Error
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

// ListChains 分页获取鼓励链列表，支持按 status / category 过滤
func (r *EncourageRepo) ListChains(status, category string, page, pageSize int) ([]model.EncourageChain, int64, error) {
	var items []model.EncourageChain
	var total int64

	q := r.db.Model(&model.EncourageChain{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}

	q.Count(&total)

	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []model.EncourageChain{}
	}
	return items, total, nil
}

// CreateLink 创建接力记录，同时递增鼓励链的 current_length
func (r *EncourageRepo) CreateLink(link *model.EncourageLink) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(link).Error; err != nil {
			return err
		}
		return tx.Model(&model.EncourageChain{}).
			Where("id = ?", link.ChainID).
			UpdateColumn("current_length", gorm.Expr("current_length + 1")).Error
	})
}

// GetLinksByChainID 获取某条鼓励链的所有接力记录（按 position 升序）
func (r *EncourageRepo) GetLinksByChainID(chainID int64) ([]model.EncourageLink, error) {
	var links []model.EncourageLink
	err := r.db.Where("chain_id = ?", chainID).Order("position ASC").Find(&links).Error
	if err != nil {
		return nil, err
	}
	if links == nil {
		links = []model.EncourageLink{}
	}
	return links, nil
}

// CheckUserJoined 检查用户是否已参与某条鼓励链
func (r *EncourageRepo) CheckUserJoined(chainID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.EncourageLink{}).
		Where("chain_id = ? AND user_id = ?", chainID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CompleteChain 将鼓励链标记为已完成
func (r *EncourageRepo) CompleteChain(chainID int64) error {
	return r.db.Model(&model.EncourageChain{}).
		Where("id = ?", chainID).
		Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": gorm.Expr("NOW()"),
		}).Error
}

// ListUserChains 获取用户发起或参与的所有鼓励链
func (r *EncourageRepo) ListUserChains(userID int64) ([]model.EncourageChain, error) {
	// 查询用户发起的链
	var chains []model.EncourageChain
	err := r.db.Model(&model.EncourageChain{}).
		Where("initiator_id = ?", userID).
		Order("id DESC").
		Find(&chains).Error
	if err != nil {
		return nil, err
	}

	// 查询用户参与的链 ID
	var linkChainIDs []int64
	r.db.Model(&model.EncourageLink{}).
		Where("user_id = ?", userID).
		Distinct("chain_id").
		Pluck("chain_id", &linkChainIDs)

	// 合并去重：把参与但不是发起的链也查出来
	existing := make(map[int64]bool)
	for _, c := range chains {
		existing[c.ID] = true
	}
	for _, cid := range linkChainIDs {
		if existing[cid] {
			continue
		}
		var c model.EncourageChain
		if err := r.db.Where("id = ?", cid).First(&c).Error; err == nil {
			chains = append(chains, c)
		}
	}

	if chains == nil {
		chains = []model.EncourageChain{}
	}
	return chains, nil
}
