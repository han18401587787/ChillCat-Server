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

// Setup 设置路由
func Setup(cfg *config.Config) *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

		// 注册自定义校验规则
		validator.RegisterCustomValidators()

	// 初始化数据库
	db := initDB(cfg)

	// 自动迁移
	autoMigrate(db)

	// 初始化 Repository
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
		feedRepo := repository.NewFeedRepo(db)
		msgRepo := repository.NewMessageRepo(db)

	// 初始化 Service
	userService := service.NewUserService(userRepo, memberRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	memberService := service.NewMemberService(memberRepo, userRepo)
		feedService := service.NewFeedService(feedRepo)
		msgService := service.NewMessageService(msgRepo)

	// 初始化 Handler
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(userService, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userHandler := handler.NewUserHandler(userService)
	memberHandler := handler.NewMemberHandler(memberService)
		feedHandler := handler.NewFeedHandler(feedService)
		msgHandler := handler.NewMessageHandler(msgService)

	// 创建路由
	r := gin.New()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit())
	r.Use(middleware.SlowRequestLog(500 * time.Millisecond))
	r.Use(gin.Recovery())

	// 健康检查（无需认证）
	r.GET("/health", healthHandler.Check)

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 认证接口（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 需要认证的接口
		authorized := v1.Group("")
		authorized.Use(middleware.Auth(cfg.JWT.Secret))
		{
			// 用户接口
			authorized.GET("/user/profile", userHandler.GetProfile)

			// 会员接口
			member := authorized.Group("/member")
			{
				member.GET("/info", memberHandler.GetInfo)
				member.GET("/products", memberHandler.GetProducts)
				member.GET("/privileges", memberHandler.GetPrivileges)
				member.POST("/purchase", memberHandler.Purchase)
				member.GET("/history", memberHandler.GetOrderHistory)
			}

			// 内容流接口
			authorized.GET("/feeds", feedHandler.ListFeeds)
			authorized.GET("/feeds/:id", feedHandler.GetDetail)
			authorized.GET("/search", feedHandler.Search)

			// 消息接口
			authorized.GET("/messages", msgHandler.List)
			authorized.GET("/messages/unread", msgHandler.UnreadCount)
			authorized.POST("/messages/:id/read", msgHandler.MarkRead)
		}
	}

	return r
}

// initDB 初始化数据库连接
func initDB(cfg *config.Config) *gorm.DB {
	logLevel := gormlogger.Warn
	if cfg.Server.Mode == "debug" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		logger.Fatalf("数据库连接失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatalf("获取数据库实例失败: %v", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)

	logger.Info("数据库连接成功")
	return db
}

// autoMigrate 自动迁移
func autoMigrate(db *gorm.DB) {
	if err := db.AutoMigrate(
		&model.User{},
		&model.MemberInfo{},
		&model.MemberOrder{},
			&model.FeedItem{},
			&model.Message{},
	); err != nil {
		logger.Fatalf("数据库迁移失败: %v", err)
	}
	logger.Info("数据库迁移完成")
}
