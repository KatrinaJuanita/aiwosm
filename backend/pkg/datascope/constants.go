package datascope

// 数据权限常量定义 对应Java后端的DataScopeAspect常量

const (
	// 数据权限范围类型 对应Java后端的DataScopeAspect常量
	DataScopeAll          = "1" // 全部数据权限
	DataScopeCustom       = "2" // 自定数据权限
	DataScopeDept         = "3" // 部门数据权限
	DataScopeDeptAndChild = "4" // 部门及以下数据权限
	DataScopeSelf         = "5" // 仅本人数据权限

	// 数据权限过滤关键字 对应Java后端的DATA_SCOPE
	DataScopeKey = "dataScope"

	// 权限上下文关键字
	PermissionContextKey = "permission"

	// 用户信息上下文关键字
	UserContextKey = "user"

	// 角色状态常量
	RoleStatusNormal  = "0" // 正常
	RoleStatusDisable = "1" // 停用
)

// DataScopeConfig 数据权限配置 对应Java后端的@DataScope注解
type DataScopeConfig struct {
	DeptAlias    string // 部门表别名
	UserAlias    string // 用户表别名
	CreatorAlias string // 创建者表别名（用于create_by字段）
	Permission   string // 权限字符
}

// GetDataScopeText 获取数据权限范围文本描述
func GetDataScopeText(dataScope string) string {
	switch dataScope {
	case DataScopeAll:
		return "全部数据权限"
	case DataScopeCustom:
		return "自定数据权限"
	case DataScopeDept:
		return "部门数据权限"
	case DataScopeDeptAndChild:
		return "部门及以下数据权限"
	case DataScopeSelf:
		return "仅本人数据权限"
	default:
		return "未知权限范围"
	}
}

// IsValidDataScope 验证数据权限范围是否有效
func IsValidDataScope(dataScope string) bool {
	validScopes := []string{
		DataScopeAll,
		DataScopeCustom,
		DataScopeDept,
		DataScopeDeptAndChild,
		DataScopeSelf,
	}

	for _, scope := range validScopes {
		if scope == dataScope {
			return true
		}
	}
	return false
}

// GetDefaultDataScope 获取默认数据权限范围
func GetDefaultDataScope() string {
	return DataScopeAll
}

// DataScopeOrder 数据权限优先级排序（数字越小优先级越高）
var DataScopeOrder = map[string]int{
	DataScopeAll:          1, // 全部数据权限优先级最高
	DataScopeCustom:       2, // 自定数据权限
	DataScopeDept:         3, // 部门数据权限
	DataScopeDeptAndChild: 4, // 部门及以下数据权限
	DataScopeSelf:         5, // 仅本人数据权限优先级最低
}

// GetDataScopePriority 获取数据权限优先级
func GetDataScopePriority(dataScope string) int {
	if priority, exists := DataScopeOrder[dataScope]; exists {
		return priority
	}
	return 999 // 未知权限范围优先级最低
}
