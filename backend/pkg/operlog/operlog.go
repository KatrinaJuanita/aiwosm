package operlog

import (
	"fmt"
	"time"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"

	"github.com/gin-gonic/gin"
)

// RecordOperLog 记录操作日志的通用方法
func RecordOperLog(ctx *gin.Context, title, businessType, content string, success bool) {
	// 获取用户信息
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		return
	}

	currentUser := loginUser.(*model.LoginUser)

	// 确定业务类型
	var businessTypeInt int
	switch businessType {
	case "新增":
		businessTypeInt = model.BusinessTypeInsert
	case "修改":
		businessTypeInt = model.BusinessTypeUpdate
	case "删除":
		businessTypeInt = model.BusinessTypeDelete
	case "导出":
		businessTypeInt = model.BusinessTypeExport
	case "导入":
		businessTypeInt = model.BusinessTypeImport
	default:
		businessTypeInt = model.BusinessTypeOther
	}

	// 构建操作日志
	now := time.Now()
	operLog := &model.SysOperLog{
		Title:         title,
		BusinessType:  businessTypeInt,
		Method:        fmt.Sprintf("%s.%s", getControllerName(ctx), getMethodNameFromBusinessType(businessType)),
		RequestMethod: ctx.Request.Method,
		OperatorType:  model.OperatorTypeManage, // 后台用户
		OperName:      currentUser.User.UserName,
		DeptName:      getDeptName(currentUser.User),
		OperURL:       ctx.Request.URL.Path,
		OperIP:        ctx.ClientIP(),
		OperLocation:  "", // 可以通过IP获取地理位置
		OperParam:     content,
		JSONResult:    "",
		Status:        model.OperStatusFail, // 默认失败
		ErrorMsg:      "",
		OperTime:      &now,
	}

	if success {
		operLog.Status = model.OperStatusSuccess // 成功
	}

	// 异步记录日志，不影响主流程
	go func() {
		operLogService := system.NewOperLogService()
		if err := operLogService.InsertOperLog(operLog); err != nil {
			fmt.Printf("记录操作日志失败: %v\n", err)
		}
	}()
}

// getControllerName 根据URL路径获取控制器名称
func getControllerName(ctx *gin.Context) string {
	path := ctx.Request.URL.Path
	if len(path) > 0 {
		switch {
		case contains(path, "/system/user"):
			return "UserController"
		case contains(path, "/system/role"):
			return "RoleController"
		case contains(path, "/system/post"):
			return "PostController"
		case contains(path, "/system/dict/type"):
			return "DictTypeController"
		case contains(path, "/system/dict/data"):
			return "DictDataController"
		case contains(path, "/system/config"):
			return "ConfigController"
		case contains(path, "/monitor/operlog"):
			return "OperLogController"
		case contains(path, "/monitor/logininfor"):
			return "LoginLogController"
		default:
			return "UnknownController"
		}
	}
	return "UnknownController"
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

// indexOf 查找子字符串在字符串中的位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// getMethodNameFromBusinessType 根据业务类型获取方法名
func getMethodNameFromBusinessType(businessType string) string {
	switch businessType {
	case "新增":
		return "Add"
	case "修改":
		return "Edit"
	case "删除":
		return "Remove"
	case "导出":
		return "Export"
	case "导入":
		return "Import"
	default:
		return "Unknown"
	}
}

// getDeptName 获取部门名称
func getDeptName(user *model.SysUser) string {
	if user.Dept != nil {
		return user.Dept.DeptName
	}
	return ""
}
