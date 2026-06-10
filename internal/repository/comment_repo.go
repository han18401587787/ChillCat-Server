package repository
import ("chillcat-server/internal/model"; "gorm.io/gorm")
type CommentRepo struct{ db *gorm.DB }
func NewCommentRepo(db *gorm.DB) *CommentRepo { return &CommentRepo{db: db} }
func (r *CommentRepo) List(courseID int64, page, pageSize int) ([]model.CourseComment, int64, error) {
	var items []model.CourseComment; var total int64
	q := r.db.Model(&model.CourseComment{}).Where("course_id = ?", courseID)
	q.Count(&total)
	if err := q.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error; err != nil { return nil, 0, err }
	if items == nil { items = []model.CourseComment{} }
	return items, total, nil
}
func (r *CommentRepo) Create(c *model.CourseComment) error { return r.db.Create(c).Error }
