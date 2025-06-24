package model

import (
	"time"
)

// SysDictType 字典类型表 对应Java后端的SysDictType实体
// 严格按照Java后端真实数据库表结构定义（基于ruoyi-java/sql/ry_20250522.sql）：
// dict_id, dict_name, dict_type, status, create_by, create_time, update_by, update_time, remark
type SysDictType struct {
	DictID     int64      `gorm:"column:dict_id;primaryKey;autoIncrement" json:"dictId" excel:"name:字典主键;sort:1;cellType:numeric"`          // 字典主键
	DictName   string     `gorm:"column:dict_name;size:100;not null" json:"dictName" excel:"name:字典名称;sort:2"`                              // 字典名称
	DictType   string     `gorm:"column:dict_type;size:100;not null" json:"dictType" excel:"name:字典类型;sort:3"`                              // 字典类型
	Status     string     `gorm:"column:status;size:1;default:0" json:"status" excel:"name:状态;sort:4;readConverterExp:0=正常,1=停用"`           // 状态（0正常 1停用）
	CreateBy   string     `gorm:"column:create_by;size:64" json:"createBy"`                                                                 // 创建者
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime" excel:"name:创建时间;sort:5;type:export;dateFormat:yyyy-MM-dd HH:mm:ss"` // 创建时间
	UpdateBy   string     `gorm:"column:update_by;size:64" json:"updateBy"`                                                                 // 更新者
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`                                                                     // 更新时间
	Remark     string     `gorm:"column:remark;size:500" json:"remark"`                                                                     // 备注
}

// TableName 指定表名
func (SysDictType) TableName() string {
	return "sys_dict_type"
}
