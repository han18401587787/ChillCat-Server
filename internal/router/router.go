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

	// Services
	userService := service.NewUserService(userRepo, memberRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	memberService := service.NewMemberService(memberRepo, userRepo)
	emotionService := service.NewEmotionService(emotionRepo)
	treeholeService := service.NewTreeHoleService(treeholeRepo)
	courseService := service.NewCourseService(courseRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(userService, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userHandler := handler.NewUserHandler(userService)
	memberHandler := handler.NewMemberHandler(memberService)
	emotionHandler := handler.NewEmotionHandler(emotionService)
	treeholeHandler := handler.NewTreeHoleHandler(treeholeService)
	courseHandler := handler.NewCourseHandler(courseService)
	healthHandler := handler.NewHealthHandler()

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
			}

			// 树洞
			th := authorized.Group("/treehole")
			{
				th.POST("/posts", treeholeHandler.CreatePost)
				th.GET("/posts", treeholeHandler.ListPosts)
				th.POST("/posts/:id/hug", treeholeHandler.AddHug)
			}

			// 课程
			authorized.GET("/courses", courseHandler.List)
			authorized.POST("/courses/:id/complete", courseHandler.MarkComplete)
		}
	}

	return r
}

func initDB(cfg *config.Config) *gorm.DB {
	logLevel := gormlogger.Warn
	if cfg.Server.Mode == "debug" { logLevel = gormlogger.Info }
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{Logger: gormlogger.Default.LogMode(logLevel)})
	if err != nil { logger.Fatalf("数据库连接失败: %v", err) }
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
		&model.Course{}, &model.UserCourseProgress{},
	); err != nil { logger.Fatalf("数据库迁移失败: %v", err) }
	logger.Info("数据库迁移完成")
}
