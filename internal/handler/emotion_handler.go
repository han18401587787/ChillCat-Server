package handler

import (
    "chillcat-server/internal/service"
    "chillcat-server/pkg/response"
    "strconv"
	"github.com/gin-gonic/gin"
)

type EmotionHandler struct{ svc *service.EmotionService }
func NewEmotionHandler(svc *service.EmotionService) *EmotionHandler { return &EmotionHandler{svc: svc} }

func (h *EmotionHandler) Checkin(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    var req service.CheckinRequest
    if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, response.ErrBadRequest); return }
    result, code, err := h.svc.Checkin(userID, &req)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}

func (h *EmotionHandler) GetToday(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    result, code, err := h.svc.GetToday(userID)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}

func (h *EmotionHandler) Journal(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    month := c.DefaultQuery("month", "")
    page := 1; pageSize := 10
    if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil { page = p }
    if ps, err := strconv.Atoi(c.DefaultQuery("page_size", "10")); err == nil { pageSize = ps }
    result, code, err := h.svc.ListJournal(userID, month, page, pageSize)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}

func (h *EmotionHandler) WeeklyStats(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    result, code, err := h.svc.WeeklyStats(userID)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}

// Alerts 情绪预警接口
// GET /api/v1/emotion/alerts — 检查当前用户是否需要预警
func (h *EmotionHandler) Alerts(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    result, code, err := h.svc.Alerts(userID)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}
