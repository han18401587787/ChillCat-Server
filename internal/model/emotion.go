package model

import "time"

type EmotionCheckin struct {
    ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    UserID    int64     `json:"user_id" gorm:"index;not null"`
    Emotion   string    `json:"emotion" gorm:"size:20;not null"`
    Note      string    `json:"note" gorm:"type:text"`
    HasAudio  bool      `json:"has_audio" gorm:"default:false"`
    AudioURL  string    `json:"audio_url" gorm:"size:500"`
    HasDoodle bool      `json:"has_doodle" gorm:"default:false"`
    CheckinDate string  `json:"checkin_date" gorm:"size:10;index"` // 2026-06-08
    CreatedAt time.Time `json:"created_at"`
}
