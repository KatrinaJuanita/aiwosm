package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"wosm/internal/repository/model"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequirePermission 权限验证中间件 对应Java后端的@PreAuthorize注解
func RequirePermission(permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("RequirePermission: 检查权限, Permission=%s\n", permission)

		// 从上下文获取当前用户信息
		userInterface, exists := ctx.Get("currentUser")
		if !exists {
			fmt.Printf("RequirePermission: 用户未登录\n")
			response.ErrorWithCode(ctx, http.StatusUnauthorized, "用户未登录")
			ctx.Abort()
			return
		}

		user, ok := userInterface.(*model.LoginUser)
		if !ok {
			fmt.Printf("RequirePermission: 用户信息格式错误\n")
			response.ErrorWithCode(ctx, http.StatusUnauthorized, "用户信息格式错误")
			ctx.Abort()
			return
		}

		// 检查权限
		if !hasPermission(user, permission) {
			fmt.Printf("RequirePermission: 权限不足, UserID=%d, Permission=%s\n", user.User.UserID, permission)
			response.ErrorWithCode(ctx, http.StatusForbidden, "权限不足")
			ctx.Abort()
			return
		}

		fmt.Printf("RequirePermission: 权限验证通过, UserID=%d, Permission=%s\n", user.User.UserID, permission)
		ctx.Next()
	}
}

// RequireRole 角色验证中间件 对应Java后端的@PreAuthorize("@ss.hasRole('admin')")
func RequireRole(role string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("RequireRole: 检查角色, Role=%s\n", role)

		// 从上下文获取当前用户信息
		userInterface, exists := ctx.Get("currentUser")
		if !exists {
			fmt.Printf("RequireRole: 用户未登录\n")
			response.ErrorWithCode(ctx, http.StatusUnauthorized, "用户未登录")
			ctx.Abort()
			return
		}

		user, ok := userInterface.(*model.LoginUser)
		if !ok {
			fmt.Printf("RequireRole: 用户信息格式错误\n")
			response.ErrorWithCode(ctx, http.StatusUnauthorized, "用户信息格式错误")
			ctx.Abort()
			return
		}

		// 检查角色
		if !hasRole(user, role) {
			fmt.Printf("RequireRole: 角色不足, UserID=%d, Role=%s\n", user.User.UserID, role)
			response.ErrorWithCode(ctx, http.StatusForbidden, "角色不足")
			ctx.Abort()
			return
		}

		fmt.Printf("RequireRole: 角色验证通过, UserID=%d, Role=%s\n", user.User.UserID, role)
		ctx.Next()
	}
}

// hasPermission 检查用户是否具有指定权限 对应Java后端的SecurityUtils.hasPermi方法
func hasPermission(user *model.LoginUser, permission string) bool {
	if user == nil {
		return false
	}

	// 超级管理员拥有所有权限
	if user.User.IsAdmin() {
		return true
	}

	// 检查用户权限列表
	for _, perm := range user.Permissions {
		if perm == permission || perm == "*:*:*" {
			return true
		}

		// 支持通配符权限检查
		if matchWildcardPermission(perm, permission) {
			return true
		}
	}

	return false
}

// hasRole 检查用户是否具有指定角色 对应Java后端的SecurityUtils.hasRole方法
func hasRole(user *model.LoginUser, role string) bool {
	if user == nil {
		return false
	}

	// 超级管理员拥有所有角色
	if user.User.IsAdmin() {
		return true
	}

	// 检查用户角色列表
	for _, userRole := range user.User.Roles {
		if userRole.RoleKey == role {
			return true
		}
	}

	return false
}

// matchWildcardPermission 通配符权限匹配 对应Java后端的权限通配符逻辑
func matchWildcardPermission(userPerm, requiredPerm string) bool {
	// 分割权限字符串
	userParts := strings.Split(userPerm, ":")
	requiredParts := strings.Split(requiredPerm, ":")

	// 权限格式：模块:功能:操作，如 system:user:list
	if len(userParts) != 3 || len(requiredParts) != 3 {
		return false
	}

	// 逐级检查权限
	for i := 0; i < 3; i++ {
		if userParts[i] == "*" {
			continue // 通配符匹配
		}
		if userParts[i] != requiredParts[i] {
			return false
		}
	}

	return true
}

// 代码生成模块权限常量
const (
	PermGenList    = "tool:gen:list"    // 查询代码生成列表
	PermGenQuery   = "tool:gen:query"   // 查询代码生成详情
	PermGenImport  = "tool:gen:import"  // 导入表结构
	PermGenEdit    = "tool:gen:edit"    // 编辑代码生成
	PermGenRemove  = "tool:gen:remove"  // 删除代码生成
	PermGenPreview = "tool:gen:preview" // 预览代码
	PermGenCode    = "tool:gen:code"    // 生成代码
)

// 角色常量
const (
	RoleAdmin = "admin" // 管理员角色
)
