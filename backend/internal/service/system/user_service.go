package system

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
	"wosm/internal/utils"
	"wosm/pkg/datascope"
)

// UserService 用户服务 对应Java后端的SysUserServiceImpl
type UserService struct {
	userDao       *dao.UserDao
	userRoleDao   *dao.UserRoleDao
	userPostDao   *dao.UserPostDao
	roleDao       *dao.RoleDao
	postDao       *dao.PostDao
	deptDao       *dao.DeptDao
	configService *ConfigService
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userDao:       dao.NewUserDao(),
		userRoleDao:   dao.NewUserRoleDao(),
		userPostDao:   dao.NewUserPostDao(),
		roleDao:       dao.NewRoleDao(),
		postDao:       dao.NewPostDao(),
		deptDao:       dao.NewDeptDao(),
		configService: NewConfigService(),
	}
}

// SelectUserList 查询用户列表 对应Java后端的selectUserList
func (s *UserService) SelectUserList(user *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	// 直接使用数据库分页查询
	return s.userDao.SelectUserList(user, pageNum, pageSize)
}

// SelectAllocatedList 查询已分配用户角色列表 对应Java后端的selectAllocatedList
func (s *UserService) SelectAllocatedList(currentUser *model.SysUser, user *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	fmt.Printf("UserService.SelectAllocatedList: 查询已分配用户角色列表, RoleID=%v\n", user.RoleID)

	if user.RoleID == nil {
		return nil, 0, fmt.Errorf("角色ID不能为空")
	}

	// 使用数据权限处理器查询
	return s.userDao.SelectAllocatedList(currentUser, user, pageNum, pageSize)
}

// SelectUnallocatedList 查询未分配用户角色列表 对应Java后端的selectUnallocatedList
func (s *UserService) SelectUnallocatedList(currentUser *model.SysUser, user *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	fmt.Printf("UserService.SelectUnallocatedList: 查询未分配用户角色列表, RoleID=%v\n", user.RoleID)

	if user.RoleID == nil {
		return nil, 0, fmt.Errorf("角色ID不能为空")
	}

	// 使用数据权限处理器查询
	return s.userDao.SelectUnallocatedList(currentUser, user, pageNum, pageSize)
}

// SelectUserListWithDataScope 查询用户列表（支持数据权限） 对应Java后端的@DataScope注解
func (s *UserService) SelectUserListWithDataScope(currentUser *model.SysUser, queryUser *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	fmt.Printf("UserService.SelectUserListWithDataScope: 查询用户列表（数据权限）\n")
	fmt.Printf("UserService.SelectUserListWithDataScope: 当前用户ID=%d, 用户名=%s\n", currentUser.UserID, currentUser.UserName)
	fmt.Printf("UserService.SelectUserListWithDataScope: 当前用户角色数量=%d\n", len(currentUser.Roles))

	// 打印当前用户的角色信息
	for i, role := range currentUser.Roles {
		fmt.Printf("UserService.SelectUserListWithDataScope: 角色%d: ID=%d, 名称=%s, 数据权限=%s, 状态=%s, 权限数量=%d\n",
			i+1, role.RoleID, role.RoleName, role.DataScope, role.Status, len(role.Permissions))
	}

	// 创建查询参数
	params := make(map[string]interface{})

	// 应用数据权限 对应Java后端的@DataScope(deptAlias = "d", userAlias = "u")
	err := datascope.ApplyDataScope(currentUser, "d", "u", "system:user:list", params)
	if err != nil {
		fmt.Printf("UserService.SelectUserListWithDataScope: 应用数据权限失败: %v\n", err)
		return nil, 0, fmt.Errorf("应用数据权限失败: %v", err)
	}

	// 打印数据权限参数
	fmt.Printf("UserService.SelectUserListWithDataScope: 数据权限参数: %+v\n", params)

	// 将数据权限SQL设置到查询用户对象中
	if queryUser == nil {
		queryUser = &model.SysUser{}
	}
	if queryUser.Params == nil {
		queryUser.Params = make(map[string]interface{})
	}

	// 合并数据权限参数
	for key, value := range params {
		queryUser.Params[key] = value
		fmt.Printf("UserService.SelectUserListWithDataScope: 设置查询参数: %s = %v\n", key, value)
	}

	// 执行查询
	users, total, err := s.userDao.SelectUserList(queryUser, pageNum, pageSize)
	fmt.Printf("UserService.SelectUserListWithDataScope: 查询结果: 总数=%d, 当前页数量=%d, 错误=%v\n", total, len(users), err)
	return users, total, err
}

// SelectUserById 根据用户ID查询用户信息 对应Java后端的selectUserById
func (s *UserService) SelectUserById(userId int64) (*model.SysUser, error) {
	return s.userDao.SelectUserById(userId)
}

// SelectUserByLoginName 根据用户名查询用户 对应Java后端的selectUserByLoginName
func (s *UserService) SelectUserByLoginName(loginName string) (*model.SysUser, error) {
	return s.userDao.SelectUserByLoginName(loginName)
}

// InsertUser 新增用户 对应Java后端的insertUser
func (s *UserService) InsertUser(user *model.SysUser, createBy string) error {
	// 校验用户名唯一性
	unique, err := s.userDao.CheckLoginNameUnique(user.UserName, 0)
	if err != nil {
		return fmt.Errorf("校验用户名失败: %v", err)
	}
	if !unique {
		return fmt.Errorf("新增用户'%s'失败，登录账号已存在", user.UserName)
	}

	// 校验手机号唯一性
	if user.Phonenumber != "" {
		phoneUnique, err := s.userDao.CheckPhoneUnique(user.Phonenumber, 0)
		if err != nil {
			return fmt.Errorf("校验手机号失败: %v", err)
		}
		if !phoneUnique {
			return fmt.Errorf("新增用户'%s'失败，手机号码已存在", user.UserName)
		}
	}

	// 校验邮箱唯一性
	if user.Email != "" {
		emailUnique, err := s.userDao.CheckEmailUnique(user.Email, 0)
		if err != nil {
			return fmt.Errorf("校验邮箱失败: %v", err)
		}
		if !emailUnique {
			return fmt.Errorf("新增用户'%s'失败，邮箱账号已存在", user.UserName)
		}
	}

	// 使用BCrypt加密密码（与Java后端一致）
	if user.Password != "" {
		hashedPassword, err := utils.BcryptPassword(user.Password)
		if err != nil {
			return fmt.Errorf("密码加密失败: %v", err)
		}
		user.Password = hashedPassword
	}

	// 设置创建信息
	now := time.Now()
	user.CreateBy = createBy
	user.CreateTime = &now
	user.DelFlag = "0"
	user.Status = "0"

	// 1. 新增用户信息 对应Java后端的userMapper.insertUser(user)
	err = s.userDao.InsertUser(user)
	if err != nil {
		return fmt.Errorf("新增用户失败: %v", err)
	}

	// 2. 新增用户岗位关联 对应Java后端的insertUserPost(user)
	if len(user.PostIDs) > 0 {
		err = s.insertUserPost(user)
		if err != nil {
			return fmt.Errorf("新增用户岗位关联失败: %v", err)
		}
	}

	// 3. 新增用户与角色管理 对应Java后端的insertUserRole(user)
	if len(user.RoleIDs) > 0 {
		err = s.insertUserRole(user)
		if err != nil {
			return fmt.Errorf("新增用户角色关联失败: %v", err)
		}
	}

	fmt.Printf("UserService.InsertUser: 新增用户成功, UserID=%d\n", user.UserID)
	return nil
}

// UpdateUser 修改用户 对应Java后端的updateUser
func (s *UserService) UpdateUser(user *model.SysUser, updateBy string) error {
	fmt.Printf("UserService.UpdateUser: 修改用户, UserID=%d\n", user.UserID)

	// 校验用户名唯一性
	unique, err := s.userDao.CheckLoginNameUnique(user.UserName, user.UserID)
	if err != nil {
		return fmt.Errorf("校验用户名失败: %v", err)
	}
	if !unique {
		return fmt.Errorf("修改用户'%s'失败，登录账号已存在", user.UserName)
	}

	// 校验手机号唯一性
	if user.Phonenumber != "" {
		phoneUnique, err := s.userDao.CheckPhoneUnique(user.Phonenumber, user.UserID)
		if err != nil {
			return fmt.Errorf("校验手机号失败: %v", err)
		}
		if !phoneUnique {
			return fmt.Errorf("修改用户'%s'失败，手机号码已存在", user.UserName)
		}
	}

	// 校验邮箱唯一性
	if user.Email != "" {
		emailUnique, err := s.userDao.CheckEmailUnique(user.Email, user.UserID)
		if err != nil {
			return fmt.Errorf("校验邮箱失败: %v", err)
		}
		if !emailUnique {
			return fmt.Errorf("修改用户'%s'失败，邮箱账号已存在", user.UserName)
		}
	}

	// 如果修改密码，使用BCrypt重新加密（与Java后端一致）
	if user.Password != "" {
		hashedPassword, err := utils.BcryptPassword(user.Password)
		if err != nil {
			return fmt.Errorf("密码加密失败: %v", err)
		}
		user.Password = hashedPassword
		// 设置密码更新时间
		now := time.Now()
		user.PwdUpdateDate = &now
	}

	// 设置更新信息
	now := time.Now()
	user.UpdateBy = updateBy
	user.UpdateTime = &now

	// 1. 删除用户与角色关联 对应Java后端的userRoleMapper.deleteUserRoleByUserId(userId)
	err = s.userRoleDao.DeleteUserRoleByUserId(user.UserID)
	if err != nil {
		return fmt.Errorf("删除用户角色关联失败: %v", err)
	}

	// 2. 新增用户与角色管理 对应Java后端的insertUserRole(user)
	if len(user.RoleIDs) > 0 {
		err = s.insertUserRole(user)
		if err != nil {
			return fmt.Errorf("新增用户角色关联失败: %v", err)
		}
	}

	// 3. 删除用户与岗位关联 对应Java后端的userPostMapper.deleteUserPostByUserId(userId)
	err = s.userPostDao.DeleteUserPostByUserId(user.UserID)
	if err != nil {
		return fmt.Errorf("删除用户岗位关联失败: %v", err)
	}

	// 4. 新增用户与岗位管理 对应Java后端的insertUserPost(user)
	if len(user.PostIDs) > 0 {
		err = s.insertUserPost(user)
		if err != nil {
			return fmt.Errorf("新增用户岗位关联失败: %v", err)
		}
	}

	// 5. 更新用户基本信息 对应Java后端的userMapper.updateUser(user)
	err = s.userDao.UpdateUser(user)
	if err != nil {
		return fmt.Errorf("更新用户信息失败: %v", err)
	}

	fmt.Printf("UserService.UpdateUser: 修改用户成功, UserID=%d\n", user.UserID)
	return nil
}

// DeleteUserByIds 批量删除用户 对应Java后端的deleteUserByIds
func (s *UserService) DeleteUserByIds(currentUser *model.SysUser, userIds []int64) error {
	fmt.Printf("UserService.DeleteUserByIds: 批量删除用户, CurrentUserID=%d, UserIDs=%v\n", currentUser.UserID, userIds)

	// 逐个校验用户 对应Java后端的for (Long userId : userIds)
	for _, userId := range userIds {
		// 校验用户是否允许操作 对应Java后端的checkUserAllowed
		user := &model.SysUser{UserID: userId}
		if err := s.CheckUserAllowed(user); err != nil {
			return err
		}

		// 校验用户数据权限 对应Java后端的checkUserDataScope
		if err := s.CheckUserDataScope(userId, currentUser); err != nil {
			return err
		}
	}

	// 删除用户与角色关联 对应Java后端的userRoleMapper.deleteUserRole(userIds)
	for _, userId := range userIds {
		err := s.userRoleDao.DeleteUserRoleByUserId(userId)
		if err != nil {
			return fmt.Errorf("删除用户角色关联失败: %v", err)
		}
	}

	// 删除用户与岗位关联 对应Java后端的userPostMapper.deleteUserPost(userIds)
	for _, userId := range userIds {
		err := s.userPostDao.DeleteUserPostByUserId(userId)
		if err != nil {
			return fmt.Errorf("删除用户岗位关联失败: %v", err)
		}
	}

	// 删除用户信息 对应Java后端的userMapper.deleteUserByIds(userIds)
	for _, userId := range userIds {
		err := s.userDao.DeleteUserById(userId)
		if err != nil {
			return fmt.Errorf("删除用户失败: %v", err)
		}
	}

	fmt.Printf("UserService.DeleteUserByIds: 批量删除用户成功, 数量=%d\n", len(userIds))
	return nil
}

// ResetPwd 重置密码 对应Java后端的resetPwd
func (s *UserService) ResetPwd(userId int64, password string, updateBy string) error {
	// 使用BCrypt加密密码（与Java后端一致）
	hashedPassword, err := utils.BcryptPassword(password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	now := time.Now()
	user := &model.SysUser{
		UserID:        userId,
		Password:      hashedPassword,
		UpdateBy:      updateBy,
		UpdateTime:    &now,
		PwdUpdateDate: &now, // 设置密码更新时间
	}

	return s.userDao.UpdateUser(user)
}

// ChangeStatus 修改用户状态 对应Java后端的changeStatus
func (s *UserService) ChangeStatus(user *model.SysUser, updateBy string) error {
	fmt.Printf("UserService.ChangeStatus: 修改用户状态 - UserID=%d, Status=%s, UpdateBy=%s\n", user.UserID, user.Status, updateBy)

	// 不能停用管理员
	if user.UserID == 1 {
		fmt.Printf("UserService.ChangeStatus: 不允许停用管理员用户\n")
		return errors.New("不允许停用管理员用户")
	}

	now := time.Now()
	user.UpdateBy = updateBy
	user.UpdateTime = &now

	err := s.userDao.UpdateUser(user)
	if err != nil {
		fmt.Printf("UserService.ChangeStatus: 修改用户状态失败: %v\n", err)
		return err
	}

	fmt.Printf("UserService.ChangeStatus: 修改用户状态成功\n")
	return nil
}

// UpdateUserStatus 修改用户状态 对应Java后端的updateUserStatus
func (s *UserService) UpdateUserStatus(user *model.SysUser) error {
	fmt.Printf("UserService.UpdateUserStatus: 修改用户状态, UserID=%d, Status=%s\n", user.UserID, user.Status)
	return s.userDao.UpdateUser(user)
}

// GetUserAuthRole 获取用户授权角色信息 对应Java后端的authRole
func (s *UserService) GetUserAuthRole(userId int64) (map[string]interface{}, error) {
	// 获取用户信息
	user, err := s.userDao.SelectUserById(userId)
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 获取所有角色列表
	roleService := NewRoleService()
	allRoles, err := roleService.SelectRoleAll()
	if err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %v", err)
	}

	// 获取用户已分配的角色ID列表
	userRoleIds, err := roleService.SelectRoleListByUserId(userId)
	if err != nil {
		return nil, fmt.Errorf("查询用户角色失败: %v", err)
	}

	// 创建角色ID映射，用于快速查找
	userRoleMap := make(map[int64]bool)
	for _, roleId := range userRoleIds {
		userRoleMap[roleId] = true
	}

	// 设置角色的flag字段，标识用户是否拥有该角色
	for i := range allRoles {
		allRoles[i].Flag = userRoleMap[allRoles[i].RoleID]
	}

	// 过滤管理员角色 对应Java后端的ajax.put("roles", SysUser.isAdmin(userId) ? roles : roles.stream().filter(r -> !r.isAdmin()).collect(Collectors.toList()))
	var filteredRoles []model.SysRole
	if userId == 1 { // 管理员用户可以看到所有角色
		filteredRoles = allRoles
	} else { // 非管理员用户过滤掉管理员角色
		for _, role := range allRoles {
			if role.RoleID != 1 { // 过滤掉管理员角色（roleId=1）
				filteredRoles = append(filteredRoles, role)
			}
		}
	}

	result := map[string]interface{}{
		"user":  user,
		"roles": filteredRoles,
	}

	fmt.Printf("UserService.GetUserAuthRole: 获取用户授权角色成功, UserID=%d, 角色数量=%d\n", userId, len(filteredRoles))
	return result, nil
}

// CheckUserNameUnique 校验用户名是否唯一 对应Java后端的checkUserNameUnique
func (s *UserService) CheckUserNameUnique(user *model.SysUser) bool {
	unique, err := s.userDao.CheckLoginNameUnique(user.UserName, user.UserID)
	if err != nil {
		return false
	}
	return unique
}

// CheckPhoneUnique 校验手机号是否唯一 对应Java后端的checkPhoneUnique
func (s *UserService) CheckPhoneUnique(user *model.SysUser) bool {
	if user.Phonenumber == "" {
		return true
	}
	unique, err := s.userDao.CheckPhoneUnique(user.Phonenumber, user.UserID)
	if err != nil {
		return false
	}
	return unique
}

// CheckEmailUnique 校验邮箱是否唯一 对应Java后端的checkEmailUnique
func (s *UserService) CheckEmailUnique(user *model.SysUser) bool {
	if user.Email == "" {
		return true
	}
	unique, err := s.userDao.CheckEmailUnique(user.Email, user.UserID)
	if err != nil {
		return false
	}
	return unique
}

// RegisterUser 注册用户 对应Java后端的registerUser
func (s *UserService) RegisterUser(user *model.SysUser) error {
	return s.userDao.InsertUser(user)
}

// SelectUserRoleGroup 查询用户所属角色组 对应Java后端的selectUserRoleGroup
func (s *UserService) SelectUserRoleGroup(userName string) (string, error) {
	return s.userDao.SelectUserRoleGroup(userName)
}

// SelectUserPostGroup 查询用户所属岗位组 对应Java后端的selectUserPostGroup
func (s *UserService) SelectUserPostGroup(userName string) (string, error) {
	return s.userDao.SelectUserPostGroup(userName)
}

// CheckUserAllowed 校验用户是否允许操作 对应Java后端的checkUserAllowed
func (s *UserService) CheckUserAllowed(user *model.SysUser) error {
	if user.UserID != 0 && user.IsAdmin() {
		return fmt.Errorf("不允许操作超级管理员用户")
	}
	return nil
}

// CheckUserDataScope 校验用户是否有数据权限 对应Java后端的checkUserDataScope
func (s *UserService) CheckUserDataScope(userId int64, currentUser *model.SysUser) error {
	if !currentUser.IsAdmin() {
		user := &model.SysUser{UserID: userId}
		users, _, err := s.SelectUserListWithDataScope(currentUser, user, 1, 1)
		if err != nil {
			return fmt.Errorf("检查用户数据权限失败: %v", err)
		}
		if len(users) == 0 {
			return fmt.Errorf("没有权限访问用户数据！")
		}
	}
	return nil
}

// UpdateUserProfile 修改用户基本信息 对应Java后端的updateUserProfile
func (s *UserService) UpdateUserProfile(user *model.SysUser) error {
	return s.userDao.UpdateUserProfile(user)
}

// UpdateUserAvatar 修改用户头像 对应Java后端的updateUserAvatar
func (s *UserService) UpdateUserAvatar(userId int64, avatar string) error {
	return s.userDao.UpdateUserAvatar(userId, avatar)
}

// ResetUserPwd 重置用户密码(指定用户ID和密码) 对应Java后端的resetUserPwd
func (s *UserService) ResetUserPwd(userId int64, password string) error {
	// 加密密码
	hashedPassword, err := utils.BcryptPassword(password)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	return s.userDao.ResetUserPwd(userId, hashedPassword)
}

// InsertUserAuth 用户授权角色 对应Java后端的insertUserAuth
func (s *UserService) InsertUserAuth(userId int64, roleIds []int64) error {
	fmt.Printf("UserService.InsertUserAuth: 用户授权角色, UserID=%d, RoleIDs=%v\n", userId, roleIds)

	// 1. 删除用户现有角色关联 对应Java后端的userRoleMapper.deleteUserRoleByUserId(userId)
	err := s.userRoleDao.DeleteUserRoleByUserId(userId)
	if err != nil {
		fmt.Printf("UserService.InsertUserAuth: 删除用户现有角色关联失败: %v\n", err)
		return fmt.Errorf("删除用户现有角色关联失败: %v", err)
	}

	// 2. 插入新的角色关联 对应Java后端的insertUserRole(userId, roleIds)
	if len(roleIds) > 0 {
		err = s.insertUserRoleByIds(userId, roleIds)
		if err != nil {
			fmt.Printf("UserService.InsertUserAuth: 插入新的角色关联失败: %v\n", err)
			return fmt.Errorf("插入新的角色关联失败: %v", err)
		}
	}

	fmt.Printf("UserService.InsertUserAuth: 用户授权角色成功, UserID=%d\n", userId)
	return nil
}

// ImportUser 导入用户数据 对应Java后端的importUser
func (s *UserService) ImportUser(userList []model.SysUser, updateSupport bool, operName string) (string, error) {
	fmt.Printf("UserService.ImportUser: 导入用户数据, 数量=%d, 更新支持=%t\n", len(userList), updateSupport)

	if len(userList) == 0 {
		return "", errors.New("导入用户数据不能为空")
	}

	successNum := 0
	failureNum := 0
	var successMsg []string
	var failureMsg []string

	for i, user := range userList {
		rowNum := i + 3 // Excel行号从3开始（第1行是标题，第2行是表头）

		// 验证必要字段
		if user.UserName == "" {
			failureNum++
			failureMsg = append(failureMsg, fmt.Sprintf("第%d行：用户名不能为空", rowNum))
			continue
		}

		if user.NickName == "" {
			failureNum++
			failureMsg = append(failureMsg, fmt.Sprintf("第%d行：用户昵称不能为空", rowNum))
			continue
		}

		// 验证部门ID是否存在
		if user.DeptID != nil && *user.DeptID > 0 {
			// TODO: 添加部门存在性验证
			// dept, err := s.deptService.SelectDeptById(*user.DeptID)
			// if err != nil || dept == nil {
			//     failureNum++
			//     failureMsg = append(failureMsg, fmt.Sprintf("第%d行：部门编号 %d 不存在", rowNum, *user.DeptID))
			//     continue
			// }
		}

		// 检查用户名是否已存在
		existingUser, err := s.userDao.SelectUserByLoginName(user.UserName)
		if err != nil {
			failureNum++
			failureMsg = append(failureMsg, fmt.Sprintf("第%d行：查询用户失败 - %v", rowNum, err))
			continue
		}

		if existingUser != nil {
			if !updateSupport {
				failureNum++
				failureMsg = append(failureMsg, fmt.Sprintf("第%d行：账号 %s 已存在", rowNum, user.UserName))
				continue
			} else {
				// 更新现有用户 - 对应Java后端的更新逻辑
				// 校验用户是否允许操作
				if err := s.CheckUserAllowed(existingUser); err != nil {
					failureNum++
					failureMsg = append(failureMsg, fmt.Sprintf("第%d行：%v", rowNum, err))
					continue
				}

				// 校验部门数据权限
				if user.DeptID != nil {
					// TODO: 实现部门数据权限校验 deptService.checkDeptDataScope(user.getDeptId())
				}

				user.UserID = existingUser.UserID
				user.UpdateBy = operName
				now := time.Now()
				user.UpdateTime = &now

				// 如果没有提供密码，保持原密码
				if user.Password == "" {
					user.Password = existingUser.Password
				} else {
					// 加密新密码
					hashedPassword, err := utils.BcryptPassword(user.Password)
					if err != nil {
						failureNum++
						failureMsg = append(failureMsg, fmt.Sprintf("第%d行：密码加密失败 - %v", rowNum, err))
						continue
					}
					user.Password = hashedPassword
					user.PwdUpdateDate = &now
				}

				err = s.userDao.UpdateUser(&user)
				if err != nil {
					failureNum++
					failureMsg = append(failureMsg, fmt.Sprintf("第%d行：更新用户失败 - %v", rowNum, err))
					continue
				}
				successNum++
				successMsg = append(successMsg, fmt.Sprintf("%d、账号 %s 更新成功", successNum, user.UserName))
			}
		} else {
			// 新增用户 - 对应Java后端的新增逻辑
			// 校验部门数据权限
			if user.DeptID != nil {
				// TODO: 实现部门数据权限校验 deptService.checkDeptDataScope(user.getDeptId())
			}

			// 获取系统默认密码
			defaultPassword := "123456" // TODO: 从配置中获取 configService.selectConfigByKey("sys.user.initPassword")
			if user.Password == "" {
				user.Password = defaultPassword
			}

			// 加密密码
			hashedPassword, err := utils.BcryptPassword(user.Password)
			if err != nil {
				failureNum++
				failureMsg = append(failureMsg, fmt.Sprintf("第%d行：密码加密失败 - %v", rowNum, err))
				continue
			}
			user.Password = hashedPassword

			// 设置创建信息
			now := time.Now()
			user.CreateBy = operName
			user.CreateTime = &now
			user.DelFlag = "0"
			if user.Status == "" {
				user.Status = "0" // 默认正常状态
			}

			err = s.userDao.InsertUser(&user)
			if err != nil {
				failureNum++
				failureMsg = append(failureMsg, fmt.Sprintf("第%d行：新增用户失败 - %v", rowNum, err))
				continue
			}
			successNum++
			successMsg = append(successMsg, fmt.Sprintf("%d、账号 %s 导入成功", successNum, user.UserName))
		}
	}

	// 构建返回消息 - 对应Java后端的消息格式
	var message string
	if failureNum > 0 {
		message = fmt.Sprintf("很抱歉，导入失败！共 %d 条数据格式不正确，错误如下：<br/>", failureNum)
		// 只显示前5条错误信息
		maxErrors := 5
		if len(failureMsg) > maxErrors {
			failureMsg = failureMsg[:maxErrors]
		}
		message += strings.Join(failureMsg, "<br/>")
		return "", errors.New(message)
	} else {
		message = fmt.Sprintf("恭喜您，数据已全部导入成功！共 %d 条，数据如下：<br/>", successNum)
		message += strings.Join(successMsg, "<br/>")
	}

	fmt.Printf("UserService.ImportUser: 导入完成, 成功=%d, 失败=%d\n", successNum, failureNum)
	return message, nil
}

// insertUserRole 新增用户角色信息 对应Java后端的insertUserRole
func (s *UserService) insertUserRole(user *model.SysUser) error {
	if len(user.RoleIDs) == 0 {
		return nil
	}

	fmt.Printf("UserService.insertUserRole: 新增用户角色关联, UserID=%d, RoleIDs=%v\n", user.UserID, user.RoleIDs)

	// 构建用户角色关联数据 对应Java后端的batchUserRole
	userRoles := make([]model.SysUserRole, len(user.RoleIDs))
	for i, roleId := range user.RoleIDs {
		userRoles[i] = model.SysUserRole{
			UserID: user.UserID,
			RoleID: roleId,
		}
	}

	// 批量插入用户角色关联
	return s.userRoleDao.BatchInsertUserRole(userRoles)
}

// insertUserPost 新增用户岗位信息 对应Java后端的insertUserPost
func (s *UserService) insertUserPost(user *model.SysUser) error {
	if len(user.PostIDs) == 0 {
		return nil
	}

	fmt.Printf("UserService.insertUserPost: 新增用户岗位关联, UserID=%d, PostIDs=%v\n", user.UserID, user.PostIDs)

	// 构建用户岗位关联数据 对应Java后端的batchUserPost
	userPosts := make([]model.SysUserPost, len(user.PostIDs))
	for i, postId := range user.PostIDs {
		userPosts[i] = model.SysUserPost{
			UserID: user.UserID,
			PostID: postId,
		}
	}

	// 批量插入用户岗位关联
	return s.userPostDao.BatchInsertUserPost(userPosts)
}

// insertUserRoleByIds 根据用户ID和角色ID数组插入用户角色关联 对应Java后端的insertUserRole(userId, roleIds)
func (s *UserService) insertUserRoleByIds(userId int64, roleIds []int64) error {
	fmt.Printf("UserService.insertUserRoleByIds: 插入用户角色关联, UserID=%d, RoleIDs=%v\n", userId, roleIds)

	// 构建用户角色关联数据 对应Java后端的batchUserRole
	userRoles := make([]model.SysUserRole, len(roleIds))
	for i, roleId := range roleIds {
		userRoles[i] = model.SysUserRole{
			UserID: userId,
			RoleID: roleId,
		}
	}

	// 批量插入用户角色关联
	return s.userRoleDao.BatchInsertUserRole(userRoles)
}
