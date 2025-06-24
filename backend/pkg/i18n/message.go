package i18n

import (
	"fmt"
	"strings"
)

// MessageUtils 消息工具类 对应Java后端的MessageUtils
type MessageUtils struct{}

// Message 获取国际化消息 对应Java后端的MessageUtils.message方法
func Message(key string, args ...interface{}) string {
	return MessageWithLang(DefaultLanguage, key, args...)
}

// MessageWithLang 获取指定语言的国际化消息
func MessageWithLang(lang, key string, args ...interface{}) string {
	manager := GetI18nManager()
	return manager.GetMessage(lang, key, args...)
}

// MessageWithContext 从上下文获取语言并返回消息
func MessageWithContext(lang, key string, args ...interface{}) string {
	// 标准化语言代码
	normalizedLang := GetLanguageCode(lang)
	return MessageWithLang(normalizedLang, key, args...)
}

// 便捷的消息获取方法

// ErrorMessage 获取错误消息
func ErrorMessage(lang, key string, args ...interface{}) string {
	return MessageWithLang(lang, key, args...)
}

// SuccessMessage 获取成功消息
func SuccessMessage(lang string) string {
	return MessageWithLang(lang, MsgSuccess)
}

// FailMessage 获取失败消息
func FailMessage(lang string) string {
	return MessageWithLang(lang, MsgFail)
}

// ValidationMessage 获取验证消息
func ValidationMessage(lang, field string, validationType string, params map[string]interface{}) string {
	var key string
	switch validationType {
	case "required":
		key = MsgNotNull
	case "length":
		key = MsgLengthNotValid
	case "email":
		key = MsgUserEmailNotValid
	case "mobile":
		key = MsgUserMobileNotValid
	case "username":
		key = MsgUserUsernameNotValid
	case "password":
		key = MsgUserPasswordNotValid
	default:
		key = MsgParamInvalid
	}
	
	return MessageWithLang(lang, key, params)
}

// PermissionMessage 获取权限相关消息
func PermissionMessage(lang, operation, resource string) string {
	var key string
	switch operation {
	case "create", "add":
		key = MsgNoCreatePermission
	case "update", "edit":
		key = MsgNoUpdatePermission
	case "delete", "remove":
		key = MsgNoDeletePermission
	case "export":
		key = MsgNoExportPermission
	case "view", "query":
		key = MsgNoViewPermission
	default:
		key = MsgNoPermission
	}
	
	return MessageWithLang(lang, key, resource)
}

// OperationMessage 获取操作相关消息
func OperationMessage(lang, operation string, success bool) string {
	var key string
	if success {
		switch operation {
		case "add", "create":
			key = MsgAddSuccess
		case "update", "edit":
			key = MsgUpdateSuccess
		case "delete", "remove":
			key = MsgDeleteSuccess
		case "query", "search":
			key = MsgQuerySuccess
		case "export":
			key = MsgExportSuccess
		case "import":
			key = MsgImportSuccess
		default:
			key = MsgOperationSuccess
		}
	} else {
		key = MsgOperationFail
	}
	
	return MessageWithLang(lang, key)
}

// UserMessage 获取用户相关消息
func UserMessage(lang, messageType string, args ...interface{}) string {
	var key string
	switch messageType {
	case "login.success":
		key = MsgUserLoginSuccess
	case "logout.success":
		key = MsgUserLogoutSuccess
	case "register.success":
		key = MsgUserRegisterSuccess
	case "not.exists":
		key = MsgUserNotExists
	case "password.not.match":
		key = MsgUserPasswordNotMatch
	case "blocked":
		key = MsgUserBlocked
	case "jcaptcha.error":
		key = MsgUserJcaptchaError
	case "jcaptcha.expire":
		key = MsgUserJcaptchaExpire
	case "password.retry.count":
		key = MsgUserPasswordRetryCount
	case "password.retry.exceed":
		key = MsgUserPasswordRetryExceed
	case "notfound":
		key = MsgUserNotfound
	case "forcelogout":
		key = MsgUserForcelogout
	default:
		key = MsgUserUnknownError
	}
	
	return MessageWithLang(lang, key, args...)
}

// SystemMessage 获取系统相关消息
func SystemMessage(lang, messageType string) string {
	var key string
	switch messageType {
	case "error":
		key = MsgSystemError
	case "busy":
		key = MsgSystemBusy
	case "maintenance":
		key = MsgSystemMaintenance
	case "network.error":
		key = MsgNetworkError
	case "timeout":
		key = MsgTimeout
	default:
		key = MsgUnknown
	}
	
	return MessageWithLang(lang, key)
}

// UploadMessage 获取文件上传相关消息
func UploadMessage(lang, messageType string, args ...interface{}) string {
	var key string
	switch messageType {
	case "exceed.maxSize":
		key = MsgUploadExceedMaxSize
	case "filename.exceed.length":
		key = MsgUploadFilenameExceedLength
	default:
		key = MsgError
	}
	
	return MessageWithLang(lang, key, args...)
}

// FormatMessage 格式化消息，支持多种参数格式
func FormatMessage(message string, args ...interface{}) string {
	if len(args) == 0 {
		return message
	}
	
	result := message
	
	// 支持位置参数 {0}, {1}, {2}
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", arg))
	}
	
	// 支持命名参数 {name}, {value}
	if len(args) > 0 {
		if argMap, ok := args[0].(map[string]interface{}); ok {
			for key, value := range argMap {
				placeholder := fmt.Sprintf("{%s}", key)
				result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
			}
		}
	}
	
	return result
}

// GetLanguageInfo 获取语言信息
func GetLanguageInfo() map[string]interface{} {
	manager := GetI18nManager()
	
	info := make(map[string]interface{})
	info["defaultLanguage"] = DefaultLanguage
	info["supportedLanguages"] = SupportedLanguages
	info["loadedLanguages"] = manager.GetLoadedLanguages()
	
	// 语言显示名称
	displayNames := make(map[string]string)
	for _, lang := range SupportedLanguages {
		displayNames[lang] = GetLanguageDisplayName(lang)
	}
	info["displayNames"] = displayNames
	
	// 消息统计
	messageStats := make(map[string]int)
	for _, lang := range manager.GetLoadedLanguages() {
		messageStats[lang] = manager.GetMessageCount(lang)
	}
	info["messageStats"] = messageStats
	
	return info
}

// ValidateLanguage 验证语言代码
func ValidateLanguage(lang string) (string, bool) {
	normalizedLang := GetLanguageCode(lang)
	return normalizedLang, IsValidLanguage(normalizedLang)
}

// GetAvailableLanguages 获取可用的语言列表（已加载的）
func GetAvailableLanguages() []map[string]string {
	manager := GetI18nManager()
	loadedLanguages := manager.GetLoadedLanguages()
	
	var languages []map[string]string
	for _, lang := range loadedLanguages {
		languages = append(languages, map[string]string{
			"code":        lang,
			"name":        GetLanguageDisplayName(lang),
			"messageCount": fmt.Sprintf("%d", manager.GetMessageCount(lang)),
		})
	}
	
	return languages
}

// ReloadMessages 重新加载所有消息
func ReloadMessages() error {
	manager := GetI18nManager()
	return manager.ReloadMessages()
}

// AddCustomMessage 添加自定义消息
func AddCustomMessage(lang, key, message string) {
	manager := GetI18nManager()
	manager.AddMessage(lang, key, message)
}

// GetMessageKeys 获取所有消息键
func GetMessageKeys(lang string) []string {
	manager := GetI18nManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	if messages, exists := manager.messages[lang]; exists {
		var keys []string
		for key := range messages {
			keys = append(keys, key)
		}
		return keys
	}
	
	return []string{}
}

// ExportMessages 导出指定语言的所有消息
func ExportMessages(lang string) map[string]string {
	manager := GetI18nManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	if messages, exists := manager.messages[lang]; exists {
		// 创建副本避免并发问题
		result := make(map[string]string)
		for key, value := range messages {
			result[key] = value
		}
		return result
	}
	
	return make(map[string]string)
}
