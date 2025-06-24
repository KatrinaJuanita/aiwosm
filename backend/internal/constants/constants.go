package constants

// 系统常量定义 对应Java后端的Constants类
const (
	// 通用状态
	SUCCESS = "0" // 成功
	FAIL    = "1" // 失败

	// 删除标志
	NOT_DELETED = "0" // 未删除
	DELETED     = "2" // 已删除

	// 用户状态
	USER_NORMAL  = "0" // 正常
	USER_DISABLE = "1" // 停用

	// 角色状态
	ROLE_NORMAL  = "0" // 正常
	ROLE_DISABLE = "1" // 停用

	// 菜单状态
	MENU_NORMAL  = "0" // 正常
	MENU_DISABLE = "1" // 停用

	// 菜单类型
	TYPE_DIR    = "M" // 目录
	TYPE_MENU   = "C" // 菜单
	TYPE_BUTTON = "F" // 按钮

	// 是否菜单外链
	YES_FRAME = "0" // 是
	NO_FRAME  = "1" // 否

	// 菜单显示状态
	SHOW = "0" // 显示
	HIDE = "1" // 隐藏

	// 数据范围
	DATA_SCOPE_ALL            = "1" // 全部数据权限
	DATA_SCOPE_CUSTOM         = "2" // 自定数据权限
	DATA_SCOPE_DEPT           = "3" // 部门数据权限
	DATA_SCOPE_DEPT_AND_CHILD = "4" // 部门及以下数据权限
	DATA_SCOPE_SELF           = "5" // 仅本人数据权限

	// 超级管理员ID
	SUPER_ADMIN_ID = 1

	// 登录相关常量 对应Java后端的Constants
	LOGIN_SUCCESS = "Success"  // 登录成功
	LOGIN_FAIL    = "Error"    // 登录失败
	LOGOUT        = "Logout"   // 注销
	REGISTER      = "Register" // 注册

	// Token相关常量
	TOKEN           = "token"          // 令牌
	TOKEN_PREFIX    = "Bearer "        // 令牌前缀
	LOGIN_USER_KEY  = "login_user_key" // 登录用户key
	JWT_USERNAME    = "sub"            // JWT用户名
	JWT_USERID      = "userid"         // JWT用户ID
	JWT_AVATAR      = "avatar"         // JWT头像
	JWT_CREATED     = "created"        // JWT创建时间
	JWT_AUTHORITIES = "authorities"    // JWT权限

	// 权限相关常量
	ALL_PERMISSION       = "*:*:*" // 所有权限标识
	SUPER_ADMIN          = "admin" // 管理员角色权限标识
	ROLE_DELIMETER       = ","     // 角色权限分隔符
	PERMISSION_DELIMETER = ","     // 权限标识分隔符

	// 验证码相关常量
	CAPTCHA_EXPIRATION = 2 // 验证码有效期（分钟）

	// 资源相关常量 对应Java后端的Constants.RESOURCE_PREFIX
	RESOURCE_PREFIX = "/profile" // 资源映射路径前缀

	// 定时任务安全相关常量 对应Java后端的Constants
	LOOKUP_RMI   = "rmi:"     // RMI 远程方法调用
	LOOKUP_LDAP  = "ldap:"    // LDAP 远程方法调用
	LOOKUP_LDAPS = "ldaps:"   // LDAPS 远程方法调用
	HTTP         = "http://"  // HTTP请求
	HTTPS        = "https://" // HTTPS请求

	// 用户名密码长度限制 对应Java后端的UserConstants
	USERNAME_MIN_LENGTH = 2  // 用户名最小长度
	USERNAME_MAX_LENGTH = 20 // 用户名最大长度
	PASSWORD_MIN_LENGTH = 5  // 密码最小长度
	PASSWORD_MAX_LENGTH = 20 // 密码最大长度

	// 唯一性校验返回标识 对应Java后端的UserConstants
	UNIQUE     = true  // 唯一
	NOT_UNIQUE = false // 不唯一

	// 是否为系统默认 对应Java后端的UserConstants
	YES = "Y" // 是
	NO  = "N" // 否
)

// 缓存键常量 对应Java后端的CacheConstants
const (
	LOGIN_TOKEN_KEY   = "login_tokens:"  // 登录用户 redis key
	CAPTCHA_CODE_KEY  = "captcha_codes:" // 验证码 redis key
	SYS_CONFIG_KEY    = "sys_config:"    // 参数管理 cache key
	SYS_DICT_KEY      = "sys_dict:"      // 字典管理 cache key
	REPEAT_SUBMIT_KEY = "repeat_submit:" // 防重提交 redis key
	RATE_LIMIT_KEY    = "rate_limit:"    // 限流 redis key
	PWD_ERR_CNT_KEY   = "pwd_err_cnt:"   // 登录账户密码错误次数 redis key
)

// 错误消息常量 对应Java后端的messages.properties
const (
	MSG_NOT_NULL                = "必须填写"
	MSG_USER_JCAPTCHA_ERROR     = "验证码错误"
	MSG_USER_JCAPTCHA_EXPIRE    = "验证码已失效"
	MSG_USER_NOT_EXISTS         = "用户不存在/密码错误"
	MSG_USER_PASSWORD_NOT_MATCH = "用户不存在/密码错误"
	MSG_USER_BLOCKED            = "用户已封禁，请联系管理员"
	MSG_ROLE_BLOCKED            = "角色已封禁，请联系管理员"
	MSG_LOGIN_BLOCKED           = "很遗憾，访问IP已被列入系统黑名单"
	MSG_USER_LOGOUT_SUCCESS     = "退出成功"
	MSG_USER_LOGIN_SUCCESS      = "登录成功"
	MSG_USER_REGISTER_SUCCESS   = "注册成功"
	MSG_USER_NOTFOUND           = "请重新登录"
	MSG_USER_FORCELOGOUT        = "管理员强制退出，请重新登录"
	MSG_USER_UNKNOWN_ERROR      = "未知错误，请重新登录"
)

// 定时任务白名单配置 对应Java后端的Constants.JOB_WHITELIST_STR
var JOB_WHITELIST_STR = []string{"com.ruoyi.quartz.task", "wosm.task"}

// 定时任务违规字符 对应Java后端的Constants.JOB_ERROR_STR
var JOB_ERROR_STR = []string{
	"java.net.URL",
	"javax.naming.InitialContext",
	"org.yaml.snakeyaml",
	"org.springframework",
	"org.apache",
	"com.ruoyi.common.utils.file",
	"com.ruoyi.common.config",
	"com.ruoyi.generator",
	"os.exec",
	"os/exec",
	"syscall",
	"unsafe",
}
