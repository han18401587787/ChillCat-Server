package validator

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators 在 Gin 的 binding validator 引擎上注册自定义校验规则。
// 应在路由初始化之前调用一次。
func RegisterCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("username", validateUsername)
		_ = v.RegisterValidation("password", validatePassword)
		_ = v.RegisterValidation("member_type", validateMemberType)
	}
}

// validateUsername 用户名：4-32位，字母数字下划线
func validateUsername(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]{4,32}$`, fl.Field().String())
	return matched
}

// validatePassword 密码：6-64位
func validatePassword(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^.{6,64}$`, fl.Field().String())
	return matched
}

// validateMemberType 会员类型校验
func validateMemberType(fl validator.FieldLevel) bool {
	validTypes := map[string]bool{
		"monthly":   true,
		"quarterly": true,
		"yearly":    true,
		"permanent": true,
	}
	return validTypes[fl.Field().String()]
}
