package main

import (
	"fmt"
	"log"
	"os"
	"wosm/internal/api/middleware"
	"wosm/internal/api/v1/auth"
	"wosm/internal/api/v1/common"
	"wosm/internal/api/v1/monitor"
	"wosm/internal/api/v1/system"
	"wosm/internal/api/v1/tool"
	"wosm/internal/config"
	"wosm/internal/constants"
	authService "wosm/internal/service/auth"
	systemService "wosm/internal/service/system"
	"wosm/pkg/database"
	"wosm/pkg/logger"
	"wosm/pkg/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WOSM Go Backend Main Entry Point
// 企业级Go后端主程序入口

func main() {
	// 1. 加载配置 - 支持命令行参数指定配置文件
	configFile := "configs/config.yaml" // 默认配置文件

	// 检查命令行参数
	if len(os.Args) > 1 {
		for i, arg := range os.Args {
			if arg == "--config" && i+1 < len(os.Args) {
				configFile = os.Args[i+1]
				break
			}
		}
	}

	// 检查环境变量
	if envConfig := os.Getenv("CONFIG_FILE"); envConfig != "" {
		configFile = envConfig
	}

	if err := config.LoadConfig(configFile); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Printf("配置加载成功: %s", configFile)

	// 2. 初始化日志系统
	if err := logger.InitLogger(); err != nil {
		log.Fatalf("日志系统初始化失败: %v", err)
	}
	logger.Info("日志系统初始化成功")
	defer logger.Close()

	// 3. 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	}
	logger.Info("数据库连接成功")
	defer database.Close()

	// 4. 初始化Redis连接
	if err := redis.InitRedis(); err != nil {
		logger.Fatal("Redis初始化失败", zap.Error(err))
	}
	logger.Info("Redis连接成功")
	defer redis.Close()

	// 5. 初始化验证器
	config.InitValidator()
	logger.Info("验证器初始化成功")

	// 6. 初始化路由
	router := setupRouter()

	// 7. 启动服务器
	port := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("WOSM Go Backend 启动成功，监听端口: %s", port)

	if err := router.Run(port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// setupRouter 设置路由 对应Java后端的路由配置
func setupRouter() *gin.Engine {
	// 设置Gin模式为调试模式以便看到请求日志
	gin.SetMode(gin.DebugMode)

	// 创建Gin引擎
	router := gin.New()

	// 设置信任的代理（解决安全警告）
	router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CorsMiddleware())
	router.Use(middleware.I18nMiddleware())
	router.Use(middleware.MessageMiddleware())
	router.Use(middleware.RepeatSubmitMiddleware()) // 防重复提交中间件 对应Java后端的RepeatSubmitInterceptor

	// 加载字典缓存
	dictTypeService := systemService.NewDictTypeService()
	dictTypeService.LoadingDictCache()

	// 初始化和加载参数配置缓存
	configService := systemService.NewConfigService()
	configService.InitDefaultConfigs()
	configService.LoadingConfigCache()

	// 初始化国际化系统
	i18nService := systemService.NewI18nService()
	i18nService.InitializeI18n()

	// 创建带密码验证的认证服务
	authServiceWithPassword := authService.NewAuthServiceWithPassword(configService, redis.GetRedis(), config.AppConfig)
	authController := auth.NewAuthControllerWithPassword(authServiceWithPassword)

	// 创建其他控制器
	registerController := auth.NewRegisterController()
	fileController := common.NewFileController()
	indexController := system.NewIndexController()
	userController := system.NewUserController()
	profileController := system.NewProfileController()
	roleController := system.NewRoleController()
	menuController := system.NewMenuController()
	deptController := system.NewDeptController()
	postController := system.NewPostController()
	dictTypeController := system.NewDictTypeController()
	dictDataController := system.NewDictDataController()
	configController := system.NewConfigController()
	i18nController := system.NewI18nController()
	noticeController := system.NewNoticeController()
	operLogController := monitor.NewOperLogController()
	loginLogController := monitor.NewLoginLogController()
	onlineController := monitor.NewOnlineController()
	serverController := monitor.NewServerController()
	cacheController := monitor.NewCacheController()
	druidController := monitor.NewDruidController()
	jobController := monitor.NewJobController()
	jobLogController := monitor.NewJobLogController() // 新增定时任务调度日志控制器
	genController := tool.NewGenController()
	swaggerController := tool.NewSwaggerController()
	testController := tool.NewTestController()

	// 公开路由（不需要认证）
	public := router.Group("/")
	{
		// 系统首页
		public.GET("/", indexController.Index)
		public.GET("/getSystemInfo", indexController.GetSystemInfo)

		public.GET("/captchaImage", authController.CaptchaImage)
		public.POST("/login", authController.Login)
		public.POST("/register", registerController.Register)

		// 国际化公开接口
		public.POST("/i18n/change", i18nController.ChangeLanguage)
		public.GET("/i18n/languages", i18nController.GetAvailableLanguages)
		public.GET("/i18n/info", i18nController.GetLanguageInfo)
	}

	// 兼容前端dev-api路径的公开路由
	devApiPublic := router.Group("/dev-api")
	{
		devApiPublic.GET("/captchaImage", authController.CaptchaImage)
		devApiPublic.POST("/login", authController.Login)
		devApiPublic.POST("/register", registerController.Register)

		// 国际化公开接口
		devApiPublic.POST("/i18n/change", i18nController.ChangeLanguage)
		devApiPublic.GET("/i18n/languages", i18nController.GetAvailableLanguages)
		devApiPublic.GET("/i18n/info", i18nController.GetLanguageInfo)
	}

	// 静态资源路由 对应Java后端的ResourcesConfig.addResourceHandlers
	// 映射 /profile/** 到文件系统路径
	router.Static(constants.RESOURCE_PREFIX, config.AppConfig.File.UploadPath)

	// Druid监控路由 对应Java后端的DruidConfig.statViewServlet
	druid := router.Group("/druid")
	{
		// Druid监控页面（公开访问）
		druid.GET("/", druidController.Index)
		druid.GET("/login.html", druidController.Login)
		druid.POST("/submitLogin", druidController.Auth)

		// Druid监控API（需要认证）
		druid.GET("/datasource.json", druidController.DataSource)
		druid.GET("/sql.json", druidController.SQL)
		druid.POST("/reset-all.json", druidController.ResetAll)
	}

	// Swagger文档路由 对应Java后端的SwaggerConfig
	swagger := router.Group("/swagger-ui")
	{
		// Swagger UI页面（公开访问）
		swagger.GET("/index.html", swaggerController.Index)
		swagger.GET("/api-docs", swaggerController.Spec)
	}

	// 通用文件管理路由（部分需要认证）
	common := router.Group("/common")
	{
		// 文件下载（公开访问）
		common.GET("/download", fileController.Download)
		common.GET("/download/resource", fileController.ResourceDownload)
	}

	// 需要认证的文件管理路由
	commonProtected := router.Group("/common")
	commonProtected.Use(middleware.AuthMiddleware())
	{
		// 文件上传
		commonProtected.POST("/upload", fileController.Upload)
		commonProtected.POST("/uploads", fileController.Uploads)
		// 文件信息
		commonProtected.GET("/fileInfo", fileController.GetFileInfo)
		commonProtected.GET("/listFiles", fileController.ListFiles)
	}

	// 需要认证的路由
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	protected.Use(middleware.OperationLogMiddleware()) // 添加操作日志中间件
	{
		// 认证相关
		protected.GET("/getInfo", authController.GetInfo)
		protected.GET("/getRouters", authController.GetRouters)
		protected.POST("/logout", authController.Logout)
	}

	// 兼容前端dev-api路径的需要认证的路由
	devApiProtected := router.Group("/dev-api")
	devApiProtected.Use(middleware.AuthMiddleware())
	devApiProtected.Use(middleware.OperationLogMiddleware()) // 添加操作日志中间件
	{
		// 认证相关
		devApiProtected.GET("/getInfo", authController.GetInfo)
		devApiProtected.GET("/getRouters", authController.GetRouters)
		devApiProtected.POST("/logout", authController.Logout)

		// 系统管理 - 用户管理
		systemUser := protected.Group("/system/user")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:list')")
			systemUser.GET("/list", middleware.WithPermissionAndDataScope("system:user:list", "d", "u", userController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:query')")
			systemUser.GET("/", middleware.WithPermission("system:user:query", userController.GetInfo))        // 新增用户时获取初始化数据
			systemUser.GET("/:userId", middleware.WithPermission("system:user:query", userController.GetInfo)) // 编辑用户时获取用户详情
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:add')")
			systemUser.POST("", middleware.WithPermission("system:user:add", userController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:edit')")
			systemUser.PUT("", middleware.WithPermission("system:user:edit", userController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:remove')")
			systemUser.DELETE("/:ids", middleware.WithPermission("system:user:remove", userController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:resetPwd')")
			systemUser.PUT("/resetPwd", middleware.WithPermission("system:user:resetPwd", userController.ResetPwd))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:edit')")
			systemUser.PUT("/changeStatus", middleware.WithPermission("system:user:edit", userController.ChangeStatus))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:query')")
			systemUser.GET("/authRole/:userId", middleware.WithPermission("system:user:query", userController.AuthRole))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:edit')")
			systemUser.PUT("/authRole", middleware.WithPermission("system:user:edit", userController.InsertAuthRole))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:list')")
			systemUser.GET("/deptTree", middleware.WithPermission("system:user:list", userController.DeptTree))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:export')")
			systemUser.POST("/export", middleware.WithPermissionAndDataScope("system:user:export", "d", "u", userController.Export))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:user:import')")
			systemUser.POST("/importData", middleware.WithPermission("system:user:import", userController.ImportData))
			systemUser.POST("/importTemplate", middleware.WithPermission("system:user:import", userController.ImportTemplate))

			// 个人中心 - 不需要特殊权限，只需要登录即可
			systemUser.GET("/profile", profileController.Profile)
			systemUser.PUT("/profile", profileController.UpdateProfile)
			systemUser.PUT("/profile/updatePwd", profileController.UpdatePwd)
			systemUser.POST("/profile/avatar", profileController.Avatar)
		}

		// 系统管理 - 角色管理
		systemRole := protected.Group("/system/role")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:list')")
			systemRole.GET("/list", middleware.WithPermission("system:role:list", roleController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:query')")
			systemRole.GET("/:roleId", middleware.WithPermission("system:role:query", roleController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:add')")
			systemRole.POST("", middleware.WithPermission("system:role:add", roleController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:edit')")
			systemRole.PUT("", middleware.WithPermission("system:role:edit", roleController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:remove')")
			systemRole.DELETE("/:roleIds", middleware.WithPermission("system:role:remove", roleController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:edit')")
			systemRole.PUT("/changeStatus", middleware.WithPermission("system:role:edit", roleController.ChangeStatus))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:edit')")
			systemRole.PUT("/dataScope", middleware.WithPermission("system:role:edit", roleController.DataScope))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:query')")
			systemRole.GET("/deptTree/:roleId", middleware.WithPermission("system:role:query", roleController.DeptTree))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:query')")
			systemRole.GET("/optionselect", middleware.WithPermission("system:role:query", roleController.OptionSelect))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:list')")
			systemRole.GET("/authUser/allocatedList", middleware.WithPermissionAndDataScope("system:role:list", "d", "u", roleController.AllocatedList))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:list')")
			systemRole.GET("/authUser/unallocatedList", middleware.WithPermissionAndDataScope("system:role:list", "d", "u", roleController.UnallocatedList))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:edit')")
			systemRole.PUT("/authUser/cancel", middleware.WithPermission("system:role:edit", roleController.CancelAuthUser))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:edit')")
			systemRole.PUT("/authUser/cancelAll", middleware.WithPermission("system:role:edit", roleController.CancelAuthUserAll))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:edit')")
			systemRole.PUT("/authUser/selectAll", middleware.WithPermission("system:role:edit", roleController.SelectAuthUserAll))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:role:export')")
			systemRole.POST("/export", middleware.WithPermission("system:role:export", roleController.Export))
		}

		// 系统管理 - 菜单管理
		systemMenu := protected.Group("/system/menu")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:list')")
			systemMenu.GET("/list", middleware.WithPermission("system:menu:list", menuController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:query')")
			systemMenu.GET("/:menuId", middleware.WithPermission("system:menu:query", menuController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:add')")
			systemMenu.POST("", middleware.WithPermission("system:menu:add", menuController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:edit')")
			systemMenu.PUT("", middleware.WithPermission("system:menu:edit", menuController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:remove')")
			systemMenu.DELETE("/:menuId", middleware.WithPermission("system:menu:remove", menuController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:query')")
			systemMenu.GET("/treeselect", middleware.WithPermission("system:menu:query", menuController.TreeSelect))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:menu:query')")
			systemMenu.GET("/roleMenuTreeselect/:roleId", middleware.WithPermission("system:menu:query", menuController.RoleMenuTreeSelect))
		}

		// 系统管理 - 部门管理
		systemDept := protected.Group("/system/dept")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:list')")
			systemDept.GET("/list", middleware.WithPermission("system:dept:list", deptController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:list')")
			systemDept.GET("/list/exclude/:deptId", middleware.WithPermission("system:dept:list", deptController.ExcludeChild))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:query')")
			systemDept.GET("/:deptId", middleware.WithPermission("system:dept:query", deptController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:add')")
			systemDept.POST("", middleware.WithPermission("system:dept:add", deptController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:edit')")
			systemDept.PUT("", middleware.WithPermission("system:dept:edit", deptController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:remove')")
			systemDept.DELETE("/:deptId", middleware.WithPermission("system:dept:remove", deptController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:query')")
			systemDept.GET("/treeselect", middleware.WithPermission("system:dept:query", deptController.TreeSelect))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:query')")
			systemDept.GET("/roleDeptTreeselect/:roleId", middleware.WithPermission("system:dept:query", deptController.RoleDeptTreeSelect))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dept:export')")
			systemDept.POST("/export", middleware.WithPermission("system:dept:export", deptController.Export))
		}

		// 系统管理 - 岗位管理
		systemPost := protected.Group("/system/post")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:list')")
			systemPost.GET("/list", middleware.WithPermission("system:post:list", postController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:query')")
			systemPost.GET("/:postId", middleware.WithPermission("system:post:query", postController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:query')")
			systemPost.GET("/optionselect", middleware.WithPermission("system:post:query", postController.OptionSelect))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:add')")
			systemPost.POST("", middleware.WithPermission("system:post:add", postController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:edit')")
			systemPost.PUT("", middleware.WithPermission("system:post:edit", postController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:remove')")
			systemPost.DELETE("/:postIds", middleware.WithPermission("system:post:remove", postController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:post:export')")
			systemPost.POST("/export", middleware.WithPermission("system:post:export", postController.Export))
		}

		// 系统管理 - 字典类型管理
		systemDictType := protected.Group("/system/dict/type")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:list')")
			systemDictType.GET("/list", middleware.WithPermission("system:dict:list", dictTypeController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:query')")
			systemDictType.GET("/:dictId", middleware.WithPermission("system:dict:query", dictTypeController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:add')")
			systemDictType.POST("", middleware.WithPermission("system:dict:add", dictTypeController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:edit')")
			systemDictType.PUT("", middleware.WithPermission("system:dict:edit", dictTypeController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:remove')")
			systemDictType.DELETE("/:dictIds", middleware.WithPermission("system:dict:remove", dictTypeController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:remove')")
			systemDictType.DELETE("/refreshCache", middleware.WithPermission("system:dict:remove", dictTypeController.RefreshCache))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:query')")
			systemDictType.GET("/optionselect", middleware.WithPermission("system:dict:query", dictTypeController.OptionSelect))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:export')")
			systemDictType.POST("/export", middleware.WithPermission("system:dict:export", dictTypeController.Export))
		}

		// 系统管理 - 字典数据管理
		systemDictData := protected.Group("/system/dict/data")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:list')")
			systemDictData.GET("/list", middleware.WithPermission("system:dict:list", dictDataController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:query')")
			systemDictData.GET("/:dictCode", middleware.WithPermission("system:dict:query", dictDataController.GetInfo))
			// 字典类型查询不需要权限验证，前端需要使用
			systemDictData.GET("/type/:dictType", dictDataController.DictType)
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:add')")
			systemDictData.POST("", middleware.WithPermission("system:dict:add", dictDataController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:edit')")
			systemDictData.PUT("", middleware.WithPermission("system:dict:edit", dictDataController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:remove')")
			systemDictData.DELETE("/:dictCodes", middleware.WithPermission("system:dict:remove", dictDataController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:dict:export')")
			systemDictData.POST("/export", middleware.WithPermission("system:dict:export", dictDataController.Export))
		}

		// 系统管理 - 参数配置管理
		systemConfig := protected.Group("/system/config")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:list')")
			systemConfig.GET("/list", middleware.WithPermission("system:config:list", configController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:query')")
			systemConfig.GET("/:configId", middleware.WithPermission("system:config:query", configController.GetInfo))
			// 配置键查询不需要权限验证，系统内部使用
			systemConfig.GET("/configKey/:configKey", configController.GetConfigKey)
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:add')")
			systemConfig.POST("", middleware.WithPermission("system:config:add", configController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:edit')")
			systemConfig.PUT("", middleware.WithPermission("system:config:edit", configController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:remove')")
			systemConfig.DELETE("/:configIds", middleware.WithPermission("system:config:remove", configController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:remove')")
			systemConfig.DELETE("/refreshCache", middleware.WithPermission("system:config:remove", configController.RefreshCache))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:config:export')")
			systemConfig.POST("/export", middleware.WithPermission("system:config:export", configController.Export))
		}

		// 系统管理 - 国际化管理
		systemI18n := protected.Group("/system/i18n")
		{
			// 国际化功能暂时不需要特殊权限，只需要登录即可
			systemI18n.GET("/info", i18nController.GetLanguageInfo)
			systemI18n.GET("/languages", i18nController.GetAvailableLanguages)
			systemI18n.POST("/change", i18nController.ChangeLanguage)
			systemI18n.GET("/message", i18nController.GetMessage)
			systemI18n.POST("/messages", i18nController.GetMessages)
			systemI18n.GET("/export", i18nController.ExportMessages)
			systemI18n.POST("/reload", i18nController.ReloadMessages)
			systemI18n.GET("/integrity", i18nController.CheckIntegrity)
			systemI18n.GET("/statistics", i18nController.GetStatistics)
			systemI18n.GET("/keys", i18nController.GetMessageKeys)
		}

		// 系统管理 - 通知公告管理
		systemNotice := protected.Group("/system/notice")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:notice:list')")
			systemNotice.GET("/list", middleware.WithPermission("system:notice:list", noticeController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:notice:query')")
			systemNotice.GET("/:noticeId", middleware.WithPermission("system:notice:query", noticeController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:notice:add')")
			systemNotice.POST("", middleware.WithPermission("system:notice:add", noticeController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:notice:edit')")
			systemNotice.PUT("", middleware.WithPermission("system:notice:edit", noticeController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:notice:remove')")
			systemNotice.DELETE("/:noticeIds", middleware.WithPermission("system:notice:remove", noticeController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('system:notice:export')")
			systemNotice.POST("/export", middleware.WithPermission("system:notice:export", noticeController.Export))
		}

		// 系统监控 - 操作日志管理
		monitorOperLog := protected.Group("/monitor/operlog")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:operlog:list')")
			monitorOperLog.GET("/list", middleware.WithPermission("monitor:operlog:list", operLogController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:operlog:query')")
			monitorOperLog.GET("/:operId", middleware.WithPermission("monitor:operlog:query", operLogController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:operlog:remove')")
			monitorOperLog.DELETE("/:operIds", middleware.WithPermission("monitor:operlog:remove", operLogController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:operlog:remove')")
			monitorOperLog.DELETE("/clean", middleware.WithPermission("monitor:operlog:remove", operLogController.Clean))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:operlog:export')")
			monitorOperLog.POST("/export", middleware.WithPermission("monitor:operlog:export", operLogController.Export))
		}

		// 系统监控 - 登录日志管理
		monitorLoginLog := protected.Group("/monitor/logininfor")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:logininfor:list')")
			monitorLoginLog.GET("/list", middleware.WithPermission("monitor:logininfor:list", loginLogController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:logininfor:remove')")
			monitorLoginLog.DELETE("/:infoIds", middleware.WithPermission("monitor:logininfor:remove", loginLogController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:logininfor:remove')")
			monitorLoginLog.DELETE("/clean", middleware.WithPermission("monitor:logininfor:remove", loginLogController.Clean))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:logininfor:export')")
			monitorLoginLog.POST("/export", middleware.WithPermission("monitor:logininfor:export", loginLogController.Export))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:logininfor:unlock')")
			monitorLoginLog.GET("/unlock/:userName", middleware.WithPermission("monitor:logininfor:unlock", loginLogController.Unlock))
		}

		// 系统监控 - 在线用户管理
		monitorOnline := protected.Group("/monitor/online")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:online:list')")
			monitorOnline.GET("/list", middleware.WithPermission("monitor:online:list", onlineController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:online:forceLogout')")
			monitorOnline.DELETE("/:tokenId", middleware.WithPermission("monitor:online:forceLogout", onlineController.ForceLogout))
			// Go后端扩展功能：导出在线用户数据 (Java后端没有此功能)
			monitorOnline.POST("/export", middleware.WithPermission("monitor:online:export", onlineController.Export))
		}

		// 系统监控 - 服务器监控
		monitorServer := protected.Group("/monitor/server")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:server:list')")
			monitorServer.GET("", middleware.WithPermission("monitor:server:list", serverController.GetInfo))
		}

		// 系统监控 - 缓存监控
		monitorCache := protected.Group("/monitor/cache")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:cache:list')")
			monitorCache.GET("", middleware.WithPermission("monitor:cache:list", cacheController.GetInfo))
			monitorCache.GET("/getNames", middleware.WithPermission("monitor:cache:list", cacheController.GetNames))
			monitorCache.GET("/getKeys/:cacheName", middleware.WithPermission("monitor:cache:list", cacheController.GetKeys))
			monitorCache.GET("/getValue/:cacheName/:cacheKey", middleware.WithPermission("monitor:cache:list", cacheController.GetValue))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:cache:clear')")
			monitorCache.DELETE("/clearCacheName/:cacheName", middleware.WithPermission("monitor:cache:clear", cacheController.ClearCacheName))
			monitorCache.DELETE("/clearCacheKey/:cacheKey", middleware.WithPermission("monitor:cache:clear", cacheController.ClearCacheKey))
			monitorCache.DELETE("/clearCacheAll", middleware.WithPermission("monitor:cache:clear", cacheController.ClearCacheAll))
		}

		// 系统监控 - 定时任务管理
		monitorJob := protected.Group("/monitor/job")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:list')")
			monitorJob.GET("/list", middleware.WithPermission("monitor:job:list", jobController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:query')")
			monitorJob.GET("/:jobId", middleware.WithPermission("monitor:job:query", jobController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:add')")
			monitorJob.POST("", middleware.WithPermission("monitor:job:add", jobController.Add))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:edit')")
			monitorJob.PUT("", middleware.WithPermission("monitor:job:edit", jobController.Edit))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:remove')")
			monitorJob.DELETE("/:jobIds", middleware.WithPermission("monitor:job:remove", jobController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:changeStatus')")
			monitorJob.PUT("/changeStatus", middleware.WithPermission("monitor:job:changeStatus", jobController.ChangeStatus))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:changeStatus')")
			monitorJob.PUT("/run", middleware.WithPermission("monitor:job:changeStatus", jobController.Run))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:export')")
			monitorJob.POST("/export", middleware.WithPermission("monitor:job:export", jobController.Export))
		}

		// 系统监控 - 定时任务调度日志管理
		monitorJobLog := protected.Group("/monitor/jobLog")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:list')")
			monitorJobLog.GET("/list", middleware.WithPermission("monitor:job:list", jobLogController.List))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:query')")
			monitorJobLog.GET("/:jobLogId", middleware.WithPermission("monitor:job:query", jobLogController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:remove')")
			monitorJobLog.DELETE("/:jobLogIds", middleware.WithPermission("monitor:job:remove", jobLogController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:remove')")
			monitorJobLog.DELETE("/clean", middleware.WithPermission("monitor:job:remove", jobLogController.Clean))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('monitor:job:export')")
			monitorJobLog.POST("/export", middleware.WithPermission("monitor:job:export", jobLogController.Export))
		}

		// 系统工具 - 代码生成
		toolGen := protected.Group("/tool/gen")
		{
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:list')")
			toolGen.GET("/list", middleware.WithPermission("tool:gen:list", genController.GenList))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:list')")
			toolGen.GET("/db/list", middleware.WithPermission("tool:gen:list", genController.DbList))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:query')")
			toolGen.GET("/:tableId", middleware.WithPermission("tool:gen:query", genController.GetInfo))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:edit')")
			toolGen.PUT("", middleware.WithPermission("tool:gen:edit", genController.EditSave))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:import')")
			toolGen.POST("/importTable", middleware.WithPermission("tool:gen:import", genController.ImportTable))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:edit')")
			toolGen.POST("/createTable", middleware.WithPermission("tool:gen:edit", genController.CreateTable))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:remove')")
			toolGen.DELETE("/:tableIds", middleware.WithPermission("tool:gen:remove", genController.Remove))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:preview')")
			toolGen.GET("/preview/:tableId", middleware.WithPermission("tool:gen:preview", genController.Preview))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:code')")
			toolGen.GET("/genCode/:tableName", middleware.WithPermission("tool:gen:code", genController.GenCode))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:edit')")
			toolGen.GET("/synchDb/:tableName", middleware.WithPermission("tool:gen:edit", genController.SynchDb))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:code')")
			toolGen.GET("/download/:tableName", middleware.WithPermission("tool:gen:code", genController.Download))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:code')")
			toolGen.GET("/batchGenCode", middleware.WithPermission("tool:gen:code", genController.BatchGenCode))
			// 对应Java后端 @PreAuthorize("@ss.hasPermi('tool:gen:query')")
			toolGen.GET("/column/:tableId", middleware.WithPermission("tool:gen:query", genController.ColumnList))
		}

		// 系统工具 - 测试接口
		toolTest := protected.Group("/test")
		{
			// 测试接口不需要特殊权限，只需要登录即可
			toolTest.GET("/info", testController.GetTestInfo)
			toolTest.POST("/reset", testController.ResetTestData)

			// 测试用户接口
			testUser := toolTest.Group("/user")
			{
				testUser.GET("/list", testController.UserList)
				testUser.GET("/:userId", testController.GetUser)
				testUser.POST("/save", testController.SaveUser)
				testUser.PUT("/update", testController.UpdateUser)
				testUser.DELETE("/:userId", testController.DeleteUser)
			}
		}
	}

	return router
}
