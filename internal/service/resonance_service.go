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

type ResonanceService struct {
	repo *repository.ResonanceRepo
	rdb  *cache.RedisClient // Redis 客户端（可能为 nil）
}

func NewResonanceService(repo *repository.ResonanceRepo, rdb *cache.RedisClient) *ResonanceService {
	return &ResonanceService{repo: repo, rdb: rdb}
}

// CreateStoryRequest 发布共鸣故事请求
type CreateStoryRequest struct {
	Content     string `json:"content" binding:"required"`
	EmotionType string `json:"emotion_type" binding:"required"`
	IsAnonymous bool   `json:"is_anonymous"`
}

// StoryVO 共鸣故事展示对象
type StoryVO struct {
	ID             int64  `json:"id"`
	Content        string `json:"content"`
	EmotionType    string `json:"emotion_type"`
	IsAnonymous    bool   `json:"is_anonymous"`
	ResonanceCount int64  `json:"resonance_count"`
	DisplayName    string `json:"display_name"`
	CreatedAt      string `json:"created_at"`
}

// ResonatorVO 共鸣者展示对象
type ResonatorVO struct {
	UserID    int64  `json:"user_id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

// CreateStory 发布一条共鸣故事
func (s *ResonanceService) CreateStory(userID int64, req *CreateStoryRequest) (*StoryVO, int, error) {
	story := &model.ResonanceStory{
		UserID:      userID,
		Content:     req.Content,
		EmotionType: req.EmotionType,
		IsAnonymous: req.IsAnonymous,
	}
	if err := s.repo.CreateStory(story); err != nil {
		return nil, response.ErrInternal, err
	}
	return s.toStoryVO(story), response.CodeSuccess, nil
}

// GetStory 获取单条共鸣故事详情
func (s *ResonanceService) GetStory(storyID int64) (*StoryVO, int, error) {
	story, err := s.repo.GetStoryByID(storyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.ErrNotFound, errors.New("故事不存在")
		}
		return nil, response.ErrInternal, err
	}
	return s.toStoryVO(story), response.CodeSuccess, nil
}

// ListStories 分页获取共鸣故事列表
// 返回: 分页数据, 在线人数, 错误码, 错误
func (s *ResonanceService) ListStories(page, pageSize int) (*response.Page, int64, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	items, total, err := s.repo.ListStories(page, pageSize)
	if err != nil {
		return nil, 0, response.ErrInternal, err
	}
	vos := make([]StoryVO, 0, len(items))
	for _, story := range items {
		vos = append(vos, *s.toStoryVO(&story))
	}

	// 获取在线人数（从 Redis 或估算）
	onlineCount := s.getOnlineCount()

	return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, onlineCount, response.CodeSuccess, nil
}

// getOnlineCount 获取当前在线人数（Redis 在线用户数或降级为固定估算值）
func (s *ResonanceService) getOnlineCount() int64 {
	if s.rdb != nil && s.rdb.IsOK() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		// 尝试从 Redis 获取在线人数
		if val, err := s.rdb.Get(ctx, "chillcat:online_count"); err == nil && val != "" {
			var count int64
			if json.Unmarshal([]byte(val), &count) == nil {
				return count
			}
		}
	}
	// Redis 不可用时返回合理默认值
	return 42
}

// ResonateRequest 共鸣请求
type ResonateRequest struct {
	Message string `json:"message"`
}

// Resonate 对一条故事表达共鸣（一人只能共鸣一次）
func (s *ResonanceService) Resonate(storyID, userID int64, req *ResonateRequest) (int, error) {
	// 检查故事是否存在
	_, err := s.repo.GetStoryByID(storyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ErrNotFound, errors.New("故事不存在")
		}
		return response.ErrInternal, err
	}

	// 检查是否已共鸣过
	resonated, err := s.repo.CheckUserResonated(storyID, userID)
	if err != nil {
		return response.ErrInternal, err
	}
	if resonated {
		return response.ErrUserExists, errors.New("你已经表达过共鸣了")
	}

	// 创建共鸣记录
	msg := req.Message
	if len(msg) > 200 {
		msg = msg[:200]
	}
	record := &model.ResonanceRecord{
		StoryID: storyID,
		UserID:  userID,
		Message: msg,
	}
	if err := s.repo.CreateRecord(record); err != nil {
		return response.ErrInternal, err
	}

	// 发布共鸣事件到 Redis（用于 WebSocket 实时推送）
	s.publishResonanceEvent(storyID, userID, msg)

	return response.CodeSuccess, nil
}

// GetResonators 获取共鸣者列表
func (s *ResonanceService) GetResonators(storyID int64) ([]ResonatorVO, int, error) {
	records, err := s.repo.GetResonators(storyID)
	if err != nil {
		return nil, response.ErrInternal, err
	}
	vos := make([]ResonatorVO, 0, len(records))
	for _, r := range records {
		vos = append(vos, ResonatorVO{
			UserID:    r.UserID,
			Message:   r.Message,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}
	return vos, response.CodeSuccess, nil
}

// toStoryVO 将 model 转换为 StoryVO
func (s *ResonanceService) toStoryVO(story *model.ResonanceStory) *StoryVO {
	name := "我"
	if story.IsAnonymous {
		name = "匿名用户"
	}
	return &StoryVO{
		ID:             story.ID,
		Content:        story.Content,
		EmotionType:    story.EmotionType,
		IsAnonymous:    story.IsAnonymous,
		ResonanceCount: story.ResonanceCount,
		DisplayName:    name,
		CreatedAt:      story.CreatedAt.Format(time.RFC3339),
	}
}

// publishResonanceEvent 通过 Redis Pub/Sub 发布共鸣事件
func (s *ResonanceService) publishResonanceEvent(storyID, userID int64, message string) {
	if s.rdb == nil || !s.rdb.IsOK() {
		return
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"type": "resonance.new",
		"data": map[string]interface{}{
			"story_id": storyID,
			"user_id":  userID,
			"message":  message,
			"time":     time.Now().Format(time.RFC3339),
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.rdb.Publish(ctx, "chillcat:events", string(payload)); err != nil {
		logger.Warnf("Redis 发布共鸣事件失败: %v", err)
	}
}
