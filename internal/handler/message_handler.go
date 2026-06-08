package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct{ msgService *service.MessageService }

func NewMessageHandler(msgService *service.MessageService) *MessageHandler {
	return &MessageHandler{msgService: msgService}
}

func (h *MessageHandler) List(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok { return }
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	result, code, err := h.msgService.List(userID, page, pageSize)
	if err != nil { response.Error(c, code); return }
	response.Success(c, result)
}

func (h *MessageHandler) UnreadCount(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok { return }
	count, code, err := h.msgService.UnreadCount(userID)
	if err != nil { response.Error(c, code); return }
	response.Success(c, gin.H{"count": count})
}

func (h *MessageHandler) MarkRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok { return }
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, response.ErrBadRequest); return }
	code, err := h.msgService.MarkRead(userID, id)
	if err != nil { response.Error(c, code); return }
	response.Success(c, nil)
}
