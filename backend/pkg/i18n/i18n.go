package i18n

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// I18nManager 国际化管理器 对应Java后端的MessageSource
type I18nManager struct {
	messages map[string]map[string]string // [language][key]message
	mutex    sync.RWMutex
	basePath string
}

// 全局国际化管理器实例
var globalI18nManager *I18nManager
var once sync.Once

// GetI18nManager 获取全局国际化管理器实例
func GetI18nManager() *I18nManager {
	once.Do(func() {
		globalI18nManager = NewI18nManager(MessageBasePath)
	})
	return globalI18nManager
}

// NewI18nManager 创建国际化管理器
func NewI18nManager(basePath string) *I18nManager {
	manager := &I18nManager{
		messages: make(map[string]map[string]string),
		basePath: basePath,
	}
	
	// 加载所有语言的消息文件
	manager.LoadAllMessages()
	
	return manager
}

// LoadAllMessages 加载所有语言的消息文件
func (m *I18nManager) LoadAllMessages() error {
	fmt.Printf("I18nManager.LoadAllMessages: 开始加载国际化消息文件\n")
	
	for _, lang := range SupportedLanguages {
		if err := m.LoadMessages(lang); err != nil {
			fmt.Printf("I18nManager.LoadAllMessages: 加载语言 %s 失败: %v\n", lang, err)
			// 继续加载其他语言，不因为一个语言失败而停止
		}
	}
	
	fmt.Printf("I18nManager.LoadAllMessages: 国际化消息文件加载完成\n")
	return nil
}

// LoadMessages 加载指定语言的消息文件
func (m *I18nManager) LoadMessages(lang string) error {
	filePath := GetMessageFilePath(lang)
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("I18nManager.LoadMessages: 消息文件不存在: %s\n", filePath)
		// 如果是默认语言文件不存在，创建默认消息
		if lang == DefaultLanguage {
			return m.createDefaultMessages(lang)
		}
		return nil // 非默认语言文件不存在时不报错
	}
	
	// 读取文件内容
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取消息文件失败: %v", err)
	}
	
	// 解析JSON内容
	var messages map[string]string
	if err := json.Unmarshal(content, &messages); err != nil {
		return fmt.Errorf("解析消息文件失败: %v", err)
	}
	
	// 存储消息
	m.mutex.Lock()
	m.messages[lang] = messages
	m.mutex.Unlock()
	
	fmt.Printf("I18nManager.LoadMessages: 成功加载语言 %s，消息数量: %d\n", lang, len(messages))
	return nil
}

// createDefaultMessages 创建默认消息
func (m *I18nManager) createDefaultMessages(lang string) error {
	fmt.Printf("I18nManager.createDefaultMessages: 创建默认消息文件: %s\n", lang)
	
	// 默认中文消息 对应Java后端的messages.properties
	defaultMessages := map[string]string{
		MsgNotNull:                    "* 必须填写",
		MsgSuccess:                    "操作成功",
		MsgFail:                       "操作失败",
		MsgError:                      "系统错误",
		MsgUnknown:                    "未知错误",
		MsgUserJcaptchaError:          "验证码错误",
		MsgUserJcaptchaExpire:         "验证码已失效",
		MsgUserNotExists:              "用户不存在/密码错误",
		MsgUserPasswordNotMatch:       "用户不存在/密码错误",
		MsgUserPasswordRetryCount:     "密码输入错误{0}次",
		MsgUserPasswordRetryExceed:    "密码输入错误{0}次，帐户锁定{1}分钟",
		MsgUserPasswordDelete:         "对不起，您的账号已被删除",
		MsgUserBlocked:                "用户已封禁，请联系管理员",
		MsgUserLogoutSuccess:          "退出成功",
		MsgUserLoginSuccess:           "登录成功",
		MsgUserRegisterSuccess:        "注册成功",
		MsgUserNotfound:               "请重新登录",
		MsgUserForcelogout:            "管理员强制退出，请重新登录",
		MsgUserUnknownError:           "未知错误，请重新登录",
		MsgRoleBlocked:                "角色已封禁，请联系管理员",
		MsgLoginBlocked:               "很遗憾，访问IP已被列入系统黑名单",
		MsgLengthNotValid:             "长度必须在{min}到{max}个字符之间",
		MsgUserUsernameNotValid:       "* 2到20个汉字、字母、数字或下划线组成，且必须以非数字开头",
		MsgUserPasswordNotValid:       "* 5-50个字符",
		MsgUserEmailNotValid:          "邮箱格式错误",
		MsgUserMobileNotValid:         "手机号格式错误",
		MsgUploadExceedMaxSize:        "上传的文件大小超出限制的文件大小！允许的文件最大大小是：{0}MB！",
		MsgUploadFilenameExceedLength: "上传的文件名最长{0}个字符",
		MsgNoPermission:               "您没有数据的权限，请联系管理员添加权限 [{0}]",
		MsgNoCreatePermission:         "您没有创建数据的权限，请联系管理员添加权限 [{0}]",
		MsgNoUpdatePermission:         "您没有修改数据的权限，请联系管理员添加权限 [{0}]",
		MsgNoDeletePermission:         "您没有删除数据的权限，请联系管理员添加权限 [{0}]",
		MsgNoExportPermission:         "您没有导出数据的权限，请联系管理员添加权限 [{0}]",
		MsgNoViewPermission:           "您没有查看数据的权限，请联系管理员添加权限 [{0}]",
		MsgOperationSuccess:           "操作成功",
		MsgOperationFail:              "操作失败",
		MsgAddSuccess:                 "新增成功",
		MsgUpdateSuccess:              "修改成功",
		MsgDeleteSuccess:              "删除成功",
		MsgQuerySuccess:               "查询成功",
		MsgExportSuccess:              "导出成功",
		MsgImportSuccess:              "导入成功",
		MsgDataNotExists:              "数据不存在",
		MsgDataExists:                 "数据已存在",
		MsgDataInvalid:                "数据无效",
		MsgParamInvalid:               "参数无效",
		MsgParamMissing:               "参数缺失",
		MsgSystemError:                "系统错误",
		MsgSystemBusy:                 "系统繁忙，请稍后再试",
		MsgSystemMaintenance:          "系统维护中",
		MsgNetworkError:               "网络错误",
		MsgTimeout:                    "请求超时",
	}
	
	// 存储到内存
	m.mutex.Lock()
	m.messages[lang] = defaultMessages
	m.mutex.Unlock()
	
	// 保存到文件
	return m.saveMessages(lang, defaultMessages)
}

// saveMessages 保存消息到文件
func (m *I18nManager) saveMessages(lang string, messages map[string]string) error {
	filePath := GetMessageFilePath(lang)
	
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	
	// 序列化为JSON
	content, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}
	
	// 写入文件
	if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("写入消息文件失败: %v", err)
	}
	
	fmt.Printf("I18nManager.saveMessages: 成功保存消息文件: %s\n", filePath)
	return nil
}

// GetMessage 获取消息 对应Java后端的MessageUtils.message方法
func (m *I18nManager) GetMessage(lang, key string, args ...interface{}) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 获取指定语言的消息
	if messages, exists := m.messages[lang]; exists {
		if message, found := messages[key]; found {
			return m.formatMessage(message, args...)
		}
	}
	
	// 如果指定语言没有找到，尝试默认语言
	if lang != DefaultLanguage {
		if messages, exists := m.messages[DefaultLanguage]; exists {
			if message, found := messages[key]; found {
				return m.formatMessage(message, args...)
			}
		}
	}
	
	// 如果都没有找到，返回键名
	return key
}

// formatMessage 格式化消息，支持参数替换
func (m *I18nManager) formatMessage(message string, args ...interface{}) string {
	if len(args) == 0 {
		return message
	}
	
	// 简单的参数替换，支持 {0}, {1}, {2} 等格式
	result := message
	for i, arg := range args {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", arg))
	}
	
	// 支持命名参数，如 {min}, {max}
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

// GetSupportedLanguages 获取支持的语言列表
func (m *I18nManager) GetSupportedLanguages() []string {
	return SupportedLanguages
}

// GetLoadedLanguages 获取已加载的语言列表
func (m *I18nManager) GetLoadedLanguages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var languages []string
	for lang := range m.messages {
		languages = append(languages, lang)
	}
	return languages
}

// ReloadMessages 重新加载消息文件
func (m *I18nManager) ReloadMessages() error {
	fmt.Printf("I18nManager.ReloadMessages: 重新加载国际化消息文件\n")
	
	m.mutex.Lock()
	m.messages = make(map[string]map[string]string)
	m.mutex.Unlock()
	
	return m.LoadAllMessages()
}

// AddMessage 添加消息
func (m *I18nManager) AddMessage(lang, key, message string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.messages[lang] == nil {
		m.messages[lang] = make(map[string]string)
	}
	m.messages[lang][key] = message
}

// GetMessageCount 获取消息数量
func (m *I18nManager) GetMessageCount(lang string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if messages, exists := m.messages[lang]; exists {
		return len(messages)
	}
	return 0
}
