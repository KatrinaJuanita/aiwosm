package model

import (
	"time"
)

// SysDept 部门表 对应Java后端的SysDept实体
type SysDept struct {
	DeptID     int64      `gorm:"column:dept_id;primaryKey;autoIncrement" json:"deptId"`                             // 部门id
	ParentID   int64      `gorm:"column:parent_id;default:0" json:"parentId"`                                        // 父部门id
	Ancestors  string     `gorm:"column:ancestors;size:50" json:"ancestors"`                                         // 祖级列表
	DeptName   string     `gorm:"column:dept_name;size:30;not null" json:"deptName" binding:"required,min=1,max=30"` // 部门名称 对应Java后端@NotBlank @Size(max=30)
	OrderNum   int        `gorm:"column:order_num;default:0" json:"orderNum" binding:"required"`                     // 显示顺序 对应Java后端@NotNull
	Leader     string     `gorm:"column:leader;size:20" json:"leader"`                                               // 负责人
	Phone      string     `gorm:"column:phone;size:11" json:"phone" binding:"omitempty,max=11"`                      // 联系电话 对应Java后端@Size(max=11)
	Email      string     `gorm:"column:email;size:50" json:"email" binding:"omitempty,email,max=50"`                // 邮箱 对应Java后端@Email @Size(max=50)
	Status     string     `gorm:"column:status;size:1;default:0" json:"status"`                                      // 部门状态（0正常 1停用）
	DelFlag    string     `gorm:"column:del_flag;size:1;default:0" json:"delFlag"`                                   // 删除标志（0代表存在 2代表删除）
	CreateBy   string     `gorm:"column:create_by;size:64" json:"createBy"`                                          // 创建者
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime"`                                              // 创建时间
	UpdateBy   string     `gorm:"column:update_by;size:64" json:"updateBy"`                                          // 更新者
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`                                              // 更新时间

	// 关联字段（不存储在数据库中）
	ParentName string                 `gorm:"-" json:"parentName,omitempty"` // 父部门名称（不存储在数据库中，用于显示）
	Children   []SysDept              `gorm:"-" json:"children,omitempty"`   // 子部门
	Params     map[string]interface{} `gorm:"-" json:"-"`                    // 请求参数（用于数据权限等）
}

// TableName 指定表名
func (SysDept) TableName() string {
	return "sys_dept"
}

// SysDeptExport 部门导出结构体 对应Java后端的Excel导出格式
type SysDeptExport struct {
	DeptID     int64  `excel:"name:部门编号;sort:1"`
	ParentID   int64  `excel:"name:上级部门;sort:2"`
	DeptName   string `excel:"name:部门名称;sort:3"`
	OrderNum   int    `excel:"name:显示顺序;sort:4"`
	Leader     string `excel:"name:负责人;sort:5"`
	Phone      string `excel:"name:联系电话;sort:6"`
	Email      string `excel:"name:邮箱;sort:7"`
	Status     string `excel:"name:部门状态;sort:8;readConverterExp:0=正常,1=停用"`
	CreateTime string `excel:"name:创建时间;sort:9"`
}
