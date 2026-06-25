package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LetterHandler 感谢信处理器
type LetterHandler struct{ svc *service.LetterService }

// NewLetterHandler 创建感谢信处理器
func NewLetterHandler(svc *service.LetterService) *LetterHandler {
	return &LetterHandler{svc: svc}
}

// Create 写感谢信
// POST /api/v1/letters
func (h *LetterHandler) Create(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req service.CreateLetterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	result, code, err := h.svc.Create(userID, &req)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// Sent 我发出的感谢信
// GET /api/v1/letters/sent
func (h *LetterHandler) Sent(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page := 1
	pageSize := 10
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil {
		pageSize = ps
	}

	result, code, err := h.svc.Sent(userID, page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// Received 我收到的感谢信
// GET /api/v1/letters/received
func (h *LetterHandler) Received(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	page := 1
	pageSize := 10
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil {
		page = p
	}
	if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil {
		pageSize = ps
	}

	result, code, err := h.svc.Received(userID, page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// Get 查看感谢信详情
// GET /api/v1/letters/:id
func (h *LetterHandler) Get(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	result, code, err := h.svc.Get(id, userID)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}
