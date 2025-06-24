package system

import (
	"fmt"
	"wosm/internal/api/middleware"
	"wosm/internal/service/system"
	"wosm/pkg/i18n"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// I18nController 国际化控制器 对应Java后端的国际化相关控制器
type I18nController struct {
	i18nService *system.I18nService
}

// NewI18nController 创建国际化控制器实例
func NewI18nController() *I18nController {
	return &I18nController{
		i18nService: system.NewI18nService(),
	}
}

// GetLanguageInfo 获取语言信息 对应Java后端的语言信息接口
// @Summary 获取语言信息
// @Description 获取系统支持的语言信息和统计数据
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/i18n/info [get]
func (c *I18nController) GetLanguageInfo(ctx *gin.Context) {
	fmt.Printf("I18nController.GetLanguageInfo: 获取语言信息\n")

	// 获取当前语言
	currentLang := middleware.GetLanguageFromContext(ctx)

	// 获取语言信息
	info := c.i18nService.GetLanguageInfo()
	info["currentLanguage"] = currentLang

	// 获取统计信息
	stats := c.i18nService.GetLanguageStatistics()
	info["statistics"] = stats

	response.SuccessWithData(ctx, info)
}

// GetAvailableLanguages 获取可用语言列表
// @Summary 获取可用语言列表
// @Description 获取系统中可用的语言列表
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/i18n/languages [get]
func (c *I18nController) GetAvailableLanguages(ctx *gin.Context) {
	fmt.Printf("I18nController.GetAvailableLanguages: 获取可用语言列表\n")

	languages := c.i18nService.GetAvailableLanguages()
	response.SuccessWithData(ctx, languages)
}

// ChangeLanguage 切换语言 对应Java后端的语言切换接口
// @Summary 切换语言
// @Description 切换系统显示语言
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Param lang query string true "语言代码"
// @Success 200 {object} response.Response
// @Router /system/i18n/change [post]
func (c *I18nController) ChangeLanguage(ctx *gin.Context) {
	fmt.Printf("I18nController.ChangeLanguage: 切换语言\n")

	lang := ctx.Query("lang")
	if lang == "" {
		lang = ctx.PostForm("lang")
	}

	if lang == "" {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamMissing))
		return
	}

	// 验证语言
	normalizedLang, isValid := c.i18nService.ValidateLanguage(lang)
	if !isValid {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamInvalid))
		return
	}

	// 设置Cookie
	ctx.SetCookie(i18n.LanguageCookieName, normalizedLang, 86400*30, "/", "", false, true)

	// 设置到上下文
	middleware.SetLanguageToContext(ctx, normalizedLang)

	response.SuccessWithData(ctx, map[string]interface{}{
		"language":    normalizedLang,
		"displayName": i18n.GetLanguageDisplayName(normalizedLang),
		"message":     middleware.GetMessage(ctx, i18n.MsgOperationSuccess),
	})
}

// GetMessage 获取国际化消息
// @Summary 获取国际化消息
// @Description 根据消息键获取国际化消息
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Param key query string true "消息键"
// @Param args query string false "参数（JSON格式）"
// @Success 200 {object} response.Response
// @Router /system/i18n/message [get]
func (c *I18nController) GetMessage(ctx *gin.Context) {
	fmt.Printf("I18nController.GetMessage: 获取国际化消息\n")

	key := ctx.Query("key")
	if key == "" {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamMissing))
		return
	}

	// 获取当前语言
	lang := middleware.GetLanguageFromContext(ctx)

	// 获取消息
	message := c.i18nService.GetMessage(lang, key)

	response.SuccessWithData(ctx, map[string]interface{}{
		"key":      key,
		"message":  message,
		"language": lang,
	})
}

// GetMessages 批量获取国际化消息
// @Summary 批量获取国际化消息
// @Description 批量获取多个消息键的国际化消息
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Param keys body []string true "消息键列表"
// @Success 200 {object} response.Response
// @Router /system/i18n/messages [post]
func (c *I18nController) GetMessages(ctx *gin.Context) {
	fmt.Printf("I18nController.GetMessages: 批量获取国际化消息\n")

	var keys []string
	if err := ctx.ShouldBindJSON(&keys); err != nil {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamInvalid))
		return
	}

	if len(keys) == 0 {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamMissing))
		return
	}

	// 获取当前语言
	lang := middleware.GetLanguageFromContext(ctx)

	// 批量获取消息
	messages := make(map[string]string)
	for _, key := range keys {
		messages[key] = c.i18nService.GetMessage(lang, key)
	}

	response.SuccessWithData(ctx, map[string]interface{}{
		"messages": messages,
		"language": lang,
		"count":    len(messages),
	})
}

// ExportMessages 导出消息
// @Summary 导出消息
// @Description 导出指定语言的所有消息
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Param lang query string false "语言代码，默认为当前语言"
// @Success 200 {object} response.Response
// @Router /system/i18n/export [get]
func (c *I18nController) ExportMessages(ctx *gin.Context) {
	fmt.Printf("I18nController.ExportMessages: 导出消息\n")

	// 获取语言参数
	lang := ctx.Query("lang")
	if lang == "" {
		lang = middleware.GetLanguageFromContext(ctx)
	}

	// 验证语言
	normalizedLang, isValid := c.i18nService.ValidateLanguage(lang)
	if !isValid {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamInvalid))
		return
	}

	// 导出消息
	messages := c.i18nService.ExportMessages(normalizedLang)

	response.SuccessWithData(ctx, map[string]interface{}{
		"language":     normalizedLang,
		"displayName":  i18n.GetLanguageDisplayName(normalizedLang),
		"messages":     messages,
		"count":        len(messages),
		"exportTime":   fmt.Sprintf("%d", ctx.GetInt64("timestamp")),
	})
}

// ReloadMessages 重新加载消息
// @Summary 重新加载消息
// @Description 重新加载所有语言的消息文件
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/i18n/reload [post]
func (c *I18nController) ReloadMessages(ctx *gin.Context) {
	fmt.Printf("I18nController.ReloadMessages: 重新加载消息\n")

	if err := c.i18nService.ReloadMessages(); err != nil {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgSystemError)+": "+err.Error())
		return
	}

	// 获取重新加载后的统计信息
	stats := c.i18nService.GetLanguageStatistics()

	response.SuccessWithData(ctx, map[string]interface{}{
		"message":    middleware.GetMessage(ctx, i18n.MsgOperationSuccess),
		"statistics": stats,
	})
}

// CheckIntegrity 检查消息完整性
// @Summary 检查消息完整性
// @Description 检查各语言消息的完整性
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/i18n/integrity [get]
func (c *I18nController) CheckIntegrity(ctx *gin.Context) {
	fmt.Printf("I18nController.CheckIntegrity: 检查消息完整性\n")

	result := c.i18nService.CheckMessageIntegrity()
	response.SuccessWithData(ctx, result)
}

// GetStatistics 获取统计信息
// @Summary 获取统计信息
// @Description 获取国际化系统的统计信息
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /system/i18n/statistics [get]
func (c *I18nController) GetStatistics(ctx *gin.Context) {
	fmt.Printf("I18nController.GetStatistics: 获取统计信息\n")

	stats := c.i18nService.GetLanguageStatistics()
	response.SuccessWithData(ctx, stats)
}

// GetMessageKeys 获取消息键列表
// @Summary 获取消息键列表
// @Description 获取指定语言的所有消息键
// @Tags 国际化管理
// @Accept json
// @Produce json
// @Param lang query string false "语言代码，默认为当前语言"
// @Success 200 {object} response.Response
// @Router /system/i18n/keys [get]
func (c *I18nController) GetMessageKeys(ctx *gin.Context) {
	fmt.Printf("I18nController.GetMessageKeys: 获取消息键列表\n")

	// 获取语言参数
	lang := ctx.Query("lang")
	if lang == "" {
		lang = middleware.GetLanguageFromContext(ctx)
	}

	// 验证语言
	normalizedLang, isValid := c.i18nService.ValidateLanguage(lang)
	if !isValid {
		response.ErrorWithMessage(ctx, middleware.GetMessage(ctx, i18n.MsgParamInvalid))
		return
	}

	// 获取消息键
	keys := c.i18nService.GetMessageKeys(normalizedLang)

	response.SuccessWithData(ctx, map[string]interface{}{
		"language": normalizedLang,
		"keys":     keys,
		"count":    len(keys),
	})
}
