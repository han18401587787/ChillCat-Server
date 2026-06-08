package main

import (
	"chillcat-server/internal/config"
	"chillcat-server/internal/model"
	"chillcat-server/pkg/logger"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, _ := config.Load()
	logger.Init("info", "console")
	logger.Info("🌱 填充绪安种子数据...")

	db, _ := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.MemberInfo{}, &model.MemberOrder{},
		&model.EmotionCheckin{}, &model.TreeHolePost{},
		&model.Course{}, &model.UserCourseProgress{})

	// ── 用户 ──────────────────────────────────────
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	users := []struct {
		username, nickname string
	}{
		{"test", "安静的云雀"},
		{"demo_xiaoming", "小明"},
		{"demo_xiaohong", "小红"},
	}
	var userIDs []int64
	for _, u := range users {
		user := &model.User{Username: u.username, Email: u.username + "@xuanpeace.app", Password: string(hash), Nickname: u.nickname, Status: 1}
		db.Where("username = ?", u.username).FirstOrCreate(user)
		userIDs = append(userIDs, user.ID)
		logger.Infof("用户: %s / 123456 (ID=%d)", u.username, user.ID)
	}

	// ── 30天情绪打卡 ─────────────────────────────
	emotions := []string{"平静", "开心", "疲惫", "焦虑", "委屈", "孤独", "烦躁", "迷茫", "易怒", "内耗"}
	notes := []string{
		"今天开会又被批评了…", "拿到了那个项目的正向反馈，整个人都轻了一截",
		"下午开会的时候leader说我的方案不够细致，我知道他说的有道理，但就是委屈…",
		"", "为什么明明没做错，还是觉得对不起所有人？",
		"周末去了公园，坐在草地上发呆了一个小时，脑子难得空空的",
		"今天又是被否定的一天，感觉自己什么都做不好",
		"", "今天不对自己说任何负面的话",
		"收到了朋友的鼓励，感觉好多了", "",
		"完成了今天的任务，很有成就感",
		"下雨天，一个人在家听了很久的歌",
		"好久没这么放松了", "",
	}
	emotionCount := 0
	for i := 0; i < 30; i++ {
		// 周末更偏正面，周三更偏负面
		date := time.Now().AddDate(0, 0, -i)
		weekday := int(date.Weekday())
		ei := rand.Intn(len(emotions))
		if weekday == 0 || weekday == 6 { ei = (ei % 3) + 0 }        // 周末偏正面
		if weekday == 3 { ei = (ei % 4) + 3 }                       // 周三偏负面
		ni := rand.Intn(len(notes))
		db.Create(&model.EmotionCheckin{
			UserID: userIDs[0], Emotion: emotions[ei], Note: notes[ni],
			HasDoodle: rand.Intn(4) == 0, CheckinDate: date.Format("2006-01-02"),
			CreatedAt: date,
		})
		emotionCount++
	}
	logger.Infof("情绪打卡: %d 条 (30天)", emotionCount)

	// ── 树洞帖子 ─────────────────────────────────
	posts := []struct {
		content string
		hugs    int64
	}{
		{"下午开会的时候leader当着所有人说我的方案不够细致，我知道他说的有道理，但就是委屈…", 24},
		{"周末去了公园，坐在草地上发呆了一个小时，脑子难得空空的，感觉很舒服。", 18},
		{"为什么明明没做错，还是觉得对不起所有人？这种感觉好难受。", 31},
		{"今天拿到了项目的正向反馈，整个人都轻了一截！努力没有白费。", 42},
		{"刚入职三个月，每天都好焦虑，怕自己做不好被开掉…有没有过来人说说怎么调整心态？", 15},
		{"和男朋友吵架了，他说我不够理解他，但我觉得我已经很努力了…", 27},
		{"今天不对自己说任何负面的话。打卡第一天！", 56},
		{"一个人在外地工作，下班回到出租屋就觉得特别孤独。想家了。", 33},
		{"最近在看《被讨厌的勇气》，里面说'你之所以不幸，是因为你缺乏获得幸福的勇气'，这句话触动到我了。", 19},
		{"失眠第三天了，脑子里一大堆事转来转去，有没有好的助眠方法推荐？", 12},
		{"妈妈今天打电话来说做了我最爱吃的红烧肉，挂了电话眼泪就下来了。", 45},
		{"认真工作了一周，周五晚上奖励自己一顿火锅，一个人的火锅也挺香的。", 38},
		{"看完了一本关于情绪管理的书，学到最重要的：情绪不是敌人，是信使。", 21},
		{"被同事误会了，明明不是我做的却被甩锅，解释又显得很小气…", 16},
		{"今天去做了一次冥想，感觉整个人都轻了。推荐大家一起试试。", 29},
	}
	for i, p := range posts {
		uid := userIDs[i%len(userIDs)]
		db.Where("content = ?", p.content).FirstOrCreate(&model.TreeHolePost{
			UserID: uid, Content: p.content, Scope: "public",
			IsAnonymous: i%3 != 0, Hugs: p.hugs,
			CreatedAt: time.Now().Add(-time.Duration(i*8) * time.Hour),
		})
	}
	logger.Infof("树洞帖子: %d 条", len(posts))

	// ── 课程 ─────────────────────────────────────
	courses := []model.Course{
		{Title: "如何在职场中设置情绪边界", Duration: 480, Category: "情绪管理", Tag: "职场场景", SortOrder: 20},
		{Title: "情绪爆发之后，怎么修复关系？", Duration: 360, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 19},
		{Title: "自我接纳：允许自己偶尔脆弱", Duration: 300, Category: "成长", Tag: "成长", SortOrder: 18},
		{Title: "你通常在周三情绪最低", Duration: 240, Category: "情绪管理", Tag: "绪安洞察", SortOrder: 17},
		{Title: "独居场景：一个人的放松指南", Duration: 420, Category: "焦虑治愈", Tag: "独居场景", SortOrder: 16},
		{Title: "睡前记录：今天想说…", Duration: 180, Category: "睡前助眠", Tag: "睡前助眠", SortOrder: 15},
		{Title: "呼吸训练入门：4-7-8呼吸法", Duration: 300, Category: "睡前助眠", Tag: "练习计划", SortOrder: 14},
		{Title: "职场解压：5分钟快速放松", Duration: 300, Category: "职场解压", Tag: "职场场景", SortOrder: 13},
		{Title: "恋爱中的情绪沟通", Duration: 540, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 12},
		{Title: "10天稳情计划：从内耗到松弛", Duration: 600, Category: "成长", Tag: "练习计划", SortOrder: 11},
	}
	for _, c := range courses {
		db.Where("title = ?", c.Title).FirstOrCreate(&c)
	}
	logger.Infof("课程: %d 门", len(courses))

	// ── 会员数据 ─────────────────────────────────
	vipUser := userIDs[1] // demo_xiaoming 是会员
	db.Where("user_id = ?", vipUser).FirstOrCreate(&model.MemberInfo{
		UserID: vipUser, MemberType: "yearly", Status: "active",
		StartDate: timePtr(time.Now().AddDate(0, -3, 0)),
		EndDate:   timePtr(time.Now().AddDate(0, 9, 0)),
		AutoRenew: true,
	})
	logger.Infof("会员: %s (年度会员)", users[1].nickname)

	logger.Info("✅ 绪安种子数据填充完成！")
}

func timePtr(t time.Time) *time.Time { return &t }
