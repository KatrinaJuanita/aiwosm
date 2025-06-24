package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"wosm/internal/config"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
	systemService "wosm/internal/service/system"
	"wosm/internal/utils"
	"wosm/pkg/redis"

	"github.com/google/uuid"
	"github.com/mssola/useragent"
	redisv9 "github.com/redis/go-redis/v9"
)

// AuthService 认证服务 对应Java后端的SysLoginService
type AuthService struct {
	userDao         *dao.UserDao
	menuDao         *dao.MenuDao
	configService   *system.ConfigService
	passwordService *PasswordService // 密码验证服务 对应Java后端的SysPasswordService
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		userDao: dao.NewUserDao(),
		menuDao: dao.NewMenuDao(),
	}
}

// NewAuthServiceWithPassword 创建带密码验证的认证服务
func NewAuthServiceWithPassword(configService *system.ConfigService, redisClient *redisv9.Client, cfg *config.Config) *AuthService {
	return &AuthService{
		userDao:         dao.NewUserDao(),
		menuDao:         dao.NewMenuDao(),
		configService:   configService,
		passwordService: NewPasswordService(redisClient, cfg),
	}
}

// Login 用户登录 对应Java后端的SysLoginService.login
func (s *AuthService) Login(loginBody *model.LoginBody, userAgent, ipAddr string) (string, error) {
	fmt.Printf("开始登录验证: 用户名=%s\n", loginBody.Username)

	// 验证验证码 对应Java后端的validateCaptcha
	fmt.Printf("验证验证码: UUID=%s, Code=%s\n", loginBody.UUID, loginBody.Code)
	err := s.validateCaptcha(loginBody.Username, loginBody.Code, loginBody.UUID)
	if err != nil {
		fmt.Printf("验证码验证失败: %v\n", err)
		// 记录登录失败日志
		s.recordLoginLog(loginBody.Username, model.LoginStatusFail, err.Error(), ipAddr, userAgent)
		return "", err
	}
	fmt.Printf("验证码验证成功\n")

	// 登录前置校验 对应Java后端的loginPreCheck
	err = s.loginPreCheck(loginBody.Username, loginBody.Password, ipAddr)
	if err != nil {
		fmt.Printf("登录前置校验失败: %v\n", err)
		// 记录登录失败日志
		s.recordLoginLog(loginBody.Username, model.LoginStatusFail, err.Error(), ipAddr, userAgent)
		return "", err
	}
	fmt.Printf("登录前置校验成功\n")

	// 查询用户信息
	fmt.Printf("查询用户信息: %s\n", loginBody.Username)
	user, err := s.userDao.SelectUserByLoginName(loginBody.Username)
	if err != nil {
		fmt.Printf("查询用户失败: %v\n", err)
		return "", fmt.Errorf("查询用户失败: %v", err)
	}

	if user == nil {
		fmt.Printf("用户不存在: %s\n", loginBody.Username)
		// 记录登录失败日志
		s.recordLoginLog(loginBody.Username, model.LoginStatusFail, model.LoginMsgUserNotExists, ipAddr, userAgent)
		return "", errors.New("用户不存在/密码错误")
	}
	fmt.Printf("找到用户: ID=%d, 用户名=%s, 状态=%s, 删除标志=%s\n", user.UserID, user.UserName, user.Status, user.DelFlag)

	// 验证用户状态
	if user.DelFlag == "2" {
		fmt.Printf("用户已被删除: %s\n", loginBody.Username)
		// 记录登录失败日志
		s.recordLoginLog(loginBody.Username, model.LoginStatusFail, "对不起，您的账号已被删除", ipAddr, userAgent)
		return "", errors.New("对不起，您的账号已被删除")
	}

	if user.Status == "1" {
		fmt.Printf("用户已被停用: %s\n", loginBody.Username)
		// 记录登录失败日志
		s.recordLoginLog(loginBody.Username, model.LoginStatusFail, model.LoginMsgUserDisabled, ipAddr, userAgent)
		return "", errors.New("对不起，您的账号已停用")
	}

	// 验证密码 - 使用密码验证服务（对应Java后端的SysPasswordService.validate）
	fmt.Printf("验证密码: 输入密码=%s, 数据库密码长度=%d\n", loginBody.Password, len(user.Password))
	if s.passwordService != nil {
		// 使用密码验证服务，包含错误次数限制和用户锁定功能
		err = s.passwordService.Validate(user, loginBody.Password)
		if err != nil {
			fmt.Printf("密码验证失败: %v\n", err)
			// 记录登录失败日志
			s.recordLoginLog(loginBody.Username, model.LoginStatusFail, err.Error(), ipAddr, userAgent)
			return "", err
		}
	} else {
		// 降级到简单密码验证（兼容性）
		if !utils.MatchesPassword(loginBody.Password, user.Password) {
			fmt.Printf("密码验证失败\n")
			// 记录登录失败日志
			s.recordLoginLog(loginBody.Username, model.LoginStatusFail, model.LoginMsgPasswordError, ipAddr, userAgent)
			return "", errors.New("用户不存在/密码错误")
		}
	}
	fmt.Printf("密码验证成功\n")

	// 记录登录信息 对应Java后端的recordLoginInfo
	s.recordLoginInfo(user.UserID, ipAddr)

	// 记录登录成功日志
	s.recordLoginLog(loginBody.Username, model.LoginStatusSuccess, model.LoginMsgLoginSuccess, ipAddr, userAgent)

	// 生成UUID Token（对应Java后端的IdUtils.fastUUID()）
	token := s.generateUUIDToken()

	// 确保用户角色信息已加载（数据权限需要）
	if len(user.Roles) == 0 {
		fmt.Printf("Login: 用户角色信息为空，重新加载用户信息\n")
		fullUser, err := s.userDao.SelectUserById(user.UserID)
		if err == nil && fullUser != nil {
			user = fullUser
			fmt.Printf("Login: 重新加载用户角色成功, 角色数量=%d\n", len(user.Roles))
		}
	}

	// 创建登录用户信息
	loginUser := &model.LoginUser{
		UserID:        user.UserID,
		DeptID:        user.DeptID,
		Token:         token,
		LoginTime:     time.Now().UnixMilli(),                    // 使用毫秒时间戳，对应Java后端
		ExpireTime:    time.Now().Add(2 * time.Hour).UnixMilli(), // 使用毫秒时间戳
		IPAddr:        ipAddr,
		LoginLocation: s.getLoginLocation(ipAddr),
		Browser:       s.getBrowser(userAgent),
		OS:            s.getOS(userAgent),
		User:          user,
	}

	// 获取用户权限
	permissions, err := s.menuDao.SelectMenuPermsByUserId(user.UserID)
	if err != nil {
		return "", fmt.Errorf("获取用户权限失败: %v", err)
	}
	loginUser.Permissions = permissions

	// 清除该用户之前的所有会话（对应Java后端的单用户单会话机制）
	s.clearUserPreviousSessions(user.UserID)

	// 将登录用户信息存储到Redis
	err = s.storeLoginUser(token, loginUser)
	if err != nil {
		return "", fmt.Errorf("存储用户会话失败: %v", err)
	}

	// 生成JWT Token用于前端认证（包含UUID token）
	jwtToken, err := s.createJWTToken(token, user.UserName)
	if err != nil {
		return "", fmt.Errorf("生成JWT Token失败: %v", err)
	}

	return jwtToken, nil
}

// GetLoginUser 获取登录用户信息 对应Java后端的TokenService.getLoginUser
func (s *AuthService) GetLoginUser(token string) (*model.LoginUser, error) {
	// 直接使用UUID token从Redis获取用户信息（对应Java后端逻辑）
	key := fmt.Sprintf("login_tokens:%s", token)
	loginUser, err := s.getLoginUserFromRedis(key)
	if err != nil {
		return nil, fmt.Errorf("获取用户会话失败: %v", err)
	}

	if loginUser == nil {
		return nil, errors.New("用户会话已过期")
	}

	// 验证token是否匹配
	if loginUser.Token != token {
		return nil, errors.New("用户会话不匹配")
	}

	return loginUser, nil
}

// VerifyToken 验证令牌有效期，相差不足20分钟，自动刷新缓存 对应Java后端的verifyToken
func (s *AuthService) VerifyToken(loginUser *model.LoginUser) error {
	expireTime := loginUser.ExpireTime
	currentTime := time.Now().UnixMilli()

	// 如果相差不足20分钟，自动刷新缓存
	if expireTime-currentTime <= 20*60*1000 { // 20分钟的毫秒数
		fmt.Printf("VerifyToken: Token即将过期，自动刷新缓存\n")
		return s.RefreshToken(loginUser)
	}

	return nil
}

// RefreshToken 刷新令牌有效期 对应Java后端的refreshToken
func (s *AuthService) RefreshToken(loginUser *model.LoginUser) error {
	loginUser.LoginTime = time.Now().UnixMilli()
	loginUser.ExpireTime = loginUser.LoginTime + 2*60*60*1000 // 2小时的毫秒数

	// 根据uuid将loginUser缓存到Redis
	key := fmt.Sprintf("login_tokens:%s", loginUser.Token)
	data, err := json.Marshal(loginUser)
	if err != nil {
		return fmt.Errorf("序列化用户信息失败: %v", err)
	}

	// 设置过期时间为2小时
	err = redis.Set(key, string(data), 2*time.Hour)
	if err != nil {
		return fmt.Errorf("刷新Token缓存失败: %v", err)
	}

	fmt.Printf("RefreshToken: Token刷新成功, 新过期时间: %d\n", loginUser.ExpireTime)
	return nil
}

// validateCaptcha 校验验证码 对应Java后端的validateCaptcha
func (s *AuthService) validateCaptcha(username, code, uuid string) error {
	// 检查验证码是否启用 - 使用数据库配置
	configService := systemService.NewConfigService()
	captchaEnabled, err := configService.SelectCaptchaEnabled()
	if err != nil {
		fmt.Printf("获取验证码配置失败: %v，使用配置文件默认值\n", err)
		captchaEnabled = config.AppConfig.Captcha.Enabled
	}

	if !captchaEnabled {
		fmt.Printf("验证码已禁用，跳过验证\n")
		return nil
	}

	// 验证验证码是否为空
	if code == "" {
		return errors.New("验证码不能为空")
	}
	if uuid == "" {
		return errors.New("验证码已失效")
	}

	// 验证验证码
	if !utils.VerifyCaptcha(uuid, code) {
		return errors.New("验证码错误")
	}

	return nil
}

// loginPreCheck 登录前置校验 对应Java后端的loginPreCheck
func (s *AuthService) loginPreCheck(username, password, ipAddr string) error {
	// 用户名或密码为空 错误
	if username == "" || password == "" {
		return errors.New("用户名/密码必须填写")
	}

	// 密码如果不在指定范围内 错误 对应Java后端的UserConstants
	if len(password) < 5 || len(password) > 20 {
		return errors.New("用户密码不在指定范围")
	}

	// 用户名不在指定范围内 错误
	if len(username) < 2 || len(username) > 20 {
		return errors.New("用户名不在指定范围")
	}

	// 检查用户名是否包含非法字符 对应Java后端和前端的验证规则
	if !isValidUsername(username) {
		return errors.New("用户名包含非法字符")
	}

	// 密码错误次数校验 对应Java后端的checkPasswordErrorCount
	if s.passwordService != nil && s.passwordService.IsLocked(username) {
		lockTime := s.passwordService.GetLockTime()
		maxRetryCount := s.passwordService.GetMaxRetryCount()
		return fmt.Errorf("密码输入错误%d次，帐户锁定%d分钟", maxRetryCount, lockTime)
	}

	// IP黑名单校验 对应Java后端的checkBlackIPList
	err := s.checkBlackIPList(ipAddr)
	if err != nil {
		return err
	}

	return nil
}

// isValidUsername 检查用户名是否有效 对应前端的验证规则
func isValidUsername(username string) bool {
	// 根据前端验证规则：不能包含 < > " ' \ | 这些危险字符
	// 对应前端正则：/^[^<>"'|\\]+$/

	// 检查是否包含危险字符
	dangerousChars := []string{"<", ">", "\"", "'", "\\", "|"}
	for _, char := range dangerousChars {
		if strings.Contains(username, char) {
			return false
		}
	}

	// 检查是否只包含安全字符：字母、数字、下划线、中文、连字符
	// 这个规则比前端更严格，确保安全性
	matched, err := regexp.MatchString("^[a-zA-Z0-9_\u4e00-\u9fa5-]+$", username)
	if err != nil {
		return false
	}

	return matched
}

// checkBlackIPList 检查IP黑名单 对应Java后端的checkBlackIPList
func (s *AuthService) checkBlackIPList(ipAddr string) error {
	// 获取系统配置中的IP黑名单
	blackIPList, err := s.configService.SelectConfigByKey("sys.login.blackIPList")
	if err != nil {
		// 如果获取配置失败，记录日志但不阻止登录
		fmt.Printf("checkBlackIPList: 获取IP黑名单配置失败: %v\n", err)
		return nil
	}

	// 如果黑名单为空，直接返回
	if blackIPList == "" {
		return nil
	}

	// 检查当前IP是否在黑名单中
	if s.isMatchedIP(blackIPList, ipAddr) {
		fmt.Printf("checkBlackIPList: IP %s 在黑名单中\n", ipAddr)
		return errors.New("很抱歉，您的IP已被列入系统黑名单")
	}

	return nil
}

// isMatchedIP 检查IP是否匹配黑名单 对应Java后端的IpUtils.isMatchedIp
func (s *AuthService) isMatchedIP(blackIPList, ipAddr string) bool {
	if blackIPList == "" || ipAddr == "" {
		return false
	}

	// 分割IP列表（支持逗号、分号、换行符分割）
	ipList := strings.FieldsFunc(blackIPList, func(c rune) bool {
		return c == ',' || c == ';' || c == '\n' || c == '\r'
	})

	for _, ip := range ipList {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}

		// 精确匹配
		if ip == ipAddr {
			return true
		}

		// 支持通配符匹配（如 192.168.1.*）
		if strings.Contains(ip, "*") {
			pattern := strings.ReplaceAll(ip, "*", ".*")
			matched, err := regexp.MatchString("^"+pattern+"$", ipAddr)
			if err == nil && matched {
				return true
			}
		}

		// 支持CIDR格式（如 192.168.1.0/24）
		if strings.Contains(ip, "/") {
			// 这里可以实现CIDR匹配，暂时简化处理
			// 实际项目中可以使用net包的ParseCIDR和Contains方法
		}
	}

	return false
}

// GetUserInfo 获取用户信息 对应Java后端的getInfo
func (s *AuthService) GetUserInfo(loginUser *model.LoginUser) (*model.UserInfoResponse, error) {
	user := loginUser.User

	// 获取角色列表
	roles := make([]string, 0)
	if user.IsAdmin() {
		roles = append(roles, "admin")
	} else {
		for _, role := range user.Roles {
			roles = append(roles, role.RoleKey)
		}
	}

	return &model.UserInfoResponse{
		User:        user,
		Roles:       roles,
		Permissions: loginUser.Permissions,
	}, nil
}

// GetRouters 获取路由信息 对应Java后端的getRouters
func (s *AuthService) GetRouters(userId int64) ([]model.RouterVo, error) {
	// 获取用户菜单
	menus, err := s.menuDao.SelectMenusByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("获取用户菜单失败: %v", err)
	}

	// 映射数据库字段到前端字段
	for i := range menus {
		s.mapMenuFields(&menus[i])
	}

	// 构建菜单树
	routers := s.buildMenuTreeVo(menus, 0)

	// 调试：打印路由信息
	fmt.Printf("GetRouters: 构建的路由结构:\n")
	for i, router := range routers {
		fmt.Printf("  路由%d: Name=%s, Path=%s, Component=%s, Hidden=%t\n",
			i+1, router.Name, router.Path, router.Component, router.Hidden)
		if len(router.Children) > 0 {
			for j, child := range router.Children {
				fmt.Printf("    子路由%d: Name=%s, Path=%s, Component=%s\n",
					j+1, child.Name, child.Path, child.Component)
			}
		}
	}

	return routers, nil
}

// mapMenuFields 映射数据库字段到前端字段
func (s *AuthService) mapMenuFields(menu *model.SysMenu) {
	// 新数据库已经有正确的字段，无需复杂的映射逻辑
	fmt.Printf("  处理菜单: %s, Path='%s', Component='%s', 类型: %s\n",
		menu.MenuName, menu.Path, menu.Component, menu.MenuType)

	// 数据库字段已经正确，只需要设置一些默认值
	if menu.Query == "" {
		menu.Query = ""
	}

	// IsFrame和IsCache字段已经是int类型，无需转换

	fmt.Printf("    处理完成: Path='%s', Component='%s', IsFrame=%s, IsCache=%s\n",
		menu.Path, menu.Component, menu.IsFrame, menu.IsCache)
}

// Logout 用户登出 对应Java后端的logout
func (s *AuthService) Logout(token string) error {
	if token == "" {
		return nil
	}

	// 删除Redis中的用户会话
	key := fmt.Sprintf("login_tokens:%s", token)
	return redis.Del(key)
}

// storeLoginUser 存储登录用户信息到Redis
func (s *AuthService) storeLoginUser(token string, loginUser *model.LoginUser) error {
	key := fmt.Sprintf("login_tokens:%s", token)
	data, err := json.Marshal(loginUser)
	if err != nil {
		return err
	}

	// 设置过期时间为2小时
	return redis.Set(key, string(data), 2*time.Hour)
}

// getLoginUserFromRedis 从Redis获取登录用户信息
func (s *AuthService) getLoginUserFromRedis(key string) (*model.LoginUser, error) {
	data, err := redis.Get(key)
	if err != nil {
		return nil, err
	}

	if data == "" {
		return nil, nil
	}

	var loginUser model.LoginUser
	err = json.Unmarshal([]byte(data), &loginUser)
	if err != nil {
		return nil, err
	}

	return &loginUser, nil
}

// generateUUIDToken 生成UUID Token 对应Java后端的IdUtils.fastUUID()
func (s *AuthService) generateUUIDToken() string {
	return uuid.New().String()
}

// clearUserPreviousSessions 清除用户之前的所有会话 对应Java后端的单用户单会话机制
func (s *AuthService) clearUserPreviousSessions(userID int64) {
	fmt.Printf("clearUserPreviousSessions: 清除用户%d的之前会话\n", userID)

	// 获取所有登录token的key
	ctx := context.Background()
	keys, err := redis.GetRedis().Keys(ctx, "login_tokens:*").Result()
	if err != nil {
		fmt.Printf("clearUserPreviousSessions: 获取Redis keys失败: %v\n", err)
		return
	}

	// 遍历所有会话，删除同一用户的其他会话
	for _, key := range keys {
		loginUserData, err := redis.Get(key)
		if err != nil {
			continue
		}

		if loginUserData == "" {
			continue
		}

		var loginUser model.LoginUser
		if err := json.Unmarshal([]byte(loginUserData), &loginUser); err != nil {
			continue
		}

		// 如果是同一用户，删除之前的会话
		if loginUser.UserID == userID {
			redis.Del(key)
			fmt.Printf("clearUserPreviousSessions: 删除用户%d的会话: %s\n", userID, key)
		}
	}
}

// createJWTToken 创建JWT Token 对应Java后端的createToken方法
func (s *AuthService) createJWTToken(uuidToken, username string) (string, error) {
	// 简化实现：直接返回UUID token，因为前端主要使用这个
	// 这样与Java后端的行为更一致
	return uuidToken, nil
}

// getLoginLocation 获取登录地点
func (s *AuthService) getLoginLocation(ipAddr string) string {
	// 简单实现，实际项目中可以集成IP地址库
	if ipAddr == "127.0.0.1" || ipAddr == "::1" {
		return "内网IP"
	}
	return "未知"
}

// getBrowser 获取浏览器信息
func (s *AuthService) getBrowser(userAgent string) string {
	ua := useragent.New(userAgent)
	name, version := ua.Browser()
	if name == "" {
		return "Unknown"
	}
	return fmt.Sprintf("%s %s", name, version)
}

// getOS 获取操作系统信息
func (s *AuthService) getOS(userAgent string) string {
	ua := useragent.New(userAgent)
	os := ua.OS()
	if os == "" {
		return "Unknown"
	}
	return os
}

// buildMenuTree 构建菜单树 对应Java后端的菜单树构建逻辑
func (s *AuthService) buildMenuTree(menus []model.SysMenu, parentId int64) []model.RouterResponse {
	var routers []model.RouterResponse

	for _, menu := range menus {
		if menu.ParentID == parentId {
			router := s.buildRouter(menu)

			// 递归构建子菜单
			children := s.buildMenuTree(menus, menu.MenuID)
			if len(children) > 0 {
				router.Children = children
			}

			routers = append(routers, router)
		}
	}

	return routers
}

// buildMenuTreeVo 构建菜单树Vo 对应Java后端的菜单树构建逻辑
func (s *AuthService) buildMenuTreeVo(menus []model.SysMenu, parentId int64) []model.RouterVo {
	var routers []model.RouterVo

	for _, menu := range menus {
		if menu.ParentID == parentId {
			router := s.buildRouterVoComplete(menu)

			// 递归构建子菜单
			children := s.buildMenuTreeVo(menus, menu.MenuID)
			if len(children) > 0 {
				router.Children = children
			}

			routers = append(routers, router)
		}
	}

	return routers
}

// buildRouter 构建路由对象 对应Java后端的路由构建逻辑
func (s *AuthService) buildRouter(menu model.SysMenu) model.RouterResponse {
	router := model.RouterResponse{
		Name:      getRouteName(menu),
		Path:      getRouterPath(menu),
		Hidden:    menu.Visible == "1",
		Component: getComponent(menu),
		Query:     menu.Query,
		Meta: model.RouterMeta{
			Title:   menu.MenuName,
			Icon:    menu.Icon,
			NoCache: menu.IsCache == "1",
			Link:    "",
		},
	}

	// 处理外链
	if menu.IsFrame == "0" && utils.IsNotEmpty(menu.Path) {
		router.Meta.Link = menu.Path
	}

	return router
}

// buildRouterVo 构建路由对象Vo 对应Java后端的路由构建逻辑
func (s *AuthService) buildRouterVo(menu model.SysMenu) model.RouterVo {
	router := model.RouterVo{
		Name:      getRouteName(menu),
		Path:      getRouterPath(menu),
		Hidden:    menu.Visible == "1",
		Component: getComponent(menu),
		Query:     menu.Query,
		Meta: &model.MetaVo{
			Title:   menu.MenuName,
			Icon:    menu.Icon,
			NoCache: menu.IsCache == "1",
			Link:    "",
		},
	}

	// 处理外链 - 对应Java后端的外链处理逻辑
	if menu.IsFrame == "0" && utils.IsNotEmpty(menu.Path) {
		router.Meta.Link = menu.Path
	}

	return router
}

// buildRouterVoComplete 构建完整的路由对象Vo 包含所有业务逻辑
func (s *AuthService) buildRouterVoComplete(menu model.SysMenu) model.RouterVo {
	router := model.RouterVo{
		Name:      getRouteName(menu),
		Path:      getRouterPath(menu),
		Hidden:    menu.Visible == "1",
		Component: getComponent(menu),
		Query:     menu.Query,
		Meta: &model.MetaVo{
			Title:   menu.MenuName,
			Icon:    menu.Icon,
			NoCache: menu.IsCache == "1",
			Link:    "",
		},
	}

	// 处理外链 - 设置Meta.Link
	if menu.IsFrame == "0" && utils.IsNotEmpty(menu.Path) {
		router.Meta.Link = menu.Path
	}

	// 处理子菜单 - 对应Java后端的子菜单处理逻辑
	cMenus := menu.Children
	if len(cMenus) > 0 && menu.MenuType == "M" {
		router.AlwaysShow = true
		router.Redirect = "noRedirect"
		// 注意：这里不能递归调用，因为在buildMenuTreeVo中已经处理了递归
	} else if isMenuFrame(menu) {
		// 内链处理 - 对应Java后端的isMenuFrame处理逻辑
		// 当菜单为内部跳转时，需要设置Meta为null并创建子路由
		router.Meta = nil
		var childrenList []model.RouterVo
		children := model.RouterVo{
			Path:      menu.Path,
			Component: menu.Component,
			Name:      getRouteNameWithPath(menu.RouteName, menu.Path),
			Meta: &model.MetaVo{
				Title:   menu.MenuName,
				Icon:    menu.Icon,
				NoCache: menu.IsCache == "1",
				Link:    menu.Path,
			},
			Query: menu.Query,
		}
		childrenList = append(childrenList, children)
		router.Children = childrenList
	} else if menu.ParentID == 0 && isInnerLink(menu) {
		// 外链特殊处理 - 对应Java后端的外链处理逻辑
		// 当父节点为0且为内链时的特殊处理
		router.Meta = &model.MetaVo{
			Title: menu.MenuName,
			Icon:  menu.Icon,
		}
		router.Path = "/"
		var childrenList []model.RouterVo
		children := model.RouterVo{
			Path:      innerLinkReplaceEach(menu.Path),
			Component: "InnerLink",
			Name:      getRouteNameWithPath(menu.RouteName, innerLinkReplaceEach(menu.Path)),
			Meta: &model.MetaVo{
				Title: menu.MenuName,
				Icon:  menu.Icon,
				Link:  menu.Path,
			},
		}
		childrenList = append(childrenList, children)
		router.Children = childrenList
	}

	return router
}

// getRouteName 获取路由名称 对应Java后端的getRouteName
func getRouteName(menu model.SysMenu) string {
	// 优先使用路由名称
	if menu.RouteName != "" {
		return menu.RouteName
	}

	// 使用路径生成路由名称
	routerName := menu.Path
	if routerName != "" {
		// 简单的首字母大写处理
		routerName = strings.ToUpper(routerName[:1]) + routerName[1:]
	}

	// 如果是非外链并且是一级目录（类型为目录）
	if isMenuFrame(menu) {
		routerName = ""
	}
	return routerName
}

// getRouteNameWithPath 获取带路径的路由名称 对应Java后端的getRouteName重载方法
func getRouteNameWithPath(routeName, path string) string {
	if routeName != "" {
		return routeName
	}
	// 如果没有路由名称，使用路径生成
	if path != "" {
		// 简单处理：去掉特殊字符，首字母大写
		name := strings.ReplaceAll(path, "/", "")
		name = strings.ReplaceAll(name, "-", "")
		name = strings.ReplaceAll(name, "_", "")
		if len(name) > 0 {
			return strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return ""
}

// getRouterPath 获取路由地址
func getRouterPath(menu model.SysMenu) string {
	routerPath := menu.Path
	// 内链打开外网方式
	if menu.ParentID != 0 && isInnerLink(menu) {
		routerPath = innerLinkReplaceEach(routerPath)
	}
	// 非外链并且是一级目录（类型为目录）
	if menu.ParentID == 0 && menu.MenuType == "M" && menu.IsFrame == "1" {
		routerPath = "/" + menu.Path
	}
	// 非外链并且是一级目录（类型为菜单）
	if menu.ParentID == 0 && isMenuFrame(menu) {
		routerPath = "/"
	}
	return routerPath
}

// getComponent 获取组件信息 对应Java后端的getComponent
func getComponent(menu model.SysMenu) string {
	component := "Layout"

	// 如果有自定义组件且不是菜单框架
	if utils.IsNotEmpty(menu.Component) && !isMenuFrame(menu) {
		component = menu.Component
	} else if utils.IsEmpty(menu.Component) && menu.ParentID != 0 && isInnerLink(menu) {
		component = "InnerLink"
	} else if utils.IsEmpty(menu.Component) && isParentView(menu) {
		component = "ParentView"
	}

	return component
}

// isMenuFrame 是否为菜单内部跳转
func isMenuFrame(menu model.SysMenu) bool {
	return menu.ParentID == 0 && menu.MenuType == "C" && menu.IsFrame == "1"
}

// isInnerLink 是否为内链组件
func isInnerLink(menu model.SysMenu) bool {
	return menu.IsFrame == "1" && utils.IsNotEmpty(menu.Path) &&
		(strings.HasPrefix(menu.Path, "http://") || strings.HasPrefix(menu.Path, "https://"))
}

// isParentView 是否为parent_view组件
func isParentView(menu model.SysMenu) bool {
	return menu.ParentID != 0 && menu.MenuType == "M"
}

// innerLinkReplaceEach 内链域名特殊字符替换
func innerLinkReplaceEach(path string) string {
	path = strings.ReplaceAll(path, "http://", "")
	path = strings.ReplaceAll(path, "https://", "")
	path = strings.ReplaceAll(path, "www.", "")
	path = strings.ReplaceAll(path, ".", "/")
	path = strings.ReplaceAll(path, ":", "/")
	return path
}

// recordLoginInfo 记录登录信息 对应Java后端的recordLoginInfo
func (s *AuthService) recordLoginInfo(userId int64, ipAddr string) {
	fmt.Printf("AuthService.recordLoginInfo: 记录用户登录信息, UserID=%d, IP=%s\n", userId, ipAddr)

	// 异步更新用户登录信息，避免影响登录性能
	go func() {
		err := s.userDao.UpdateUserLoginInfo(userId, ipAddr)
		if err != nil {
			fmt.Printf("AuthService.recordLoginInfo: 更新用户登录信息失败: %v\n", err)
		} else {
			fmt.Printf("AuthService.recordLoginInfo: 更新用户登录信息成功\n")
		}
	}()
}

// recordLoginLog 记录登录日志
func (s *AuthService) recordLoginLog(userName, status, message, ipAddr, userAgent string) {
	// 创建登录日志服务实例
	loginLogService := systemService.NewLoginLogService()

	// 异步记录登录日志
	go func() {
		loginLogService.RecordLoginInfo(userName, status, message, ipAddr, userAgent)
	}()
}
