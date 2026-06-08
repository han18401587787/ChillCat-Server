package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/response"
)

// FeedService 内容流服务
type FeedService struct {
	feedRepo *repository.FeedRepo
}

// NewFeedService 创建内容流服务
func NewFeedService(feedRepo *repository.FeedRepo) *FeedService {
	return &FeedService{feedRepo: feedRepo}
}

// FeedItemVO 内容条目视图对象
type FeedItemVO struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Subtitle    string `json:"subtitle"`
	ImageURL    string `json:"image_url"`
	ContentType string `json:"content_type"`
}

func toFeedItemVO(item *model.FeedItem) FeedItemVO {
	return FeedItemVO{
		ID:          item.ID,
		Title:       item.Title,
		Subtitle:    item.Subtitle,
		ImageURL:    item.ImageURL,
		ContentType: item.ContentType,
	}
}

// ListFeeds 获取内容列表
func (s *FeedService) ListFeeds(page, pageSize int) (*response.Page, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	items, total, err := s.feedRepo.List(page, pageSize)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	var vos []FeedItemVO
	for i := range items {
		vos = append(vos, toFeedItemVO(&items[i]))
	}

	return &response.Page{
		List:     vos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, response.CodeSuccess, nil
}

// GetFeedDetail 获取内容详情
func (s *FeedService) GetFeedDetail(id int64) (*model.FeedItem, int, error) {
	item, err := s.feedRepo.GetByID(id)
	if err != nil {
		return nil, response.ErrNotFound, err
	}
	return item, response.CodeSuccess, nil
}


// SearchFeeds 搜索内容
func (s *FeedService) SearchFeeds(query string, page, pageSize int) (*response.Page, int, error) {
	if query == "" {
		return s.ListFeeds(page, pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	items, total, err := s.feedRepo.Search(query, page, pageSize)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	var vos []FeedItemVO
	for i := range items {
		vos = append(vos, toFeedItemVO(&items[i]))
	}

	return &response.Page{
		List:     vos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, response.CodeSuccess, nil
}
