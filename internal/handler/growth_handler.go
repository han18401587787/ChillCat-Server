package handler

import (
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GrowthHandler 成长与成就相关接口
type GrowthHandler struct{}

func NewGrowthHandler() *GrowthHandler {
	return &GrowthHandler{}
}

// Achievement 成就
type Achievement struct {
	ID              int64   `json:"id"`
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	IconName        string  `json:"icon_name"`
	Category        string  `json:"category"`
	IsUnlocked      bool    `json:"is_unlocked"`
	Progress        int     `json:"progress"`
	TargetValue     int     `json:"target_value"`
	ProgressPercent float64 `json:"progress_percent"`
	UnlockedAt      *string `json:"unlocked_at"`
}

// GetAchievements 获取所有成就列表
func (h *GrowthHandler) GetAchievements(c *gin.Context) {
	achievements := []Achievement{
		{ID: 1, Code: "first_checkin", Name: "初次打卡", Description: "完成第一次情绪打卡", IconName: "star.fill", Category: "milestone", IsUnlocked: true, Progress: 1, TargetValue: 1, ProgressPercent: 100, UnlockedAt: strPtr("2026-06-01")},
		{ID: 2, Code: "streak_7", Name: "坚持一周", Description: "连续打卡7天", IconName: "flame.fill", Category: "streak", IsUnlocked: true, Progress: 7, TargetValue: 7, ProgressPercent: 100, UnlockedAt: strPtr("2026-06-07")},
		{ID: 3, Code: "streak_30", Name: "月度达人", Description: "连续打卡30天", IconName: "crown.fill", Category: "streak", IsUnlocked: false, Progress: 12, TargetValue: 30, ProgressPercent: 40, UnlockedAt: nil},
		{ID: 4, Code: "resonance_10", Name: "共鸣使者", Description: "收到10次共鸣", IconName: "heart.fill", Category: "social", IsUnlocked: false, Progress: 3, TargetValue: 10, ProgressPercent: 30, UnlockedAt: nil},
		{ID: 5, Code: "treehole_5", Name: "树洞常客", Description: "发布5篇树洞帖子", IconName: "bubble.left.fill", Category: "social", IsUnlocked: false, Progress: 2, TargetValue: 5, ProgressPercent: 40, UnlockedAt: nil},
		{ID: 6, Code: "meditation_10", Name: "冥想修行者", Description: "完成10次冥想练习", IconName: "leaf.fill", Category: "tool", IsUnlocked: false, Progress: 4, TargetValue: 10, ProgressPercent: 40, UnlockedAt: nil},
	}
	response.Success(c, gin.H{"achievements": achievements})
}

// Milestone 里程碑
type Milestone struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	MilestoneType string `json:"milestone_type"`
	CreatedAt     string `json:"created_at"`
}

// GetMilestones 获取里程碑列表
func (h *GrowthHandler) GetMilestones(c *gin.Context) {
	milestones := []Milestone{
		{ID: 1, Title: "开启绪安之旅", Description: "首次登录并完成匿名注册", MilestoneType: "onboarding", CreatedAt: "2026-06-01T10:00:00Z"},
		{ID: 2, Title: "第一次倾诉", Description: "在树洞发布了第一条心声", MilestoneType: "social", CreatedAt: "2026-06-03T14:20:00Z"},
	}
	response.Success(c, gin.H{"milestones": milestones})
}

// GrowthStats 成长统计
type GrowthStats struct {
	TotalCheckins         int64 `json:"total_checkins"`
	StreakDays            int64 `json:"streak_days"`
	EmotionTypes          int64 `json:"emotion_types"`
	ToolUsageCount        int64 `json:"tool_usage_count"`
	CommunityInteractions int64 `json:"community_interactions"`
	TotalDays             int64 `json:"total_days"`
}

// GetGrowthStats 获取成长统计数据
func (h *GrowthHandler) GetGrowthStats(c *gin.Context) {
	stats := GrowthStats{
		TotalCheckins:         28,
		StreakDays:            7,
		EmotionTypes:          5,
		ToolUsageCount:        15,
		CommunityInteractions: 8,
		TotalDays:             14,
	}
	response.Success(c, stats)
}

func strPtr(s string) *string { return &s }
