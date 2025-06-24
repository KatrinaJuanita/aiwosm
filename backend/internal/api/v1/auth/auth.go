package auth

import (
	"fmt"
	"strings"
	"wosm/internal/config"
	"wosm/internal/repository/model"
	"wosm/internal/service/auth"
	"wosm/internal/service/system"
	"wosm/internal/utils"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthController 认证控制器 对应Java后端的SysLoginController
type AuthController struct {
	authService *auth.AuthService
}

// NewAuthController 创建认证控制器
func NewAuthController() *AuthController {
	return &AuthController{
		authService: auth.NewAuthService(),
	}
}

// NewAuthControllerWithPassword 创建带密码验证的认证控制器
func NewAuthControllerWithPassword(authService *auth.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// CaptchaImage 获取验证码 对应Java后端的getCode
// @Summary 获取验证码
// @Description 生成验证码图片
// @Tags 认证接口
// @Accept json
// @Produce json
// @Success 200 {object} response.Result{data=model.CaptchaResponse}
// @Router /captchaImage [get]
func (c *AuthController) CaptchaImage(ctx *gin.Context) {
	// 检查验证码是否启用 对应Java后端的configService.selectCaptchaEnabled()
	configService := system.NewConfigService()
	captchaEnabled, err := configService.SelectCaptchaEnabled()
	if err != nil {
		fmt.Printf("获取验证码配置失败: %v，使用配置文件默认值\n", err)
		captchaEnabled = config.AppConfig.Captcha.Enabled
	}

	// 按照Java后端的格式返回，直接在根级别设置字段
	// Java: ajax.put("captchaEnabled", captchaEnabled); ajax.put("uuid", uuid); ajax.put("img", img);
	fields := map[string]interface{}{
		"captchaEnabled": captchaEnabled,
	}

	// 如果验证码未启用，直接返回
	if !captchaEnabled {
		fmt.Printf("验证码已禁用，跳过生成\n")
		response.SuccessWithFields(ctx, fields)
		return
	}

	// 生成验证码
	uuid, img, err := utils.GenerateCaptcha()
	if err != nil {
		fmt.Printf("验证码生成失败: %v\n", err)
		response.ErrorWithMessage(ctx, "验证码生成失败")
		return
	}

	fmt.Printf("验证码生成成功: UUID=%s, IMG长度=%d\n", uuid, len(img))

	// 添加验证码信息
	fields["uuid"] = uuid
	fields["img"] = img

	response.SuccessWithFields(ctx, fields)
}

// Login 用户登录 对应Java后端的login
// @Summary 用户登录
// @Description 用户登录验证
// @Tags 认证接口
// @Accept json
// @Produce json
// @Param loginBody body model.LoginBody true "登录信息"
// @Success 200 {object} response.Result{data=string}
// @Router /login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var loginBody model.LoginBody
	if err := ctx.ShouldBindJSON(&loginBody); err != nil {
		fmt.Printf("登录参数解析失败: %v\n", err)
		response.ErrorWithMessage(ctx, "参数错误")
		return
	}

	fmt.Printf("登录请求: 用户名=%s, 验证码=%s, UUID=%s\n", loginBody.Username, loginBody.Code, loginBody.UUID)

	// 获取客户端信息
	userAgent := ctx.GetHeader("User-Agent")
	ipAddr := c.getClientIP(ctx)

	// 执行登录
	token, err := c.authService.Login(&loginBody, userAgent, ipAddr)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	fmt.Printf("登录成功: 用户=%s, Token=%s\n", loginBody.Username, token[:20]+"...")

	// 返回token - 按照Java后端格式返回
	// Java: ajax.put(Constants.TOKEN, token); 其中 Constants.TOKEN = "token"
	fields := map[string]interface{}{
		"token": token,
	}
	fmt.Printf("登录响应: %+v\n", fields)
	response.SuccessWithFields(ctx, fields)
}

// GetInfo 获取用户信息 对应Java后端的getInfo
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 认证接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result{data=model.UserInfoResponse}
// @Router /getInfo [get]
func (c *AuthController) GetInfo(ctx *gin.Context) {
	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		fmt.Printf("GetInfo: 用户未登录\n")
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	user := loginUser.(*model.LoginUser)
	fmt.Printf("GetInfo: 获取用户信息, UserID=%d\n", user.UserID)

	// 获取用户详细信息
	userInfo, err := c.authService.GetUserInfo(user)
	if err != nil {
		fmt.Printf("GetInfo: 获取用户信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "获取用户信息失败")
		return
	}

	// 按照Java后端格式返回，直接在根级别设置字段
	// Java: ajax.put("user", user); ajax.put("roles", roles); ajax.put("permissions", permissions);
	fields := map[string]interface{}{
		"user":               userInfo.User,
		"roles":              userInfo.Roles,
		"permissions":        userInfo.Permissions,
		"isDefaultModifyPwd": false, // 暂时设为false
		"isPasswordExpired":  false, // 暂时设为false
	}

	fmt.Printf("GetInfo: 返回用户信息成功\n")
	response.SuccessWithFields(ctx, fields)
}

// GetRouters 获取路由信息 对应Java后端的getRouters
// @Summary 获取路由信息
// @Description 获取当前用户的菜单路由信息
// @Tags 认证接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result{data=[]model.RouterResponse}
// @Router /getRouters [get]
func (c *AuthController) GetRouters(ctx *gin.Context) {
	// 获取当前登录用户
	loginUser, exists := ctx.Get("loginUser")
	if !exists {
		fmt.Printf("GetRouters: 用户未登录\n")
		response.ErrorWithMessage(ctx, "用户未登录")
		return
	}

	user := loginUser.(*model.LoginUser)
	fmt.Printf("GetRouters: 获取路由信息, UserID=%d\n", user.UserID)

	// 获取用户路由信息
	routers, err := c.authService.GetRouters(user.UserID)
	if err != nil {
		fmt.Printf("GetRouters: 获取路由信息失败: %v\n", err)
		response.ErrorWithMessage(ctx, "获取路由信息失败")
		return
	}

	fmt.Printf("GetRouters: 返回路由信息成功, 路由数量=%d\n", len(routers))
	response.SuccessWithData(ctx, routers)
}

// Logout 用户登出 对应Java后端的logout
// @Summary 用户登出
// @Description 用户登出，清除会话信息
// @Tags 认证接口
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} response.Result
// @Router /logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// 获取token
	token := c.getTokenFromHeader(ctx)

	// 执行登出
	err := c.authService.Logout(token)
	if err != nil {
		response.ErrorWithMessage(ctx, "登出失败")
		return
	}

	response.SuccessWithMessage(ctx, "退出成功")
}

// getClientIP 获取客户端IP地址
func (c *AuthController) getClientIP(ctx *gin.Context) string {
	// 尝试从X-Forwarded-For头获取
	xForwardedFor := ctx.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP头获取
	xRealIP := ctx.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 使用RemoteAddr
	return ctx.ClientIP()
}

// getTokenFromHeader 从请求头获取Token
func (c *AuthController) getTokenFromHeader(ctx *gin.Context) string {
	token := ctx.GetHeader("Authorization")
	if token != "" && strings.HasPrefix(token, "Bearer ") {
		return strings.TrimPrefix(token, "Bearer ")
	}
	return token
}
