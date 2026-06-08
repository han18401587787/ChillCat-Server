package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 通用错误码
const (
	CodeSuccess        = 0
	ErrBadRequest      = 10001
	ErrUnauthorized    = 10002
	ErrForbidden       = 10003
	ErrNotFound        = 10004
	ErrRateLimit       = 10005
	ErrInternal        = 10006
	ErrValidation      = 10007

	// 用户模块 (20000-29999)
	ErrUserNotFound    = 20001
	ErrUserExists      = 20002
	ErrPasswordWrong   = 20003
	ErrUserDisabled    = 20004

	// 会员模块 (30000-39999)
	ErrMemberNotFound  = 30001
	ErrMemberExpired   = 30002
	ErrPurchaseFailed  = 30003
	ErrProductNotFound = 30004
	ErrInvalidProduct  = 30005
)

var codeMsg = map[int]string{
	CodeSuccess:            "success",
	ErrBadRequest:      "请求参数错误",
	ErrUnauthorized:    "登录已过期，请重新登录",
	ErrForbidden:       "无权限访问",
	ErrNotFound:        "请求的资源不存在",
	ErrRateLimit:       "请求过于频繁，请稍后再试",
	ErrInternal:        "服务器内部错误",
	ErrValidation:      "参数校验失败",

	ErrUserNotFound:    "用户不存在",
	ErrUserExists:      "用户已存在",
	ErrPasswordWrong:   "密码错误",
	ErrUserDisabled:    "账号已被禁用",

	ErrMemberNotFound:  "会员信息不存在",
	ErrMemberExpired:   "会员已过期",
	ErrPurchaseFailed:  "购买失败，请稍后再试",
	ErrProductNotFound: "商品不存在",
	ErrInvalidProduct:  "无效的商品",
}

// GetMsg 获取错误码对应的消息
func GetMsg(code int) string {
	if msg, ok := codeMsg[code]; ok {
		return msg
	}
	return "未知错误"
}

// JSON 返回 JSON 响应
func JSON(c *gin.Context, code int, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: GetMsg(code),
		Data:    data,
	})
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	JSON(c, CodeSuccess, data)
}

// Error 错误响应
func Error(c *gin.Context, code int) {
	JSON(c, code, nil)
}

// ErrorWithMsg 带自定义消息的错误响应
func ErrorWithMsg(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	})
}

// Page 分页响应
type Page struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
