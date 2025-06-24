package system

import (
	"fmt"
	"wosm/internal/repository/model"
)

// PermissionService 权限服务 对应Java后端的SysPermissionService
type PermissionService struct {
	roleService *RoleService
	menuService *MenuService
}

// NewPermissionService 创建权限服务实例
func NewPermissionService() *PermissionService {
	return &PermissionService{
		roleService: NewRoleService(),
		menuService: NewMenuService(),
	}
}

// GetRolePermission 获取角色数据权限 对应Java后端的getRolePermission
func (s *PermissionService) GetRolePermission(user *model.SysUser) ([]string, error) {
	fmt.Printf("PermissionService.GetRolePermission: 获取用户角色权限, UserID=%d\n", user.UserID)

	var roles []string

	// 管理员拥有所有权限 对应Java后端的user.isAdmin()
	if user.IsAdmin() {
		roles = append(roles, "admin")
		fmt.Printf("PermissionService.GetRolePermission: 超级管理员，返回admin权限\n")
		return roles, nil
	}

	// 查询用户角色权限 对应Java后端的roleService.selectRolePermissionByUserId(user.getUserId())
	rolePermissions, err := s.roleService.SelectRolePermissionByUserId(user.UserID)
	if err != nil {
		fmt.Printf("PermissionService.GetRolePermission: 查询角色权限失败: %v\n", err)
		return nil, err
	}

	roles = append(roles, rolePermissions...)
	fmt.Printf("PermissionService.GetRolePermission: 查询到角色权限, UserID=%d, Roles=%v\n", user.UserID, roles)
	return roles, nil
}

// GetMenuPermission 获取菜单权限 对应Java后端的getMenuPermission
func (s *PermissionService) GetMenuPermission(user *model.SysUser) ([]string, error) {
	fmt.Printf("PermissionService.GetMenuPermission: 获取用户菜单权限, UserID=%d\n", user.UserID)

	var permissions []string

	// 管理员拥有所有权限 对应Java后端的user.isAdmin()
	if user.IsAdmin() {
		permissions = append(permissions, "*:*:*")
		fmt.Printf("PermissionService.GetMenuPermission: 超级管理员，返回所有权限\n")
		return permissions, nil
	}

	// 查询用户菜单权限 对应Java后端的menuService.selectMenuPermsByUserId(user.getUserId())
	menuPermissions, err := s.menuService.SelectMenuPermsByUserId(user.UserID)
	if err != nil {
		fmt.Printf("PermissionService.GetMenuPermission: 查询菜单权限失败: %v\n", err)
		return nil, err
	}

	permissions = append(permissions, menuPermissions...)
	fmt.Printf("PermissionService.GetMenuPermission: 查询到菜单权限, UserID=%d, Permissions=%v\n", user.UserID, permissions)
	return permissions, nil
}

// RefreshUserPermissions 刷新用户权限缓存 对应Java后端的权限刷新逻辑
func (s *PermissionService) RefreshUserPermissions(user *model.SysUser) (*model.LoginUser, error) {
	fmt.Printf("PermissionService.RefreshUserPermissions: 刷新用户权限缓存, UserID=%d\n", user.UserID)

	// 获取菜单权限
	permissions, err := s.GetMenuPermission(user)
	if err != nil {
		return nil, fmt.Errorf("获取菜单权限失败: %v", err)
	}

	// 创建登录用户对象
	loginUser := &model.LoginUser{
		UserID:      user.UserID,
		DeptID:      user.DeptID,
		User:        user,
		Permissions: permissions,
	}

	fmt.Printf("PermissionService.RefreshUserPermissions: 权限刷新成功, UserID=%d, PermissionCount=%d\n", 
		user.UserID, len(permissions))
	return loginUser, nil
}

// HasPermission 检查用户是否拥有指定权限 对应Java后端的hasPermi
func (s *PermissionService) HasPermission(user *model.SysUser, permission string) (bool, error) {
	fmt.Printf("PermissionService.HasPermission: 检查用户权限, UserID=%d, Permission=%s\n", user.UserID, permission)

	if permission == "" {
		return false, nil
	}

	// 管理员拥有所有权限
	if user.IsAdmin() {
		fmt.Printf("PermissionService.HasPermission: 超级管理员，拥有所有权限\n")
		return true, nil
	}

	// 获取用户权限列表
	permissions, err := s.GetMenuPermission(user)
	if err != nil {
		return false, err
	}

	// 检查权限
	for _, perm := range permissions {
		if perm == permission || perm == "*:*:*" {
			fmt.Printf("PermissionService.HasPermission: 权限匹配成功\n")
			return true, nil
		}
	}

	fmt.Printf("PermissionService.HasPermission: 权限不足\n")
	return false, nil
}

// LacksPermission 检查用户是否不具备某权限 对应Java后端的lacksPermi
func (s *PermissionService) LacksPermission(user *model.SysUser, permission string) (bool, error) {
	hasPermission, err := s.HasPermission(user, permission)
	if err != nil {
		return false, err
	}
	return !hasPermission, nil
}

// HasAnyPermissions 检查用户是否拥有任意一个权限 对应Java后端的hasAnyPermi
func (s *PermissionService) HasAnyPermissions(user *model.SysUser, permissions []string) (bool, error) {
	fmt.Printf("PermissionService.HasAnyPermissions: 检查用户任意权限, UserID=%d, Permissions=%v\n", user.UserID, permissions)

	for _, permission := range permissions {
		hasPermission, err := s.HasPermission(user, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			fmt.Printf("PermissionService.HasAnyPermissions: 找到匹配权限: %s\n", permission)
			return true, nil
		}
	}

	fmt.Printf("PermissionService.HasAnyPermissions: 没有任何匹配权限\n")
	return false, nil
}

// HasRole 检查用户是否拥有指定角色 对应Java后端的hasRole
func (s *PermissionService) HasRole(user *model.SysUser, role string) (bool, error) {
	fmt.Printf("PermissionService.HasRole: 检查用户角色, UserID=%d, Role=%s\n", user.UserID, role)

	if role == "" {
		return false, nil
	}

	// 获取用户角色列表
	roles, err := s.GetRolePermission(user)
	if err != nil {
		return false, err
	}

	// 检查角色
	for _, userRole := range roles {
		if userRole == role {
			fmt.Printf("PermissionService.HasRole: 角色匹配成功\n")
			return true, nil
		}
	}

	fmt.Printf("PermissionService.HasRole: 角色不匹配\n")
	return false, nil
}

// HasAnyRoles 检查用户是否拥有任意一个角色 对应Java后端的hasAnyRoles
func (s *PermissionService) HasAnyRoles(user *model.SysUser, roles []string) (bool, error) {
	fmt.Printf("PermissionService.HasAnyRoles: 检查用户任意角色, UserID=%d, Roles=%v\n", user.UserID, roles)

	for _, role := range roles {
		hasRole, err := s.HasRole(user, role)
		if err != nil {
			return false, err
		}
		if hasRole {
			fmt.Printf("PermissionService.HasAnyRoles: 找到匹配角色: %s\n", role)
			return true, nil
		}
	}

	fmt.Printf("PermissionService.HasAnyRoles: 没有任何匹配角色\n")
	return false, nil
}
