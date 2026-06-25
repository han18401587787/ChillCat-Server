package handler

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/internal/service"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseHandler struct {
	svc         *service.CourseService
	commentRepo *repository.CommentRepo
}

func NewCourseHandler(svc *service.CourseService, commentRepo *repository.CommentRepo) *CourseHandler {
	return &CourseHandler{svc: svc, commentRepo: commentRepo}
}

func (h *CourseHandler) List(c *gin.Context) {
	category := c.DefaultQuery("category", "")
	result, code, err := h.svc.List(category)
	if err != nil {
		response.Error(c, code)
		return
	}
	response.Success(c, result)
}

func (h *CourseHandler) MarkComplete(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok { return }
	courseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, response.ErrBadRequest); return }
	code, err := h.svc.MarkComplete(userID, courseID)
	if err != nil { response.Error(c, code); return }
	response.Success(c, nil)
}

func (h *CourseHandler) ListComments(c *gin.Context) {
	courseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil { response.Error(c, response.ErrBadRequest); return }
	comments, total, err := h.commentRepo.List(courseID, 1, 50)
	if err != nil { response.Error(c, response.ErrInternal); return }
	response.Success(c, gin.H{"list": comments, "total": total})
}

func (h *CourseHandler) AddComment(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok { return }
	courseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, response.ErrBadRequest); return }
	if err := h.commentRepo.Create(&model.CourseComment{CourseID: courseID, UserID: userID, Content: req.Content}); err != nil {
		response.Error(c, response.ErrInternal)
		return
	}
	response.Success(c, nil)
}
