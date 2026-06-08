package handler
import ("chillcat-server/internal/service"; "chillcat-server/pkg/response"; "strconv"; "github.com/gin-gonic/gin")
type CourseHandler struct{ svc *service.CourseService }
func NewCourseHandler(svc *service.CourseService) *CourseHandler { return &CourseHandler{svc: svc} }
func (h *CourseHandler) List(c *gin.Context) {
    category := c.DefaultQuery("category", "")
    result, code, err := h.svc.List(category)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}
func (h *CourseHandler) MarkComplete(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    courseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil { response.Error(c, response.ErrBadRequest); return }
    code, err := h.svc.MarkComplete(userID, courseID)
    if err != nil { response.Error(c, code); return }
    response.Success(c, nil)
}
