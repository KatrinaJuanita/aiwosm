package config

import (
	customValidator "wosm/pkg/validator"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// InitValidator 初始化验证器
func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册自定义验证器
		customValidator.RegisterCustomValidators(v)
	}
}
