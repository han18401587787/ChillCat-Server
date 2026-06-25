package handler

import (
	"chillcat-server/internal/model"
	"chillcat-server/pkg/response"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VoiceHandler 语音情绪日记处理器
type VoiceHandler struct{ db *gorm.DB }

// NewVoiceHandler 创建语音处理器
func NewVoiceHandler(db *gorm.DB) *VoiceHandler { return &VoiceHandler{db: db} }

// Upload 上传音频文件
// POST /api/v1/voice/upload
func (h *VoiceHandler) Upload(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	file, err := c.FormFile("audio")
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	// 校验文件大小 (最大 10MB)
	if file.Size > 10*1024*1024 {
		response.ErrorWithMsg(c, response.ErrValidation, "音频文件大小不能超过 10MB")
		return
	}

	// 校验文件类型
	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{".mp3": true, ".wav": true, ".m4a": true, ".aac": true, ".ogg": true, ".webm": true}
	if !allowed[ext] {
		response.ErrorWithMsg(c, response.ErrValidation, "不支持的音频格式，支持: mp3/wav/m4a/aac/ogg/webm")
		return
	}

	// 确保上传目录存在
	uploadDir := "/tmp/uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Error(c, response.ErrInternal)
		return
	}

	// 保存文件
	filename := fmt.Sprintf("%d_%d_%s", userID, time.Now().UnixNano(), file.Filename)
	savePath := filepath.Join(uploadDir, filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, response.ErrInternal)
		return
	}

	today := time.Now().Format("2006-01-02")
	checkinIDStr := c.PostForm("checkin_id")

	var checkin model.EmotionCheckin

	// 如果传了 checkin_id，关联到已有打卡
	if checkinIDStr != "" {
		if id, err := strconv.ParseInt(checkinIDStr, 10, 64); err == nil {
			if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&checkin).Error; err != nil {
				response.Error(c, response.ErrNotFound)
				return
			}
		}
	}

	if checkin.ID != 0 {
		// 更新已有打卡
		checkin.HasAudio = true
		checkin.AudioURL = savePath
		checkin.AudioStatus = "pending"
		h.db.Save(&checkin)
	} else {
		// 新建打卡
		checkin = model.EmotionCheckin{
			UserID:      userID,
			Emotion:     "语音记录",
			Note:        "",
			HasAudio:    true,
			AudioURL:    savePath,
			AudioStatus: "pending",
			CheckinDate: today,
			CreatedAt:   time.Now(),
		}
		h.db.Create(&checkin)
	}

	// 异步模拟语音识别 + 情绪分析
	go h.mockProcess(checkin.ID)

	response.Success(c, gin.H{
		"checkin_id":   checkin.ID,
		"audio_url":    savePath,
		"audio_status": "pending",
	})
}

// Status 查询语音处理状态
// GET /api/v1/voice/:id/status
func (h *VoiceHandler) Status(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	var checkin model.EmotionCheckin
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&checkin).Error; err != nil {
		response.Error(c, response.ErrNotFound)
		return
	}

	result := gin.H{
		"checkin_id":   checkin.ID,
		"audio_status": checkin.AudioStatus,
	}

	// 分析完成时返回转录和情绪结果
	if checkin.AudioStatus == "analyzed" {
		result["transcription"] = checkin.Note
		result["emotion_analysis"] = checkin.Emotion
	}

	response.Success(c, result)
}

// mockProcess 模拟异步语音识别和情绪分析（后续接入真实 Whisper）
func (h *VoiceHandler) mockProcess(checkinID int64) {
	// 模拟 2-3 秒处理延迟
	time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)

	// 先标记为转录中
	h.db.Model(&model.EmotionCheckin{}).Where("id = ?", checkinID).Update("audio_status", "transcribing")
	time.Sleep(500 * time.Millisecond)

	// 模拟转录结果
	transcriptions := []string{
		"今天工作很顺利，完成了项目的一个重要里程碑，感觉特别有成就感。",
		"下午和朋友一起喝咖啡聊天，分享了很多有趣的事情，心情变得很好。",
		"今天遇到了一些困难，但冷静下来想了想，其实也没那么糟糕，我会继续努力的。",
		"阳光很好，午饭后去公园散了步，听着鸟叫声觉得生活还是很美好的。",
		"今天有点疲惫，但晚上做了一顿好吃的犒劳自己，简单的幸福也很满足。",
		"和家人通了电话，听到他们的声音让我觉得特别温暖和安心。",
	}

	emotions := []string{"开心", "放松", "平静", "满足", "温暖", "感恩"}

	note := transcriptions[rand.Intn(len(transcriptions))]
	emotion := emotions[rand.Intn(len(emotions))]

	// 更新打卡记录
	h.db.Model(&model.EmotionCheckin{}).Where("id = ?", checkinID).Updates(map[string]interface{}{
		"note":         note,
		"emotion":      emotion,
		"audio_status": "analyzed",
	})
}
