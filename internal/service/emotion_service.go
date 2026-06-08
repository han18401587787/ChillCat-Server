package service

import (
    "chillcat-server/internal/model"
    "chillcat-server/internal/repository"
    "chillcat-server/pkg/response"
    "errors"
    "time"
    "gorm.io/gorm"
)

type EmotionService struct{ repo *repository.EmotionRepo }
func NewEmotionService(repo *repository.EmotionRepo) *EmotionService { return &EmotionService{repo: repo} }

type CheckinRequest struct {
    Emotion string `json:"emotion" binding:"required"`
    Note    string `json:"note"`
}

type CheckinVO struct {
    ID      int64  `json:"id"`
    Emotion string `json:"emotion"`
    Note    string `json:"note"`
    Date    string `json:"checkin_date"`
    Streak  int64  `json:"streak_days"`
}

func (s *EmotionService) Checkin(userID int64, req *CheckinRequest) (*CheckinVO, int, error) {
    today := time.Now().Format("2006-01-02")
    if _, err := s.repo.GetToday(userID, today); err == nil {
        return nil, response.ErrUserExists, errors.New("今日已打卡")
    }
    c := &model.EmotionCheckin{UserID: userID, Emotion: req.Emotion, Note: req.Note, CheckinDate: today}
    if err := s.repo.Create(c); err != nil { return nil, response.ErrInternal, err }
    return &CheckinVO{ID: c.ID, Emotion: c.Emotion, Note: c.Note, Date: today, Streak: s.repo.StreakDays(userID)}, response.CodeSuccess, nil
}

func (s *EmotionService) GetToday(userID int64) (*CheckinVO, int, error) {
    today := time.Now().Format("2006-01-02")
    c, err := s.repo.GetToday(userID, today)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return &CheckinVO{Date: today}, response.CodeSuccess, nil
        }
        return nil, response.ErrInternal, err
    }
    return &CheckinVO{ID: c.ID, Emotion: c.Emotion, Note: c.Note, Date: today, Streak: s.repo.StreakDays(userID)}, response.CodeSuccess, nil
}

type JournalEntryVO struct {
    ID        int64  `json:"id"`
    Emotion   string `json:"emotion"`
    Note      string `json:"note"`
    HasDoodle bool   `json:"has_doodle"`
    Date      string `json:"checkin_date"`
    Created   string `json:"created_at"`
}

func (s *EmotionService) ListJournal(userID int64, month string, page, pageSize int) (*response.Page, int, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 50 { pageSize = 10 }
    items, total, err := s.repo.List(userID, month, page, pageSize)
    if err != nil { return nil, response.ErrInternal, err }
    var vos []JournalEntryVO
    for _, c := range items {
        vos = append(vos, JournalEntryVO{ID: c.ID, Emotion: c.Emotion, Note: c.Note, HasDoodle: c.HasDoodle, Date: c.CheckinDate, Created: c.CreatedAt.Format("2006-01-02 15:04")})
    }
    return &response.Page{List: vos, Total: total, Page: page, PageSize: pageSize}, response.CodeSuccess, nil
}

type WeeklyStatsVO struct {
    Entries    []JournalEntryVO `json:"entries"`
    TotalCount int64            `json:"total_count"`
    StreakDays int64            `json:"streak_days"`
    TopEmotion string           `json:"top_emotion"`
    Insight    string           `json:"insight"`
}

func (s *EmotionService) WeeklyStats(userID int64) (*WeeklyStatsVO, int, error) {
    items, err := s.repo.WeeklyStats(userID)
    if err != nil { return nil, response.ErrInternal, err }
    var vos []JournalEntryVO
    counts := map[string]int64{}
    for _, c := range items {
        vos = append(vos, JournalEntryVO{ID: c.ID, Emotion: c.Emotion, Note: c.Note, Date: c.CheckinDate, Created: c.CreatedAt.Format("2006-01-02 15:04")})
        counts[c.Emotion]++
    }
    top := "平静"
    max := int64(0)
    for e, c := range counts { if c > max { max, top = c, e } }
    return &WeeklyStatsVO{Entries: vos, TotalCount: int64(len(vos)), StreakDays: s.repo.StreakDays(userID), TopEmotion: top, Insight: "这周你的情绪以「" + top + "」为主"}, response.CodeSuccess, nil
}
