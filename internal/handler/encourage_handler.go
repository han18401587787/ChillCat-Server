package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EncourageHandler struct{ svc *service.EncourageService }

func NewEncourageHandler(svc *service.EncourageService) *EncourageHandler {
	return &EncourageHandler{svc: svc}
}

// CreateChain 发起鼓励链
// POST /api/v1/encourage/chains
func (h *EncourageHandler) CreateChain(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	var req service.CreateChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	result, code, err := h.svc.CreateChain(userID, &req)
	if err != nil {
		response.ErrorWithMsg(c, code, err.Error())
		return
	}
	response.Success(c, result)
}

// ListChains 获取鼓励链列表
// GET /api/v1/encourage/chains
func (h *EncourageHandler) ListChains(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	category := c.DefaultQuery("category", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, code, err := h.svc.ListChains(status, category, page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// GetChain 获取单条鼓励链详情 + 接力链
// GET /api/v1/encourage/chains/:id
func (h *EncourageHandler) GetChain(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	result, code, err := h.svc.GetChain(id)
	if err != nil {
		response.ErrorWithMsg(c, code, err.Error())
		return
	}
	response.Success(c, result)
}

// JoinChain 接力加入
// POST /api/v1/encourage/chains/:id/join
func (h *EncourageHandler) JoinChain(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	var req service.JoinChainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}
	result, code, err := h.svc.JoinChain(id, userID, &req)
	if err != nil {
		response.ErrorWithMsg(c, code, err.Error())
		return
	}
	response.Success(c, result)
}

// ListMyChains 获取我参与/发起的鼓励链
// GET /api/v1/encourage/my-chains
func (h *EncourageHandler) ListMyChains(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	result, code, err := h.svc.ListMyChains(userID)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}
