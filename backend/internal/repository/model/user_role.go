package model

// SysUserRole 用户和角色关联表 对应Java后端的SysUserRole实体
// 严格按照Java后端真实数据库表结构定义（基于ruoyi-java/sql/ry_20250522.sql）：
// user_id, role_id
type SysUserRole struct {
	UserID int64 `gorm:"column:user_id;primaryKey" json:"userId"` // 用户ID
	RoleID int64 `gorm:"column:role_id;primaryKey" json:"roleId"` // 角色ID
}

// TableName 设置表名
func (SysUserRole) TableName() string {
	return "sys_user_role"
}
