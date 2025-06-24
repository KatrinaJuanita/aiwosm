package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"wosm/internal/constants"
	"wosm/pkg/redis"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// RepeatSubmitConfig 防重复提交配置 对应Java后端的RepeatSubmit注解
type RepeatSubmitConfig struct {
	Interval int    `json:"interval"` // 间隔时间(ms)，小于此时间视为重复提交
	Message  string `json:"message"`  // 提示消息
}

// DefaultRepeatSubmitConfig 默认防重复提交配置
var DefaultRepeatSubmitConfig = RepeatSubmitConfig{
	Interval: 5000,            // 5秒
	Message:  "不允许重复提交，请稍候再试", // 对应Java后端的默认消息
}

// RepeatSubmitData 重复提交数据 对应Java后端的SameUrlDataInterceptor中的数据结构
type RepeatSubmitData struct {
	RepeatParams string `json:"repeatParams"` // 请求参数
	RepeatTime   int64  `json:"repeatTime"`   // 请求时间
}

// RepeatSubmitMiddleware 防重复提交中间件 对应Java后端的RepeatSubmitInterceptor
// 实现与Java后端SameUrlDataInterceptor相同的逻辑
func RepeatSubmitMiddleware() gin.HandlerFunc {
	return RepeatSubmitMiddlewareWithConfig(DefaultRepeatSubmitConfig)
}

// RepeatSubmitMiddlewareWithConfig 带配置的防重复提交中间件
func RepeatSubmitMiddlewareWithConfig(config RepeatSubmitConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 只对POST和PUT请求进行防重复提交检查 对应Java后端的逻辑
		if ctx.Request.Method != http.MethodPost && ctx.Request.Method != http.MethodPut {
			ctx.Next()
			return
		}

		// 检查是否需要跳过防重复提交检查
		if shouldSkipRepeatSubmitCheck(ctx) {
			ctx.Next()
			return
		}

		// 执行防重复提交检查 对应Java后端的isRepeatSubmit方法
		if isRepeatSubmit(ctx, config) {
			fmt.Printf("RepeatSubmitMiddleware: 检测到重复提交, URL=%s\n", ctx.Request.URL.Path)
			response.ErrorWithMessage(ctx, config.Message)
			ctx.Abort()
			return
		}

		fmt.Printf("RepeatSubmitMiddleware: 防重复提交检查通过, 继续执行, URL=%s\n", ctx.Request.URL.Path)
		ctx.Next()
		fmt.Printf("RepeatSubmitMiddleware: 请求处理完成, URL=%s\n", ctx.Request.URL.Path)
	}
}

// shouldSkipRepeatSubmitCheck 检查是否应该跳过防重复提交检查
func shouldSkipRepeatSubmitCheck(ctx *gin.Context) bool {
	// 检查请求头中是否设置了跳过标志 对应前端的repeatSubmit: false
	if ctx.GetHeader("repeatSubmit") == "false" {
		return true
	}

	// 对于某些特殊接口，可以跳过检查
	skipPaths := []string{
		"/login",                      // 登录接口
		"/logout",                     // 登出接口
		"/captchaImage",               // 验证码接口
		"/getInfo",                    // 获取用户信息
		"/getRouters",                 // 获取路由信息
		"/system/user/export",         // 用户数据导出
		"/system/user/importTemplate", // 用户导入模板
		"/system/role/export",         // 角色数据导出
		"/system/dept/export",         // 部门数据导出
		"/system/post/export",         // 岗位数据导出
		"/system/dict/data/export",    // 字典数据导出
		"/system/config/export",       // 参数配置导出
		"/system/notice/export",       // 通知公告导出
		"/monitor/operlog/export",     // 操作日志导出
		"/monitor/logininfor/export",  // 登录日志导出
	}

	for _, path := range skipPaths {
		if ctx.Request.URL.Path == path {
			return true
		}
	}

	return false
}

// isRepeatSubmit 检查是否为重复提交 对应Java后端的SameUrlDataInterceptor.isRepeatSubmit方法
func isRepeatSubmit(ctx *gin.Context, config RepeatSubmitConfig) bool {
	// 获取请求参数 对应Java后端的获取nowParams逻辑
	nowParams := getRequestParams(ctx)

	// 构建当前请求数据 对应Java后端的nowDataMap
	nowData := RepeatSubmitData{
		RepeatParams: nowParams,
		RepeatTime:   time.Now().UnixMilli(),
	}

	// 获取请求URL 对应Java后端的url
	url := ctx.Request.URL.Path

	// 获取唯一标识 对应Java后端的submitKey（使用Authorization头）
	submitKey := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if submitKey == "" {
		// 如果没有Authorization头，使用客户端IP作为标识
		submitKey = ctx.ClientIP()
	}

	// 构建缓存键 对应Java后端的cacheRepeatKey
	cacheKey := constants.REPEAT_SUBMIT_KEY + url + ":" + generateKeyHash(submitKey)

	fmt.Printf("RepeatSubmitMiddleware: 检查重复提交, CacheKey=%s\n", cacheKey)

	// 从Redis获取之前的请求数据 对应Java后端的redisCache.getCacheObject
	redisClient := redis.GetRedis()
	previousDataStr, err := redisClient.Get(ctx, cacheKey).Result()

	if err == nil && previousDataStr != "" {
		// 解析之前的请求数据
		var previousData RepeatSubmitData
		if json.Unmarshal([]byte(previousDataStr), &previousData) == nil {
			// 比较参数和时间 对应Java后端的compareParams和compareTime
			if compareParams(nowData, previousData) && compareTime(nowData, previousData, config.Interval) {
				return true // 检测到重复提交
			}
		}
	}

	// 将当前请求数据存入Redis 对应Java后端的redisCache.setCacheObject
	nowDataBytes, _ := json.Marshal(nowData)
	redisClient.Set(ctx, cacheKey, string(nowDataBytes), time.Duration(config.Interval)*time.Millisecond)

	return false // 不是重复提交
}

// getRequestParams 获取请求参数 对应Java后端的获取nowParams逻辑
func getRequestParams(ctx *gin.Context) string {
	var params string

	// 首先尝试获取Body参数 对应Java后端的HttpHelper.getBodyString
	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err == nil && len(bodyBytes) > 0 {
			params = string(bodyBytes)
			// 重新设置Body，以便后续处理可以继续读取
			ctx.Request.Body = io.NopCloser(strings.NewReader(params))
		}
	}

	// 如果Body为空，获取URL参数 对应Java后端的request.getParameterMap()
	if params == "" {
		if len(ctx.Request.URL.RawQuery) > 0 {
			params = ctx.Request.URL.RawQuery
		} else {
			// 获取Form参数
			ctx.Request.ParseForm()
			if len(ctx.Request.Form) > 0 {
				formData, _ := json.Marshal(ctx.Request.Form)
				params = string(formData)
			}
		}
	}

	return params
}

// compareParams 比较请求参数 对应Java后端的compareParams方法
func compareParams(nowData, previousData RepeatSubmitData) bool {
	return nowData.RepeatParams == previousData.RepeatParams
}

// compareTime 比较请求时间 对应Java后端的compareTime方法
func compareTime(nowData, previousData RepeatSubmitData, interval int) bool {
	timeDiff := nowData.RepeatTime - previousData.RepeatTime
	return timeDiff < int64(interval)
}

// generateKeyHash 生成键的哈希值，用于缩短Redis键长度
func generateKeyHash(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])[:8] // 取前8位
}

// WithRepeatSubmit 为特定路由设置防重复提交配置的辅助函数
func WithRepeatSubmit(interval int, message string) gin.HandlerFunc {
	config := RepeatSubmitConfig{
		Interval: interval,
		Message:  message,
	}
	return RepeatSubmitMiddlewareWithConfig(config)
}

// 预定义的防重复提交配置
var (
	// 用户操作相关的防重复提交配置
	UserRepeatSubmitConfig = RepeatSubmitConfig{
		Interval: 3000,
		Message:  "用户操作过于频繁，请稍候再试",
	}

	// 文件上传相关的防重复提交配置
	FileUploadRepeatSubmitConfig = RepeatSubmitConfig{
		Interval: 10000,
		Message:  "文件上传中，请勿重复提交",
	}

	// 数据导入相关的防重复提交配置
	DataImportRepeatSubmitConfig = RepeatSubmitConfig{
		Interval: 30000,
		Message:  "数据导入中，请勿重复提交",
	}
)
