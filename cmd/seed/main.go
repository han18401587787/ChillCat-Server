package main

import (
	"chillcat-server/internal/config"
	"chillcat-server/internal/model"
	"chillcat-server/pkg/logger"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	logger.Init("info", "console")
	logger.Info("开始填充种子数据...")

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		logger.Fatalf("数据库连接失败: %v", err)
	}

	db.AutoMigrate(&model.User{}, &model.MemberInfo{}, &model.MemberOrder{}, &model.FeedItem{}, &model.Message{})

	// --- 测试用户 ---
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := &model.User{Username: "test", Email: "test@chillcat.app", Password: string(hash), Nickname: "测试用户", Status: 1}
	db.Where("username = ?", "test").FirstOrCreate(user)
	logger.Infof("测试用户: test / 123456 (ID=%d)", user.ID)

	// --- Feed 内容 ---
	feeds := []model.FeedItem{
		{Title: "夏日清凉歌单推荐", Subtitle: "精选30首清凉歌曲，陪你度过炎热夏天", ImageURL: "https://picsum.photos/400/300?1", ContentType: "music", SortOrder: 100, Status: 1},
		{Title: "周杰伦新专辑首发", Subtitle: "时隔六年，周杰伦携全新专辑《夏日狂想曲》回归", ImageURL: "https://picsum.photos/400/300?2", ContentType: "music", SortOrder: 99, Status: 1},
		{Title: "ChillCat 会员权益全面升级", Subtitle: "新增无损音质、专属歌单、线下活动优先权等重磅权益", ImageURL: "https://picsum.photos/400/300?3", ContentType: "article", SortOrder: 98, Status: 1},
		{Title: "晚间助眠白噪音合集", Subtitle: "雨声、海浪、森林鸟鸣…帮你快速入眠的自然之声", ImageURL: "https://picsum.photos/400/300?4", ContentType: "music", SortOrder: 97, Status: 1},
		{Title: "2026年度金曲盘点", Subtitle: "回顾上半年最受欢迎的100首华语金曲", ImageURL: "https://picsum.photos/400/300?5", ContentType: "music", SortOrder: 96, Status: 1},
		{Title: "玩转 ChillCat：10个隐藏功能", Subtitle: "你可能不知道的实用小技巧，让你的音乐体验更上一层楼", ImageURL: "https://picsum.photos/400/300?6", ContentType: "article", SortOrder: 95, Status: 1},
		{Title: "独立音乐人扶持计划启动", Subtitle: "ChillCat 联合多家厂牌，为原创音乐人提供全方位支持", ImageURL: "https://picsum.photos/400/300?7", ContentType: "article", SortOrder: 94, Status: 1},
		{Title: "周末Live：线上音乐会直播", Subtitle: "本周六晚8点，三位独立音乐人带来不插电现场", ImageURL: "https://picsum.photos/400/300?8", ContentType: "video", SortOrder: 93, Status: 1},
		{Title: "经典老歌：80年代回忆杀", Subtitle: "那些曾经陪伴我们成长的经典旋律，每一首都是青春的记忆", ImageURL: "https://picsum.photos/400/300?9", ContentType: "music", SortOrder: 92, Status: 1},
		{Title: "通勤路上听什么？", Subtitle: "适合地铁公交的通勤歌单，让你的一天从好音乐开始", ImageURL: "https://picsum.photos/400/300?10", ContentType: "music", SortOrder: 91, Status: 1},
		{Title: "运动跑步BGM推荐", Subtitle: "高能节奏助你突破极限，适合不同运动强度的音乐推荐", ImageURL: "https://picsum.photos/400/300?11", ContentType: "music", SortOrder: 90, Status: 1},
		{Title: "咖啡厅氛围音乐", Subtitle: "适合阅读、工作的轻音乐，营造舒适的咖啡厅氛围", ImageURL: "https://picsum.photos/400/300?12", ContentType: "music", SortOrder: 89, Status: 1},
		{Title: "Hi-Res无损音质体验指南", Subtitle: "如何用ChillCat享受录音室级别的音质体验", ImageURL: "https://picsum.photos/400/300?13", ContentType: "article", SortOrder: 88, Status: 1},
		{Title: "民谣新生代推荐", Subtitle: "这些年轻的声音，正在重新定义华语民谣", ImageURL: "https://picsum.photos/400/300?14", ContentType: "music", SortOrder: 87, Status: 1},
		{Title: "深夜电台：治愈系女声", Subtitle: "温柔嗓音伴你入眠，精选十位治愈系女歌手作品", ImageURL: "https://picsum.photos/400/300?15", ContentType: "music", SortOrder: 86, Status: 1},
	}
	for i := range feeds {
		db.Where("title = ?", feeds[i].Title).FirstOrCreate(&feeds[i])
	}
	logger.Infof("Feed 内容: %d 条", len(feeds))

	// --- 消息通知 ---
	messages := []model.Message{
		{UserID: user.ID, Title: "欢迎加入 ChillCat", Content: "感谢注册ChillCat！我们为你准备了精选歌单，快来探索吧。", MsgType: "system", IsRead: false},
		{UserID: user.ID, Title: "会员权益升级通知", Content: "ChillCat会员权益已全面升级，新增无损音质和专属歌单功能，快来看看吧！", MsgType: "member", IsRead: false},
		{UserID: user.ID, Title: "你有一份年度报告", Content: "2026年上半年听歌报告已生成，回顾你的音乐旅程。", MsgType: "activity", IsRead: true},
		{UserID: user.ID, Title: "新版本发布 v2.0", Content: "ChillCat v2.0已发布，支持暗色模式和搜索功能，快去更新体验吧！", MsgType: "system", IsRead: true},
		{UserID: user.ID, Title: "周杰伦《夏日狂想曲》已上线", Content: "你关注的艺人周杰伦发布了新专辑，快来抢先收听！", MsgType: "activity", IsRead: false},
	}
	for i := range messages {
		messages[i].CreatedAt = time.Now().Add(-time.Duration(i*24) * time.Hour)
		db.Create(&messages[i])
	}
	logger.Infof("消息通知: %d 条", len(messages))

	logger.Info("种子数据填充完成！")
}
