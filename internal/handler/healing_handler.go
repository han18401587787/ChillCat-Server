package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HealingHandler 稳情计划处理器
type HealingHandler struct {
	svc *service.HealingService
}

// NewHealingHandler 创建稳情计划处理器实例
func NewHealingHandler(svc *service.HealingService) *HealingHandler {
	return &HealingHandler{svc: svc}
}

// GetPlan 获取当前稳情计划
// GET /api/v1/healing/plan
func (h *HealingHandler) GetPlan(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	result, code, err := h.svc.GetCurrentPlan(userID)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// GeneratePlan 生成一周稳情计划
// POST /api/v1/healing/plan/generate
func (h *HealingHandler) GeneratePlan(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	result, code, err := h.svc.GeneratePlan(userID)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

// CompleteTask 完成稳情计划中的某个任务
// POST /api/v1/healing/plan/tasks/:id/complete
func (h *HealingHandler) CompleteTask(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	result, code, err := h.svc.CompleteTask(userID, taskID)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}
