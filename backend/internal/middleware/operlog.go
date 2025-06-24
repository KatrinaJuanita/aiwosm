package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"

	"github.com/gin-gonic/gin"
)

// OperLogConfig 操作日志配置
type OperLogConfig struct {
	Title          string // 操作标题
	BusinessType   int    // 业务类型
	IsSaveReqData  bool   // 是否保存请求数据
	IsSaveRespData bool   // 是否保存响应数据
}

// 业务类型常量 对应Java后端的BusinessType枚举
const (
	BusinessTypeOther   = 0 // 其它
	BusinessTypeInsert  = 1 // 新增
	BusinessTypeUpdate  = 2 // 修改
	BusinessTypeDelete  = 3 // 删除
	BusinessTypeGrant   = 4 // 授权
	BusinessTypeExport  = 5 // 导出
	BusinessTypeImport  = 6 // 导入
	BusinessTypeForce   = 7 // 强退
	BusinessTypeGencode = 8 // 生成代码
	BusinessTypeClean   = 9 // 清空数据
)

// RecordOperLog 记录操作日志中间件 对应Java后端的@Log注解
func RecordOperLog(config OperLogConfig) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 读取请求体
		var requestBody []byte
		if config.IsSaveReqData && ctx.Request.Body != nil {
			requestBody, _ = io.ReadAll(ctx.Request.Body)
			// 重新设置请求体，以便后续处理
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器包装器
		responseWriter := &responseBodyWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
		}
		ctx.Writer = responseWriter

		// 继续处理请求
		ctx.Next()

		// 记录结束时间
		endTime := time.Now()
		costTime := endTime.Sub(startTime).Milliseconds()

		// 异步记录操作日志
		go func() {
			recordOperationLog(ctx, config, requestBody, responseWriter.body.Bytes(), costTime, startTime)
		}()
	}
}

// responseBodyWriter 响应体写入器包装器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// recordOperationLog 记录操作日志
func recordOperationLog(ctx *gin.Context, config OperLogConfig, requestBody, responseBody []byte, costTime int64, operTime time.Time) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("RecordOperLog: 记录操作日志异常: %v\n", r)
		}
	}()

	// 获取当前用户信息
	var operName string
	var deptName string
	userInterface, exists := ctx.Get("currentUser")
	if exists {
		if user, ok := userInterface.(*model.LoginUser); ok {
			operName = user.User.UserName
			if user.User.Dept != nil {
				deptName = user.User.Dept.DeptName
			}
		}
	}
	if operName == "" {
		operName = "匿名用户"
	}

	// 构建操作日志对象
	operLog := &model.SysOperLog{
		Title:         config.Title,
		BusinessType:  config.BusinessType,
		Method:        ctx.Request.Method,
		RequestMethod: ctx.Request.Method,
		OperatorType:  1, // 后台用户
		OperName:      operName,
		DeptName:      deptName,
		OperURL:       ctx.Request.URL.Path,
		OperIP:        getClientIP(ctx),
		OperLocation:  "", // TODO: 根据IP获取地理位置
		Status:        getOperStatus(ctx.Writer.Status()),
		CostTime:      costTime,
		OperTime:      &operTime,
	}

	// 设置请求参数
	if config.IsSaveReqData {
		operLog.OperParam = getOperParam(ctx, requestBody)
	}

	// 设置响应结果
	if config.IsSaveRespData {
		operLog.JSONResult = string(responseBody)
	}

	// 设置错误消息
	if ctx.Writer.Status() >= 400 {
		operLog.ErrorMsg = getErrorMessage(responseBody)
	}

	// 保存操作日志
	operLogService := system.NewOperLogService()
	if err := operLogService.InsertOperLog(operLog); err != nil {
		fmt.Printf("RecordOperLog: 保存操作日志失败: %v\n", err)
	} else {
		fmt.Printf("RecordOperLog: 操作日志记录成功, Title=%s, OperName=%s\n", config.Title, operName)
	}
}

// getClientIP 获取客户端IP地址
func getClientIP(ctx *gin.Context) string {
	// 尝试从X-Forwarded-For头获取
	if ip := ctx.GetHeader("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP头获取
	if ip := ctx.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// 从RemoteAddr获取
	return ctx.ClientIP()
}

// getOperStatus 获取操作状态
func getOperStatus(statusCode int) int {
	if statusCode >= 200 && statusCode < 300 {
		return 0 // 成功
	}
	return 1 // 失败
}

// getOperParam 获取操作参数
func getOperParam(ctx *gin.Context, requestBody []byte) string {
	params := make(map[string]any)

	// 添加查询参数
	for key, values := range ctx.Request.URL.Query() {
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			params[key] = values
		}
	}

	// 添加路径参数
	for _, param := range ctx.Params {
		params[param.Key] = param.Value
	}

	// 添加请求体参数（如果是JSON）
	if len(requestBody) > 0 && strings.Contains(ctx.GetHeader("Content-Type"), "application/json") {
		var bodyParams map[string]any
		if err := json.Unmarshal(requestBody, &bodyParams); err == nil {
			for key, value := range bodyParams {
				params[key] = value
			}
		}
	}

	// 转换为JSON字符串
	if len(params) > 0 {
		if paramBytes, err := json.Marshal(params); err == nil {
			return string(paramBytes)
		}
	}

	return ""
}

// getErrorMessage 获取错误消息
func getErrorMessage(responseBody []byte) string {
	if len(responseBody) == 0 {
		return ""
	}

	// 尝试解析响应体中的错误消息
	var response map[string]any
	if err := json.Unmarshal(responseBody, &response); err == nil {
		if msg, exists := response["msg"]; exists {
			if msgStr, ok := msg.(string); ok {
				return msgStr
			}
		}
		if message, exists := response["message"]; exists {
			if messageStr, ok := message.(string); ok {
				return messageStr
			}
		}
	}

	return string(responseBody)
}

// 代码生成模块操作日志配置
var (
	LogGenImport = OperLogConfig{
		Title:         "代码生成",
		BusinessType:  BusinessTypeImport,
		IsSaveReqData: true,
	}
	LogGenUpdate = OperLogConfig{
		Title:         "代码生成",
		BusinessType:  BusinessTypeUpdate,
		IsSaveReqData: true,
	}
	LogGenDelete = OperLogConfig{
		Title:         "代码生成",
		BusinessType:  BusinessTypeDelete,
		IsSaveReqData: true,
	}
	LogGenCode = OperLogConfig{
		Title:         "代码生成",
		BusinessType:  BusinessTypeGencode,
		IsSaveReqData: true,
	}
	LogCreateTable = OperLogConfig{
		Title:         "创建表",
		BusinessType:  BusinessTypeOther,
		IsSaveReqData: true,
	}
)
