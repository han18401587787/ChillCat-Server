package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/logger"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	logger.Init("error", "console")
}

// 使用内存 SQLite 做单元测试
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	// 自动建表
	err = db.AutoMigrate(
		&model.User{},
		&model.MemberInfo{},
		&model.MemberOrder{},
			&model.FeedItem{},
	)
	if err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}

	return db
}

// =============================================
// 用户服务测试
// =============================================

func TestRegister_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewUserService(userRepo, memberRepo, "test-secret", 72)

	resp, code, err := svc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "test@chillcat.app",
		Password: "123456",
		Nickname: "测试用户",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.NotNil(t, resp)
	assert.Equal(t, "testuser", resp.Username)
	assert.NotEmpty(t, resp.Token)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewUserService(userRepo, memberRepo, "test-secret", 72)

	// 第一次注册
	_, _, _ = svc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "test@chillcat.app",
		Password: "123456",
	})

	// 重复用户名
	_, code, err := svc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "other@chillcat.app",
		Password: "123456",
	})

	assert.Error(t, err)
	assert.Equal(t, 20002, code) // ErrUserExists
}

func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewUserService(userRepo, memberRepo, "test-secret", 72)

	// 先注册
	_, _, _ = svc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "test@chillcat.app",
		Password: "123456",
	})

	// 登录
	resp, code, err := svc.Login(&LoginRequest{
		Username: "testuser",
		Password: "123456",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "testuser", resp.Username)
	assert.NotEmpty(t, resp.Token)
}

func TestLogin_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewUserService(userRepo, memberRepo, "test-secret", 72)

	// 先注册
	_, _, _ = svc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "test@chillcat.app",
		Password: "123456",
	})

	// 错误密码
	_, code, err := svc.Login(&LoginRequest{
		Username: "testuser",
		Password: "wrong_pass",
	})

	assert.Error(t, err)
	assert.Equal(t, 20003, code) // ErrPasswordWrong
}

func TestLogin_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewUserService(userRepo, memberRepo, "test-secret", 72)

	_, code, err := svc.Login(&LoginRequest{
		Username: "not_exist",
		Password: "123456",
	})

	assert.Error(t, err)
	assert.Equal(t, 20001, code) // ErrUserNotFound
}

// =============================================
// 会员服务测试
// =============================================

func TestGetMemberInfo_NonMember(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewMemberService(memberRepo, userRepo)

	// 非会员查询
	info, code, err := svc.GetMemberInfo(999)

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "none", info.Status)
	assert.Equal(t, false, info.IsValid)
}

func TestGetProducts(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewMemberService(memberRepo, userRepo)

	products, code, err := svc.GetProducts()

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, 4, len(products))

	// 验证商品类型
	types := make(map[string]bool)
	for _, p := range products {
		types[p.MemberType] = true
	}
	assert.True(t, types["monthly"])
	assert.True(t, types["quarterly"])
	assert.True(t, types["yearly"])
	assert.True(t, types["permanent"])
}

func TestGetPrivileges(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	svc := NewMemberService(memberRepo, userRepo)

	privileges, code, err := svc.GetPrivileges()

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, 5, len(privileges))
}

func TestPurchase_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	userSvc := NewUserService(userRepo, memberRepo, "test-secret", 72)
	memberSvc := NewMemberService(memberRepo, userRepo)

	// 先注册用户
	regResp, _, _ := userSvc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "test@chillcat.app",
		Password: "123456",
	})

	// 购买月度会员
	resp, code, err := memberSvc.Purchase(regResp.UserID, &PurchaseRequest{
		ProductID: "monthly_001",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.NotEmpty(t, resp.OrderNo)
	assert.True(t, resp.MemberInfo.IsValid)
	assert.Equal(t, "monthly", resp.MemberInfo.MemberType)

	// 验证会员信息
	info, _, _ := memberSvc.GetMemberInfo(regResp.UserID)
	assert.Equal(t, "active", info.Status)
	assert.True(t, info.IsValid)
}

func TestPurchase_Permanent(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	userSvc := NewUserService(userRepo, memberRepo, "test-secret", 72)
	memberSvc := NewMemberService(memberRepo, userRepo)

	regResp, _, _ := userSvc.Register(&RegisterRequest{
		Username: "perm_user",
		Email:    "perm@chillcat.app",
		Password: "123456",
	})

	resp, code, err := memberSvc.Purchase(regResp.UserID, &PurchaseRequest{
		ProductID: "permanent_001",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "permanent", resp.MemberInfo.MemberType)
	assert.True(t, resp.MemberInfo.IsValid)
	assert.Equal(t, -1, resp.MemberInfo.Remaining)
}

func TestPurchase_ProductNotFound(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	memberSvc := NewMemberService(memberRepo, userRepo)

	_, code, err := memberSvc.Purchase(1, &PurchaseRequest{
		ProductID: "not_exist",
	})

	assert.Error(t, err)
	assert.Equal(t, 30004, code) // ErrProductNotFound
}

func TestGetOrderHistory(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepo(db)
	memberRepo := repository.NewMemberRepo(db)
	userSvc := NewUserService(userRepo, memberRepo, "test-secret", 72)
	memberSvc := NewMemberService(memberRepo, userRepo)

	regResp, _, _ := userSvc.Register(&RegisterRequest{
		Username: "testuser",
		Email:    "test@chillcat.app",
		Password: "123456",
	})

	// 购买两次
	_, _, _ = memberSvc.Purchase(regResp.UserID, &PurchaseRequest{ProductID: "monthly_001"})
	_, _, _ = memberSvc.Purchase(regResp.UserID, &PurchaseRequest{ProductID: "yearly_001"})

	// 查询历史
	result, code, err := memberSvc.GetOrderHistory(regResp.UserID, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, code)

	data := result.(map[string]interface{})
	assert.Equal(t, int64(2), data["total"])
}

// =============================================
// 内容流服务测试

func TestFeedService_EmptyList(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepo(db)
	svc := NewFeedService(feedRepo)

	result, code, err := svc.ListFeeds(1, 10)

	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Total)
	list, ok := result.List.([]FeedItemVO)
	assert.True(t, ok)
	assert.Equal(t, 0, len(list))
}

func TestFeedService_Pagination(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepo(db)
	svc := NewFeedService(feedRepo)

	for i := 0; i < 15; i++ {
		db.Create(&model.FeedItem{
			Title:       fmt.Sprintf("测试内容 %d", i+1),
			Subtitle:    fmt.Sprintf("描述 %d", i+1),
			ContentType: "article",
			Status:      1,
			SortOrder:   i,
		})
	}

	result, code, err := svc.ListFeeds(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, int64(15), result.Total)
	list, _ := result.List.([]FeedItemVO)
	assert.Equal(t, 10, len(list))

	result, code, err = svc.ListFeeds(2, 10)
	assert.NoError(t, err)
	list2, _ := result.List.([]FeedItemVO)
	assert.Equal(t, 5, len(list2))
}

func TestFeedService_DefaultPageSize(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepo(db)
	svc := NewFeedService(feedRepo)

	result, code, err := svc.ListFeeds(0, 100)
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
}

func TestFeedService_GetDetail(t *testing.T) {
	db := setupTestDB(t)

	item := &model.FeedItem{
		Title:       "详情测试",
		Subtitle:    "副标题",
		ContentType: "music",
		Status:      1,
	}
	db.Create(item)

	feedRepo := repository.NewFeedRepo(db)
	svc := NewFeedService(feedRepo)

	result, code, err := svc.GetFeedDetail(item.ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, "详情测试", result.Title)
	assert.Equal(t, "music", result.ContentType)
}

func TestFeedService_GetDetail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepo(db)
	svc := NewFeedService(feedRepo)

	_, code, err := svc.GetFeedDetail(9999)
	assert.Error(t, err)
	assert.Equal(t, 10004, code)
}
