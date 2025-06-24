package middleware

import (
	"fmt"
	"strings"
	"wosm/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// I18nMiddleware 国际化中间件 对应Java后端的LocaleChangeInterceptor
func I18nMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("I18nMiddleware: 处理国际化语言设置\n")
		
		// 获取语言设置，优先级：URL参数 > Header > Cookie > 默认语言
		lang := detectLanguage(ctx)
		
		// 验证和标准化语言代码
		normalizedLang, isValid := i18n.ValidateLanguage(lang)
		if !isValid {
			normalizedLang = i18n.DefaultLanguage
		}
		
		// 设置语言到上下文
		ctx.Set("language", normalizedLang)
		ctx.Set("lang", normalizedLang)
		
		// 设置响应头
		ctx.Header("Content-Language", normalizedLang)
		
		fmt.Printf("I18nMiddleware: 设置语言为 %s\n", normalizedLang)
		
		ctx.Next()
	}
}

// detectLanguage 检测客户端语言 对应Java后端的语言检测逻辑
func detectLanguage(ctx *gin.Context) string {
	// 1. 优先从URL参数获取 对应Java后端的LocaleChangeInterceptor
	if lang := ctx.Query(i18n.LanguageParamName); lang != "" {
		fmt.Printf("detectLanguage: 从URL参数获取语言: %s\n", lang)
		return lang
	}
	
	// 2. 从POST表单获取
	if lang := ctx.PostForm(i18n.LanguageParamName); lang != "" {
		fmt.Printf("detectLanguage: 从POST表单获取语言: %s\n", lang)
		return lang
	}
	
	// 3. 从Cookie获取
	if lang, err := ctx.Cookie(i18n.LanguageCookieName); err == nil && lang != "" {
		fmt.Printf("detectLanguage: 从Cookie获取语言: %s\n", lang)
		return lang
	}
	
	// 4. 从Accept-Language Header获取
	if lang := parseAcceptLanguage(ctx.GetHeader(i18n.LanguageHeaderName)); lang != "" {
		fmt.Printf("detectLanguage: 从Header获取语言: %s\n", lang)
		return lang
	}
	
	// 5. 返回默认语言
	fmt.Printf("detectLanguage: 使用默认语言: %s\n", i18n.DefaultLanguage)
	return i18n.DefaultLanguage
}

// parseAcceptLanguage 解析Accept-Language头
func parseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return ""
	}
	
	// 解析Accept-Language头，格式如：zh-CN,zh;q=0.9,en;q=0.8
	languages := strings.Split(acceptLang, ",")
	for _, lang := range languages {
		// 移除权重信息
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])
		if lang != "" {
			// 检查是否为支持的语言
			if normalizedLang := i18n.GetLanguageCode(lang); i18n.IsValidLanguage(normalizedLang) {
				return normalizedLang
			}
		}
	}
	
	return ""
}

// GetLanguageFromContext 从上下文获取语言
func GetLanguageFromContext(ctx *gin.Context) string {
	if lang, exists := ctx.Get("language"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return i18n.DefaultLanguage
}

// SetLanguageToContext 设置语言到上下文
func SetLanguageToContext(ctx *gin.Context, lang string) {
	normalizedLang, isValid := i18n.ValidateLanguage(lang)
	if !isValid {
		normalizedLang = i18n.DefaultLanguage
	}
	
	ctx.Set("language", normalizedLang)
	ctx.Set("lang", normalizedLang)
	ctx.Header("Content-Language", normalizedLang)
}

// MessageMiddleware 消息中间件，为响应添加国际化消息支持
func MessageMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 添加消息获取方法到上下文
		ctx.Set("getMessage", func(key string, args ...interface{}) string {
			lang := GetLanguageFromContext(ctx)
			return i18n.MessageWithLang(lang, key, args...)
		})
		
		ctx.Set("getErrorMessage", func(key string, args ...interface{}) string {
			lang := GetLanguageFromContext(ctx)
			return i18n.ErrorMessage(lang, key, args...)
		})
		
		ctx.Set("getSuccessMessage", func() string {
			lang := GetLanguageFromContext(ctx)
			return i18n.SuccessMessage(lang)
		})
		
		ctx.Set("getFailMessage", func() string {
			lang := GetLanguageFromContext(ctx)
			return i18n.FailMessage(lang)
		})
		
		ctx.Next()
	}
}

// GetMessage 从上下文获取国际化消息
func GetMessage(ctx *gin.Context, key string, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.MessageWithLang(lang, key, args...)
}

// GetErrorMessage 从上下文获取错误消息
func GetErrorMessage(ctx *gin.Context, key string, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.ErrorMessage(lang, key, args...)
}

// GetSuccessMessage 从上下文获取成功消息
func GetSuccessMessage(ctx *gin.Context) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.SuccessMessage(lang)
}

// GetFailMessage 从上下文获取失败消息
func GetFailMessage(ctx *gin.Context) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.FailMessage(lang)
}

// GetValidationMessage 从上下文获取验证消息
func GetValidationMessage(ctx *gin.Context, field, validationType string, params map[string]interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.ValidationMessage(lang, field, validationType, params)
}

// GetPermissionMessage 从上下文获取权限消息
func GetPermissionMessage(ctx *gin.Context, operation, resource string) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.PermissionMessage(lang, operation, resource)
}

// GetOperationMessage 从上下文获取操作消息
func GetOperationMessage(ctx *gin.Context, operation string, success bool) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.OperationMessage(lang, operation, success)
}

// GetUserMessage 从上下文获取用户相关消息
func GetUserMessage(ctx *gin.Context, messageType string, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.UserMessage(lang, messageType, args...)
}

// GetSystemMessage 从上下文获取系统消息
func GetSystemMessage(ctx *gin.Context, messageType string) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.SystemMessage(lang, messageType)
}

// GetUploadMessage 从上下文获取上传消息
func GetUploadMessage(ctx *gin.Context, messageType string, args ...interface{}) string {
	lang := GetLanguageFromContext(ctx)
	return i18n.UploadMessage(lang, messageType, args...)
}

// LanguageChangeHandler 语言切换处理器 对应Java后端的语言切换功能
func LanguageChangeHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lang := ctx.Query(i18n.LanguageParamName)
		if lang == "" {
			lang = ctx.PostForm(i18n.LanguageParamName)
		}
		
		if lang == "" {
			ctx.JSON(400, gin.H{
				"code":    400,
				"message": "语言参数不能为空",
			})
			return
		}
		
		// 验证语言
		normalizedLang, isValid := i18n.ValidateLanguage(lang)
		if !isValid {
			ctx.JSON(400, gin.H{
				"code":    400,
				"message": "不支持的语言: " + lang,
			})
			return
		}
		
		// 设置Cookie
		ctx.SetCookie(i18n.LanguageCookieName, normalizedLang, 86400*30, "/", "", false, true)
		
		// 设置到上下文
		SetLanguageToContext(ctx, normalizedLang)
		
		ctx.JSON(200, gin.H{
			"code":     200,
			"message":  "语言设置成功",
			"language": normalizedLang,
		})
	}
}

// GetLanguageInfo 获取语言信息处理器
func GetLanguageInfo() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		info := i18n.GetLanguageInfo()
		info["currentLanguage"] = GetLanguageFromContext(ctx)
		
		ctx.JSON(200, gin.H{
			"code":    200,
			"message": "获取语言信息成功",
			"data":    info,
		})
	}
}

// ReloadMessages 重新加载消息处理器
func ReloadMessages() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := i18n.ReloadMessages(); err != nil {
			ctx.JSON(500, gin.H{
				"code":    500,
				"message": "重新加载消息失败: " + err.Error(),
			})
			return
		}
		
		ctx.JSON(200, gin.H{
			"code":    200,
			"message": "重新加载消息成功",
		})
	}
}
