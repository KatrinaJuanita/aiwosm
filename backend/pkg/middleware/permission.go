package middleware

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// RequirePermission 权限验证中间件 对应Java后端的@PreAuthorize注解
func RequirePermission(permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("RequirePermission: 检查权限 %s\n", permission)

		// 获取登录用户信息
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			response.ErrorWithMessage(ctx, "用户未登录")
			ctx.Abort()
			return
		}

		currentUser := loginUser.(*model.LoginUser)

		// 检查权限
		if !hasPermission(currentUser, permission) {
			response.ErrorWithMessage(ctx, "没有权限执行此操作")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// hasPermission 检查用户是否具有指定权限 对应Java后端的权限检查逻辑
func hasPermission(user *model.LoginUser, permission string) bool {
	// 超级管理员拥有所有权限
	if user.User.UserID == 1 {
		return true
	}

	// 检查用户权限列表
	for _, userPermission := range user.Permissions {
		if userPermission == "*:*:*" || userPermission == permission {
			return true
		}
	}

	return false
}

// RequireAnyPermission 检查用户是否具有任意一个权限
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("RequireAnyPermission: 检查权限 %v\n", permissions)

		// 获取登录用户信息
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			response.ErrorWithMessage(ctx, "用户未登录")
			ctx.Abort()
			return
		}

		currentUser := loginUser.(*model.LoginUser)

		// 检查是否具有任意一个权限
		hasAnyPermission := false
		for _, permission := range permissions {
			if hasPermission(currentUser, permission) {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			response.ErrorWithMessage(ctx, "没有权限执行此操作")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// RequireAllPermissions 检查用户是否具有所有权限
func RequireAllPermissions(permissions ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("RequireAllPermissions: 检查权限 %v\n", permissions)

		// 获取登录用户信息
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			response.ErrorWithMessage(ctx, "用户未登录")
			ctx.Abort()
			return
		}

		currentUser := loginUser.(*model.LoginUser)

		// 检查是否具有所有权限
		for _, permission := range permissions {
			if !hasPermission(currentUser, permission) {
				response.ErrorWithMessage(ctx, "没有权限执行此操作")
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}

// RequireRole 检查用户是否具有指定角色
func RequireRole(role string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("RequireRole: 检查角色 %s\n", role)

		// 获取登录用户信息
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			response.ErrorWithMessage(ctx, "用户未登录")
			ctx.Abort()
			return
		}

		currentUser := loginUser.(*model.LoginUser)

		// 检查角色
		hasRole := false
		for _, userRole := range currentUser.User.Roles {
			if userRole.RoleKey == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			response.ErrorWithMessage(ctx, "没有权限执行此操作")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
