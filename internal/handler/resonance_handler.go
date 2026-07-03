package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResonanceHandler struct{ svc *service.ResonanceService }

func NewResonanceHandler(svc *service.ResonanceService) *ResonanceHandler {
	return &ResonanceHandler{svc: svc}
}

// CreateStory 发布共鸣故事
// POST /api/v1/resonance/stories
func (h *ResonanceHandler) CreateStory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	var req service.CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	result, code, err := h.svc.CreateStory(userID, &req)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// ListStories 获取共鸣墙列表
// GET /api/v1/resonance/stories
func (h *ResonanceHandler) ListStories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	result, onlineCount, code, err := h.svc.ListStories(page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}
	// 返回带 online_count 的扩展响应
	response.Success(c, gin.H{
		"list":         result.List,
		"total":        result.Total,
		"page":         result.Page,
		"page_size":    result.PageSize,
		"online_count": onlineCount,
	})
}

// GetStory 获取单条共鸣故事详情
// GET /api/v1/resonance/stories/:id
func (h *ResonanceHandler) GetStory(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	result, code, err := h.svc.GetStory(id)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// Resonate 表达共鸣
// POST /api/v1/resonance/stories/:id/resonate
func (h *ResonanceHandler) Resonate(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	var req service.ResonateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	code, err := h.svc.Resonate(id, userID, &req)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, nil)
}

// GetResonators 查看共鸣者列表
// GET /api/v1/resonance/stories/:id/resonators
func (h *ResonanceHandler) GetResonators(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	result, code, err := h.svc.GetResonators(id)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}
