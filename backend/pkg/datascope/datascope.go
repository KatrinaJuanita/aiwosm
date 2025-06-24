package datascope

import (
	"fmt"
	"strconv"
	"strings"
	"wosm/internal/repository/model"
)

// DataScopeProcessor 数据权限处理器 对应Java后端的DataScopeAspect
type DataScopeProcessor struct{}

// NewDataScopeProcessor 创建数据权限处理器实例
func NewDataScopeProcessor() *DataScopeProcessor {
	return &DataScopeProcessor{}
}

// ProcessDataScope 处理数据权限 对应Java后端的dataScopeFilter方法
func (p *DataScopeProcessor) ProcessDataScope(user *model.SysUser, config *DataScopeConfig, params map[string]interface{}) error {
	fmt.Printf("DataScopeProcessor.ProcessDataScope: 处理数据权限, UserID=%d, DeptAlias=%s, UserAlias=%s, CreatorAlias=%s\n",
		user.UserID, config.DeptAlias, config.UserAlias, config.CreatorAlias)

	// 清空之前的数据权限参数，防止注入
	p.clearDataScope(params)

	// 如果是超级管理员，不进行数据权限过滤
	if user.IsAdmin() {
		fmt.Printf("DataScopeProcessor.ProcessDataScope: 超级管理员，跳过数据权限过滤\n")
		return nil
	}

	// 生成数据权限SQL
	dataScopeSQL := p.generateDataScopeSQL(user, config)

	// 设置数据权限参数
	if dataScopeSQL != "" {
		params[DataScopeKey] = " AND (" + dataScopeSQL + ")"
		fmt.Printf("DataScopeProcessor.ProcessDataScope: 生成数据权限SQL: %s\n", params[DataScopeKey])
	}

	return nil
}

// generateDataScopeSQL 生成数据权限SQL 对应Java后端的dataScopeFilter核心逻辑
func (p *DataScopeProcessor) generateDataScopeSQL(user *model.SysUser, config *DataScopeConfig) string {
	var sqlParts []string
	var processedScopes []string
	var customRoleIds []string

	// 如果没有角色，返回限制性条件
	if len(user.Roles) == 0 {
		fmt.Printf("DataScopeProcessor.generateDataScopeSQL: 用户没有角色信息, UserID=%d\n", user.UserID)
		return fmt.Sprintf("%s.dept_id = 0", config.DeptAlias)
	}

	// 第一遍遍历：收集自定义数据权限的角色ID
	for _, role := range user.Roles {
		if role.DataScope == DataScopeCustom &&
			role.Status == RoleStatusNormal &&
			p.hasPermission(&role, config.Permission) {
			customRoleIds = append(customRoleIds, strconv.FormatInt(role.RoleID, 10))
		}
	}

	// 第二遍遍历：处理各种数据权限
	for _, role := range user.Roles {
		dataScope := role.DataScope

		// 跳过已处理的权限范围或停用的角色
		if p.contains(processedScopes, dataScope) || role.Status == RoleStatusDisable {
			continue
		}

		// 检查角色是否有对应的权限
		if !p.hasPermission(&role, config.Permission) {
			continue
		}

		// 根据数据权限范围生成对应的SQL条件
		switch dataScope {
		case DataScopeAll:
			// 全部数据权限，清空所有条件并跳出循环
			sqlParts = []string{}
			processedScopes = append(processedScopes, dataScope)
			goto end

		case DataScopeCustom:
			// 自定数据权限 - 完整实现
			if len(customRoleIds) > 0 {
				// 构建自定数据权限SQL
				sql := fmt.Sprintf("%s.dept_id IN (SELECT dept_id FROM sys_role_dept WHERE role_id IN (%s))",
					config.DeptAlias, strings.Join(customRoleIds, ","))
				sqlParts = append(sqlParts, sql)
			} else {
				// 如果没有自定权限，不查询任何数据
				sql := fmt.Sprintf("%s.dept_id = 0", config.DeptAlias)
				sqlParts = append(sqlParts, sql)
			}

		case DataScopeDept:
			// 部门数据权限
			if user.DeptID != nil {
				sql := fmt.Sprintf("%s.dept_id = %d", config.DeptAlias, *user.DeptID)
				sqlParts = append(sqlParts, sql)
			} else {
				// 没有部门ID时，不查询任何数据
				sql := fmt.Sprintf("%s.dept_id = 0", config.DeptAlias)
				sqlParts = append(sqlParts, sql)
			}

		case DataScopeDeptAndChild:
			// 部门及以下数据权限 - 修复SQL Server语法
			if user.DeptID != nil {
				// 使用SQL Server的CHARINDEX函数查找部门层级关系
				sql := fmt.Sprintf(`%s.dept_id IN (
					SELECT dept_id FROM sys_dept
					WHERE dept_id = %d
					OR CHARINDEX(',%d,', ','+ISNULL(ancestors,'')+',') > 0
				)`, config.DeptAlias, *user.DeptID, *user.DeptID)
				sqlParts = append(sqlParts, sql)
			} else {
				// 没有部门ID时，不查询任何数据
				sql := fmt.Sprintf("%s.dept_id = 0", config.DeptAlias)
				sqlParts = append(sqlParts, sql)
			}

		case DataScopeSelf:
			// 仅本人数据权限 - 支持user_id和create_by两种方式
			var selfConditions []string

			// 优先使用user_id字段（如果配置了UserAlias）
			if config.UserAlias != "" {
				userCondition := fmt.Sprintf("%s.user_id = %d", config.UserAlias, user.UserID)
				selfConditions = append(selfConditions, userCondition)
			}

			// 支持create_by字段（如果配置了CreatorAlias）
			if config.CreatorAlias != "" {
				creatorCondition := fmt.Sprintf("%s.create_by = '%s'", config.CreatorAlias, user.UserName)
				selfConditions = append(selfConditions, creatorCondition)
			}

			// 如果有任何条件，使用OR连接
			if len(selfConditions) > 0 {
				sql := strings.Join(selfConditions, " OR ")
				sqlParts = append(sqlParts, sql)
			} else {
				// 没有任何别名时，不查询任何数据
				sql := fmt.Sprintf("%s.dept_id = 0", config.DeptAlias)
				sqlParts = append(sqlParts, sql)
			}
		}

		processedScopes = append(processedScopes, dataScope)
	}

end:
	// 如果没有任何有效的权限条件，限制不查询任何数据
	if len(processedScopes) == 0 {
		return fmt.Sprintf("%s.dept_id = 0", config.DeptAlias)
	}

	// 如果有全部数据权限，返回空字符串（不添加任何限制）
	if p.contains(processedScopes, DataScopeAll) {
		return ""
	}

	// 组合所有SQL条件
	if len(sqlParts) > 0 {
		return strings.Join(sqlParts, " OR ")
	}

	return ""
}

// hasPermission 检查角色是否有指定权限 对应Java后端的权限检查逻辑
func (p *DataScopeProcessor) hasPermission(role *model.SysRole, permission string) bool {
	// 如果没有指定权限要求，默认通过
	if permission == "" {
		return true
	}

	// 检查角色状态
	if role.Status != RoleStatusNormal {
		return false
	}

	// 对应Java后端的权限检查逻辑：
	// if (!StringUtils.containsAny(role.getPermissions(), Convert.toStrArray(permission)))
	// 检查角色权限字符串是否包含指定权限
	permissions := strings.Split(permission, ",")
	rolePermissions := role.Permissions

	fmt.Printf("DataScopeProcessor.hasPermission: 检查角色权限, RoleID=%d, 需要权限=%s, 角色权限数量=%d\n",
		role.RoleID, permission, len(rolePermissions))

	for _, perm := range permissions {
		perm = strings.TrimSpace(perm)
		if perm == "" {
			continue
		}

		for _, rolePerm := range rolePermissions {
			rolePerm = strings.TrimSpace(rolePerm)
			if rolePerm == perm {
				fmt.Printf("DataScopeProcessor.hasPermission: 权限匹配成功, RoleID=%d, 权限=%s\n", role.RoleID, perm)
				return true
			}
		}
	}

	fmt.Printf("DataScopeProcessor.hasPermission: 权限匹配失败, RoleID=%d, 需要权限=%s\n", role.RoleID, permission)
	return false
}

// contains 检查字符串切片是否包含指定元素
func (p *DataScopeProcessor) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// clearDataScope 清空数据权限参数 对应Java后端的clearDataScope方法
func (p *DataScopeProcessor) clearDataScope(params map[string]interface{}) {
	if params != nil {
		params[DataScopeKey] = ""
	}
}

// ApplyDataScope 应用数据权限到查询参数 便捷方法
func ApplyDataScope(user *model.SysUser, deptAlias, userAlias, permission string, params map[string]interface{}) error {
	if user == nil || params == nil {
		return fmt.Errorf("用户信息或参数不能为空")
	}

	processor := NewDataScopeProcessor()
	config := &DataScopeConfig{
		DeptAlias:  deptAlias,
		UserAlias:  userAlias,
		Permission: permission,
	}

	return processor.ProcessDataScope(user, config, params)
}

// ApplyDataScopeWithCreator 应用数据权限到查询参数（支持创建者权限） 扩展方法
func ApplyDataScopeWithCreator(user *model.SysUser, deptAlias, userAlias, creatorAlias, permission string, params map[string]interface{}) error {
	if user == nil || params == nil {
		return fmt.Errorf("用户信息或参数不能为空")
	}

	processor := NewDataScopeProcessor()
	config := &DataScopeConfig{
		DeptAlias:    deptAlias,
		UserAlias:    userAlias,
		CreatorAlias: creatorAlias,
		Permission:   permission,
	}

	return processor.ProcessDataScope(user, config, params)
}

// GetDataScopeSQL 获取数据权限SQL（仅用于调试和测试）
func GetDataScopeSQL(user *model.SysUser, deptAlias, userAlias, permission string) string {
	if user == nil {
		return ""
	}

	processor := NewDataScopeProcessor()
	config := &DataScopeConfig{
		DeptAlias:  deptAlias,
		UserAlias:  userAlias,
		Permission: permission,
	}

	return processor.generateDataScopeSQL(user, config)
}

// GetDataScopeSQLWithCreator 获取数据权限SQL（支持创建者权限，仅用于调试和测试）
func GetDataScopeSQLWithCreator(user *model.SysUser, deptAlias, userAlias, creatorAlias, permission string) string {
	if user == nil {
		return ""
	}

	processor := NewDataScopeProcessor()
	config := &DataScopeConfig{
		DeptAlias:    deptAlias,
		UserAlias:    userAlias,
		CreatorAlias: creatorAlias,
		Permission:   permission,
	}

	return processor.generateDataScopeSQL(user, config)
}

// ValidateDataScopeConfig 验证数据权限配置
func ValidateDataScopeConfig(config *DataScopeConfig) error {
	if config == nil {
		return fmt.Errorf("数据权限配置不能为空")
	}

	if config.DeptAlias == "" {
		return fmt.Errorf("部门表别名不能为空")
	}

	// 用户表别名和创建者别名可以为空（某些场景下不需要用户权限过滤）
	// 但对于"仅本人数据"权限，至少需要配置其中一个

	return nil
}

// GetUserDataScopes 获取用户的所有数据权限范围
func GetUserDataScopes(user *model.SysUser) []string {
	if user == nil || len(user.Roles) == 0 {
		return []string{}
	}

	var scopes []string
	scopeMap := make(map[string]bool)

	for _, role := range user.Roles {
		if role.Status == RoleStatusNormal && !scopeMap[role.DataScope] {
			scopes = append(scopes, role.DataScope)
			scopeMap[role.DataScope] = true
		}
	}

	return scopes
}

// HasDataScope 检查用户是否有指定的数据权限范围
func HasDataScope(user *model.SysUser, dataScope string) bool {
	if user == nil || len(user.Roles) == 0 {
		return false
	}

	for _, role := range user.Roles {
		if role.Status == RoleStatusNormal && role.DataScope == dataScope {
			return true
		}
	}

	return false
}

// GetHighestDataScope 获取用户的最高数据权限范围
func GetHighestDataScope(user *model.SysUser) string {
	if user == nil || len(user.Roles) == 0 {
		return DataScopeSelf // 默认最低权限
	}

	highestScope := DataScopeSelf
	highestPriority := GetDataScopePriority(DataScopeSelf)

	for _, role := range user.Roles {
		if role.Status == RoleStatusNormal {
			priority := GetDataScopePriority(role.DataScope)
			if priority < highestPriority {
				highestScope = role.DataScope
				highestPriority = priority
			}
		}
	}

	return highestScope
}
