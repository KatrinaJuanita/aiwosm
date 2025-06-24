package model

import (
	"time"
)

// SysRole 角色信息表 对应Java后端的SysRole实体
//
// 业务说明：
// 角色是权限管理的核心概念，用于将用户与权限进行关联。每个角色可以分配不同的菜单权限和数据权限。
// 系统支持多种数据权限范围：全部数据、自定义数据、本部门数据、本部门及以下数据、仅本人数据。
//
// 数据库表结构（基于ruoyi-java/sql/ry_20250522.sql）：
// - role_id: 角色主键，自增
// - role_name: 角色名称，如"管理员"、"普通用户"
// - role_key: 角色权限字符，如"admin"、"common"，用于代码中的权限判断
// - role_sort: 显示顺序，数字越小越靠前
// - data_scope: 数据权限范围（1-5）
// - menu_check_strictly: 菜单树选择时是否父子关联
// - dept_check_strictly: 部门树选择时是否父子关联
// - status: 角色状态（0正常 1停用）
// - del_flag: 删除标志（0存在 2删除）
// - 基础字段: create_by, create_time, update_by, update_time, remark
//
// 使用示例：
//
//	role := &SysRole{
//	    RoleName: "测试角色",
//	    RoleKey:  "test",
//	    RoleSort: 3,
//	    DataScope: "2", // 自定义数据权限
//	    Status:   "0",  // 正常状态
//	}
type SysRole struct {
	RoleID            int64      `gorm:"column:role_id;primaryKey;autoIncrement" json:"roleId" excel:"name:角色序号;sort:1;cellType:numeric"`                                                   // 角色ID
	RoleName          string     `gorm:"column:role_name;size:30;not null" json:"roleName" excel:"name:角色名称;sort:2" validate:"required,role_name"`                                          // 角色名称
	RoleKey           string     `gorm:"column:role_key;size:100;not null" json:"roleKey" excel:"name:角色权限;sort:3" validate:"required,role_key"`                                            // 角色权限字符串
	RoleSort          int        `gorm:"column:role_sort;not null" json:"roleSort" excel:"name:角色排序;sort:4" validate:"required,min=0"`                                                      // 显示顺序
	DataScope         string     `gorm:"column:data_scope;size:1;default:1" json:"dataScope" excel:"name:数据范围;sort:5;readConverterExp:1=所有数据权限,2=自定义数据权限,3=本部门数据权限,4=本部门及以下数据权限,5=仅本人数据权限"` // 数据范围（1：全部数据权限 2：自定数据权限 3：本部门数据权限 4：本部门及以下数据权限 5：仅本人数据权限）
	MenuCheckStrictly bool       `gorm:"column:menu_check_strictly;type:bit;default:1" json:"menuCheckStrictly"`                                                                            // 菜单树选择项是否关联显示（0：父子不互相关联显示 1：父子互相关联显示）
	DeptCheckStrictly bool       `gorm:"column:dept_check_strictly;type:bit;default:1" json:"deptCheckStrictly"`                                                                            // 部门树选择项是否关联显示（0：父子不互相关联显示 1：父子互相关联显示）
	Status            string     `gorm:"column:status;size:1;not null" json:"status" excel:"name:角色状态;sort:6;readConverterExp:0=正常,1=停用"`                                                   // 角色状态（0正常 1停用）
	DelFlag           string     `gorm:"column:del_flag;size:1;default:0" json:"delFlag"`                                                                                                   // 删除标志（0代表存在 2代表删除）
	CreateBy          string     `gorm:"column:create_by;size:64" json:"createBy"`                                                                                                          // 创建者
	CreateTime        *time.Time `gorm:"column:create_time" json:"createTime"`                                                                                                              // 创建时间
	UpdateBy          string     `gorm:"column:update_by;size:64" json:"updateBy"`                                                                                                          // 更新者
	UpdateTime        *time.Time `gorm:"column:update_time" json:"updateTime"`                                                                                                              // 更新时间
	Remark            string     `gorm:"column:remark;size:500" json:"remark"`                                                                                                              // 备注

	// 关联字段（不存储在数据库中）
	MenuIDs     []int64                `gorm:"-" json:"menuIds,omitempty"`     // 菜单组
	DeptIDs     []int64                `gorm:"-" json:"deptIds,omitempty"`     // 部门组（数据权限）
	Permissions []string               `gorm:"-" json:"permissions,omitempty"` // 权限列表
	Flag        bool                   `gorm:"-" json:"flag"`                  // 用户是否存在此角色标识 默认不存在
	Admin       bool                   `gorm:"-" json:"admin"`                 // 是否为管理员角色
	Params      map[string]interface{} `gorm:"-" json:"-"`                     // 请求参数（用于数据权限等）
}

// TableName 设置表名
func (SysRole) TableName() string {
	return "sys_role"
}

// IsAdmin 判断是否为管理员角色 对应Java后端的isAdmin方法
func (r *SysRole) IsAdmin() bool {
	return r.RoleID == 1
}
