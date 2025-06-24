package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/model"
	systemService "wosm/internal/service/system"

	"github.com/gin-gonic/gin"
)

// OperationLogMiddleware 操作日志记录中间件 对应Java后端的LogAspect
func OperationLogMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 记录开始时间
		startTime := time.Now()

		// 创建响应写入器来捕获响应内容
		responseWriter := &responseBodyWriter{
			ResponseWriter: ctx.Writer,
			body:           &bytes.Buffer{},
		}
		ctx.Writer = responseWriter

		// 读取请求体
		var requestBody []byte
		if ctx.Request.Body != nil {
			requestBody, _ = io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 继续处理请求
		ctx.Next()

		// 计算耗时
		costTime := time.Since(startTime).Milliseconds()

		// 异步记录操作日志
		go func() {
			recordOperationLog(ctx, requestBody, responseWriter.body.Bytes(), costTime)
		}()
	}
}

// responseBodyWriter 响应体写入器
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// recordOperationLog 记录操作日志
func recordOperationLog(ctx *gin.Context, requestBody, responseBody []byte, costTime int64) {
	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		return // 未登录用户不记录日志
	}

	currentUser := loginUser.(*model.LoginUser)

	// 判断是否需要记录日志
	if !shouldRecordLog(ctx.Request.Method, ctx.Request.URL.Path) {
		return
	}

	// 获取部门名称（安全处理）
	deptName := ""
	if currentUser.User.Dept != nil {
		deptName = currentUser.User.Dept.DeptName
	}

	// 构建操作日志
	operLog := &model.SysOperLog{
		Title:         getOperationTitle(ctx.Request.URL.Path),
		BusinessType:  getBusinessType(ctx.Request.Method, ctx.Request.URL.Path),
		Method:        fmt.Sprintf("%s.%s", getControllerName(ctx.Request.URL.Path), getMethodName(ctx.Request.Method)),
		RequestMethod: ctx.Request.Method,
		OperatorType:  model.OperatorTypeManage, // 后台用户
		OperName:      currentUser.User.UserName,
		DeptName:      deptName,
		OperURL:       ctx.Request.URL.Path,
		OperIP:        getClientIP(ctx),
		OperLocation:  getIPLocation(getClientIP(ctx)),
		OperParam:     string(requestBody),
		JSONResult:    string(responseBody),
		Status:        getOperationStatus(ctx.Writer.Status()),
		ErrorMsg:      getErrorMessage(ctx.Writer.Status(), responseBody),
		CostTime:      costTime,
	}

	// 限制参数长度
	if len(operLog.OperParam) > 2000 {
		operLog.OperParam = operLog.OperParam[:2000] + "..."
	}
	if len(operLog.JSONResult) > 2000 {
		operLog.JSONResult = operLog.JSONResult[:2000] + "..."
	}

	// 保存操作日志
	operLogService := systemService.NewOperLogService()
	if err := operLogService.InsertOperLog(operLog); err != nil {
		fmt.Printf("记录操作日志失败: %v\n", err)
	}
}

// shouldRecordLog 判断是否需要记录日志
func shouldRecordLog(method, path string) bool {
	// 排除不需要记录的路径
	excludePaths := []string{
		"/login",
		"/logout",
		"/captchaImage",
		"/getInfo",
		"/getRouters",
		"/list", // GET请求的列表查询不记录
	}

	// GET请求的查询操作不记录日志
	if method == "GET" {
		for _, excludePath := range excludePaths {
			if strings.Contains(path, excludePath) {
				return false
			}
		}
		return false
	}

	// 只记录POST、PUT、DELETE操作
	return method == "POST" || method == "PUT" || method == "DELETE"
}

// getOperationTitle 获取操作标题
func getOperationTitle(path string) string {
	if strings.Contains(path, "/user") {
		return "用户管理"
	} else if strings.Contains(path, "/role") {
		return "角色管理"
	} else if strings.Contains(path, "/menu") {
		return "菜单管理"
	} else if strings.Contains(path, "/dept") {
		return "部门管理"
	} else if strings.Contains(path, "/dict") {
		return "字典管理"
	} else if strings.Contains(path, "/operlog") {
		return "操作日志"
	}
	return "系统操作"
}

// getBusinessType 获取业务类型
func getBusinessType(method, path string) int {
	switch method {
	case "POST":
		if strings.Contains(path, "/export") {
			return model.BusinessTypeExport
		}
		return model.BusinessTypeInsert
	case "PUT":
		return model.BusinessTypeUpdate
	case "DELETE":
		if strings.Contains(path, "/clean") {
			return model.BusinessTypeClean
		}
		return model.BusinessTypeDelete
	default:
		return model.BusinessTypeOther
	}
}

// getControllerName 获取控制器名称
func getControllerName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		return parts[len(parts)-2] + "Controller"
	}
	return "UnknownController"
}

// getMethodName 获取方法名称
func getMethodName(method string) string {
	switch method {
	case "POST":
		return "add"
	case "PUT":
		return "edit"
	case "DELETE":
		return "remove"
	case "GET":
		return "list"
	default:
		return "unknown"
	}
}

// getClientIP 获取客户端IP
func getClientIP(ctx *gin.Context) string {
	// 尝试从X-Forwarded-For头获取
	if ip := ctx.GetHeader("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 && ips[0] != "" {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP头获取
	if ip := ctx.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}

	// 从RemoteAddr获取
	ip, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err != nil {
		return ctx.Request.RemoteAddr
	}
	return ip
}

// getIPLocation 获取IP地址位置 对应Java后端的AddressUtils.getRealAddressByIP
func getIPLocation(ip string) string {
	// 内网IP直接返回 对应Java后端的IpUtils.internalIp判断
	if isInternalIP(ip) {
		return "内网IP"
	}

	// TODO: 可集成第三方IP地址库获取真实地理位置
	// Java后端使用 http://whois.pconline.com.cn/ipJson.jsp 接口
	// 这里暂时返回未知位置，避免外部依赖
	return "未知位置"
}

// isInternalIP 判断是否为内网IP 对应Java后端的IpUtils.internalIp
func isInternalIP(ip string) bool {
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	// 检查私有IP地址段
	// 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	if strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "192.168.") ||
		(strings.HasPrefix(ip, "172.") && isInRange172(ip)) {
		return true
	}

	return false
}

// isInRange172 检查是否在172.16.0.0/12范围内
func isInRange172(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	if parts[0] != "172" {
		return false
	}

	second, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	return second >= 16 && second <= 31
}

// getOperationStatus 获取操作状态
func getOperationStatus(statusCode int) int {
	if statusCode >= 200 && statusCode < 300 {
		return model.OperStatusSuccess
	}
	return model.OperStatusFail
}

// getErrorMessage 获取错误消息
func getErrorMessage(statusCode int, responseBody []byte) string {
	if statusCode >= 200 && statusCode < 300 {
		return ""
	}

	// 尝试解析响应体中的错误信息
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err == nil {
		if msg, ok := response["msg"].(string); ok {
			return msg
		}
		if message, ok := response["message"].(string); ok {
			return message
		}
	}

	return fmt.Sprintf("HTTP %d", statusCode)
}
