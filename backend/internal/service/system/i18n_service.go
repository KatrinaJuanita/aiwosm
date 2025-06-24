package system

import (
	"fmt"
	"wosm/pkg/i18n"
)

// I18nService 国际化服务 对应Java后端的国际化相关服务
type I18nService struct {
	i18nManager *i18n.I18nManager
}

// NewI18nService 创建国际化服务实例
func NewI18nService() *I18nService {
	return &I18nService{
		i18nManager: i18n.GetI18nManager(),
	}
}

// GetMessage 获取国际化消息
func (s *I18nService) GetMessage(lang, key string, args ...interface{}) string {
	fmt.Printf("I18nService.GetMessage: 获取消息, Lang=%s, Key=%s\n", lang, key)
	return s.i18nManager.GetMessage(lang, key, args...)
}

// GetSupportedLanguages 获取支持的语言列表
func (s *I18nService) GetSupportedLanguages() []string {
	fmt.Printf("I18nService.GetSupportedLanguages: 获取支持的语言列表\n")
	return s.i18nManager.GetSupportedLanguages()
}

// GetLoadedLanguages 获取已加载的语言列表
func (s *I18nService) GetLoadedLanguages() []string {
	fmt.Printf("I18nService.GetLoadedLanguages: 获取已加载的语言列表\n")
	return s.i18nManager.GetLoadedLanguages()
}

// GetLanguageInfo 获取语言信息
func (s *I18nService) GetLanguageInfo() map[string]interface{} {
	fmt.Printf("I18nService.GetLanguageInfo: 获取语言信息\n")
	return i18n.GetLanguageInfo()
}

// ReloadMessages 重新加载消息
func (s *I18nService) ReloadMessages() error {
	fmt.Printf("I18nService.ReloadMessages: 重新加载消息\n")
	return s.i18nManager.ReloadMessages()
}

// ValidateLanguage 验证语言代码
func (s *I18nService) ValidateLanguage(lang string) (string, bool) {
	fmt.Printf("I18nService.ValidateLanguage: 验证语言代码, Lang=%s\n", lang)
	return i18n.ValidateLanguage(lang)
}

// GetAvailableLanguages 获取可用的语言列表
func (s *I18nService) GetAvailableLanguages() []map[string]string {
	fmt.Printf("I18nService.GetAvailableLanguages: 获取可用的语言列表\n")
	return i18n.GetAvailableLanguages()
}

// AddCustomMessage 添加自定义消息
func (s *I18nService) AddCustomMessage(lang, key, message string) {
	fmt.Printf("I18nService.AddCustomMessage: 添加自定义消息, Lang=%s, Key=%s\n", lang, key)
	s.i18nManager.AddMessage(lang, key, message)
}

// GetMessageCount 获取指定语言的消息数量
func (s *I18nService) GetMessageCount(lang string) int {
	fmt.Printf("I18nService.GetMessageCount: 获取消息数量, Lang=%s\n", lang)
	return s.i18nManager.GetMessageCount(lang)
}

// ExportMessages 导出指定语言的所有消息
func (s *I18nService) ExportMessages(lang string) map[string]string {
	fmt.Printf("I18nService.ExportMessages: 导出消息, Lang=%s\n", lang)
	return i18n.ExportMessages(lang)
}

// GetMessageKeys 获取指定语言的所有消息键
func (s *I18nService) GetMessageKeys(lang string) []string {
	fmt.Printf("I18nService.GetMessageKeys: 获取消息键, Lang=%s\n", lang)
	return i18n.GetMessageKeys(lang)
}

// FormatMessage 格式化消息
func (s *I18nService) FormatMessage(message string, args ...interface{}) string {
	return i18n.FormatMessage(message, args...)
}

// GetUserMessage 获取用户相关消息
func (s *I18nService) GetUserMessage(lang, messageType string, args ...interface{}) string {
	fmt.Printf("I18nService.GetUserMessage: 获取用户消息, Lang=%s, Type=%s\n", lang, messageType)
	return i18n.UserMessage(lang, messageType, args...)
}

// GetSystemMessage 获取系统相关消息
func (s *I18nService) GetSystemMessage(lang, messageType string) string {
	fmt.Printf("I18nService.GetSystemMessage: 获取系统消息, Lang=%s, Type=%s\n", lang, messageType)
	return i18n.SystemMessage(lang, messageType)
}

// GetValidationMessage 获取验证相关消息
func (s *I18nService) GetValidationMessage(lang, field, validationType string, params map[string]interface{}) string {
	fmt.Printf("I18nService.GetValidationMessage: 获取验证消息, Lang=%s, Field=%s, Type=%s\n", lang, field, validationType)
	return i18n.ValidationMessage(lang, field, validationType, params)
}

// GetPermissionMessage 获取权限相关消息
func (s *I18nService) GetPermissionMessage(lang, operation, resource string) string {
	fmt.Printf("I18nService.GetPermissionMessage: 获取权限消息, Lang=%s, Operation=%s, Resource=%s\n", lang, operation, resource)
	return i18n.PermissionMessage(lang, operation, resource)
}

// GetOperationMessage 获取操作相关消息
func (s *I18nService) GetOperationMessage(lang, operation string, success bool) string {
	fmt.Printf("I18nService.GetOperationMessage: 获取操作消息, Lang=%s, Operation=%s, Success=%v\n", lang, operation, success)
	return i18n.OperationMessage(lang, operation, success)
}

// GetUploadMessage 获取上传相关消息
func (s *I18nService) GetUploadMessage(lang, messageType string, args ...interface{}) string {
	fmt.Printf("I18nService.GetUploadMessage: 获取上传消息, Lang=%s, Type=%s\n", lang, messageType)
	return i18n.UploadMessage(lang, messageType, args...)
}

// InitializeI18n 初始化国际化系统
func (s *I18nService) InitializeI18n() error {
	fmt.Printf("I18nService.InitializeI18n: 初始化国际化系统\n")
	
	// 加载所有消息文件
	if err := s.i18nManager.LoadAllMessages(); err != nil {
		return fmt.Errorf("初始化国际化系统失败: %v", err)
	}
	
	fmt.Printf("I18nService.InitializeI18n: 国际化系统初始化完成\n")
	return nil
}

// GetLanguageStatistics 获取语言统计信息
func (s *I18nService) GetLanguageStatistics() map[string]interface{} {
	fmt.Printf("I18nService.GetLanguageStatistics: 获取语言统计信息\n")
	
	stats := make(map[string]interface{})
	
	// 支持的语言数量
	stats["supportedLanguageCount"] = len(s.i18nManager.GetSupportedLanguages())
	
	// 已加载的语言数量
	loadedLanguages := s.i18nManager.GetLoadedLanguages()
	stats["loadedLanguageCount"] = len(loadedLanguages)
	
	// 每种语言的消息数量
	messageStats := make(map[string]int)
	for _, lang := range loadedLanguages {
		messageStats[lang] = s.i18nManager.GetMessageCount(lang)
	}
	stats["messageStats"] = messageStats
	
	// 总消息数量
	totalMessages := 0
	for _, count := range messageStats {
		totalMessages += count
	}
	stats["totalMessages"] = totalMessages
	
	// 默认语言
	stats["defaultLanguage"] = i18n.DefaultLanguage
	
	// 语言显示名称
	displayNames := make(map[string]string)
	for _, lang := range s.i18nManager.GetSupportedLanguages() {
		displayNames[lang] = i18n.GetLanguageDisplayName(lang)
	}
	stats["languageDisplayNames"] = displayNames
	
	return stats
}

// CheckMessageIntegrity 检查消息完整性
func (s *I18nService) CheckMessageIntegrity() map[string]interface{} {
	fmt.Printf("I18nService.CheckMessageIntegrity: 检查消息完整性\n")
	
	result := make(map[string]interface{})
	loadedLanguages := s.i18nManager.GetLoadedLanguages()
	
	if len(loadedLanguages) == 0 {
		result["status"] = "error"
		result["message"] = "没有加载任何语言"
		return result
	}
	
	// 以默认语言为基准检查其他语言的消息完整性
	defaultLang := i18n.DefaultLanguage
	defaultKeys := i18n.GetMessageKeys(defaultLang)
	
	missingKeys := make(map[string][]string)
	extraKeys := make(map[string][]string)
	
	for _, lang := range loadedLanguages {
		if lang == defaultLang {
			continue
		}
		
		langKeys := i18n.GetMessageKeys(lang)
		langKeyMap := make(map[string]bool)
		for _, key := range langKeys {
			langKeyMap[key] = true
		}
		
		// 检查缺失的键
		var missing []string
		for _, key := range defaultKeys {
			if !langKeyMap[key] {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			missingKeys[lang] = missing
		}
		
		// 检查多余的键
		defaultKeyMap := make(map[string]bool)
		for _, key := range defaultKeys {
			defaultKeyMap[key] = true
		}
		
		var extra []string
		for _, key := range langKeys {
			if !defaultKeyMap[key] {
				extra = append(extra, key)
			}
		}
		if len(extra) > 0 {
			extraKeys[lang] = extra
		}
	}
	
	result["status"] = "success"
	result["defaultLanguage"] = defaultLang
	result["defaultKeyCount"] = len(defaultKeys)
	result["missingKeys"] = missingKeys
	result["extraKeys"] = extraKeys
	
	// 计算完整性得分
	totalIssues := len(missingKeys) + len(extraKeys)
	if totalIssues == 0 {
		result["integrityScore"] = 100
		result["message"] = "所有语言的消息完整性良好"
	} else {
		result["integrityScore"] = 100 - (totalIssues * 10) // 简单的评分算法
		result["message"] = fmt.Sprintf("发现 %d 个完整性问题", totalIssues)
	}
	
	return result
}
