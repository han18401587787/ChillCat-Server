package handler

import (
	"chillcat-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommunityHandler 社区相关接口（温暖模板、互助小组）
type CommunityHandler struct{}

func NewCommunityHandler() *CommunityHandler {
	return &CommunityHandler{}
}

// WarmTemplate 温暖模板响应
type WarmTemplate struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	Emoji    string `json:"emoji"`
	Category string `json:"category"`
}

// GetWarmTemplates 获取温暖回应模板列表
func (h *CommunityHandler) GetWarmTemplates(c *gin.Context) {
	templates := []WarmTemplate{
		{ID: "1", Content: "你不是一个人，我在这里", Emoji: "💚", Category: "support"},
		{ID: "2", Content: "我也有过类似的感觉", Emoji: "🤝", Category: "empathy"},
		{ID: "3", Content: "谢谢你愿意分享这些", Emoji: "🙏", Category: "gratitude"},
		{ID: "4", Content: "慢慢来，不着急", Emoji: "🌱", Category: "comfort"},
		{ID: "5", Content: "你已经很勇敢了", Emoji: "💪", Category: "encourage"},
		{ID: "6", Content: "给自己一点时间", Emoji: "⏳", Category: "comfort"},
		{ID: "7", Content: "今天辛苦了", Emoji: "🌸", Category: "care"},
		{ID: "8", Content: "一切都会好起来的", Emoji: "🌈", Category: "hope"},
	}
	response.Success(c, templates)
}

// MutualAidGroup 互助小组
type MutualAidGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	MemberCount int64  `json:"member_count"`
	IconName    string `json:"icon_name"`
	IsJoined    bool   `json:"is_joined"`
}

// GetMutualAidGroups 获取互助小组列表
func (h *CommunityHandler) GetMutualAidGroups(c *gin.Context) {
	groups := []MutualAidGroup{
		{ID: 1, Name: "焦虑治愈小组", Description: "一起面对焦虑，分享放松技巧", Category: "anxiety", MemberCount: 128, IconName: "leaf.circle.fill", IsJoined: false},
		{ID: 2, Name: "职场解压站", Description: "职场压力释放，互相支持", Category: "workplace", MemberCount: 89, IconName: "briefcase.fill", IsJoined: false},
		{ID: 3, Name: "睡前故事会", Description: "分享助眠方法，互道晚安", Category: "sleep", MemberCount: 203, IconName: "moon.stars.fill", IsJoined: false},
		{ID: 4, Name: "成长加油站", Description: "记录成长瞬间，互相鼓励", Category: "growth", MemberCount: 56, IconName: "sparkles", IsJoined: false},
	}
	response.Success(c, groups)
}

// JoinMutualAidGroup 加入互助小组
func (h *CommunityHandler) JoinMutualAidGroup(c *gin.Context) {
	response.Success(c, gin.H{"joined": true})
}

// LeaveMutualAidGroup 退出互助小组
func (h *CommunityHandler) LeaveMutualAidGroup(c *gin.Context) {
	response.Success(c, gin.H{"joined": false})
}
