package service
import ("chillcat-server/internal/repository"; "chillcat-server/pkg/response")
type CourseService struct{ repo *repository.CourseRepo }
func NewCourseService(repo *repository.CourseRepo) *CourseService { return &CourseService{repo: repo} }

type CourseVO struct {
    ID int64 `json:"id"`; Title string `json:"title"`; Description string `json:"description"`
    Duration int `json:"duration"`; Category string `json:"category"`; Tag string `json:"tag"`
}

func (s *CourseService) List(category string) ([]CourseVO, int, error) {
    items, err := s.repo.List(category)
    if err != nil { return nil, response.ErrInternal, err }
    var vos []CourseVO
    for _, c := range items {
        vos = append(vos, CourseVO{ID: c.ID, Title: c.Title, Description: c.Description, Duration: c.Duration, Category: c.Category, Tag: c.Tag})
    }
    return vos, response.CodeSuccess, nil
}

func (s *CourseService) MarkComplete(userID, courseID int64) (int, error) {
    if err := s.repo.MarkComplete(userID, courseID); err != nil { return response.ErrInternal, err }
    return response.CodeSuccess, nil
}
