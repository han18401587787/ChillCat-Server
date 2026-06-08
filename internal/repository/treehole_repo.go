package repository
import ("chillcat-server/internal/model"; "gorm.io/gorm")
type TreeHoleRepo struct{ db *gorm.DB }
func NewTreeHoleRepo(db *gorm.DB) *TreeHoleRepo { return &TreeHoleRepo{db: db} }
func (r *TreeHoleRepo) Create(post *model.TreeHolePost) error { return r.db.Create(post).Error }
func (r *TreeHoleRepo) List(page, pageSize int) ([]model.TreeHolePost, int64, error) {
    var items []model.TreeHolePost; var total int64
    q := r.db.Model(&model.TreeHolePost{})
    q.Count(&total)
    if err := q.Order("id DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error; err != nil { return nil, 0, err }
    if items == nil { items = []model.TreeHolePost{} }
    return items, total, nil
}
func (r *TreeHoleRepo) AddHug(id int64) error { return r.db.Model(&model.TreeHolePost{}).Where("id = ?", id).UpdateColumn("hugs", gorm.Expr("hugs + 1")).Error }
