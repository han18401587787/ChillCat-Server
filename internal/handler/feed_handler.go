package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FeedHandler 内容流接口
type FeedHandler struct {
	feedService *service.FeedService
}

// NewFeedHandler 创建内容流接口
func NewFeedHandler(feedService *service.FeedService) *FeedHandler {
	return &FeedHandler{feedService: feedService}
}

// ListFeeds 获取内容列表
func (h *FeedHandler) ListFeeds(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, code, err := h.feedService.ListFeeds(page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, result)
}

// GetDetail 获取内容详情
func (h *FeedHandler) GetDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	item, code, err := h.feedService.GetFeedDetail(id)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, item)
}


// Search 搜索内容
func (h *FeedHandler) Search(c *gin.Context) {
	query := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, code, err := h.feedService.SearchFeeds(query, page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, result)
}
