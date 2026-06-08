package handler
import ("chillcat-server/internal/service"; "chillcat-server/pkg/response"; "strconv"; "github.com/gin-gonic/gin")
type TreeHoleHandler struct{ svc *service.TreeHoleService }
func NewTreeHoleHandler(svc *service.TreeHoleService) *TreeHoleHandler { return &TreeHoleHandler{svc: svc} }
func (h *TreeHoleHandler) CreatePost(c *gin.Context) {
    userID, ok := getUserID(c); if !ok { return }
    var req service.PostRequest
    if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, response.ErrBadRequest); return }
    result, code, err := h.svc.CreatePost(userID, &req)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}
func (h *TreeHoleHandler) ListPosts(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    result, code, err := h.svc.ListPosts(page, pageSize)
    if err != nil { response.Error(c, code); return }
    response.Success(c, result)
}
func (h *TreeHoleHandler) AddHug(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil { response.Error(c, response.ErrBadRequest); return }
    code, err := h.svc.AddHug(id)
    if err != nil { response.Error(c, code); return }
    response.Success(c, nil)
}
