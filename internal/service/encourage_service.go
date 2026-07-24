package service

import (
	"chillcat-server/internal/cache"
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type EncourageService struct {
	repo *repository.EncourageRepo
	rdb  *cache.RedisClient // Redis 客户端（可能为 nil）
}

func NewEncourageService(repo *repository.EncourageRepo, rdb *cache.RedisClient) *EncourageService {
	return &EncourageService{repo: repo, rdb: rdb}
}

// ---------- 请求结构 ----------

// CreateChainRequest 发起鼓励链请求
type CreateChainRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	MaxLength   int    `json:"max_length"`
	Category    string `json:"category" binding:"required"`
}

// ChainResponseVO 鼓励链展示对象（与 iOS 客户端 ChainResponse 对齐）
// 客户端解码为 [ChainResponse]，字段：chain_id / links / participant_count。
// 返回数组而非分页对象，避免客户端解码 "Expected to decode Array<Any> but found a dictionary"。
type ChainResponseVO struct {
	ChainID          int64        `json:"chain_id"`
	ParticipantCount int64        `json:"participant_count"`
	Title            string       `json:"title"`
	Description      string       `json:"description"`
	Category         string       `json:"category"`
	Status           string       `json:"status"`
	CreatedAt        string       `json:"created_at"`
	Links            []LinkItemVO `json:"links"`
}

// LinkItemVO 接力展示对象（与 iOS 客户端 ChainLink 对齐）
// 客户端解码为 ChainLink，必需字段为 id，其余可选。
type LinkItemVO struct {
	ID        int64  `json:"id"`
	ChainID   int64  `json:"chain_id"`
	Content   string `json:"content"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

// JoinChainRequest 接力请求
type JoinChainRequest struct {
	Content     string `json:"content" binding:"required"`
	IsAnonymous bool   `json:"is_anonymous"`
}

// ---------- 业务方法 ----------

// CreateChain 发起一条鼓励链
func (s *EncourageService) CreateChain(userID int64, req *CreateChainRequest) (*ChainResponseVO, int, error) {
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
	return s.toChainResponseVO(chain, nil), response.CodeSuccess, nil
}

// GetChain 获取单条鼓励链详情（含接力链）
func (s *EncourageService) GetChain(chainID int64) (*ChainResponseVO, int, error) {
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

	return s.toChainResponseVO(chain, links), response.CodeSuccess, nil
}

// ListChains 获取鼓励链列表（返回数组，与 iOS 客户端 [ChainResponse] 对齐）
func (s *EncourageService) ListChains(status, category string, page, pageSize int) ([]ChainResponseVO, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	items, _, err := s.repo.ListChains(status, category, page, pageSize)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	return s.buildChainResponses(items), response.CodeSuccess, nil
}

// JoinChain 加入鼓励链接力
func (s *EncourageService) JoinChain(chainID, userID int64, req *JoinChainRequest) (*LinkItemVO, int, error) {
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

	// 发布鼓励接力事件到 Redis（用于 WebSocket 实时推送）
	s.publishJoinEvent(chainID, userID, content)

	return s.toLinkItemVO(link), response.CodeSuccess, nil
}

// ListMyChains 获取我参与/发起的鼓励链（返回数组，与 iOS 客户端 [ChainResponse] 对齐）
func (s *EncourageService) ListMyChains(userID int64) ([]ChainResponseVO, int, error) {
	chains, err := s.repo.ListUserChains(userID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	return s.buildChainResponses(chains), response.CodeSuccess, nil
}

// buildChainResponses 批量构建 ChainResponseVO，并为每条链加载接力列表。
// 始终返回非 nil 切片，保证客户端解码为数组（空数组而非 null）。
func (s *EncourageService) buildChainResponses(chains []model.EncourageChain) []ChainResponseVO {
	vos := make([]ChainResponseVO, 0, len(chains))
	for i := range chains {
		var links []model.EncourageLink
		if ls, e := s.repo.GetLinksByChainID(chains[i].ID); e == nil {
			links = ls
		}
		vos = append(vos, *s.toChainResponseVO(&chains[i], links))
	}
	return vos
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

// toChainResponseVO 将 model + 接力列表转为客户端契约结构
func (s *EncourageService) toChainResponseVO(c *model.EncourageChain, links []model.EncourageLink) *ChainResponseVO {
	linkVOs := make([]LinkItemVO, 0, len(links))
	for i := range links {
		linkVOs = append(linkVOs, *s.toLinkItemVO(&links[i]))
	}
	return &ChainResponseVO{
		ChainID:          c.ID,
		ParticipantCount: int64(c.CurrentLength),
		Title:            c.Title,
		Description:      c.Description,
		Category:         c.Category,
		Status:           c.Status,
		CreatedAt:        c.CreatedAt.Format(time.RFC3339),
		Links:            linkVOs,
	}
}

// toLinkItemVO 将 model 转为客户端契约的接力结构
func (s *EncourageService) toLinkItemVO(l *model.EncourageLink) *LinkItemVO {
	return &LinkItemVO{
		ID:        l.ID,
		ChainID:   l.ChainID,
		Content:   l.Content,
		Position:  l.Position,
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
	}
}

// publishJoinEvent 通过 Redis Pub/Sub 发布鼓励接力事件
func (s *EncourageService) publishJoinEvent(chainID, userID int64, content string) {
	if s.rdb == nil || !s.rdb.IsOK() {
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type": "encourage.joined",
		"data": map[string]interface{}{
			"chain_id": chainID,
			"user_id":  userID,
			"content":  content,
			"time":     time.Now().Format(time.RFC3339),
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.rdb.Publish(ctx, "chillcat:events", string(payload)); err != nil {
		logger.Warnf("Redis 发布鼓励接力事件失败: %v", err)
	}
}
