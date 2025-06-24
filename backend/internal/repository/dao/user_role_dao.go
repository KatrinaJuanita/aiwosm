package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// UserRoleDao 用户角色关联数据访问层 对应Java后端的SysUserRoleMapper
type UserRoleDao struct {
	db *gorm.DB
}

// NewUserRoleDao 创建用户角色关联数据访问层实例
func NewUserRoleDao() *UserRoleDao {
	return &UserRoleDao{
		db: database.GetDB(),
	}
}

// SelectRoleListByUserId 根据用户ID查询角色ID列表 对应Java后端的selectRoleListByUserId
func (d *UserRoleDao) SelectRoleListByUserId(userId int64) ([]int64, error) {
	var userRoles []model.SysUserRole
	err := d.db.Where("user_id = ?", userId).Find(&userRoles).Error
	if err != nil {
		fmt.Printf("SelectRoleListByUserId: 查询用户角色关联失败: %v\n", err)
		return nil, err
	}

	roleIds := make([]int64, len(userRoles))
	for i, ur := range userRoles {
		roleIds[i] = ur.RoleID
	}

	fmt.Printf("SelectRoleListByUserId: 查询到角色数量=%d\n", len(roleIds))
	return roleIds, nil
}

// DeleteUserRoleByUserId 删除用户角色关联 对应Java后端的deleteUserRoleByUserId
func (d *UserRoleDao) DeleteUserRoleByUserId(userId int64) error {
	err := d.db.Where("user_id = ?", userId).Delete(&model.SysUserRole{}).Error
	if err != nil {
		fmt.Printf("DeleteUserRoleByUserId: 删除用户角色关联失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteUserRoleByUserId: 删除用户角色关联成功, UserID=%d\n", userId)
	return nil
}

// BatchInsertUserRole 批量新增用户角色关联 对应Java后端的batchUserRole
func (d *UserRoleDao) BatchInsertUserRole(userRoles []model.SysUserRole) error {
	if len(userRoles) == 0 {
		return nil
	}

	err := d.db.Create(&userRoles).Error
	if err != nil {
		fmt.Printf("BatchInsertUserRole: 批量新增用户角色关联失败: %v\n", err)
		return err
	}

	fmt.Printf("BatchInsertUserRole: 批量新增用户角色关联成功, 数量=%d\n", len(userRoles))
	return nil
}

// CountUserRoleByRoleId 查询角色使用数量 对应Java后端的countUserRoleByRoleId
func (d *UserRoleDao) CountUserRoleByRoleId(roleId int64) (int64, error) {
	var count int64
	// 只统计有效用户（未删除的用户）的角色关联
	err := d.db.Table("sys_user_role ur").
		Joins("INNER JOIN sys_user u ON ur.user_id = u.user_id").
		Where("ur.role_id = ? AND u.del_flag = '0'", roleId).
		Count(&count).Error
	if err != nil {
		fmt.Printf("CountUserRoleByRoleId: 查询角色使用数量失败: %v\n", err)
		return 0, err
	}

	fmt.Printf("CountUserRoleByRoleId: 角色使用数量=%d, RoleID=%d\n", count, roleId)
	return count, nil
}

// DeleteUserRole 删除用户和角色关联信息 对应Java后端的deleteUserRole
func (d *UserRoleDao) DeleteUserRole(userIds []int64) error {
	err := d.db.Where("user_id IN ?", userIds).Delete(&model.SysUserRole{}).Error
	if err != nil {
		fmt.Printf("DeleteUserRole: 删除用户角色关联失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteUserRole: 删除用户角色关联成功, UserIDs=%v\n", userIds)
	return nil
}

// DeleteAuthUser 取消授权用户角色 对应Java后端的deleteAuthUser
func (d *UserRoleDao) DeleteAuthUser(userRole *model.SysUserRole) error {
	fmt.Printf("UserRoleDao.DeleteAuthUser: 开始取消授权, UserID=%d, RoleID=%d\n", userRole.UserID, userRole.RoleID)

	// 对应Java后端的SQL: delete from sys_user_role where user_id=#{userId} and role_id=#{roleId}
	result := d.db.Where("user_id = ? AND role_id = ?", userRole.UserID, userRole.RoleID).Delete(&model.SysUserRole{})
	if result.Error != nil {
		fmt.Printf("UserRoleDao.DeleteAuthUser: 取消授权用户角色失败: %v\n", result.Error)
		return result.Error
	}

	fmt.Printf("UserRoleDao.DeleteAuthUser: 取消授权用户角色成功, UserID=%d, RoleID=%d, 影响行数=%d\n",
		userRole.UserID, userRole.RoleID, result.RowsAffected)

	// 检查是否真的删除了记录
	if result.RowsAffected == 0 {
		fmt.Printf("UserRoleDao.DeleteAuthUser: 警告 - 没有找到要删除的记录, UserID=%d, RoleID=%d\n",
			userRole.UserID, userRole.RoleID)
		return fmt.Errorf("未找到要取消的授权关系")
	}

	return nil
}

// DeleteAuthUsers 批量取消授权用户角色 对应Java后端的deleteAuthUsers
func (d *UserRoleDao) DeleteAuthUsers(roleId int64, userIds []int64) error {
	err := d.db.Where("role_id = ? AND user_id IN ?", roleId, userIds).Delete(&model.SysUserRole{}).Error
	if err != nil {
		fmt.Printf("DeleteAuthUsers: 批量取消授权用户角色失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteAuthUsers: 批量取消授权用户角色成功, RoleID=%d, UserIDs=%v\n", roleId, userIds)
	return nil
}

// SelectAuthUserAll 批量选择授权用户角色 对应Java后端的selectAuthUserAll
func (d *UserRoleDao) SelectAuthUserAll(roleId int64, userIds []int64) error {
	userRoles := make([]model.SysUserRole, len(userIds))
	for i, userId := range userIds {
		userRoles[i] = model.SysUserRole{
			UserID: userId,
			RoleID: roleId,
		}
	}

	return d.BatchInsertUserRole(userRoles)
}
