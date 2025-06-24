package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// RoleDao 角色数据访问层 对应Java后端的SysRoleMapper
type RoleDao struct {
	db *gorm.DB
}

// NewRoleDao 创建角色数据访问层实例
func NewRoleDao() *RoleDao {
	return &RoleDao{
		db: database.GetDB(),
	}
}

// SelectRoleList 查询角色列表 对应Java后端的selectRoleList
func (d *RoleDao) SelectRoleList(role *model.SysRole) ([]model.SysRole, error) {
	var roles []model.SysRole
	query := d.db.Model(&model.SysRole{})

	// 构建查询条件 - 严格按照Java后端的查询逻辑
	if role.RoleID != 0 {
		query = query.Where("role_id = ?", role.RoleID)
	}
	if role.RoleName != "" {
		query = query.Where("role_name LIKE ?", "%"+role.RoleName+"%")
	}
	if role.Status != "" {
		query = query.Where("status = ?", role.Status)
	}
	if role.RoleKey != "" {
		query = query.Where("role_key LIKE ?", "%"+role.RoleKey+"%")
	}

	// 时间范围查询 对应Java后端的params.beginTime和params.endTime
	if role.Params != nil {
		if beginTime, ok := role.Params["beginTime"].(string); ok && beginTime != "" {
			query = query.Where("DATE_FORMAT(create_time,'%Y%m%d') >= DATE_FORMAT(?,'%Y%m%d')", beginTime)
		}
		if endTime, ok := role.Params["endTime"].(string); ok && endTime != "" {
			query = query.Where("DATE_FORMAT(create_time,'%Y%m%d') <= DATE_FORMAT(?,'%Y%m%d')", endTime)
		}

		// 数据权限过滤 对应Java后端的${params.dataScope}
		if dataScope, ok := role.Params["dataScope"].(string); ok && dataScope != "" {
			// 直接添加数据权限SQL条件
			query = query.Where(dataScope)
		}
	}

	// 添加默认条件：未删除的角色
	query = query.Where("del_flag = '0'")

	// 排序 - 按照Java后端的排序规则
	query = query.Order("role_sort, role_id")

	err := query.Find(&roles).Error
	if err != nil {
		fmt.Printf("SelectRoleList: 查询角色列表失败: %v\n", err)
		return nil, err
	}

	// 设置前端需要的虚拟字段
	for i := range roles {
		roles[i].Admin = roles[i].IsAdmin()
	}

	fmt.Printf("SelectRoleList: 查询到角色数量=%d\n", len(roles))
	return roles, nil
}

// SelectRoleById 根据角色ID查询角色信息 对应Java后端的selectRoleById
func (d *RoleDao) SelectRoleById(roleId int64) (*model.SysRole, error) {
	var role model.SysRole
	err := d.db.Where("role_id = ? AND del_flag = '0'", roleId).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		fmt.Printf("SelectRoleById: 查询角色失败: %v\n", err)
		return nil, err
	}

	// 设置前端需要的虚拟字段
	role.Admin = role.IsAdmin()

	return &role, nil
}

// SelectRoleAll 查询所有角色 对应Java后端的selectRoleAll
func (d *RoleDao) SelectRoleAll() ([]model.SysRole, error) {
	var roles []model.SysRole
	err := d.db.Where("del_flag = '0'").Order("role_sort, role_id").Find(&roles).Error
	if err != nil {
		fmt.Printf("SelectRoleAll: 查询所有角色失败: %v\n", err)
		return nil, err
	}

	// 设置前端需要的虚拟字段
	for i := range roles {
		roles[i].Admin = roles[i].IsAdmin()
	}

	fmt.Printf("SelectRoleAll: 查询到角色数量=%d\n", len(roles))
	return roles, nil
}

// CheckRoleNameUnique 校验角色名称是否唯一 对应Java后端的checkRoleNameUnique
func (d *RoleDao) CheckRoleNameUnique(roleName string, roleId int64) (bool, error) {
	var count int64
	query := d.db.Model(&model.SysRole{}).Where("role_name = ? AND del_flag = '0'", roleName)

	// 如果是更新操作，排除当前角色ID
	if roleId > 0 {
		query = query.Where("role_id != ?", roleId)
	}

	err := query.Count(&count).Error
	if err != nil {
		fmt.Printf("CheckRoleNameUnique: 校验角色名称失败: %v\n", err)
		return false, err
	}

	return count == 0, nil
}

// CheckRoleKeyUnique 校验角色权限字符串是否唯一 对应Java后端的checkRoleKeyUnique
func (d *RoleDao) CheckRoleKeyUnique(roleKey string, roleId int64) (bool, error) {
	var count int64
	query := d.db.Model(&model.SysRole{}).Where("role_key = ? AND del_flag = '0'", roleKey)

	// 如果是更新操作，排除当前角色ID
	if roleId > 0 {
		query = query.Where("role_id != ?", roleId)
	}

	err := query.Count(&count).Error
	if err != nil {
		fmt.Printf("CheckRoleKeyUnique: 校验角色权限字符串失败: %v\n", err)
		return false, err
	}

	return count == 0, nil
}

// InsertRole 新增角色 对应Java后端的insertRole
func (d *RoleDao) InsertRole(role *model.SysRole) error {
	// 基于Java后端真实数据库结构（ry_20250522.sql），包含menu_check_strictly和dept_check_strictly字段
	err := d.db.Create(role).Error
	if err != nil {
		fmt.Printf("InsertRole: 新增角色失败: %v\n", err)
		return err
	}

	fmt.Printf("InsertRole: 新增角色成功, RoleID=%d\n", role.RoleID)
	return nil
}

// UpdateRole 修改角色 对应Java后端的updateRole
func (d *RoleDao) UpdateRole(role *model.SysRole) error {
	// 基于Java后端真实数据库结构，包含menu_check_strictly和dept_check_strictly字段
	err := d.db.Model(&model.SysRole{}).Where("role_id = ?", role.RoleID).Updates(role).Error
	if err != nil {
		fmt.Printf("UpdateRole: 修改角色失败: %v\n", err)
		return err
	}

	fmt.Printf("UpdateRole: 修改角色成功, RoleID=%d\n", role.RoleID)
	return nil
}

// DeleteRoleById 删除角色 对应Java后端的deleteRoleById
func (d *RoleDao) DeleteRoleById(roleId int64) error {
	// 软删除：设置del_flag为'2'
	err := d.db.Model(&model.SysRole{}).Where("role_id = ?", roleId).Update("del_flag", "2").Error
	if err != nil {
		fmt.Printf("DeleteRoleById: 删除角色失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteRoleById: 删除角色成功, RoleID=%d\n", roleId)
	return nil
}

// DeleteRoleByIds 批量删除角色 对应Java后端的deleteRoleByIds
func (d *RoleDao) DeleteRoleByIds(roleIds []int64) error {
	// 软删除：设置del_flag为'2'
	err := d.db.Model(&model.SysRole{}).Where("role_id IN ?", roleIds).Update("del_flag", "2").Error
	if err != nil {
		fmt.Printf("DeleteRoleByIds: 批量删除角色失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteRoleByIds: 批量删除角色成功, 数量=%d\n", len(roleIds))
	return nil
}

// SelectRolesByIds 根据角色ID列表查询角色信息 对应Java后端的selectRolesByIds
func (d *RoleDao) SelectRolesByIds(roleIds []int64) ([]model.SysRole, error) {
	var roles []model.SysRole
	err := d.db.Where("role_id IN ? AND del_flag = '0'", roleIds).Find(&roles).Error
	if err != nil {
		fmt.Printf("SelectRolesByIds: 查询角色列表失败: %v\n", err)
		return nil, err
	}

	// 设置前端需要的虚拟字段
	for i := range roles {
		roles[i].Admin = roles[i].IsAdmin()
	}

	fmt.Printf("SelectRolesByIds: 查询到角色数量=%d\n", len(roles))
	return roles, nil
}

// SelectRolePermissionByUserId 根据用户ID查询角色权限 对应Java后端的selectRolePermissionByUserId
func (d *RoleDao) SelectRolePermissionByUserId(userId int64) ([]model.SysRole, error) {
	var roles []model.SysRole

	// 对应Java后端的SQL查询：
	// select distinct r.role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.menu_check_strictly, r.dept_check_strictly,
	// r.status, r.del_flag, r.create_time, r.remark
	// from sys_role r
	// left join sys_user_role ur on ur.role_id = r.role_id
	// left join sys_user u on u.user_id = ur.user_id
	// where u.user_id = ? and r.del_flag = '0'
	err := d.db.Table("sys_role r").
		Select("DISTINCT r.role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.menu_check_strictly, r.dept_check_strictly, r.status, r.del_flag, r.create_time, r.remark").
		Joins("LEFT JOIN sys_user_role ur ON ur.role_id = r.role_id").
		Joins("LEFT JOIN sys_user u ON u.user_id = ur.user_id").
		Where("u.user_id = ? AND r.del_flag = '0'", userId).
		Find(&roles).Error

	if err != nil {
		fmt.Printf("SelectRolePermissionByUserId: 查询用户角色权限失败: %v\n", err)
		return nil, err
	}

	// 设置前端需要的虚拟字段
	for i := range roles {
		roles[i].Admin = roles[i].IsAdmin()
	}

	fmt.Printf("SelectRolePermissionByUserId: 查询到用户角色权限数量=%d\n", len(roles))
	return roles, nil
}

// SelectRoleListWithPage 分页查询角色列表 对应Java后端的selectRoleList + PageHelper
func (d *RoleDao) SelectRoleListWithPage(role *model.SysRole, pageNum, pageSize int) ([]model.SysRole, int64, error) {
	var roles []model.SysRole
	var total int64

	// 构建查询条件 - 严格按照Java后端的查询逻辑
	query := d.db.Model(&model.SysRole{})

	// 基本查询条件
	if role.RoleID != 0 {
		query = query.Where("role_id = ?", role.RoleID)
	}
	if role.RoleName != "" {
		query = query.Where("role_name LIKE ?", "%"+role.RoleName+"%")
	}
	if role.Status != "" {
		query = query.Where("status = ?", role.Status)
	}
	if role.RoleKey != "" {
		query = query.Where("role_key LIKE ?", "%"+role.RoleKey+"%")
	}

	// 时间范围查询 对应Java后端的params.beginTime和params.endTime
	if role.Params != nil {
		if beginTime, ok := role.Params["beginTime"].(string); ok && beginTime != "" {
			query = query.Where("DATE_FORMAT(create_time,'%Y%m%d') >= DATE_FORMAT(?,'%Y%m%d')", beginTime)
		}
		if endTime, ok := role.Params["endTime"].(string); ok && endTime != "" {
			query = query.Where("DATE_FORMAT(create_time,'%Y%m%d') <= DATE_FORMAT(?,'%Y%m%d')", endTime)
		}

		// 数据权限过滤 对应Java后端的${params.dataScope}
		if dataScope, ok := role.Params["dataScope"].(string); ok && dataScope != "" {
			// 直接添加数据权限SQL条件
			query = query.Where(dataScope)
		}
	}

	// 添加默认条件：未删除的角色
	query = query.Where("del_flag = '0'")

	// 先查询总数
	err := query.Count(&total).Error
	if err != nil {
		fmt.Printf("SelectRoleListWithPage: 查询角色总数失败: %v\n", err)
		return nil, 0, err
	}

	// 计算偏移量
	offset := (pageNum - 1) * pageSize

	// 分页查询数据
	err = query.Order("role_sort, role_id").Offset(offset).Limit(pageSize).Find(&roles).Error
	if err != nil {
		fmt.Printf("SelectRoleListWithPage: 分页查询角色列表失败: %v\n", err)
		return nil, 0, err
	}

	// 设置前端需要的虚拟字段
	for i := range roles {
		roles[i].Admin = roles[i].IsAdmin()
	}

	fmt.Printf("SelectRoleListWithPage: 查询到角色数量=%d, 总数=%d\n", len(roles), total)
	return roles, total, nil
}

// SelectRolesByUserName 根据用户名查询角色列表 对应Java后端的selectRolesByUserName
func (d *RoleDao) SelectRolesByUserName(userName string) ([]model.SysRole, error) {
	var roles []model.SysRole

	// 对应Java后端的SQL查询：
	// select distinct r.role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.menu_check_strictly, r.dept_check_strictly,
	// r.status, r.del_flag, r.create_time, r.remark
	// from sys_role r
	// left join sys_user_role ur on ur.role_id = r.role_id
	// left join sys_user u on u.user_id = ur.user_id
	// where r.del_flag = '0' and u.user_name = ?
	err := d.db.Table("sys_role r").
		Select("DISTINCT r.role_id, r.role_name, r.role_key, r.role_sort, r.data_scope, r.menu_check_strictly, r.dept_check_strictly, r.status, r.del_flag, r.create_time, r.remark").
		Joins("LEFT JOIN sys_user_role ur ON ur.role_id = r.role_id").
		Joins("LEFT JOIN sys_user u ON u.user_id = ur.user_id").
		Where("r.del_flag = '0' AND u.user_name = ?", userName).
		Find(&roles).Error

	if err != nil {
		fmt.Printf("SelectRolesByUserName: 根据用户名查询角色失败: %v\n", err)
		return nil, err
	}

	// 设置前端需要的虚拟字段
	for i := range roles {
		roles[i].Admin = roles[i].IsAdmin()
	}

	fmt.Printf("SelectRolesByUserName: 根据用户名查询到角色数量=%d\n", len(roles))
	return roles, nil
}
