package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/response"
	"errors"
	"time"

	"gorm.io/gorm"
)

type EncourageService struct {
	repo *repository.EncourageRepo
}

func NewEncourageService(repo *repository.EncourageRepo) *EncourageService {
	return &EncourageService{repo: repo}
}

// ---------- 请求 / 展示结构 ----------

// CreateChainRequest 发起鼓励链请求
type CreateChainRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	MaxLength   int    `json:"max_length"`
	Category    string `json:"category" binding:"required"`
}

// ChainVO 鼓励链展示对象
type ChainVO struct {
	ID            int64  `json:"id"`
	InitiatorID   int64  `json:"initiator_id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	MaxLength     int    `json:"max_length"`
	CurrentLength int    `json:"current_length"`
	Category      string `json:"category"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	CompletedAt   string `json:"completed_at,omitempty"`
}

// ChainDetailVO 鼓励链详情（含接力链）
type ChainDetailVO struct {
	Chain ChainVO  `json:"chain"`
	Links []LinkVO `json:"links"`
}

// LinkVO 接力展示对象
type LinkVO struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Content     string `json:"content"`
	Position    int    `json:"position"`
	IsAnonymous bool   `json:"is_anonymous"`
	CreatedAt   string `json:"created_at"`
}

// JoinChainRequest 接力请求
type JoinChainRequest struct {
	Content     string `json:"content" binding:"required"`
	IsAnonymous bool   `json:"is_anonymous"`
}

// ---------- 业务方法 ----------

// CreateChain 发起一条鼓励链
func (s *EncourageService) CreateChain(userID int64, req *CreateChainRequest) (*ChainVO, int, error) {
	// 分类校验
	if !isValidCategory(req.Category) {
		return nil, response.ErrBadRequest, errors.New("无效的分类")
	}

	maxLen := req.MaxLength
	if maxLen <= 0 || maxLen > 100 {
		maxLen = 20
	}

	chain := &model.EncourageChain{
		InitiatorID:   userID,
		Title:         req.Title,
		Description:   req.Description,
		MaxLength:     maxLen,
		CurrentLength: 0,
		Category:      req.Category,
		Status:        "active",
	}
	if err := s.repo.CreateChain(chain); err != nil {
		return nil, response.ErrInternal, err
	}
	return s.toChainVO(chain), response.CodeSuccess, nil
}

// GetChain 获取单条鼓励链详情（含接力链）
func (s *EncourageService) GetChain(chainID int64) (*ChainDetailVO, int, error) {
	chain, err := s.repo.GetChainByID(chainID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrNotFound, errors.New("鼓励链不存在")
		}
		return nil, response.ErrInternal, err
	}

	links, err := s.repo.GetLinksByChainID(chainID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	linkVOs := make([]LinkVO, 0, len(links))
	for _, l := range links {
		linkVOs = append(linkVOs, *s.toLinkVO(&l))
	}

	return &ChainDetailVO{
		Chain: *s.toChainVO(chain),
		Links: linkVOs,
	}, response.CodeSuccess, nil
}

// ListChains 分页获取鼓励链列表
func (s *EncourageService) ListChains(status, category string, page, pageSize int) (*response.Page, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	items, total, err := s.repo.ListChains(status, category, page, pageSize)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	vos := make([]ChainVO, 0, len(items))
	for _, c := range items {
		vos = append(vos, *s.toChainVO(&c))
	}

	return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, response.CodeSuccess, nil
}

// JoinChain 加入鼓励链接力
func (s *EncourageService) JoinChain(chainID, userID int64, req *JoinChainRequest) (*LinkVO, int, error) {
	chain, err := s.repo.GetChainByID(chainID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrNotFound, errors.New("鼓励链不存在")
		}
		return nil, response.ErrInternal, err
	}

	// 检查鼓励链状态
	if chain.Status != "active" {
		return nil, response.ErrBadRequest, errors.New("该鼓励链已结束，无法接力")
	}

	// 一人一链只能接力一次
	joined, err := s.repo.CheckUserJoined(chainID, userID)
	if err != nil {
		return nil, response.ErrInternal, err
	}
	if joined {
		return nil, response.ErrUserExists, errors.New("你已经参与过这条鼓励链了")
	}

	// 内容长度限制
	content := req.Content
	if len(content) > 140 {
		content = content[:140]
	}

	link := &model.EncourageLink{
		ChainID:     chainID,
		UserID:      userID,
		Content:     content,
		Position:    chain.CurrentLength + 1,
		IsAnonymous: req.IsAnonymous,
	}

	if err := s.repo.CreateLink(link); err != nil {
		return nil, response.ErrInternal, err
	}

	// 接力后检查是否达到 max_length，若达到则自动完成
	if chain.CurrentLength+1 >= chain.MaxLength {
		_ = s.repo.CompleteChain(chainID)
	}

	return s.toLinkVO(link), response.CodeSuccess, nil
}

// ListMyChains 获取我参与/发起的鼓励链
func (s *EncourageService) ListMyChains(userID int64) ([]ChainVO, int, error) {
	chains, err := s.repo.ListUserChains(userID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	vos := make([]ChainVO, 0, len(chains))
	for _, c := range chains {
		vos = append(vos, *s.toChainVO(&c))
	}
	return vos, response.CodeSuccess, nil
}

// ---------- 内部辅助 ----------

// isValidCategory 校验分类是否合法
func isValidCategory(cat string) bool {
	switch cat {
	case "焦虑", "孤独", "工作", "学业", "其他":
		return true
	default:
		return false
	}
}

// toChainVO 将 model 转为 ChainVO
func (s *EncourageService) toChainVO(c *model.EncourageChain) *ChainVO {
	vo := &ChainVO{
		ID:            c.ID,
		InitiatorID:   c.InitiatorID,
		Title:         c.Title,
		Description:   c.Description,
		MaxLength:     c.MaxLength,
		CurrentLength: c.CurrentLength,
		Category:      c.Category,
		Status:        c.Status,
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
	}
	if c.CompletedAt != nil {
		vo.CompletedAt = c.CompletedAt.Format(time.RFC3339)
	}
	return vo
}

// toLinkVO 将 model 转为 LinkVO
func (s *EncourageService) toLinkVO(l *model.EncourageLink) *LinkVO {
	return &LinkVO{
		ID:          l.ID,
		UserID:      l.UserID,
		Content:     l.Content,
		Position:    l.Position,
		IsAnonymous: l.IsAnonymous,
		CreatedAt:   l.CreatedAt.Format(time.RFC3339),
	}
}
