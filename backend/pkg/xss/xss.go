package xss

import (
	"html"
	"regexp"
	"strings"
)

// XSSFilter XSS过滤器 对应Java后端的@Xss注解功能
type XSSFilter struct {
	// 危险标签正则表达式
	scriptPattern *regexp.Regexp
	// 危险属性正则表达式
	attrPattern *regexp.Regexp
}

// NewXSSFilter 创建XSS过滤器实例
func NewXSSFilter() *XSSFilter {
	return &XSSFilter{
		// 匹配script、iframe、object等危险标签（简化版本，避免反向引用）
		scriptPattern: regexp.MustCompile(`(?i)<\s*(script|iframe|object|embed|form|input|textarea|select|option|button|link|style|meta|base|frame|frameset|applet|bgsound|blink|comment|ilayer|layer|xml|import)[^>]*>.*?</\s*(script|iframe|object|embed|form|input|textarea|select|option|button|link|style|meta|base|frame|frameset|applet|bgsound|blink|comment|ilayer|layer|xml|import)\s*>|<\s*(script|iframe|object|embed|form|input|textarea|select|option|button|link|style|meta|base|frame|frameset|applet|bgsound|blink|comment|ilayer|layer|xml|import)[^>]*/?>`),
		// 匹配危险属性
		attrPattern: regexp.MustCompile(`(?i)\s*(on\w+|javascript:|vbscript:|data:|expression\(|@import|behavior:)\s*=`),
	}
}

// FilterString 过滤字符串中的XSS攻击代码 对应Java后端@Xss注解的功能
func (f *XSSFilter) FilterString(input string) string {
	if input == "" {
		return input
	}

	// 1. HTML实体编码
	filtered := html.EscapeString(input)

	// 2. 移除危险标签
	filtered = f.scriptPattern.ReplaceAllString(filtered, "")

	// 3. 移除危险属性
	filtered = f.attrPattern.ReplaceAllString(filtered, "")

	// 4. 移除其他危险字符序列
	filtered = f.removeDangerousSequences(filtered)

	return filtered
}

// removeDangerousSequences 移除其他危险字符序列
func (f *XSSFilter) removeDangerousSequences(input string) string {
	// 危险字符序列列表
	dangerousSequences := []string{
		"javascript:",
		"vbscript:",
		"data:",
		"expression(",
		"@import",
		"behavior:",
		"<script",
		"</script>",
		"<iframe",
		"</iframe>",
		"<object",
		"</object>",
		"<embed",
		"</embed>",
		"<form",
		"</form>",
		"<input",
		"<textarea",
		"<select",
		"<option",
		"<button",
		"<link",
		"<style",
		"</style>",
		"<meta",
		"<base",
		"<frame",
		"<frameset",
		"</frameset>",
		"<applet",
		"</applet>",
		"onload=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onchange=",
		"onsubmit=",
		"onerror=",
	}

	result := input
	for _, seq := range dangerousSequences {
		// 不区分大小写替换
		result = strings.ReplaceAll(strings.ToLower(result), strings.ToLower(seq), "")
	}

	return result
}

// ValidateXSS 验证字符串是否包含XSS攻击代码 对应Java后端@Xss注解的验证功能
func (f *XSSFilter) ValidateXSS(input string) bool {
	if input == "" {
		return true
	}

	// 检查是否包含危险标签
	if f.scriptPattern.MatchString(input) {
		return false
	}

	// 检查是否包含危险属性
	if f.attrPattern.MatchString(input) {
		return false
	}

	// 检查其他危险字符序列
	lowerInput := strings.ToLower(input)
	dangerousKeywords := []string{
		"javascript:",
		"vbscript:",
		"data:",
		"expression(",
		"@import",
		"behavior:",
		"<script",
		"<iframe",
		"<object",
		"<embed",
		"onload=",
		"onclick=",
		"onmouseover=",
		"onerror=",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(lowerInput, keyword) {
			return false
		}
	}

	return true
}

// 全局XSS过滤器实例
var globalXSSFilter = NewXSSFilter()

// FilterXSS 全局XSS过滤函数 对应Java后端@Xss注解的功能
func FilterXSS(input string) string {
	return globalXSSFilter.FilterString(input)
}

// ValidateXSS 全局XSS验证函数 对应Java后端@Xss注解的验证功能
func ValidateXSS(input string) bool {
	return globalXSSFilter.ValidateXSS(input)
}

// XSSValidationError XSS验证错误
type XSSValidationError struct {
	Field   string
	Message string
}

func (e *XSSValidationError) Error() string {
	return e.Message
}

// ValidateXSSForStruct 为结构体字段进行XSS验证 对应Java后端@Xss注解的结构体验证
func ValidateXSSForStruct(data interface{}, fieldName string, value string) error {
	if !ValidateXSS(value) {
		return &XSSValidationError{
			Field:   fieldName,
			Message: fieldName + "不能包含脚本字符",
		}
	}
	return nil
}
