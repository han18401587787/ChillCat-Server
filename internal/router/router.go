package router

import (
	"chillcat-server/internal/config"
	"chillcat-server/internal/handler"
	"chillcat-server/internal/middleware"
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/internal/service"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/validator"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Setup(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)
	validator.RegisterCustomValidators()

	db := initDB(cfg)
	autoMigrate(db)

	// Repositories
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	emotionRepo := repository.NewEmotionRepo(db)
	treeholeRepo := repository.NewTreeHoleRepo(db)
	courseRepo := repository.NewCourseRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	resonanceRepo := repository.NewResonanceRepo(db)
	encourageRepo := repository.NewEncourageRepo(db)
	healingRepo := repository.NewHealingRepo(db)

	// Services
	userService := service.NewUserService(userRepo, memberRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	memberService := service.NewMemberService(memberRepo, userRepo)
	emotionService := service.NewEmotionService(emotionRepo)
	treeholeService := service.NewTreeHoleService(treeholeRepo)
	courseService := service.NewCourseService(courseRepo)
	resonanceService := service.NewResonanceService(resonanceRepo)
	encourageService := service.NewEncourageService(encourageRepo)

	// AI 服务（无数据库依赖，使用本地规则引擎）
	aiService := service.NewAIService()

	// 情绪解码服务（依赖 AI 服务）
	emotionDecodeService := service.NewEmotionDecodeService(aiService)

	// 稳情计划服务（依赖数据库 + AI 服务）
	healingService := service.NewHealingService(healingRepo, aiService)

	// Handlers
	authHandler := handler.NewAuthHandler(userService, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userHandler := handler.NewUserHandler(userService)
	memberHandler := handler.NewMemberHandler(memberService)
	emotionHandler := handler.NewEmotionHandler(emotionService)
	treeholeHandler := handler.NewTreeHoleHandler(treeholeService)
	courseHandler := handler.NewCourseHandler(courseService, commentRepo)
	resonanceHandler := handler.NewResonanceHandler(resonanceService)
	encourageHandler := handler.NewEncourageHandler(encourageService)
	healthHandler := handler.NewHealthHandler()
	aiHandler := handler.NewAIHandler(aiService)
	emotionDecodeHandler := handler.NewEmotionDecodeHandler(emotionDecodeService)
	healingHandler := handler.NewHealingHandler(healingService)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(middleware.SlowRequestLog(500 * time.Millisecond))
	r.Use(gin.Recovery())

	r.GET("/health", healthHandler.Check)

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/anonymous", authHandler.AnonymousLogin)
		}

		authorized := v1.Group("")
		authorized.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// 用户
			authorized.GET("/user/profile", userHandler.GetProfile)

			// 会员
			member := authorized.Group("/member")
			{
				member.GET("/info", memberHandler.GetInfo)
				member.GET("/products", memberHandler.GetProducts)
				member.GET("/privileges", memberHandler.GetPrivileges)
				member.POST("/purchase", memberHandler.Purchase)
				member.GET("/history", memberHandler.GetOrderHistory)
			}

			// 情绪打卡
			emotion := authorized.Group("/emotion")
			{
				emotion.POST("/checkin", emotionHandler.Checkin)
				emotion.GET("/today", emotionHandler.GetToday)
				emotion.GET("/journal", emotionHandler.Journal)
				emotion.GET("/weekly-stats", emotionHandler.WeeklyStats)
				emotion.GET("/alerts", emotionHandler.Alerts)
			}

			// 情绪解码器（无需数据库，复用 AI 服务）
			authorized.POST("/emotion/decode", emotionDecodeHandler.Decode)

			// 稳情计划
			healing := authorized.Group("/healing")
			{
				healing.GET("/plan", healingHandler.GetPlan)
				healing.POST("/plan/generate", healingHandler.GeneratePlan)
				healing.POST("/plan/tasks/:id/complete", healingHandler.CompleteTask)
			}

			// 树洞
			th := authorized.Group("/treehole")
			{
				th.POST("/posts", middleware.ContentModeration(), treeholeHandler.CreatePost)
				th.GET("/posts", treeholeHandler.ListPosts)
				th.POST("/posts/:id/hug", treeholeHandler.AddHug)
			}

			// 共鸣墙
			resonance := authorized.Group("/resonance")
			{
				resonance.POST("/stories", middleware.ContentModeration(), resonanceHandler.CreateStory)
				resonance.GET("/stories", resonanceHandler.ListStories)
				resonance.GET("/stories/:id", resonanceHandler.GetStory)
				resonance.POST("/stories/:id/resonate", resonanceHandler.Resonate)
				resonance.GET("/stories/:id/resonators", resonanceHandler.GetResonators)
			}

			// 鼓励链
			encourage := authorized.Group("/encourage")
			{
				encourage.POST("/chains", middleware.ContentModeration(), encourageHandler.CreateChain)
				encourage.GET("/chains", encourageHandler.ListChains)
				encourage.GET("/chains/:id", encourageHandler.GetChain)
				encourage.POST("/chains/:id/join", middleware.ContentModeration(), encourageHandler.JoinChain)
				encourage.GET("/my-chains", encourageHandler.ListMyChains)
			}

			// 课程
			authorized.GET("/courses", courseHandler.List)
			authorized.POST("/courses/:id/complete", courseHandler.MarkComplete)
			authorized.GET("/courses/:id/comments", courseHandler.ListComments)
			authorized.POST("/courses/:id/comments", courseHandler.AddComment)

			// AI 情绪分析
			ai := authorized.Group("/ai")
			{
				ai.POST("/empathy", aiHandler.Empathy)
				ai.POST("/analyze", aiHandler.Analyze)
			}
		}
	}

	return r
}

func initDB(cfg *config.Config) *gorm.DB {
	logLevel := gormlogger.Warn
	if cfg.Server.Mode == "debug" {
		logLevel = gormlogger.Info
	}
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil {
		logger.Fatalf("数据库连接失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	logger.Info("数据库连接成功")
	return db
}

func autoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&model.User{}, &model.MemberInfo{}, &model.MemberOrder{},
		&model.EmotionCheckin{}, &model.TreeHolePost{},
		&model.Course{}, &model.UserCourseProgress{}, &model.CourseComment{},
		&model.ResonanceStory{}, &model.ResonanceRecord{},
		&model.EncourageChain{}, &model.EncourageLink{},
		&model.HealingPlan{}, &model.HealingPlanTask{},
	); err != nil {
		logger.Fatalf("数据库迁移失败: %v", err)
	}
	logger.Info("数据库迁移完成")
}
