package main

import (
	"chillcat-server/internal/config"
	"chillcat-server/internal/model"
	"chillcat-server/pkg/logger"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, _ := config.Load()
	logger.Init("info", "console")
	logger.Info("填充绪安种子数据...")

	db, _ := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	db.AutoMigrate(&model.User{}, &model.MemberInfo{}, &model.MemberOrder{},
		&model.EmotionCheckin{}, &model.TreeHolePost{},
		&model.Course{}, &model.UserCourseProgress{})

	hash, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	user := &model.User{Username: "test", Email: "test@xuanpeace.app", Password: string(hash), Nickname: "安静的云雀", Status: 1}
	db.Where("username = ?", "test").FirstOrCreate(user)
	logger.Infof("测试用户: test / 123456 (ID=%d)", user.ID)

	emotions := []string{"平静", "开心", "疲惫", "焦虑", "委屈", "孤独", "烦躁", "迷茫", "易怒", "内耗"}
	notes := []string{"今天开会又被批评了…", "拿到了那个项目的正向反馈", "下午开会的时候leader说我的方案不够细致", "", "为什么明明没做错", "周末去了公园发呆", ""}
	for i := 0; i < 7; i++ {
		db.Create(&model.EmotionCheckin{UserID: user.ID, Emotion: emotions[i%10], Note: notes[i%7], CheckinDate: time.Now().AddDate(0, 0, -i).Format("2006-01-02")})
	}
	logger.Info("情绪打卡: 7 条")

	posts := []string{"下午开会的时候leader当着所有人说我的方案不够细致…", "周末去了公园，坐在草地上发呆了一个小时", "为什么明明没做错，还是觉得对不起所有人？"}
	for _, p := range posts {
		db.Create(&model.TreeHolePost{UserID: user.ID, Content: p, Scope: "public", IsAnonymous: true, Hugs: int64(len(p) % 20)})
	}
	logger.Info("树洞帖子: 3 条")

	courses := []model.Course{
		{Title: "如何在职场中设置情绪边界", Duration: 480, Category: "情绪管理", Tag: "职场场景", SortOrder: 10},
		{Title: "情绪爆发之后，怎么修复关系？", Duration: 360, Category: "情绪管理", Tag: "恋爱情绪", SortOrder: 9},
		{Title: "自我接纳：允许自己偶尔脆弱", Duration: 300, Category: "成长", Tag: "成长", SortOrder: 8},
		{Title: "独居场景：一个人的放松指南", Duration: 420, Category: "焦虑治愈", Tag: "独居场景", SortOrder: 7},
		{Title: "睡前记录：今天想说…", Duration: 180, Category: "睡前助眠", Tag: "睡前助眠", SortOrder: 6},
		{Title: "呼吸训练入门", Duration: 300, Category: "睡前助眠", Tag: "练习计划", SortOrder: 5},
	}
	for _, c := range courses { db.Where("title = ?", c.Title).FirstOrCreate(&c) }
	logger.Infof("课程: %d 门", len(courses))
	logger.Info("绪安种子数据填充完成！")
}
