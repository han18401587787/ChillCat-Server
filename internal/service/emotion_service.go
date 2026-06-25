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
    vos := make([]JournalEntryVO, 0, len(items))
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
    vos := make([]JournalEntryVO, 0, len(items))
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

// ─── 情绪预警 ──────────────────────────────────────────────────

// AlertLevel 预警等级
type AlertLevel string

const (
	AlertNormal  AlertLevel = "normal"
	AlertWarning AlertLevel = "warning"
	AlertAlert   AlertLevel = "alert"
)

// AlertVO 情绪预警视图对象
type AlertVO struct {
	Level       AlertLevel `json:"level"`        // 预警等级: normal/warning/alert
	Message     string     `json:"message"`      // 预警提示文案
	NegativePct float64    `json:"negative_pct"` // 负面情绪占比 (0-100)
}

// negativeEmotions 负面情绪集合
var negativeEmotions = map[string]bool{
	"焦虑": true, "愤怒": true, "悲伤": true, "孤独": true, "疲惫": true,
	"anxiety": true, "anger": true, "sadness": true, "loneliness": true, "fatigue": true,
}

// Alerts 情绪预警：检查近 7 天负面情绪占比
func (s *EmotionService) Alerts(userID int64) (*AlertVO, int, error) {
	items, err := s.repo.WeeklyStats(userID)
	if err != nil {
		return nil, response.ErrInternal, err
	}

	// 没有足够数据，返回 normal
	total := len(items)
	if total == 0 {
		return &AlertVO{
			Level:       AlertNormal,
			Message:     "暂无足够数据，请保持情绪打卡",
			NegativePct: 0,
		}, response.CodeSuccess, nil
	}

	// 统计负面情绪数量
	negativeCount := 0
	for _, c := range items {
		if negativeEmotions[c.Emotion] {
			negativeCount++
		}
	}

	negativePct := float64(negativeCount) / float64(total) * 100

	// 预警规则
	switch {
	case negativeCount == total:
		// 近 7 天全部负面 → alert
		return &AlertVO{
			Level:       AlertAlert,
			Message:     "近7天情绪持续低落，建议寻求专业支持或与信任的人聊聊",
			NegativePct: negativePct,
		}, response.CodeSuccess, nil
	case negativePct > 60:
		// 负面情绪占比 > 60% → warning
		return &AlertVO{
			Level:       AlertWarning,
			Message:     "近期负面情绪较多，试试稳情计划或深呼吸放松一下",
			NegativePct: negativePct,
		}, response.CodeSuccess, nil
	default:
		// 正常
		return &AlertVO{
			Level:       AlertNormal,
			Message:     "情绪状态良好，继续保持",
			NegativePct: negativePct,
		}, response.CodeSuccess, nil
	}
}
