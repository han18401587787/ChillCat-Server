package service
import ("chillcat-server/internal/model"; "chillcat-server/internal/repository"; "chillcat-server/pkg/response"; "time")
type TreeHoleService struct{ repo *repository.TreeHoleRepo }
func NewTreeHoleService(repo *repository.TreeHoleRepo) *TreeHoleService { return &TreeHoleService{repo: repo} }

type PostRequest struct {
    Content     string `json:"content" binding:"required"`
    Scope       string `json:"scope"`
    IsAnonymous bool   `json:"is_anonymous"`
}
type PostVO struct {
    ID int64 `json:"id"`; Content string `json:"content"`; Scope string `json:"scope"`
    IsAnonymous bool `json:"is_anonymous"`; Hugs int64 `json:"hugs"`; CreatedAt string `json:"created_at"`
    DisplayName string `json:"display_name"`
}

func (s *TreeHoleService) CreatePost(userID int64, req *PostRequest) (*PostVO, int, error) {
    scope := req.Scope; if scope == "" { scope = "public" }
    p := &model.TreeHolePost{UserID: userID, Content: req.Content, Scope: scope, IsAnonymous: req.IsAnonymous}
    if err := s.repo.Create(p); err != nil { return nil, response.ErrInternal, err }
    name := "我"; if p.IsAnonymous { name = "匿名用户" }
    return &PostVO{ID: p.ID, Content: p.Content, Scope: p.Scope, IsAnonymous: p.IsAnonymous, Hugs: 0, CreatedAt: p.CreatedAt.Format(time.RFC3339), DisplayName: name}, response.CodeSuccess, nil
}

func (s *TreeHoleService) ListPosts(page, pageSize int) (*response.Page, int, error) {
    if page < 1 { page = 1 }; if pageSize < 1 || pageSize > 50 { pageSize = 10 }
    items, total, err := s.repo.List(page, pageSize)
    if err != nil { return nil, response.ErrInternal, err }
    var vos []PostVO
    for _, p := range items {
        name := "我"; if p.IsAnonymous { name = "匿名用户" }
        vos = append(vos, PostVO{ID: p.ID, Content: p.Content, Scope: p.Scope, IsAnonymous: p.IsAnonymous, Hugs: p.Hugs, CreatedAt: p.CreatedAt.Format(time.RFC3339), DisplayName: name})
    }
    return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, response.CodeSuccess, nil
}

func (s *TreeHoleService) AddHug(id int64) (int, error) {
    if err := s.repo.AddHug(id); err != nil { return response.ErrInternal, err }
    return response.CodeSuccess, nil
}
