package model

// SysRoleMenu 角色和菜单关联表 对应Java后端的SysRoleMenu实体
// 严格按照Java后端真实数据库表结构定义（基于ruoyi-java/sql/ry_20250522.sql）：
// role_id, menu_id
type SysRoleMenu struct {
	RoleID int64 `gorm:"column:role_id;primaryKey" json:"roleId"` // 角色ID
	MenuID int64 `gorm:"column:menu_id;primaryKey" json:"menuId"` // 菜单ID
}

// TableName 设置表名
func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}
