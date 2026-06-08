package service

import (
	"chillcat-server/internal/model"
	"chillcat-server/internal/repository"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"gorm.io/gorm"
)

// MemberService 会员服务
type MemberService struct {
	memberRepo *repository.MemberRepo
	userRepo   *repository.UserRepo
}

// NewMemberService 创建会员服务
func NewMemberService(memberRepo *repository.MemberRepo, userRepo *repository.UserRepo) *MemberService {
	return &MemberService{
		memberRepo: memberRepo,
		userRepo:   userRepo,
	}
}

// MemberInfoResponse 会员信息响应
type MemberInfoResponse struct {
	MemberType string `json:"member_type"`
	Status     string `json:"status"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
	AutoRenew  bool   `json:"auto_renew"`
	IsValid    bool   `json:"is_valid"`
	Remaining  int    `json:"remaining_days"` // 永久会员为 -1
}

// GetMemberInfo 获取会员信息
func (s *MemberService) GetMemberInfo(userID int64) (*MemberInfoResponse, int, error) {
	info, err := s.memberRepo.GetByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 非会员返回默认信息
			return &MemberInfoResponse{
				MemberType: "",
				Status:     "none",
				IsValid:    false,
				Remaining:  0,
			}, response.CodeSuccess, nil
		}
		logger.Errorf("查询会员信息失败: %v", err)
		return nil, response.ErrInternal, err
	}

	resp := &MemberInfoResponse{
		MemberType: info.MemberType,
		Status:     info.Status,
		AutoRenew:  info.AutoRenew,
		IsValid:    info.IsValid(),
		Remaining:  info.RemainingDays(),
	}

	if info.StartDate != nil {
		resp.StartDate = info.StartDate.Format(time.RFC3339)
	}
	if info.EndDate != nil {
		resp.EndDate = info.EndDate.Format(time.RFC3339)
	}

	return resp, response.CodeSuccess, nil
}

// Product 商品定义
type Product struct {
	ID            string  `json:"id"`
	MemberType    string  `json:"member_type"`
	DisplayName   string  `json:"display_name"`
	Price         float64 `json:"price"`
	OriginalPrice float64 `json:"original_price,omitempty"`
	DiscountTag   string  `json:"discount_tag,omitempty"`
	DurationDays  int     `json:"duration_days"` // 永久会员为 -1
}

// GetProducts 获取商品列表
func (s *MemberService) GetProducts() ([]Product, int, error) {
	products := []Product{
		{
			ID:           "monthly_001",
			MemberType:   "monthly",
			DisplayName:  "月度会员",
			Price:        15.00,
			DurationDays: 30,
		},
		{
			ID:           "quarterly_001",
			MemberType:   "quarterly",
			DisplayName:  "季度会员",
			Price:        40.00,
			OriginalPrice: 45.00,
			DiscountTag:  "省 11%",
			DurationDays: 90,
		},
		{
			ID:           "yearly_001",
			MemberType:   "yearly",
			DisplayName:  "年度会员",
			Price:        128.00,
			OriginalPrice: 180.00,
			DiscountTag:  "省 29%",
			DurationDays: 365,
		},
		{
			ID:           "permanent_001",
			MemberType:   "permanent",
			DisplayName:  "永久会员",
			Price:        399.00,
			OriginalPrice: 0,
			DiscountTag:  "一次购买，永久有效",
			DurationDays: -1,
		},
	}

	return products, response.CodeSuccess, nil
}

// Privilege 权益定义
type Privilege struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Icon          string   `json:"icon"`
	IsHighlight   bool     `json:"is_highlight"`
	AvailableFor  []string `json:"available_for"` // 可用会员类型
}

// GetPrivileges 获取权益列表
func (s *MemberService) GetPrivileges() ([]Privilege, int, error) {
	privileges := []Privilege{
		{
			ID:           "unlimited_play",
			Title:        "无限畅听",
			Description:  "畅享全曲库，无限制播放",
			Icon:         "music_note",
			IsHighlight:  true,
			AvailableFor: []string{"monthly", "quarterly", "yearly", "permanent"},
		},
		{
			ID:           "high_quality",
			Title:        "无损音质",
			Description:  "享受 Hi-Res 无损音质",
			Icon:         "high_quality",
			IsHighlight:  true,
			AvailableFor: []string{"quarterly", "yearly", "permanent"},
		},
		{
			ID:           "download",
			Title:        "离线下载",
			Description:  "歌曲离线下载，无网络也能听",
			Icon:         "download",
			IsHighlight:  false,
			AvailableFor: []string{"monthly", "quarterly", "yearly", "permanent"},
		},
		{
			ID:           "no_ad",
			Title:        "去广告",
			Description:  "免广告打扰，纯净体验",
			Icon:         "block",
			IsHighlight:  false,
			AvailableFor: []string{"monthly", "quarterly", "yearly", "permanent"},
		},
		{
			ID:           "exclusive_content",
			Title:        "专属内容",
			Description:  "会员专属歌曲、歌单、直播回放",
			Icon:         "star",
			IsHighlight:  false,
			AvailableFor: []string{"yearly", "permanent"},
		},
	}

	return privileges, response.CodeSuccess, nil
}

// PurchaseRequest 购买请求
type PurchaseRequest struct {
	ProductID string `json:"product_id" binding:"required"`
}

// PurchaseResponse 购买响应
type PurchaseResponse struct {
	OrderNo    string `json:"order_no"`
	MemberInfo *MemberInfoResponse `json:"member_info"`
}

// Purchase 购买会员
func (s *MemberService) Purchase(userID int64, req *PurchaseRequest) (*PurchaseResponse, int, error) {
	// 查找商品
	products, _, _ := s.GetProducts()
	var selectedProduct *Product
	for _, p := range products {
		if p.ID == req.ProductID {
			selectedProduct = &p
			break
		}
	}

	if selectedProduct == nil {
		return nil, response.ErrProductNotFound, fmt.Errorf("商品不存在: %s", req.ProductID)
	}

	// 生成订单号（毫秒时间戳+用户ID+随机数确保唯一）
	orderNo := fmt.Sprintf("CC%d%d%d", time.Now().UnixMilli(), userID%1000, rand.IntN(10000))

	now := time.Now()
	var startDate, endDate *time.Time
	startDate = &now

	if selectedProduct.DurationDays == -1 {
		// 永久会员，endDate 为 nil
		endDate = nil
	} else {
		t := now.AddDate(0, 0, selectedProduct.DurationDays)
		endDate = &t
	}

	// 创建订单
	order := &model.MemberOrder{
		UserID:        userID,
		OrderNo:       orderNo,
		MemberType:    selectedProduct.MemberType,
		OrderStatus:   "success",
		Amount:        selectedProduct.Price,
		PaymentMethod: "mock", // 阶段一使用 mock 支付
		StartDate:     startDate,
		EndDate:       endDate,
		PaidAt:        &now,
	}

	if err := s.memberRepo.CreateOrder(order); err != nil {
		logger.Errorf("创建订单失败: %v", err)
		return nil, response.ErrInternal, err
	}

	// 更新会员信息
	memberInfo := &model.MemberInfo{
		UserID:     userID,
		MemberType: selectedProduct.MemberType,
		Status:     "active",
		StartDate:  startDate,
		EndDate:    endDate,
		AutoRenew:  selectedProduct.DurationDays != -1, // 永久会员不自动续费
	}

	if err := s.memberRepo.Upsert(memberInfo); err != nil {
		logger.Errorf("更新会员信息失败: %v", err)
		return nil, response.ErrInternal, err
	}

	// 重新查询会员信息
	info, _, _ := s.GetMemberInfo(userID)

	return &PurchaseResponse{
		OrderNo:    orderNo,
		MemberInfo: info,
	}, response.CodeSuccess, nil
}

// GetOrderHistory 获取购买历史
func (s *MemberService) GetOrderHistory(userID int64, page, pageSize int) (interface{}, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}

	orders, total, err := s.memberRepo.GetOrdersByUserID(userID, page, pageSize)
	if err != nil {
		logger.Errorf("查询订单历史失败: %v", err)
		return nil, response.ErrInternal, err
	}

	type OrderVO struct {
		OrderNo       string  `json:"order_no"`
		MemberType    string  `json:"member_type"`
		OrderStatus   string  `json:"order_status"`
		Amount        float64 `json:"amount"`
		PaymentMethod string  `json:"payment_method"`
		CreatedAt     string  `json:"created_at"`
	}

	var list []OrderVO
	for _, o := range orders {
		list = append(list, OrderVO{
			OrderNo:       o.OrderNo,
			MemberType:    o.MemberType,
			OrderStatus:   o.OrderStatus,
			Amount:        o.Amount,
			PaymentMethod: o.PaymentMethod,
			CreatedAt:     o.CreatedAt.Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}, response.CodeSuccess, nil
}
