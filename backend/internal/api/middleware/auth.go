package middleware

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"wosm/internal/repository/model"
	"wosm/internal/service/auth"
	"wosm/pkg/datascope"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT认证中间件 对应Java后端的JwtAuthenticationTokenFilter
func AuthMiddleware() gin.HandlerFunc {
	authService := auth.NewAuthService()

	return func(ctx *gin.Context) {
		// 获取token
		token := getTokenFromRequest(ctx)
		if token == "" {
			response.ErrorWithDetailed(ctx, http.StatusUnauthorized, "未提供认证令牌")
			ctx.Abort()
			return
		}

		// 验证token并获取用户信息
		loginUser, err := authService.GetLoginUser(token)
		if err != nil {
			response.ErrorWithDetailed(ctx, http.StatusUnauthorized, "认证令牌无效")
			ctx.Abort()
			return
		}

		// 验证令牌有效期，相差不足20分钟，自动刷新缓存 对应Java后端的verifyToken
		err = authService.VerifyToken(loginUser)
		if err != nil {
			fmt.Printf("AuthMiddleware: Token验证失败: %v\n", err)
			// 这里不中断请求，因为刷新失败不应该影响当前请求
		}

		// 将用户信息存储到上下文中
		ctx.Set("loginUser", loginUser)
		ctx.Set("userId", loginUser.UserID)
		ctx.Set("username", loginUser.User.UserName)

		ctx.Next()
	}
}

// PermissionMiddleware 权限验证中间件 对应Java后端的权限验证
func PermissionMiddleware(permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		fmt.Printf("PermissionMiddleware: 检查权限 %s, URL: %s\n", permission, ctx.Request.URL.Path)

		// 获取当前登录用户
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			fmt.Printf("PermissionMiddleware: 用户未登录, URL: %s\n", ctx.Request.URL.Path)
			response.ErrorWithDetailed(ctx, http.StatusUnauthorized, "用户未登录")
			ctx.Abort()
			return
		}

		user := loginUser.(*model.LoginUser)
		fmt.Printf("PermissionMiddleware: 用户 %s 请求权限 %s\n", user.User.UserName, permission)

		// 管理员拥有所有权限
		if user.User.IsAdmin() {
			fmt.Printf("PermissionMiddleware: 超级管理员，允许访问\n")
			ctx.Next()
			return
		}

		// 如果用户没有任何权限，直接拒绝
		if len(user.Permissions) == 0 {
			fmt.Printf("PermissionMiddleware: 用户 %s 没有任何权限\n", user.User.UserName)
			response.ErrorWithDetailed(ctx, http.StatusForbidden, "权限不足")
			ctx.Abort()
			return
		}

		// 检查用户是否拥有指定权限
		hasPermission := false
		fmt.Printf("PermissionMiddleware: 用户权限列表: %v\n", user.Permissions)

		// 检查是否有超级权限 *:*:*
		if slices.Contains(user.Permissions, "*:*:*") {
			hasPermission = true
		}

		// 如果没有超级权限，检查具体权限
		if !hasPermission {
			for _, perm := range user.Permissions {
				if perm == permission {
					hasPermission = true
					break
				}
				// 支持通配符权限，例如 system:user:* 可以匹配 system:user:list
				if strings.HasSuffix(perm, ":*") {
					prefix := strings.TrimSuffix(perm, "*")
					if strings.HasPrefix(permission, prefix) {
						hasPermission = true
						break
					}
				}
			}
		}

		if !hasPermission {
			fmt.Printf("PermissionMiddleware: 权限验证失败，用户 %s 没有权限 %s，现有权限: %v\n",
				user.User.UserName, permission, user.Permissions)
			response.ErrorWithDetailed(ctx, http.StatusForbidden, "权限不足")
			ctx.Abort()
			return
		}

		fmt.Printf("PermissionMiddleware: 权限验证通过\n")
		ctx.Next()
	}
}

// getTokenFromRequest 从请求中获取Token
func getTokenFromRequest(ctx *gin.Context) string {
	// 从Authorization头获取
	token := ctx.GetHeader("Authorization")
	if token != "" && strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer ")
	}

	// 从查询参数获取
	token = ctx.Query("token")
	if token != "" {
		return token
	}

	return ""
}

// WithPermission 权限验证装饰器 对应Java后端的@PreAuthorize注解
func WithPermission(permission string, handler gin.HandlerFunc) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		// 先应用权限验证中间件
		middleware := PermissionMiddleware(permission)
		middleware(ctx)

		// 如果权限验证通过，继续执行处理函数
		if !ctx.IsAborted() {
			handler(ctx)
		}
	})
}

// WithDataScope 数据权限装饰器 对应Java后端的@DataScope注解
// 注意：这个函数需要在datascope_middleware.go中实现DataScopeMiddleware
func WithDataScope(deptAlias, userAlias, permission string, handler gin.HandlerFunc) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		// 应用数据权限处理逻辑
		// 获取当前用户信息
		loginUser, exists := ctx.Get("loginUser")
		if !exists {
			fmt.Printf("WithDataScope: 用户未登录, URL: %s\n", ctx.Request.URL.Path)
			response.ErrorWithDetailed(ctx, http.StatusUnauthorized, "用户未登录")
			ctx.Abort()
			return
		}

		user := loginUser.(*model.LoginUser)

		// 先进行权限验证
		if permission != "" {
			hasPermission := false

			// 管理员拥有所有权限
			if user.User.IsAdmin() {
				hasPermission = true
			} else {
				// 检查用户是否拥有指定权限
				for _, perm := range user.Permissions {
					if perm == permission {
						hasPermission = true
						break
					}
					// 支持通配符权限，例如 system:user:* 可以匹配 system:user:list
					if strings.HasSuffix(perm, ":*") {
						prefix := strings.TrimSuffix(perm, "*")
						if strings.HasPrefix(permission, prefix) {
							hasPermission = true
							break
						}
					}
				}
			}

			if !hasPermission {
				fmt.Printf("WithDataScope: 权限验证失败，用户 %s 没有权限 %s\n", user.User.UserName, permission)
				response.ErrorWithDetailed(ctx, http.StatusForbidden, "权限不足")
				ctx.Abort()
				return
			}
		}

		// 创建查询参数
		params := make(map[string]any)

		// 应用数据权限
		err := applyDataScope(user.User, deptAlias, userAlias, permission, params)
		if err != nil {
			response.ErrorWithDetailed(ctx, http.StatusForbidden, "数据权限校验失败")
			ctx.Abort()
			return
		}

		// 将数据权限参数设置到上下文中
		for key, value := range params {
			ctx.Set(key, value)
		}

		handler(ctx)
	})
}

// WithPermissionAndDataScope 权限和数据权限组合装饰器
func WithPermissionAndDataScope(permission, deptAlias, userAlias string, handler gin.HandlerFunc) gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		// 先应用权限验证中间件
		permissionMiddleware := PermissionMiddleware(permission)
		permissionMiddleware(ctx)

		if ctx.IsAborted() {
			return
		}

		// 再应用数据权限装饰器
		dataScopeHandler := WithDataScope(deptAlias, userAlias, permission, handler)
		dataScopeHandler(ctx)
	})
}

// applyDataScope 应用数据权限 对应Java后端的DataScopeAspect处理逻辑
func applyDataScope(user *model.SysUser, deptAlias, userAlias, permission string, params map[string]any) error {
	// 使用完整的数据权限处理器
	return datascope.ApplyDataScope(user, deptAlias, userAlias, permission, params)
}
