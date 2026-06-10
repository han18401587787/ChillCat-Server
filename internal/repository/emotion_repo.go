package repository

import (
    "chillcat-server/internal/model"
    "time"

    "gorm.io/gorm"
)

type EmotionRepo struct{ db *gorm.DB }
func NewEmotionRepo(db *gorm.DB) *EmotionRepo { return &EmotionRepo{db: db} }

func (r *EmotionRepo) Create(checkin *model.EmotionCheckin) error { return r.db.Create(checkin).Error }

func (r *EmotionRepo) GetToday(userID int64, date string) (*model.EmotionCheckin, error) {
    var c model.EmotionCheckin
    err := r.db.Where("user_id = ? AND checkin_date = ?", userID, date).First(&c).Error
    if err != nil { return nil, err }
    return &c, nil
}

func (r *EmotionRepo) List(userID int64, month string, page, pageSize int) ([]model.EmotionCheckin, int64, error) {
    var items []model.EmotionCheckin
    var total int64
    q := r.db.Model(&model.EmotionCheckin{}).Where("user_id = ? AND checkin_date LIKE ?", userID, month+"%")
    q.Count(&total)
    offset := (page - 1) * pageSize
    if err := q.Order("checkin_date DESC, id DESC").Offset(offset).Limit(pageSize).Find(&items).Error; err != nil { return nil, 0, err }
    if items == nil { items = []model.EmotionCheckin{} }
    return items, total, nil
}

func (r *EmotionRepo) WeeklyStats(userID int64) ([]model.EmotionCheckin, error) {
    var items []model.EmotionCheckin
    weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
    err := r.db.Where("user_id = ? AND checkin_date >= ?", userID, weekAgo).Order("checkin_date DESC").Find(&items).Error
    if err != nil { return nil, err }
    if items == nil { items = []model.EmotionCheckin{} }
    return items, nil
}

func (r *EmotionRepo) StreakDays(userID int64) int64 {
    var count int64
    r.db.Model(&model.EmotionCheckin{}).Where("user_id = ?", userID).Select("COUNT(DISTINCT checkin_date)").Scan(&count)
    return count
}
