package i18n

// 国际化常量定义 对应Java后端的I18nConfig和Constants

const (
	// 支持的语言类型
	LanguageZhCN = "zh-CN" // 简体中文
	LanguageZhTW = "zh-TW" // 繁体中文
	LanguageEnUS = "en-US" // 英语（美国）
	LanguageJaJP = "ja-JP" // 日语
	LanguageKoKR = "ko-KR" // 韩语

	// 默认语言
	DefaultLanguage = LanguageZhCN

	// 语言参数名 对应Java后端的LocaleChangeInterceptor.setParamName
	LanguageParamName = "lang"

	// 语言Cookie名
	LanguageCookieName = "language"

	// 语言Header名
	LanguageHeaderName = "Accept-Language"

	// 国际化资源文件路径
	MessageBasePath = "configs/i18n"

	// 消息文件前缀
	MessageFilePrefix = "messages"

	// 消息文件扩展名
	MessageFileExtension = ".json"
)

// 语言映射表
var LanguageMap = map[string]string{
	"zh":    LanguageZhCN,
	"zh-cn": LanguageZhCN,
	"zh-CN": LanguageZhCN,
	"zh-tw": LanguageZhTW,
	"zh-TW": LanguageZhTW,
	"en":    LanguageEnUS,
	"en-us": LanguageEnUS,
	"en-US": LanguageEnUS,
	"ja":    LanguageJaJP,
	"ja-jp": LanguageJaJP,
	"ja-JP": LanguageJaJP,
	"ko":    LanguageKoKR,
	"ko-kr": LanguageKoKR,
	"ko-KR": LanguageKoKR,
}

// 支持的语言列表
var SupportedLanguages = []string{
	LanguageZhCN,
	LanguageZhTW,
	LanguageEnUS,
	LanguageJaJP,
	LanguageKoKR,
}

// 语言显示名称
var LanguageDisplayNames = map[string]string{
	LanguageZhCN: "简体中文",
	LanguageZhTW: "繁體中文",
	LanguageEnUS: "English",
	LanguageJaJP: "日本語",
	LanguageKoKR: "한국어",
}

// 消息键常量 对应Java后端的messages.properties
const (
	// 通用消息
	MsgNotNull = "not.null"
	MsgSuccess = "success"
	MsgFail    = "fail"
	MsgError   = "error"
	MsgUnknown = "unknown.error"

	// 用户相关消息
	MsgUserJcaptchaError       = "user.jcaptcha.error"
	MsgUserJcaptchaExpire      = "user.jcaptcha.expire"
	MsgUserNotExists           = "user.not.exists"
	MsgUserPasswordNotMatch    = "user.password.not.match"
	MsgUserPasswordRetryCount  = "user.password.retry.limit.count"
	MsgUserPasswordRetryExceed = "user.password.retry.limit.exceed"
	MsgUserPasswordDelete      = "user.password.delete"
	MsgUserBlocked             = "user.blocked"
	MsgUserLogoutSuccess       = "user.logout.success"
	MsgUserLoginSuccess        = "user.login.success"
	MsgUserRegisterSuccess     = "user.register.success"
	MsgUserNotfound            = "user.notfound"
	MsgUserForcelogout         = "user.forcelogout"
	MsgUserUnknownError        = "user.unknown.error"

	// 角色相关消息
	MsgRoleBlocked = "role.blocked"

	// 登录相关消息
	MsgLoginBlocked = "login.blocked"

	// 验证相关消息
	MsgLengthNotValid       = "length.not.valid"
	MsgUserUsernameNotValid = "user.username.not.valid"
	MsgUserPasswordNotValid = "user.password.not.valid"
	MsgUserEmailNotValid    = "user.email.not.valid"
	MsgUserMobileNotValid   = "user.mobile.phone.number.not.valid"

	// 文件上传消息
	MsgUploadExceedMaxSize        = "upload.exceed.maxSize"
	MsgUploadFilenameExceedLength = "upload.filename.exceed.length"

	// 权限相关消息
	MsgNoPermission       = "no.permission"
	MsgNoCreatePermission = "no.create.permission"
	MsgNoUpdatePermission = "no.update.permission"
	MsgNoDeletePermission = "no.delete.permission"
	MsgNoExportPermission = "no.export.permission"
	MsgNoViewPermission   = "no.view.permission"

	// 操作相关消息
	MsgOperationSuccess = "operation.success"
	MsgOperationFail    = "operation.fail"
	MsgAddSuccess       = "add.success"
	MsgUpdateSuccess    = "update.success"
	MsgDeleteSuccess    = "delete.success"
	MsgQuerySuccess     = "query.success"
	MsgExportSuccess    = "export.success"
	MsgImportSuccess    = "import.success"

	// 数据验证消息
	MsgDataNotExists = "data.not.exists"
	MsgDataExists    = "data.exists"
	MsgDataInvalid   = "data.invalid"
	MsgParamInvalid  = "param.invalid"
	MsgParamMissing  = "param.missing"

	// 系统相关消息
	MsgSystemError       = "system.error"
	MsgSystemBusy        = "system.busy"
	MsgSystemMaintenance = "system.maintenance"
	MsgNetworkError      = "network.error"
	MsgTimeout           = "timeout"
)

// GetLanguageCode 获取标准语言代码
func GetLanguageCode(lang string) string {
	if code, exists := LanguageMap[lang]; exists {
		return code
	}
	return DefaultLanguage
}

// IsValidLanguage 检查语言是否有效
func IsValidLanguage(lang string) bool {
	for _, supported := range SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// GetLanguageDisplayName 获取语言显示名称
func GetLanguageDisplayName(lang string) string {
	if name, exists := LanguageDisplayNames[lang]; exists {
		return name
	}
	return lang
}

// GetMessageFilePath 获取消息文件路径
func GetMessageFilePath(lang string) string {
	// 将语言代码转换为文件名格式，如 zh-CN -> zh
	var fileCode string
	switch lang {
	case LanguageZhCN:
		fileCode = "zh"
	case LanguageZhTW:
		fileCode = "zh_tw"
	case LanguageEnUS:
		fileCode = "en"
	case LanguageJaJP:
		fileCode = "ja"
	case LanguageKoKR:
		fileCode = "ko"
	default:
		fileCode = "zh"
	}

	return MessageBasePath + "/" + MessageFilePrefix + "_" + fileCode + MessageFileExtension
}

// GetDefaultMessageFilePath 获取默认消息文件路径
func GetDefaultMessageFilePath() string {
	return GetMessageFilePath(DefaultLanguage)
}
