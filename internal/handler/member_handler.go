package handler

import (
	"chillcat-server/internal/service"
	"chillcat-server/pkg/logger"
	"chillcat-server/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MemberHandler 会员接口
type MemberHandler struct {
	memberService *service.MemberService
}

// NewMemberHandler 创建会员接口
func NewMemberHandler(memberService *service.MemberService) *MemberHandler {
	return &MemberHandler{memberService: memberService}
}

// GetInfo 获取会员信息
func (h *MemberHandler) GetInfo(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	info, code, err := h.memberService.GetMemberInfo(userID)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, info)
}

// GetProducts 获取商品列表
func (h *MemberHandler) GetProducts(c *gin.Context) {
	products, code, err := h.memberService.GetProducts()
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, products)
}

// GetPrivileges 获取权益列表
func (h *MemberHandler) GetPrivileges(c *gin.Context) {
	privileges, code, err := h.memberService.GetPrivileges()
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, privileges)
}

// Purchase 购买会员
func (h *MemberHandler) Purchase(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req service.PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrBadRequest)
		return
	}

	resp, code, err := h.memberService.Purchase(userID, &req)
	if err != nil {
		logger.Errorf("购买会员失败 user_id=%d product_id=%s: %v", userID, req.ProductID, err)
		response.Error(c, code)
		return
	}

	response.Success(c, resp)
}

// GetOrderHistory 获取购买历史
func (h *MemberHandler) GetOrderHistory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	result, code, err := h.memberService.GetOrderHistory(userID, page, pageSize)
	if err != nil {
		response.Error(c, code)
		return
	}

	response.Success(c, result)
}
