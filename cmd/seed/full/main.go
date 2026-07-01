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
	logger.Info("🌱 绪安全类型测试账号填充...")

	db, _ := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	_ = db.AutoMigrate(
		&model.User{}, &model.MemberInfo{}, &model.MemberOrder{},
		&model.EmotionCheckin{}, &model.TreeHolePost{},
		&model.Course{}, &model.UserCourseProgress{}, &model.CourseComment{},
		&model.ResonanceStory{}, &model.ResonanceRecord{},
		&model.EncourageChain{}, &model.EncourageLink{},
		&model.HealingPlan{}, &model.HealingPlanTask{},
		&model.ThankYouLetter{},
	)

	// === 1. 创建测试用户 ===
	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	testUser := &model.User{Username: "tester", Email: "tester@xuanpeace.app", Password: string(hash), Nickname: "绪安体验官", Status: 1}
	db.Where("username = ?", "tester").FirstOrCreate(testUser)
	uid := testUser.ID
	logger.Infof("✅ 测试账号: tester / 123456 (uid=%d)", uid)

	// === 2. 会员 ===
	now := time.Now()
	db.Where("user_id = ?", uid).FirstOrCreate(&model.MemberInfo{
		UserID: uid, MemberType: "yearly", Status: "active",
		StartDate: timePtr(now.AddDate(0, -1, 0)),
		EndDate:   timePtr(now.AddDate(0, 11, 0)), AutoRenew: true,
	})
	logger.Info("✅ 年度会员")

	// === 3. 情绪打卡 (最近30天) ===
	emotions := []string{"平静", "开心", "疲惫", "焦虑", "委屈", "孤独", "烦躁", "迷茫", "易怒", "内耗"}
	notes := []string{"今天工作效率很高", "和朋友聚餐很开心", "加班到很晚有点累", "项目deadline焦虑中", "被领导批评了委屈", "一个人看电影", "堵车迟到烦躁", "对未来感到迷茫", "和伴侣吵架了", "自我怀疑的一天"}
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i)
		db.Where("user_id = ? AND checkin_date = ?", uid, date.Format("2006-01-02")).FirstOrCreate(&model.EmotionCheckin{
			UserID: uid, Emotion: emotions[i%10], Note: notes[i%10],
			HasDoodle: i%3 == 0, CheckinDate: date.Format("2006-01-02"), CreatedAt: date,
		})
	}
	logger.Info("✅ 情绪打卡: 30天")

	// === 4. 树洞帖子 ===
	posts := []string{
		"房贷还完了开心。", "辅导孩子做作业好崩溃。", "同事离职了好舍不得。",
		"今天是我戒烟第七天！", "跑步第30天打卡，坚持下来了。",
		"煮了一碗泡面加蛋，简单幸福。", "被小区保安记住名字了。",
		"刚看完一部电影哭得稀里哗啦的。", "对象说想和我结婚有点慌。",
		"爸妈催婚催到不想回家过年了。", "终于买了心心念念的耳机。",
		"一个人去医院做胃镜，孤独等级拉满。",
	}
	for i, p := range posts {
		db.Where("content = ? AND user_id = ?", p, uid).FirstOrCreate(&model.TreeHolePost{
			UserID: uid, Content: p, Scope: "public",
			IsAnonymous: i%2 == 0, Hugs: int64(rand.Intn(50) + 1),
			CreatedAt: now.Add(-time.Duration(i*4) * time.Hour),
		})
	}
	logger.Info("✅ 树洞帖子: 12条")

	// === 5. 共鸣墙故事 ===
	stories := []struct{ content, emotion string }{
		{"三十岁生日一个人过的，给自己买了个小蛋糕。有点孤独，但也挺自由的。", "孤独"},
		{"下周一就答辩了，PPT改了三遍了还是不满意。", "焦虑"},
		{"今天终于鼓起勇气和妈妈说了心里话。说着说着就哭了，但说完轻松了好多好多。", "平静"},
		{"在地铁上看到一个女孩偷偷擦眼泪，想递张纸巾又怕冒犯。希望你现在好一点了。", "委屈"},
		{"拿到了心仪公司的offer！努力没有白费。", "开心"},
		{"加班到凌晨，回家的路上看到环卫工已经在扫街了。大家都不容易。", "疲惫"},
	}
	var storyIDs []int64
	for _, s := range stories {
		story := &model.ResonanceStory{
			UserID: uid, Content: s.content, EmotionType: s.emotion,
			IsAnonymous: true, ResonanceCount: int64(rand.Intn(3000) + 100),
			CreatedAt: now.Add(-time.Duration(rand.Intn(72)) * time.Hour),
		}
		db.Where("content = ? AND user_id = ?", s.content, uid).FirstOrCreate(story)
		storyIDs = append(storyIDs, story.ID)
	}
	logger.Info("✅ 共鸣墙故事: 6条")

	// === 6. 共鸣记录 ===
	for _, sid := range storyIDs {
		for j := 0; j < rand.Intn(5)+1; j++ {
			db.Create(&model.ResonanceRecord{
				StoryID: sid, UserID: uid,
				Message:   randomMessage(),
				CreatedAt: now.Add(-time.Duration(rand.Intn(48)) * time.Hour),
			})
		}
	}
	logger.Info("✅ 共鸣回应")

	// === 7. 鼓励链 ===
	chain := &model.EncourageChain{
		InitiatorID: uid, Title: "每日鼓励接力", Description: "每人说一句温暖的话",
		MaxLength: 10, CurrentLength: 5, Category: "daily", Status: "active",
		CreatedAt: now.Add(-24 * time.Hour),
	}
	db.Where("initiator_id = ? AND status = ?", uid, "active").FirstOrCreate(chain)
	for i := 0; i < 5; i++ {
		db.Where("chain_id = ? AND position = ?", chain.ID, i+1).FirstOrCreate(&model.EncourageLink{
			ChainID: chain.ID, UserID: uid, Content: randomEncourage(),
			Position: i + 1, CreatedAt: now.Add(-time.Duration(24-i*4) * time.Hour),
		})
	}
	logger.Info("✅ 鼓励链: 1条 (5个环节)")

	// === 8. 稳情计划 ===
	today := now.Format("2006-01-02")
	endDay := now.AddDate(0, 0, 7).Format("2006-01-02")
	plan := &model.HealingPlan{
		UserID: uid, StartDate: today, EndDate: endDay, Status: "active",
		CreatedAt: now,
	}
	db.Where("user_id = ? AND status = ?", uid, "active").FirstOrCreate(plan)
	tasks := []struct {
		day                   int
		taskType, title, desc string
		completed             bool
	}{
		{1, "breathing", "4-7-8呼吸法", "跟随引导做5分钟深呼吸", true},
		{2, "journal", "情绪日记", "写下今天最强烈的情绪和触发原因", true},
		{3, "meditation", "冥想放松", "闭上眼睛跟随音频放松10分钟", true},
		{4, "active", "散步30分钟", "去户外感受阳光和自然", false},
		{5, "music", "听一首喜欢的歌", "选一首让自己心情好的音乐", false},
		{6, "active", "整理房间", "花15分钟整理一个角落", false},
		{7, "journal", "感恩练习", "写下今天最想感谢的三件事", false},
	}
	for _, t := range tasks {
		db.Where("plan_id = ? AND day_number = ?", plan.ID, t.day).FirstOrCreate(&model.HealingPlanTask{
			PlanID: plan.ID, DayNumber: t.day, TaskType: t.taskType,
			TaskTitle: t.title, TaskDesc: t.desc, IsCompleted: t.completed,
		})
	}
	logger.Info("✅ 稳情计划: 7天任务 (3完成/4未完成)")

	// === 9. 感谢信 ===
	letters := []struct {
		content, sender, receiver string
		public                    bool
	}{
		{"谢谢你一直以来的陪伴，每次和你聊天都觉得很温暖。", "匿名用户", "绪安体验官", true},
		{"你的分享给了我很大的勇气，让我知道自己不是一个人。", "小雅", "绪安体验官", true},
	}
	for _, l := range letters {
		db.Create(&model.ThankYouLetter{
			SenderID: uid, ReceiverID: uid,
			Content: l.content, IsPublic: l.public, CreatedAt: now.Add(-time.Duration(rand.Intn(48)) * time.Hour),
		})
	}
	logger.Info("✅ 感谢信: 2封")

	logger.Info("🎉 全类型测试账号创建完成!")
	logger.Info("   账号: tester / 123456")
	logger.Info("   数据: 会员+30天打卡+12帖子+6共鸣+鼓励链+稳情计划+感谢信")
}

func randomMessage() string {
	msgs := []string{"我也有过这种感觉", "加油！", "你不是一个人", "抱抱你", "一切都会好起来的", "你很勇敢", "谢谢你的分享"}
	return msgs[rand.Intn(len(msgs))]
}

func randomEncourage() string {
	words := []string{"今天也要加油💪", "你是最棒的✨", "一切都会好的🌈", "给自己一个拥抱🫂", "你值得被温柔对待🌸"}
	return words[rand.Intn(len(words))]
}

func timePtr(t time.Time) *time.Time { return &t }
