package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// XSSValidator XSS验证器 - 对应Java后端的@Xss注解
func XSSValidator(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// 如果值为空，则通过验证（由required标签处理空值）
	if value == "" {
		return true
	}

	// 检查是否包含HTML标签（参考Java后端XssValidator的实现）
	htmlPattern := `<(\S*?)[^>]*>.*?|<.*? />`
	matched, err := regexp.MatchString(htmlPattern, value)
	if err != nil {
		return false
	}

	// 如果匹配到HTML标签，则验证失败
	if matched {
		return false
	}

	// 检查常见的XSS攻击模式
	xssPatterns := []string{
		`<script[^>]*>.*?</script>`,
		`javascript:`,
		`vbscript:`,
		`onload=`,
		`onerror=`,
		`onclick=`,
		`onmouseover=`,
		`onmouseout=`,
		`onfocus=`,
		`onblur=`,
		`onchange=`,
		`onsubmit=`,
		`onreset=`,
		`<iframe[^>]*>.*?</iframe>`,
		`<object[^>]*>.*?</object>`,
		`<embed[^>]*>.*?</embed>`,
		`<form[^>]*>.*?</form>`,
		`<input[^>]*>`,
		`<textarea[^>]*>`,
		`<select[^>]*>`,
		`expression\(`,
		`url\(javascript:`,
		`url\(vbscript:`,
		`<meta[^>]*>`,
		`<link[^>]*>`,
		`<style[^>]*>.*?</style>`,
		`<base[^>]*>`,
		`document\.cookie`,
		`document\.write`,
		`window\.location`,
		`eval\(`,
		`setTimeout\(`,
		`setInterval\(`,
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range xssPatterns {
		matched, err := regexp.MatchString(pattern, lowerValue)
		if err != nil {
			continue
		}
		if matched {
			return false
		}
	}

	return true
}

// RegisterCustomValidators 注册自定义验证器
func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("xss", XSSValidator)
}

// ErrorMessages 中文错误信息映射
var ErrorMessages = map[string]string{
	"required": "不能为空",
	"email":    "邮箱格式不正确",
	"max":      "长度不能超过%s个字符",
	"min":      "长度不能少于%s个字符",
	"oneof":    "值必须是%s中的一个",
	"xss":      "不能包含脚本字符",
}

// GetFieldName 获取字段中文名称
func GetFieldName(field string) string {
	fieldNames := map[string]string{
		"UserName":       "用户账号",
		"NickName":       "用户昵称",
		"Email":          "用户邮箱",
		"RoleName":       "角色名称",
		"RoleKey":        "角色权限字符",
		"RoleSort":       "角色排序",
		"DeptName":       "部门名称",
		"PostName":       "岗位名称",
		"PostCode":       "岗位编码",
		"PostSort":       "岗位排序",
		"NoticeTitle":    "公告标题",
		"NoticeType":     "公告类型",
		"NoticeContent":  "公告内容",
		"ConfigName":     "参数名称",
		"ConfigKey":      "参数键名",
		"ConfigValue":    "参数键值",
		"DictLabel":      "字典标签",
		"DictValue":      "字典键值",
		"DictType":       "字典类型",
		"DictSort":       "字典排序",
		"MenuName":       "菜单名称",
		"MenuType":       "菜单类型",
		"OrderNum":       "显示顺序",
		"Path":           "路由地址",
		"Component":      "组件路径",
		"Perms":          "权限标识",
		"Icon":           "菜单图标",
		"JobName":        "任务名称",
		"JobGroup":       "任务组名",
		"InvokeTarget":   "调用目标字符串",
		"CronExpression": "cron执行表达式",
		"MisfirePolicy":  "计划执行错误策略",
		"Concurrent":     "是否并发执行",
		"Status":         "状态",
		"Remark":         "备注",
	}

	if name, exists := fieldNames[field]; exists {
		return name
	}
	return field
}

// TranslateError 翻译验证错误为中文
func TranslateError(err validator.ValidationErrors) map[string]string {
	errors := make(map[string]string)

	for _, e := range err {
		field := e.Field()
		tag := e.Tag()
		param := e.Param()

		switch tag {
		case "required":
			errors[field] = GetFieldName(field) + ErrorMessages["required"]
		case "email":
			errors[field] = ErrorMessages["email"]
		case "max":
			errors[field] = GetFieldName(field) + fmt.Sprintf("长度不能超过%s个字符", param)
		case "min":
			errors[field] = GetFieldName(field) + fmt.Sprintf("长度不能少于%s个字符", param)
		case "oneof":
			errors[field] = GetFieldName(field) + fmt.Sprintf("值必须是[%s]中的一个", param)
		case "xss":
			errors[field] = GetFieldName(field) + ErrorMessages["xss"]
		default:
			errors[field] = GetFieldName(field) + "格式不正确"
		}
	}

	return errors
}
