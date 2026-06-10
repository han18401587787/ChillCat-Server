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
	logger.Info("🌱 绪安 100+种子数据填充...")

	db, _ := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.MemberInfo{}, &model.MemberOrder{},
		&model.EmotionCheckin{}, &model.TreeHolePost{},
		&model.Course{}, &model.UserCourseProgress{}, &model.Message{})

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	users := []struct{ u, n, e string }{
		{"test", "安静的云雀", "test@xuanpeace.app"},
		{"xiaoming", "小明", "xiaoming@xuanpeace.app"},
		{"xiaohong", "小红", "xiaohong@xuanpeace.app"},
	}
	var uids []int64
	for _, u := range users {
		user := &model.User{Username: u.u, Email: u.e, Password: string(hash), Nickname: u.n, Status: 1}
		db.Where("username = ?", u.u).FirstOrCreate(user)
		uids = append(uids, user.ID)
		logger.Infof("用户: %s / 123456", u.u)
	}

	emotions := []string{"平静", "开心", "疲惫", "焦虑", "委屈", "孤独", "烦躁", "迷茫", "易怒", "内耗"}
	notes := []string{
		"今天开会又被批评了…", "拿到了项目的正向反馈，整个人都轻了一截",
		"下午leader说我的方案不够细致，但就是委屈", "周末去了公园，发呆了一个小时很舒服",
		"为什么明明没做错还是觉得对不起所有人", "今天不对自己说任何负面的话",
		"完成了一个大任务，很有成就感", "下雨天一个人在家听歌，很放松",
		"和好朋友聊了很久，心情好多了", "健身完浑身舒畅",
		"做了冥想练习，平静了很多", "收到妈妈的电话说想我了",
		"失眠到凌晨三点", "咖啡厅码代码效率好高",
		"被同事甩锅了，解释又显得小气", "看到一只流浪猫，喂了它",
		"今天学到了新东西，充实", "阳光很好，出门散步了",
		"有点焦虑下个月的KPI", "看完了一本好书",
	}
	for i := 0; i < 100; i++ {
		date := time.Now().AddDate(0, 0, -(i % 30))
		db.Create(&model.EmotionCheckin{
			UserID: uids[i%3], Emotion: emotions[i%10], Note: notes[i%20],
			HasDoodle: i%5 == 0, CheckinDate: date.Format("2006-01-02"), CreatedAt: date,
		})
	}
	logger.Infof("情绪打卡: 100 条")

	posts := []string{
		"下午开会的时候leader当着所有人说我的方案不够细致，我知道他说的有道理，但就是委屈…",
		"周末去了公园，坐在草地上发呆了一个小时，脑子难得空空的。",
		"为什么明明没做错，还是觉得对不起所有人？这种感觉好难受。",
		"今天拿到了项目的正向反馈，整个人都轻了一截！努力没有白费。",
		"刚入职三个月，每天都好焦虑，怕自己做不好被开掉…",
		"和男朋友吵架了，他说我不够理解他，但我觉得我已经很努力了…",
		"今天不对自己说任何负面的话。打卡第一天！",
		"一个人在外地工作，下班回到出租屋就觉得特别孤独。想家了。",
		"最近在看《被讨厌的勇气》，触动很大。",
		"失眠第三天了，脑子里一大堆事转来转去。",
		"妈妈今天打电话来说做了我最爱吃的红烧肉，挂了电话眼泪就下来了。",
		"认真工作了一周，周五晚上奖励自己一顿火锅。",
		"看完了一本情绪管理书，学到最重要的：情绪不是敌人，是信使。",
		"被同事误会了，明明不是我做的却被甩锅。",
		"今天去做了一次冥想，感觉整个人都轻了。",
		"面试挂了，虽然嘴上说没事但心里还是有点难过。",
		"减肥失败了三次，这次真的要坚持下去！",
		"在绪安上匿名说了自己的烦恼，收到了好多陌生人的抱抱。",
		"项目deadline压得喘不过气，但我知道我能搞定。",
		"30岁生日一个人过的，点了一根蜡烛。",
		"下雨天好适合发呆，什么都不想做。",
		"朋友借了钱一直不还，问了又显得自己小气。",
		"读了一句话：你不需要完美才值得被爱。",
		"在公交车上看到一个老爷爷给老奶奶系鞋带。",
		"养了一盆多肉，每天看着它慢慢长大很有成就感。",
		"刷到前男友的结婚照，说不难过是假的。",
		"老板说这个月业绩还不错，能拿奖金了！",
		"腿疼了一个星期了要不要去医院看。",
		"你们有没有那种就是不想说话的时候？",
		"地铁上让座被老奶奶夸了，开心。",
		"卷不动了，想躺平。", "有时候觉得社恐到极致了。",
		"辞职信写好了不敢发。", "暗恋一个人三年了不敢表白。",
		"今天是我戒烟第七天。", "跑步第30天打卡，坚持下来了。",
		"最近在看《繁花》太好看了推荐。", "有时候觉得自己好普通好平凡。",
		"煮了一碗泡面加蛋，简单幸福。",
		"被小区保安记住名字了，有人认识的感觉真好。",
		"刚看完一部电影哭得稀里哗啦的。", "对象说想和我结婚有点慌。",
		"爸妈催婚催到不想回家过年了。", "终于买了心心念念的耳机。",
		"一个人去医院做胃镜，孤独等级拉满。",
		"同事离职了好舍不得。", "辅导孩子做作业好崩溃。",
		"房贷还完了开心。",
	}
	for i, p := range posts {
		db.Where("content = ?", p).FirstOrCreate(&model.TreeHolePost{
			UserID: uids[i%3], Content: p, Scope: "public",
			IsAnonymous: i%3 != 0, Hugs: int64(rand.Intn(50) + 1),
			CreatedAt: time.Now().Add(-time.Duration(i*3) * time.Hour),
		})
	}
	logger.Infof("树洞帖子: %d 条", len(posts))

	courses := []model.Course{
		{Title: "如何在职场中设置情绪边界", Duration: 480, Category: "情绪管理", Tag: "职场场景", SortOrder: 30},
		{Title: "情绪爆发之后，怎么修复关系？", Duration: 360, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 29},
		{Title: "自我接纳：允许自己偶尔脆弱", Duration: 300, Category: "成长", Tag: "成长", SortOrder: 28},
		{Title: "你通常在周三情绪最低", Duration: 240, Category: "情绪管理", Tag: "绪安洞察", SortOrder: 27},
		{Title: "独居场景：一个人的放松指南", Duration: 420, Category: "焦虑治愈", Tag: "独居场景", SortOrder: 26},
		{Title: "睡前记录：今天想说…", Duration: 180, Category: "睡前助眠", Tag: "睡前助眠", SortOrder: 25},
		{Title: "呼吸训练入门：4-7-8呼吸法", Duration: 300, Category: "睡前助眠", Tag: "练习计划", SortOrder: 24},
		{Title: "职场解压：5分钟快速放松", Duration: 300, Category: "职场解压", Tag: "职场场景", SortOrder: 23},
		{Title: "恋爱中的情绪沟通", Duration: 540, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 22},
		{Title: "10天稳情计划：从内耗到松弛", Duration: 600, Category: "成长", Tag: "练习计划", SortOrder: 21},
		{Title: "社交焦虑：你不是一个人", Duration: 400, Category: "焦虑治愈", Tag: "社交场景", SortOrder: 20},
		{Title: "学会说不的七个技巧", Duration: 350, Category: "成长", Tag: "职场场景", SortOrder: 19},
		{Title: "正念饮食：用吃治愈情绪", Duration: 250, Category: "成长", Tag: "成长", SortOrder: 18},
		{Title: "失去之后如何走出来", Duration: 500, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 17},
		{Title: "5分钟办公室冥想", Duration: 300, Category: "职场解压", Tag: "练习计划", SortOrder: 16},
		{Title: "自卑与超越：找到内在力量", Duration: 450, Category: "成长", Tag: "成长", SortOrder: 15},
		{Title: "睡前身体扫描放松法", Duration: 200, Category: "睡前助眠", Tag: "练习计划", SortOrder: 14},
		{Title: "焦虑发作时的紧急自救", Duration: 280, Category: "焦虑治愈", Tag: "绪安洞察", SortOrder: 13},
		{Title: "高敏感人格的生存指南", Duration: 520, Category: "成长", Tag: "成长", SortOrder: 12},
		{Title: "和压力做朋友", Duration: 380, Category: "职场解压", Tag: "职场场景", SortOrder: 11},
		{Title: "情绪日记怎么写才有用", Duration: 220, Category: "情绪管理", Tag: "练习计划", SortOrder: 10},
		{Title: "原生家庭与你的情绪模式", Duration: 600, Category: "成长", Tag: "成长", SortOrder: 9},
		{Title: "愤怒管理：不伤害的表达", Duration: 400, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 8},
		{Title: "睡眠卫生：今晚开始好眠", Duration: 180, Category: "睡前助眠", Tag: "睡前助眠", SortOrder: 7},
		{Title: "停止精神内耗的五种方法", Duration: 350, Category: "焦虑治愈", Tag: "绪安洞察", SortOrder: 6},
		{Title: "用颜色治愈情绪", Duration: 420, Category: "成长", Tag: "成长", SortOrder: 5},
		{Title: "考试面试前的放松练习", Duration: 200, Category: "职场解压", Tag: "练习计划", SortOrder: 4},
		{Title: "独处的艺术：享受一个人的时光", Duration: 320, Category: "成长", Tag: "独居场景", SortOrder: 3},
		{Title: "替代性创伤：如何保护你的情绪", Duration: 480, Category: "情绪管理", Tag: "绪安洞察", SortOrder: 2},
		{Title: "感恩日记：每天都值得被记录", Duration: 160, Category: "成长", Tag: "练习计划", SortOrder: 1},
	}
	for _, c := range courses {
		db.Where("title = ?", c.Title).FirstOrCreate(&c)
	}
	logger.Infof("课程: %d 门", len(courses))

	now := time.Now()
	db.Where("user_id = ?", uids[1]).FirstOrCreate(&model.MemberInfo{
		UserID: uids[1], MemberType: "yearly", Status: "active",
		StartDate: timePtr(now.AddDate(0, -3, 0)),
		EndDate: timePtr(now.AddDate(0, 9, 0)), AutoRenew: true,
	})
	db.Where("user_id = ?", uids[2]).FirstOrCreate(&model.MemberInfo{
		UserID: uids[2], MemberType: "monthly", Status: "active",
		StartDate: timePtr(now.AddDate(0, 0, -10)),
		EndDate: timePtr(now.AddDate(0, 0, 20)), AutoRenew: true,
	})
	logger.Info("会员: 小明(年度) + 小红(月度)")

	msgs := []struct{ t, c, m string }{
		{"欢迎加入绪安", "感谢来到绪安。这里不评判，只陪伴。", "system"},
		{"会员权益升级通知", "你已开通会员，解锁全部课程和情绪趋势分析。", "member"},
		{"你有一份周报", "本周你记录了5次情绪，来看看变化吧。", "activity"},
		{"新课程上线", "《停止精神内耗的五种方法》已上线。", "system"},
		{"树洞回应了你", "你发布的帖子收到了18个抱抱。", "activity"},
		{"连续打卡提醒", "你已经连续打卡7天了！明天也要来。", "activity"},
		{"冥想完成", "完成了一次睡前助眠冥想，大脑放松了吗？", "activity"},
		{"隐私政策更新", "绪安隐私政策已于2026年6月更新。", "system"},
		{"情绪洞察", "你在周三情绪最低，周末会好一些。", "activity"},
		{"节日祝福", "记得吃粽子哦 🎋", "system"},
	}
	for i, uid := range uids {
		for j, m := range msgs {
			db.Create(&model.Message{UserID: uid, Title: m.t, Content: m.c, MsgType: m.m, IsRead: j > 3, CreatedAt: time.Now().Add(-time.Duration((i*10+j)*6) * time.Hour)})
		}
	}
	logger.Infof("消息: %d 条/人", len(msgs))
	logger.Info("✅ 绪安 100+ 种子数据填充完成！")
}

func timePtr(t time.Time) *time.Time { return &t }
