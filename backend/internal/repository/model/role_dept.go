package model

// SysRoleDept 角色和部门关联表 对应Java后端的SysRoleDept实体
// 严格按照Java后端真实数据库表结构定义（基于ruoyi-java/sql/ry_20250522.sql）：
// role_id, dept_id
type SysRoleDept struct {
	RoleID int64 `gorm:"column:role_id;primaryKey" json:"roleId"` // 角色ID
	DeptID int64 `gorm:"column:dept_id;primaryKey" json:"deptId"` // 部门ID
}

// TableName 设置表名
func (SysRoleDept) TableName() string {
	return "sys_role_dept"
}
