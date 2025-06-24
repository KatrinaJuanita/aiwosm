package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// CustomValidator 自定义验证器
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator 创建新的验证器实例
func NewValidator() *CustomValidator {
	v := validator.New()
	
	// 注册自定义验证规则
	v.RegisterValidation("role_name", validateRoleName)
	v.RegisterValidation("role_key", validateRoleKey)
	v.RegisterValidation("menu_name", validateMenuName)
	v.RegisterValidation("perms", validatePerms)
	
	// 注册字段名翻译
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	
	return &CustomValidator{validator: v}
}

// Validate 验证结构体
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return cv.translateError(err)
	}
	return nil
}

// translateError 翻译验证错误为用户友好的消息
func (cv *CustomValidator) translateError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrors {
			message := cv.getErrorMessage(e)
			messages = append(messages, message)
		}
		return fmt.Errorf(strings.Join(messages, "; "))
	}
	return err
}

// getErrorMessage 获取错误消息
func (cv *CustomValidator) getErrorMessage(e validator.FieldError) string {
	field := cv.getFieldName(e.Field())
	
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "min":
		return fmt.Sprintf("%s长度不能少于%s个字符", field, e.Param())
	case "max":
		return fmt.Sprintf("%s长度不能超过%s个字符", field, e.Param())
	case "len":
		return fmt.Sprintf("%s长度必须为%s个字符", field, e.Param())
	case "role_name":
		return fmt.Sprintf("%s格式不正确，只能包含中文、英文、数字和下划线", field)
	case "role_key":
		return fmt.Sprintf("%s格式不正确，只能包含英文、数字、冒号和下划线", field)
	case "menu_name":
		return fmt.Sprintf("%s格式不正确，不能包含特殊字符", field)
	case "perms":
		return fmt.Sprintf("%s格式不正确，权限标识格式应为：模块:功能:操作", field)
	default:
		return fmt.Sprintf("%s验证失败", field)
	}
}

// getFieldName 获取字段中文名称
func (cv *CustomValidator) getFieldName(field string) string {
	fieldNames := map[string]string{
		"roleName":   "角色名称",
		"roleKey":    "角色权限",
		"roleSort":   "角色排序",
		"dataScope":  "数据范围",
		"status":     "状态",
		"menuName":   "菜单名称",
		"parentId":   "父菜单ID",
		"orderNum":   "显示顺序",
		"path":       "路由地址",
		"component":  "组件路径",
		"menuType":   "菜单类型",
		"visible":    "显示状态",
		"perms":      "权限标识",
		"icon":       "菜单图标",
	}
	
	if name, exists := fieldNames[field]; exists {
		return name
	}
	return field
}

// 自定义验证规则

// validateRoleName 验证角色名称
func validateRoleName(fl validator.FieldLevel) bool {
	roleName := fl.Field().String()
	if roleName == "" {
		return false
	}
	
	// 角色名称只能包含中文、英文、数字和下划线，长度1-30
	matched, _ := regexp.MatchString(`^[\u4e00-\u9fa5a-zA-Z0-9_]{1,30}$`, roleName)
	return matched
}

// validateRoleKey 验证角色权限字符
func validateRoleKey(fl validator.FieldLevel) bool {
	roleKey := fl.Field().String()
	if roleKey == "" {
		return false
	}
	
	// 角色权限字符只能包含英文、数字、冒号和下划线，长度1-100
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9:_]{1,100}$`, roleKey)
	return matched
}

// validateMenuName 验证菜单名称
func validateMenuName(fl validator.FieldLevel) bool {
	menuName := fl.Field().String()
	if menuName == "" {
		return false
	}
	
	// 菜单名称不能包含特殊字符，长度1-50
	matched, _ := regexp.MatchString(`^[\u4e00-\u9fa5a-zA-Z0-9_\-\s]{1,50}$`, menuName)
	return matched
}

// validatePerms 验证权限标识
func validatePerms(fl validator.FieldLevel) bool {
	perms := fl.Field().String()
	if perms == "" {
		return true // 权限标识可以为空
	}
	
	// 权限标识格式：模块:功能:操作，如 system:role:list
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+:[a-zA-Z0-9_]+:[a-zA-Z0-9_]+$`, perms)
	return matched
}
