package dao

import (
	"fmt"
	"strings"
	"time"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// UserDao 用户数据访问对象 对应Java后端的SysUserMapper
type UserDao struct {
	db *gorm.DB
}

// NewUserDao 创建用户DAO
func NewUserDao() *UserDao {
	return &UserDao{
		db: database.GetDB(),
	}
}

// SelectUserByLoginName 根据用户名查询用户 对应Java后端的selectUserByLoginName
func (d *UserDao) SelectUserByLoginName(loginName string) (*model.SysUser, error) {
	var user model.SysUser
	err := d.db.Where("user_name = ? AND del_flag = '0'", loginName).
		Preload("Dept").
		Preload("Roles").
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 加载角色权限信息 对应Java后端的角色权限查询
	err = d.loadRolePermissions(&user)
	if err != nil {
		fmt.Printf("SelectUserByLoginName: 加载角色权限失败: %v\n", err)
		// 权限加载失败不影响用户查询，继续返回用户信息
	}

	return &user, nil
}

// SelectUserById 根据用户ID查询用户 对应Java后端的selectUserById
func (d *UserDao) SelectUserById(userId int64) (*model.SysUser, error) {
	var user model.SysUser
	err := d.db.Where("user_id = ? AND del_flag = '0'", userId).
		Preload("Dept").
		Preload("Roles").
		First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	// 加载角色权限信息 对应Java后端的角色权限查询
	err = d.loadRolePermissions(&user)
	if err != nil {
		fmt.Printf("SelectUserById: 加载角色权限失败: %v\n", err)
		// 权限加载失败不影响用户查询，继续返回用户信息
	}

	return &user, nil
}

// SelectUserList 查询用户列表 对应Java后端的selectUserList
func (d *UserDao) SelectUserList(user *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	var users []model.SysUser
	var total int64

	// 对应Java后端的SQL：
	// select u.user_id, u.dept_id, u.nick_name, u.user_name, u.email, u.avatar, u.phonenumber, u.sex, u.status, u.del_flag, u.login_ip, u.login_date, u.create_by, u.create_time, u.remark, d.dept_name, d.leader
	// from sys_user u left join sys_dept d on u.dept_id = d.dept_id where u.del_flag = '0'

	// 构建基础查询条件
	query := d.db.Where("del_flag = '0'")

	if user.UserName != "" {
		query = query.Where("user_name LIKE ?", "%"+user.UserName+"%")
	}
	if user.Status != "" {
		query = query.Where("status = ?", user.Status)
	}
	if user.DeptID != nil && *user.DeptID != 0 {
		// 对应Java后端的逻辑：
		// AND (u.dept_id = #{deptId} OR u.dept_id IN ( SELECT t.dept_id FROM sys_dept t WHERE find_in_set(#{deptId}, ancestors) ))
		// 查找指定部门及其所有子部门的用户
		deptID := *user.DeptID
		deptIDStr := fmt.Sprintf("%d", deptID)

		// 在SQL Server中实现find_in_set功能：
		// 1. 直接匹配部门ID
		// 2. 查找ancestors字段中包含该部门ID的所有子部门
		// ancestors格式：0,100,101 所以需要匹配 ',100,' 或者开头的 '100,' 或者结尾的 ',100'
		query = query.Where(`dept_id = ? OR dept_id IN (
			SELECT dept_id FROM sys_dept
			WHERE ancestors = ?
			   OR ancestors LIKE ?
			   OR ancestors LIKE ?
			   OR ancestors LIKE ?
		)`,
			deptID,              // 直接匹配
			deptIDStr,           // ancestors = '100'
			deptIDStr+",%",      // ancestors LIKE '100,%'  (开头)
			"%,"+deptIDStr+",%", // ancestors LIKE '%,100,%' (中间)
			"%,"+deptIDStr)      // ancestors LIKE '%,100'  (结尾)
	}
	if user.Phonenumber != "" {
		query = query.Where("phonenumber LIKE ?", "%"+user.Phonenumber+"%")
	}

	// 处理数据权限 对应Java后端的${params.dataScope}
	if user.Params != nil {
		if dataScope, exists := user.Params["dataScope"]; exists && dataScope != "" {
			dataScopeSQL := fmt.Sprintf("%v", dataScope)
			if dataScopeSQL != "" {
				// 数据权限SQL已经包含AND前缀，需要去掉前缀直接拼接
				// 格式：" AND (d.dept_id = 0)" -> "d.dept_id = 0"
				cleanSQL := strings.TrimSpace(dataScopeSQL)
				if strings.HasPrefix(cleanSQL, "AND (") && strings.HasSuffix(cleanSQL, ")") {
					cleanSQL = cleanSQL[5 : len(cleanSQL)-1] // 去掉 "AND (" 和 ")"
				}
				// 需要将别名替换为实际字段名，因为这里不使用JOIN查询
				cleanSQL = strings.ReplaceAll(cleanSQL, "d.dept_id", "dept_id")
				cleanSQL = strings.ReplaceAll(cleanSQL, "u.user_id", "user_id")
				cleanSQL = strings.ReplaceAll(cleanSQL, "c.create_by", "create_by")
				if cleanSQL != "" {
					query = query.Where(cleanSQL)
				}
			}
		}
	}

	// 先查询总数
	err := query.Model(&model.SysUser{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询，使用Preload自动加载关联数据
	offset := (pageNum - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).
		Preload("Dept").
		Preload("Roles").
		Order("user_id ASC").
		Find(&users).Error

	if err != nil {
		return nil, 0, err
	}

	// 确保所有用户的status字段都有正确的值
	for i := range users {
		if users[i].Status == "" {
			users[i].Status = "0" // 默认为正常状态
		}
		fmt.Printf("UserDao.SelectUserList: 用户%d - ID=%d, UserName=%s, Status=%s\n",
			i+1, users[i].UserID, users[i].UserName, users[i].Status)
	}

	return users, total, nil
}

// InsertUser 新增用户 对应Java后端的insertUser
func (d *UserDao) InsertUser(user *model.SysUser) error {
	return d.db.Create(user).Error
}

// UpdateUser 修改用户 对应Java后端的updateUser
func (d *UserDao) UpdateUser(user *model.SysUser) error {
	fmt.Printf("UserDao.UpdateUser: 修改用户信息, UserID=%d\n", user.UserID)

	// 构建更新字段映射，只更新非空且有效的字段
	updates := make(map[string]interface{})

	// 基本信息字段 - 只有非空时才更新
	if user.UserName != "" {
		updates["user_name"] = user.UserName
	}
	if user.NickName != "" {
		updates["nick_name"] = user.NickName
	}
	if user.UserType != "" {
		updates["user_type"] = user.UserType
	}
	if user.Email != "" {
		updates["email"] = user.Email
	}
	if user.Phonenumber != "" {
		updates["phonenumber"] = user.Phonenumber
	}
	if user.Sex != "" {
		updates["sex"] = user.Sex
	}

	// 状态字段 - 只有明确指定时才更新（防止意外修改）
	if user.Status == "0" || user.Status == "1" {
		fmt.Printf("UserDao.UpdateUser: 更新用户状态, UserID=%d, Status=%s\n", user.UserID, user.Status)
		updates["status"] = user.Status
	}

	// 管理字段
	if user.UpdateBy != "" {
		updates["update_by"] = user.UpdateBy
	}
	if user.UpdateTime != nil {
		updates["update_time"] = user.UpdateTime
	}
	if user.Remark != "" {
		updates["remark"] = user.Remark
	}

	// 只有当dept_id不为nil时才更新
	if user.DeptID != nil {
		updates["dept_id"] = *user.DeptID
	}

	// 只有当密码不为空时才更新密码相关字段
	if user.Password != "" {
		updates["password"] = user.Password
		if user.PwdUpdateDate != nil {
			updates["pwd_update_date"] = user.PwdUpdateDate
		}
	}

	// 确保至少有一个字段需要更新
	if len(updates) == 0 {
		fmt.Printf("UserDao.UpdateUser: 没有字段需要更新, UserID=%d\n", user.UserID)
		return nil
	}

	fmt.Printf("UserDao.UpdateUser: 准备更新字段: %v\n", updates)

	err := d.db.Model(&model.SysUser{}).Where("user_id = ?", user.UserID).Updates(updates).Error
	if err != nil {
		fmt.Printf("UserDao.UpdateUser: 修改用户信息失败: %v\n", err)
		return fmt.Errorf("修改用户信息失败: %v", err)
	}

	fmt.Printf("UserDao.UpdateUser: 修改用户信息成功\n")
	return nil
}

// UpdateUserLoginInfo 更新用户登录信息 对应Java后端的updateUserLoginInfo
func (d *UserDao) UpdateUserLoginInfo(userId int64, loginIP string) error {
	fmt.Printf("UserDao.UpdateUserLoginInfo: 更新用户登录信息, UserID=%d, LoginIP=%s\n", userId, loginIP)

	// 只更新登录相关字段，确保不会影响其他字段
	err := d.db.Model(&model.SysUser{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"login_ip":   loginIP,
			"login_date": gorm.Expr("GETDATE()"),
		}).Error

	if err != nil {
		fmt.Printf("UserDao.UpdateUserLoginInfo: 更新用户登录信息失败: %v\n", err)
		return fmt.Errorf("更新用户登录信息失败: %v", err)
	}

	fmt.Printf("UserDao.UpdateUserLoginInfo: 更新用户登录信息成功\n")
	return nil
}

// DeleteUserById 删除用户 对应Java后端的deleteUserById
func (d *UserDao) DeleteUserById(userId int64) error {
	return d.db.Model(&model.SysUser{}).
		Where("user_id = ?", userId).
		Update("del_flag", "2").Error
}

// DeleteUserByIds 批量删除用户 对应Java后端的deleteUserByIds
func (d *UserDao) DeleteUserByIds(userIds []int64) error {
	return d.db.Model(&model.SysUser{}).
		Where("user_id IN ?", userIds).
		Update("del_flag", "2").Error
}

// CheckLoginNameUnique 校验用户名称是否唯一 对应Java后端的checkLoginNameUnique
func (d *UserDao) CheckLoginNameUnique(loginName string, userId int64) (bool, error) {
	var count int64
	query := d.db.Model(&model.SysUser{}).
		Where("user_name = ? AND del_flag = '0'", loginName)

	if userId > 0 {
		query = query.Where("user_id != ?", userId)
	}

	err := query.Count(&count).Error
	return count == 0, err
}

// CheckPhoneUnique 校验手机号码是否唯一 对应Java后端的checkPhoneUnique
func (d *UserDao) CheckPhoneUnique(phonenumber string, userId int64) (bool, error) {
	var count int64
	query := d.db.Model(&model.SysUser{}).
		Where("phonenumber = ? AND del_flag = '0'", phonenumber)

	if userId > 0 {
		query = query.Where("user_id != ?", userId)
	}

	err := query.Count(&count).Error
	return count == 0, err
}

// CheckEmailUnique 校验邮箱是否唯一 对应Java后端的checkEmailUnique
func (d *UserDao) CheckEmailUnique(email string, userId int64) (bool, error) {
	var count int64
	query := d.db.Model(&model.SysUser{}).
		Where("email = ? AND del_flag = '0'", email)

	if userId > 0 {
		query = query.Where("user_id != ?", userId)
	}

	err := query.Count(&count).Error
	return count == 0, err
}

// SelectUserRoleGroup 查询用户所属角色组 对应Java后端的selectUserRoleGroup
func (d *UserDao) SelectUserRoleGroup(userName string) (string, error) {
	fmt.Printf("UserDao.SelectUserRoleGroup: 查询用户角色组, UserName=%s\n", userName)

	var roleNames []string
	err := d.db.Table("sys_role r").
		Select("r.role_name").
		Joins("LEFT JOIN sys_user_role ur ON ur.role_id = r.role_id").
		Joins("LEFT JOIN sys_user u ON u.user_id = ur.user_id").
		Where("u.user_name = ?", userName).
		Pluck("role_name", &roleNames).Error

	if err != nil {
		return "", fmt.Errorf("查询用户角色组失败: %v", err)
	}

	return strings.Join(roleNames, ","), nil
}

// SelectUserPostGroup 查询用户所属岗位组 对应Java后端的selectUserPostGroup
func (d *UserDao) SelectUserPostGroup(userName string) (string, error) {
	fmt.Printf("UserDao.SelectUserPostGroup: 查询用户岗位组, UserName=%s\n", userName)

	var postNames []string
	err := d.db.Table("sys_post p").
		Select("p.post_name").
		Joins("LEFT JOIN sys_user_post up ON up.post_id = p.post_id").
		Joins("LEFT JOIN sys_user u ON u.user_id = up.user_id").
		Where("u.user_name = ?", userName).
		Pluck("post_name", &postNames).Error

	if err != nil {
		return "", fmt.Errorf("查询用户岗位组失败: %v", err)
	}

	return strings.Join(postNames, ","), nil
}

// UpdateUserProfile 修改用户基本信息 对应Java后端的updateUserProfile
func (d *UserDao) UpdateUserProfile(user *model.SysUser) error {
	fmt.Printf("UserDao.UpdateUserProfile: 修改用户基本信息, UserID=%d\n", user.UserID)

	// 只更新允许修改的字段
	updates := map[string]interface{}{
		"nick_name":   user.NickName,
		"email":       user.Email,
		"phonenumber": user.Phonenumber,
		"sex":         user.Sex,
		"update_time": time.Now(),
	}

	err := d.db.Model(&model.SysUser{}).Where("user_id = ?", user.UserID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("修改用户基本信息失败: %v", err)
	}

	fmt.Printf("UserDao.UpdateUserProfile: 修改用户基本信息成功\n")
	return nil
}

// ResetUserPwd 重置用户密码 对应Java后端的resetUserPwd
func (d *UserDao) ResetUserPwd(userId int64, password string) error {
	fmt.Printf("UserDao.ResetUserPwd: 重置用户密码, UserID=%d\n", userId)

	updates := map[string]interface{}{
		"password":        password,
		"pwd_update_date": time.Now(),
		"update_time":     time.Now(),
	}

	err := d.db.Model(&model.SysUser{}).Where("user_id = ?", userId).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("重置用户密码失败: %v", err)
	}

	fmt.Printf("UserDao.ResetUserPwd: 重置用户密码成功\n")
	return nil
}

// UpdateUserAvatar 更新用户头像 对应Java后端的updateUserAvatar
func (d *UserDao) UpdateUserAvatar(userId int64, avatar string) error {
	fmt.Printf("UserDao.UpdateUserAvatar: 更新用户头像, UserID=%d\n", userId)

	updates := map[string]interface{}{
		"avatar":      avatar,
		"update_time": time.Now(),
	}

	err := d.db.Model(&model.SysUser{}).Where("user_id = ?", userId).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("更新用户头像失败: %v", err)
	}

	fmt.Printf("UserDao.UpdateUserAvatar: 更新用户头像成功\n")
	return nil
}

// SelectAllocatedList 查询已分配用户角色列表 对应Java后端的selectAllocatedList
func (d *UserDao) SelectAllocatedList(currentUser *model.SysUser, user *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	roleId := *user.RoleID
	fmt.Printf("UserDao.SelectAllocatedList: 查询已分配用户角色列表, RoleID=%d\n", roleId)

	var users []model.SysUser
	var total int64

	// 对应Java后端的SQL：
	// select distinct u.user_id, u.dept_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time
	// from sys_user u
	//      left join sys_dept d on u.dept_id = d.dept_id
	//      left join sys_user_role ur on u.user_id = ur.user_id
	//      left join sys_role r on r.role_id = ur.role_id
	// where u.del_flag = '0' and r.role_id = #{roleId}
	query := d.db.Model(&model.SysUser{}).
		Select("DISTINCT u.user_id, u.dept_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time").
		Table("sys_user u").
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.dept_id").
		Joins("LEFT JOIN sys_user_role ur ON u.user_id = ur.user_id").
		Joins("LEFT JOIN sys_role r ON r.role_id = ur.role_id").
		Where("u.del_flag = '0' AND r.role_id = ?", roleId)

	// 添加查询条件 对应Java后端的if条件
	if user.UserName != "" {
		query = query.Where("u.user_name LIKE ?", "%"+user.UserName+"%")
	}
	if user.Phonenumber != "" {
		query = query.Where("u.phonenumber LIKE ?", "%"+user.Phonenumber+"%")
	}

	// 处理数据权限 对应Java后端的${params.dataScope}
	if user.Params != nil {
		if dataScope, exists := user.Params["dataScope"]; exists && dataScope != "" {
			dataScopeSQL := fmt.Sprintf("%v", dataScope)
			if dataScopeSQL != "" {
				// 数据权限SQL已经包含AND前缀，需要去掉前缀直接拼接
				// 格式：" AND (d.dept_id = 0)" -> "d.dept_id = 0"
				cleanSQL := strings.TrimSpace(dataScopeSQL)
				if strings.HasPrefix(cleanSQL, "AND (") && strings.HasSuffix(cleanSQL, ")") {
					cleanSQL = cleanSQL[5 : len(cleanSQL)-1] // 去掉 "AND (" 和 ")"
				}
				if cleanSQL != "" {
					query = query.Where(cleanSQL)
				}
			}
		}
	}

	// 查询总数
	err := query.Count(&total).Error
	if err != nil {
		fmt.Printf("UserDao.SelectAllocatedList: 查询总数失败: %v\n", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	err = query.Order("u.user_id").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		fmt.Printf("UserDao.SelectAllocatedList: 分页查询失败: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("UserDao.SelectAllocatedList: 查询成功, 总数=%d, 当前页数量=%d\n", total, len(users))
	return users, total, nil
}

// SelectUnallocatedList 查询未分配用户角色列表 对应Java后端的selectUnallocatedList
func (d *UserDao) SelectUnallocatedList(currentUser *model.SysUser, user *model.SysUser, pageNum, pageSize int) ([]model.SysUser, int64, error) {
	roleId := *user.RoleID
	fmt.Printf("UserDao.SelectUnallocatedList: 查询未分配用户角色列表, RoleID=%d\n", roleId)

	var users []model.SysUser
	var total int64

	// 对应Java后端的SQL：
	// select distinct u.user_id, u.dept_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time
	// from sys_user u
	//      left join sys_dept d on u.dept_id = d.dept_id
	//      left join sys_user_role ur on u.user_id = ur.user_id
	//      left join sys_role r on r.role_id = ur.role_id
	// where u.del_flag = '0' and (r.role_id != #{roleId} or r.role_id IS NULL)
	// and u.user_id not in (select u.user_id from sys_user u inner join sys_user_role ur on u.user_id = ur.user_id and ur.role_id = #{roleId})
	query := d.db.Model(&model.SysUser{}).
		Select("DISTINCT u.user_id, u.dept_id, u.user_name, u.nick_name, u.email, u.phonenumber, u.status, u.create_time").
		Table("sys_user u").
		Joins("LEFT JOIN sys_dept d ON u.dept_id = d.dept_id").
		Joins("LEFT JOIN sys_user_role ur ON u.user_id = ur.user_id").
		Joins("LEFT JOIN sys_role r ON r.role_id = ur.role_id").
		Where("u.del_flag = '0' AND (r.role_id != ? OR r.role_id IS NULL)", roleId).
		Where("u.user_id NOT IN (SELECT u.user_id FROM sys_user u INNER JOIN sys_user_role ur ON u.user_id = ur.user_id AND ur.role_id = ?)", roleId)

	// 添加查询条件 对应Java后端的if条件
	if user.UserName != "" {
		query = query.Where("u.user_name LIKE ?", "%"+user.UserName+"%")
	}
	if user.Phonenumber != "" {
		query = query.Where("u.phonenumber LIKE ?", "%"+user.Phonenumber+"%")
	}

	// 处理数据权限 对应Java后端的${params.dataScope}
	if user.Params != nil {
		if dataScope, exists := user.Params["dataScope"]; exists && dataScope != "" {
			dataScopeSQL := fmt.Sprintf("%v", dataScope)
			if dataScopeSQL != "" {
				// 数据权限SQL已经包含AND前缀，需要去掉前缀直接拼接
				// 格式：" AND (d.dept_id = 0)" -> "d.dept_id = 0"
				cleanSQL := strings.TrimSpace(dataScopeSQL)
				if strings.HasPrefix(cleanSQL, "AND (") && strings.HasSuffix(cleanSQL, ")") {
					cleanSQL = cleanSQL[5 : len(cleanSQL)-1] // 去掉 "AND (" 和 ")"
				}
				if cleanSQL != "" {
					query = query.Where(cleanSQL)
				}
			}
		}
	}

	// 查询总数
	err := query.Count(&total).Error
	if err != nil {
		fmt.Printf("UserDao.SelectUnallocatedList: 查询总数失败: %v\n", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNum - 1) * pageSize
	err = query.Order("u.user_id").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		fmt.Printf("UserDao.SelectUnallocatedList: 分页查询失败: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("UserDao.SelectUnallocatedList: 查询成功, 总数=%d, 当前页数量=%d\n", total, len(users))
	return users, total, nil
}

// loadRolePermissions 加载角色权限信息 对应Java后端的角色权限查询
func (d *UserDao) loadRolePermissions(user *model.SysUser) error {
	if user == nil || len(user.Roles) == 0 {
		return nil
	}

	// 为每个角色加载权限信息
	for i := range user.Roles {
		role := &user.Roles[i]

		// 查询角色的菜单权限
		var permissions []string
		result := d.db.Table("sys_menu m").
			Select("m.perms").
			Joins("LEFT JOIN sys_role_menu rm ON m.menu_id = rm.menu_id").
			Where("rm.role_id = ? AND m.perms IS NOT NULL AND m.perms != ''", role.RoleID).
			Pluck("perms", &permissions)

		if result.Error != nil {
			fmt.Printf("loadRolePermissions: 查询角色权限失败, RoleID=%d, Error=%v\n", role.RoleID, result.Error)
			continue
		}

		// 设置角色权限
		role.Permissions = permissions
		fmt.Printf("loadRolePermissions: 角色权限加载成功, RoleID=%d, PermissionCount=%d\n", role.RoleID, len(permissions))
	}

	return nil
}
