package dao

import (
	"fmt"
	"wosm/internal/repository/model"
	"wosm/pkg/database"

	"gorm.io/gorm"
)

// RoleMenuDao 角色菜单关联数据访问层 对应Java后端的SysRoleMenuMapper
type RoleMenuDao struct {
	db *gorm.DB
}

// NewRoleMenuDao 创建角色菜单关联数据访问层实例
func NewRoleMenuDao() *RoleMenuDao {
	return &RoleMenuDao{
		db: database.GetDB(),
	}
}

// SelectMenuListByRoleId 根据角色ID查询菜单ID列表 对应Java后端的selectMenuListByRoleId
func (d *RoleMenuDao) SelectMenuListByRoleId(roleId int64, menuCheckStrictly bool) ([]int64, error) {
	var menuIds []int64

	fmt.Printf("SelectMenuListByRoleId: 查询角色菜单, RoleID=%d, MenuCheckStrictly=%t\n", roleId, menuCheckStrictly)

	query := d.db.Table("sys_role_menu rm").
		Select("rm.menu_id").
		Where("rm.role_id = ?", roleId)

	// 如果菜单树选择项关联显示，排除父级菜单
	if menuCheckStrictly {
		query = query.Where("rm.menu_id not in (select m.parent_id from sys_menu m inner join sys_role_menu rm2 on m.menu_id = rm2.menu_id and rm2.role_id = ?)", roleId)
	}

	err := query.Pluck("menu_id", &menuIds).Error
	if err != nil {
		fmt.Printf("SelectMenuListByRoleId: 查询角色菜单关联失败: %v\n", err)
		return []int64{}, err
	}

	// 确保返回的是空数组而不是nil
	if menuIds == nil {
		menuIds = []int64{}
	}

	fmt.Printf("SelectMenuListByRoleId: 查询到菜单数量=%d\n", len(menuIds))
	return menuIds, nil
}

// DeleteRoleMenuByRoleId 删除角色菜单关联 对应Java后端的deleteRoleMenuByRoleId
func (d *RoleMenuDao) DeleteRoleMenuByRoleId(roleId int64) error {
	err := d.db.Where("role_id = ?", roleId).Delete(&model.SysRoleMenu{}).Error
	if err != nil {
		fmt.Printf("DeleteRoleMenuByRoleId: 删除角色菜单关联失败: %v\n", err)
		return err
	}

	fmt.Printf("DeleteRoleMenuByRoleId: 删除角色菜单关联成功, RoleID=%d\n", roleId)
	return nil
}

// BatchInsertRoleMenu 批量新增角色菜单关联 对应Java后端的batchRoleMenu
func (d *RoleMenuDao) BatchInsertRoleMenu(roleMenus []model.SysRoleMenu) error {
	if len(roleMenus) == 0 {
		return nil
	}

	err := d.db.Create(&roleMenus).Error
	if err != nil {
		fmt.Printf("BatchInsertRoleMenu: 批量新增角色菜单关联失败: %v\n", err)
		return err
	}

	fmt.Printf("BatchInsertRoleMenu: 批量新增角色菜单关联成功, 数量=%d\n", len(roleMenus))
	return nil
}

// CheckMenuExistRole 查询菜单使用数量 对应Java后端的checkMenuExistRole
func (d *RoleMenuDao) CheckMenuExistRole(menuId int64) (int64, error) {
	var count int64
	err := d.db.Model(&model.SysRoleMenu{}).Where("menu_id = ?", menuId).Count(&count).Error
	if err != nil {
		fmt.Printf("CheckMenuExistRole: 查询菜单使用数量失败: %v\n", err)
		return 0, err
	}

	return count, nil
}
