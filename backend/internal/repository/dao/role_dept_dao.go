package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// RoleDeptDao 角色部门关联数据访问层 对应Java后端的SysRoleDeptMapper
type RoleDeptDao struct {
	db *gorm.DB
}

// NewRoleDeptDao 创建角色部门关联数据访问层实例
func NewRoleDeptDao() *RoleDeptDao {
	return &RoleDeptDao{
		db: database.GetDB(),
	}
}

// SelectDeptListByRoleId 根据角色ID查询部门ID列表 对应Java后端的selectDeptListByRoleId
func (d *RoleDeptDao) SelectDeptListByRoleId(roleId int64) ([]int64, error) {
	var roleDepts []model.SysRoleDept
	err := d.db.Where("role_id = ?", roleId).Find(&roleDepts).Error
	if err != nil {
		fmt.Printf("SelectDeptListByRoleId: 查询角色部门关联失败: %v\n", err)
		return nil, err
	}

	deptIds := make([]int64, len(roleDepts))
	for i, rd := range roleDepts {
		deptIds[i] = rd.DeptID
	}

	fmt.Printf("SelectDeptListByRoleId: 查询到部门数量=%d\n", len(deptIds))
	return deptIds, nil
}

// DeleteRoleDeptByRoleId 删除角色部门关联 对应Java后端的deleteRoleDeptByRoleId
func (d *RoleDeptDao) DeleteRoleDeptByRoleId(roleId int64) error {
	err := d.db.Where("role_id = ?", roleId).Delete(&model.SysRoleDept{}).Error
	if err != nil {
		fmt.Printf("DeleteRoleDeptByRoleId: 删除角色部门关联失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteRoleDeptByRoleId: 删除角色部门关联成功, RoleID=%d\n", roleId)
	return nil
}

// BatchInsertRoleDept 批量新增角色部门关联 对应Java后端的batchRoleDept
func (d *RoleDeptDao) BatchInsertRoleDept(roleDepts []model.SysRoleDept) error {
	if len(roleDepts) == 0 {
		return nil
	}

	err := d.db.Create(&roleDepts).Error
	if err != nil {
		fmt.Printf("BatchInsertRoleDept: 批量新增角色部门关联失败: %v\n", err)
		return err
	}

	fmt.Printf("BatchInsertRoleDept: 批量新增角色部门关联成功, 数量=%d\n", len(roleDepts))
	return nil
}

// DeleteRoleDept 删除角色和部门关联信息 对应Java后端的deleteRoleDept
func (d *RoleDeptDao) DeleteRoleDept(roleIds []int64) error {
	err := d.db.Where("role_id IN ?", roleIds).Delete(&model.SysRoleDept{}).Error
	if err != nil {
		fmt.Printf("DeleteRoleDept: 删除角色部门关联失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteRoleDept: 删除角色部门关联成功, RoleIDs=%v\n", roleIds)
	return nil
}
