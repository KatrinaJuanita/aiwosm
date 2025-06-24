package system

import (
	"fmt"
	"strings"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/pkg/datascope"
)

// RoleService 角色服务 对应Java后端的ISysRoleService
type RoleService struct {
	roleDao     *dao.RoleDao
	roleMenuDao *dao.RoleMenuDao
	roleDeptDao *dao.RoleDeptDao
	userRoleDao *dao.UserRoleDao
}

// NewRoleService 创建角色服务实例
func NewRoleService() *RoleService {
	return &RoleService{
		roleDao:     dao.NewRoleDao(),
		roleMenuDao: dao.NewRoleMenuDao(),
		roleDeptDao: dao.NewRoleDeptDao(),
		userRoleDao: dao.NewUserRoleDao(),
	}
}

// SelectRoleList 查询角色列表 对应Java后端的selectRoleList
func (s *RoleService) SelectRoleList(role *model.SysRole, pageNum, pageSize int) ([]model.SysRole, int64, error) {
	fmt.Printf("RoleService.SelectRoleList: 查询角色列表, PageNum=%d, PageSize=%d\n", pageNum, pageSize)
	return s.roleDao.SelectRoleListWithPage(role, pageNum, pageSize)
}

// SelectRoleListAll 查询所有角色列表（不分页） 对应Java后端的selectRoleList
func (s *RoleService) SelectRoleListAll(role *model.SysRole) ([]model.SysRole, error) {
	fmt.Printf("RoleService.SelectRoleListAll: 查询所有角色列表\n")
	return s.roleDao.SelectRoleList(role)
}

// SelectRoleById 根据角色ID查询角色信息 对应Java后端的selectRoleById
func (s *RoleService) SelectRoleById(roleId int64) (*model.SysRole, error) {
	fmt.Printf("RoleService.SelectRoleById: 查询角色信息, RoleID=%d\n", roleId)
	return s.roleDao.SelectRoleById(roleId)
}

// SelectRoleAll 查询所有角色 对应Java后端的selectRoleAll
func (s *RoleService) SelectRoleAll() ([]model.SysRole, error) {
	fmt.Printf("RoleService.SelectRoleAll: 查询所有角色\n")
	return s.roleDao.SelectRoleAll()
}

// SelectMenuListByRoleId 根据角色ID查询菜单树信息 对应Java后端的selectMenuListByRoleId
func (s *RoleService) SelectMenuListByRoleId(roleId int64) ([]int64, error) {
	fmt.Printf("RoleService.SelectMenuListByRoleId: 根据角色ID查询菜单, RoleID=%d\n", roleId)

	// 查询角色信息，获取MenuCheckStrictly字段
	role, err := s.SelectRoleById(roleId)
	if err != nil {
		fmt.Printf("RoleService.SelectMenuListByRoleId: 查询角色信息失败: %v\n", err)
		return []int64{}, err
	}
	if role == nil {
		fmt.Printf("RoleService.SelectMenuListByRoleId: 角色不存在, RoleID=%d\n", roleId)
		return []int64{}, nil
	}

	return s.roleMenuDao.SelectMenuListByRoleId(roleId, role.MenuCheckStrictly)
}

// CheckRoleNameUnique 校验角色名称是否唯一 对应Java后端的checkRoleNameUnique
func (s *RoleService) CheckRoleNameUnique(role *model.SysRole) (bool, error) {
	fmt.Printf("RoleService.CheckRoleNameUnique: 校验角色名称, RoleName=%s\n", role.RoleName)
	return s.roleDao.CheckRoleNameUnique(role.RoleName, role.RoleID)
}

// CheckRoleKeyUnique 校验角色权限字符串是否唯一 对应Java后端的checkRoleKeyUnique
func (s *RoleService) CheckRoleKeyUnique(role *model.SysRole) (bool, error) {
	fmt.Printf("RoleService.CheckRoleKeyUnique: 校验角色权限字符串, RoleKey=%s\n", role.RoleKey)
	return s.roleDao.CheckRoleKeyUnique(role.RoleKey, role.RoleID)
}

// CheckRoleAllowed 校验角色是否允许操作 对应Java后端的checkRoleAllowed
func (s *RoleService) CheckRoleAllowed(role *model.SysRole) error {
	if role.IsAdmin() {
		return fmt.Errorf("不允许操作超级管理员角色")
	}
	return nil
}

// CheckRoleDataScope 校验角色是否有数据权限 对应Java后端的checkRoleDataScope
func (s *RoleService) CheckRoleDataScope(currentUser *model.SysUser, roleIds ...int64) error {
	fmt.Printf("RoleService.CheckRoleDataScope: 校验角色数据权限, CurrentUserID=%d, RoleIDs=%v\n", currentUser.UserID, roleIds)

	// 超级管理员跳过数据权限校验 对应Java后端的SysUser.isAdmin(SecurityUtils.getUserId())
	if currentUser.IsAdmin() {
		fmt.Printf("RoleService.CheckRoleDataScope: 超级管理员，跳过数据权限校验\n")
		return nil
	}

	// 如果用户没有角色信息，需要重新加载
	if len(currentUser.Roles) == 0 {
		fmt.Printf("RoleService.CheckRoleDataScope: 用户角色信息为空，重新加载用户信息\n")
		userService := NewUserService()
		fullUser, err := userService.SelectUserById(currentUser.UserID)
		if err != nil {
			return fmt.Errorf("重新加载用户信息失败: %v", err)
		}
		if fullUser != nil {
			currentUser.Roles = fullUser.Roles
			currentUser.Dept = fullUser.Dept
			fmt.Printf("RoleService.CheckRoleDataScope: 重新加载用户角色成功, 角色数量=%d\n", len(currentUser.Roles))
		}
	}

	// 逐个校验角色数据权限 对应Java后端的for (Long roleId : roleIds)
	for _, roleId := range roleIds {
		// 构建查询条件 对应Java后端的SysRole role = new SysRole(); role.setRoleId(roleId);
		role := &model.SysRole{}
		role.RoleID = roleId

		// 使用数据权限查询角色列表 对应Java后端的SpringUtils.getAopProxy(this).selectRoleList(role)
		roles, err := s.SelectRoleListWithDataScope(currentUser, role)
		if err != nil {
			return fmt.Errorf("数据权限校验失败: %v", err)
		}

		// 如果查询结果为空，说明没有权限访问该角色 对应Java后端的StringUtils.isEmpty(roles)
		if len(roles) == 0 {
			return fmt.Errorf("没有权限访问角色数据！")
		}
	}

	fmt.Printf("RoleService.CheckRoleDataScope: 角色数据权限校验通过\n")
	return nil
}

// SelectRoleListWithDataScope 查询角色列表（支持数据权限） 对应Java后端的@DataScope注解
func (s *RoleService) SelectRoleListWithDataScope(currentUser *model.SysUser, queryRole *model.SysRole) ([]model.SysRole, error) {
	fmt.Printf("RoleService.SelectRoleListWithDataScope: 查询角色列表（数据权限）\n")

	// 创建查询参数
	params := make(map[string]interface{})

	// 应用数据权限 对应Java后端的@DataScope(deptAlias = "d")
	err := datascope.ApplyDataScope(currentUser, "d", "", "system:role:list", params)
	if err != nil {
		return nil, fmt.Errorf("应用数据权限失败: %v", err)
	}

	// 将数据权限SQL设置到查询角色对象中
	if queryRole == nil {
		queryRole = &model.SysRole{}
	}
	if queryRole.Params == nil {
		queryRole.Params = make(map[string]interface{})
	}

	// 复制数据权限参数
	for key, value := range params {
		queryRole.Params[key] = value
	}

	// 调用DAO层查询
	return s.roleDao.SelectRoleList(queryRole)
}

// InsertRole 新增角色 对应Java后端的insertRole
func (s *RoleService) InsertRole(role *model.SysRole) error {
	fmt.Printf("RoleService.InsertRole: 新增角色, RoleName=%s, MenuIDs=%v\n", role.RoleName, role.MenuIDs)

	// 设置创建时间
	now := time.Now()
	role.CreateTime = &now

	// 新增角色信息
	err := s.roleDao.InsertRole(role)
	if err != nil {
		fmt.Printf("RoleService.InsertRole: 新增角色失败: %v\n", err)
		return err
	}

	fmt.Printf("RoleService.InsertRole: 角色信息新增成功, RoleID=%d\n", role.RoleID)

	// 新增角色菜单关联 对应Java后端的insertRoleMenu
	err = s.insertRoleMenu(role)
	if err != nil {
		fmt.Printf("RoleService.InsertRole: 新增角色菜单关联失败: %v\n", err)
		return err
	}

	fmt.Printf("RoleService.InsertRole: 新增角色成功, RoleID=%d\n", role.RoleID)
	return nil
}

// UpdateRole 修改角色 对应Java后端的updateRole
func (s *RoleService) UpdateRole(role *model.SysRole) error {
	fmt.Printf("RoleService.UpdateRole: 修改角色, RoleID=%d\n", role.RoleID)

	// 设置更新时间
	now := time.Now()
	role.UpdateTime = &now

	// 修改角色信息
	err := s.roleDao.UpdateRole(role)
	if err != nil {
		return err
	}

	// 删除角色与菜单关联
	err = s.roleMenuDao.DeleteRoleMenuByRoleId(role.RoleID)
	if err != nil {
		return err
	}

	// 新增角色菜单关联
	return s.insertRoleMenu(role)
}

// UpdateRoleStatus 修改角色状态 对应Java后端的updateRoleStatus
func (s *RoleService) UpdateRoleStatus(role *model.SysRole) error {
	fmt.Printf("RoleService.UpdateRoleStatus: 修改角色状态, RoleID=%d, Status=%s\n", role.RoleID, role.Status)

	// 设置更新时间
	now := time.Now()
	role.UpdateTime = &now

	return s.roleDao.UpdateRole(role)
}

// DeleteRoleById 删除单个角色 对应Java后端的deleteRoleById
func (s *RoleService) DeleteRoleById(currentUser *model.SysUser, roleId int64) error {
	fmt.Printf("RoleService.DeleteRoleById: 删除角色, CurrentUserID=%d, RoleID=%d\n", currentUser.UserID, roleId)

	// 校验角色是否允许操作
	role := &model.SysRole{RoleID: roleId}
	if err := s.CheckRoleAllowed(role); err != nil {
		return err
	}

	// 校验数据权限
	if err := s.CheckRoleDataScope(currentUser, roleId); err != nil {
		return err
	}

	// 查询角色信息
	roleInfo, err := s.roleDao.SelectRoleById(roleId)
	if err != nil {
		return err
	}
	if roleInfo == nil {
		return fmt.Errorf("角色不存在")
	}

	// 检查角色是否已分配给用户
	count, err := s.userRoleDao.CountUserRoleByRoleId(roleId)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("%s已分配,不能删除", roleInfo.RoleName)
	}

	// 删除角色与菜单关联
	err = s.roleMenuDao.DeleteRoleMenuByRoleId(roleId)
	if err != nil {
		return err
	}

	// 删除角色与部门关联
	err = s.roleDeptDao.DeleteRoleDeptByRoleId(roleId)
	if err != nil {
		return err
	}

	// 删除角色信息
	return s.roleDao.DeleteRoleById(roleId)
}

// DeleteRoleByIds 批量删除角色 对应Java后端的deleteRoleByIds
func (s *RoleService) DeleteRoleByIds(currentUser *model.SysUser, roleIds []int64) error {
	fmt.Printf("RoleService.DeleteRoleByIds: 批量删除角色, CurrentUserID=%d, RoleIDs=%v\n", currentUser.UserID, roleIds)

	// 校验每个角色 对应Java后端的for (Long roleId : roleIds)
	for _, roleId := range roleIds {
		// 校验角色是否允许操作 对应Java后端的checkRoleAllowed(new SysRole(roleId))
		role := &model.SysRole{RoleID: roleId}
		if err := s.CheckRoleAllowed(role); err != nil {
			return err
		}

		// 校验数据权限 对应Java后端的checkRoleDataScope(roleId)
		if err := s.CheckRoleDataScope(currentUser, roleId); err != nil {
			return err
		}

		// 查询角色信息
		roleInfo, err := s.roleDao.SelectRoleById(roleId)
		if err != nil {
			return err
		}
		if roleInfo == nil {
			return fmt.Errorf("角色不存在")
		}

		// 检查角色是否已分配给用户
		count, err := s.userRoleDao.CountUserRoleByRoleId(roleId)
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("%s已分配,不能删除", roleInfo.RoleName)
		}
	}

	// 删除角色与菜单关联
	for _, roleId := range roleIds {
		err := s.roleMenuDao.DeleteRoleMenuByRoleId(roleId)
		if err != nil {
			return err
		}
	}

	// 删除角色与部门关联
	err := s.roleDeptDao.DeleteRoleDept(roleIds)
	if err != nil {
		return err
	}

	// 删除角色信息
	return s.roleDao.DeleteRoleByIds(roleIds)
}

// AuthDataScope 修改角色数据权限 对应Java后端的authDataScope
func (s *RoleService) AuthDataScope(role *model.SysRole) error {
	fmt.Printf("RoleService.AuthDataScope: 修改角色数据权限, RoleID=%d, DataScope=%s\n", role.RoleID, role.DataScope)

	// 设置更新时间
	now := time.Now()
	role.UpdateTime = &now

	// 修改角色信息
	err := s.roleDao.UpdateRole(role)
	if err != nil {
		return err
	}

	// 删除角色与部门关联
	err = s.roleDeptDao.DeleteRoleDeptByRoleId(role.RoleID)
	if err != nil {
		return err
	}

	// 新增角色和部门信息（数据权限）
	return s.insertRoleDept(role)
}

// insertRoleMenu 新增角色菜单信息 对应Java后端的insertRoleMenu
func (s *RoleService) insertRoleMenu(role *model.SysRole) error {
	fmt.Printf("RoleService.insertRoleMenu: 新增角色菜单关联, RoleID=%d, MenuIDs=%v\n", role.RoleID, role.MenuIDs)

	// 对应Java后端的逻辑：如果没有菜单ID，直接返回成功（rows = 1）
	if len(role.MenuIDs) == 0 {
		fmt.Printf("RoleService.insertRoleMenu: 没有菜单ID，直接返回成功\n")
		return nil
	}

	// 构建角色菜单关联数据 对应Java后端的for (Long menuId : role.getMenuIds())
	roleMenus := make([]model.SysRoleMenu, len(role.MenuIDs))
	for i, menuId := range role.MenuIDs {
		roleMenus[i] = model.SysRoleMenu{
			RoleID: role.RoleID,
			MenuID: menuId,
		}
	}

	// 批量新增角色菜单关联 对应Java后端的roleMenuMapper.batchRoleMenu(list)
	err := s.roleMenuDao.BatchInsertRoleMenu(roleMenus)
	if err != nil {
		fmt.Printf("RoleService.insertRoleMenu: 批量新增角色菜单关联失败: %v\n", err)
		return err
	}

	fmt.Printf("RoleService.insertRoleMenu: 批量新增角色菜单关联成功, 数量=%d\n", len(roleMenus))
	return nil
}

// insertRoleDept 新增角色部门信息（数据权限） 对应Java后端的insertRoleDept
func (s *RoleService) insertRoleDept(role *model.SysRole) error {
	if len(role.DeptIDs) == 0 {
		return nil
	}

	// 构建角色部门关联数据
	roleDepts := make([]model.SysRoleDept, len(role.DeptIDs))
	for i, deptId := range role.DeptIDs {
		roleDepts[i] = model.SysRoleDept{
			RoleID: role.RoleID,
			DeptID: deptId,
		}
	}

	return s.roleDeptDao.BatchInsertRoleDept(roleDepts)
}

// SelectRolesByUserId 根据用户ID查询角色列表 对应Java后端的selectRolesByUserId
func (s *RoleService) SelectRolesByUserId(userId int64) ([]model.SysRole, error) {
	fmt.Printf("RoleService.SelectRolesByUserId: 查询用户角色, UserID=%d\n", userId)

	// 查询用户已分配的角色权限 对应Java后端的roleMapper.selectRolePermissionByUserId(userId)
	userRoles, err := s.roleDao.SelectRolePermissionByUserId(userId)
	if err != nil {
		return nil, err
	}

	// 查询所有角色 对应Java后端的selectRoleAll()
	allRoles, err := s.SelectRoleAll()
	if err != nil {
		return nil, err
	}

	// 设置Flag字段标识用户已分配的角色 对应Java后端的Flag设置逻辑
	for i := range allRoles {
		allRoles[i].Flag = false // 默认未分配
		for _, userRole := range userRoles {
			if allRoles[i].RoleID == userRole.RoleID {
				allRoles[i].Flag = true // 已分配
				break
			}
		}
	}

	fmt.Printf("RoleService.SelectRolesByUserId: 查询到角色数量=%d, 用户已分配角色数量=%d\n", len(allRoles), len(userRoles))
	return allRoles, nil
}

// SelectRoleListByUserId 根据用户ID获取角色选择框列表 对应Java后端的selectRoleListByUserId
func (s *RoleService) SelectRoleListByUserId(userId int64) ([]int64, error) {
	fmt.Printf("RoleService.SelectRoleListByUserId: 查询用户角色ID列表, UserID=%d\n", userId)
	return s.userRoleDao.SelectRoleListByUserId(userId)
}

// CountUserRoleByRoleId 通过角色ID查询角色使用数量 对应Java后端的countUserRoleByRoleId
func (s *RoleService) CountUserRoleByRoleId(roleId int64) (int64, error) {
	fmt.Printf("RoleService.CountUserRoleByRoleId: 查询角色使用数量, RoleID=%d\n", roleId)
	return s.userRoleDao.CountUserRoleByRoleId(roleId)
}

// DeleteAuthUser 取消授权用户角色 对应Java后端的deleteAuthUser
func (s *RoleService) DeleteAuthUser(userRole *model.SysUserRole) error {
	fmt.Printf("RoleService.DeleteAuthUser: 取消授权用户角色, UserID=%d, RoleID=%d\n", userRole.UserID, userRole.RoleID)
	return s.userRoleDao.DeleteAuthUser(userRole)
}

// DeleteAuthUsers 批量取消授权用户角色 对应Java后端的deleteAuthUsers
func (s *RoleService) DeleteAuthUsers(roleId int64, userIds []int64) error {
	fmt.Printf("RoleService.DeleteAuthUsers: 批量取消授权用户角色, RoleID=%d, UserIDs=%v\n", roleId, userIds)
	return s.userRoleDao.DeleteAuthUsers(roleId, userIds)
}

// SelectAuthUserAll 批量选择授权用户角色 对应Java后端的selectAuthUserAll
func (s *RoleService) SelectAuthUserAll(roleId int64, userIds []int64) error {
	fmt.Printf("RoleService.SelectAuthUserAll: 批量选择授权用户角色, RoleID=%d, UserIDs=%v\n", roleId, userIds)
	return s.userRoleDao.SelectAuthUserAll(roleId, userIds)
}

// InsertAuthUsers 批量选择授权用户 对应Java后端的insertAuthUsers
func (s *RoleService) InsertAuthUsers(roleId int64, userIds []int64) error {
	fmt.Printf("RoleService.InsertAuthUsers: 批量选择授权用户, RoleID=%d, UserIDs=%v\n", roleId, userIds)

	// 构建用户角色关联列表 对应Java后端的List<SysUserRole> list = new ArrayList<SysUserRole>()
	var userRoles []model.SysUserRole
	for _, userId := range userIds {
		userRole := model.SysUserRole{
			UserID: userId,
			RoleID: roleId,
		}
		userRoles = append(userRoles, userRole)
	}

	// 批量插入用户角色关联 对应Java后端的userRoleMapper.batchUserRole(list)
	return s.userRoleDao.BatchInsertUserRole(userRoles)
}

// SelectRolePermissionByUserId 根据用户ID查询角色权限 对应Java后端的selectRolePermissionByUserId
func (s *RoleService) SelectRolePermissionByUserId(userId int64) ([]string, error) {
	fmt.Printf("RoleService.SelectRolePermissionByUserId: 查询用户角色权限, UserID=%d\n", userId)

	// 查询用户角色权限 对应Java后端的roleMapper.selectRolePermissionByUserId(userId)
	roles, err := s.roleDao.SelectRolePermissionByUserId(userId)
	if err != nil {
		return nil, err
	}

	// 提取角色权限字符串并去重 对应Java后端的Set<String> permsSet = new HashSet<>()
	permissionSet := make(map[string]bool)
	for _, role := range roles {
		if role.Status == "0" && role.RoleKey != "" { // 只有正常状态的角色才有效
			// 对应Java后端的permsSet.addAll(Arrays.asList(perm.getRoleKey().trim().split(",")))
			roleKeys := strings.Split(strings.TrimSpace(role.RoleKey), ",")
			for _, key := range roleKeys {
				key = strings.TrimSpace(key)
				if key != "" {
					permissionSet[key] = true
				}
			}
		}
	}

	// 转换为切片
	var permissions []string
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}

	fmt.Printf("RoleService.SelectRolePermissionByUserId: 查询到角色权限, UserID=%d, Permissions=%v\n", userId, permissions)
	return permissions, nil
}

// SelectRolesByUserName 根据用户名查询角色列表 对应Java后端的selectRolesByUserName
func (s *RoleService) SelectRolesByUserName(userName string) ([]model.SysRole, error) {
	fmt.Printf("RoleService.SelectRolesByUserName: 根据用户名查询角色, UserName=%s\n", userName)
	return s.roleDao.SelectRolesByUserName(userName)
}
