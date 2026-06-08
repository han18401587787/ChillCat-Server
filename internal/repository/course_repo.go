package repository
import ("chillcat-server/internal/model"; "gorm.io/gorm")
type CourseRepo struct{ db *gorm.DB }
func NewCourseRepo(db *gorm.DB) *CourseRepo { return &CourseRepo{db: db} }
func (r *CourseRepo) List(category string) ([]model.Course, error) {
    var items []model.Course
    q := r.db.Model(&model.Course{})
    if category != "" { q = q.Where("category = ?", category) }
    if err := q.Order("sort_order DESC").Find(&items).Error; err != nil { return nil, err }
    if items == nil { items = []model.Course{} }
    return items, nil
}
func (r *CourseRepo) GetByID(id int64) (*model.Course, error) {
    var c model.Course; err := r.db.First(&c, id).Error
    if err != nil { return nil, err }; return &c, nil
}
func (r *CourseRepo) MarkComplete(userID, courseID int64) error {
    var p model.UserCourseProgress
    r.db.Where("user_id = ? AND course_id = ?", userID, courseID).First(&p)
    p.UserID = userID; p.CourseID = courseID; p.Completed = true
    return r.db.Save(&p).Error
}
