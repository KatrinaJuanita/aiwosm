package model

import (
	"time"
)

// SysDictData 字典数据表 对应Java后端的SysDictData实体
// 严格按照Java后端真实数据库表结构定义（基于ruoyi-java/sql/ry_20250522.sql）：
// dict_code, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, create_by, create_time, update_by, update_time, remark
type SysDictData struct {
	DictCode   int64      `gorm:"column:dict_code;primaryKey;autoIncrement" json:"dictCode" excel:"name:字典编码;sort:1;cellType:numeric"`   // 字典编码
	DictSort   int64      `gorm:"column:dict_sort;default:0" json:"dictSort" excel:"name:字典排序;sort:2;cellType:numeric"`                  // 字典排序
	DictLabel  string     `gorm:"column:dict_label;size:100;not null" json:"dictLabel" excel:"name:字典标签;sort:3"`                         // 字典标签
	DictValue  string     `gorm:"column:dict_value;size:100;not null" json:"dictValue" excel:"name:字典键值;sort:4"`                         // 字典键值
	DictType   string     `gorm:"column:dict_type;size:100;not null" json:"dictType" excel:"name:字典类型;sort:5"`                           // 字典类型
	CssClass   string     `gorm:"column:css_class;size:100" json:"cssClass"`                                                             // 样式属性（其他样式扩展）
	ListClass  string     `gorm:"column:list_class;size:100" json:"listClass"`                                                           // 表格回显样式
	IsDefault  string     `gorm:"column:is_default;size:1;default:N" json:"isDefault" excel:"name:是否默认;sort:6;readConverterExp:Y=是,N=否"` // 是否默认（Y是 N否）
	Status     string     `gorm:"column:status;size:1;default:0" json:"status" excel:"name:状态;sort:7;readConverterExp:0=正常,1=停用"`        // 状态（0正常 1停用）
	CreateBy   string     `gorm:"column:create_by;size:64" json:"createBy"`                                                              // 创建者
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime"`                                                                  // 创建时间
	UpdateBy   string     `gorm:"column:update_by;size:64" json:"updateBy"`                                                              // 更新者
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`                                                                  // 更新时间
	Remark     string     `gorm:"column:remark;size:500" json:"remark"`                                                                  // 备注
}

// TableName 指定表名
func (SysDictData) TableName() string {
	return "sys_dict_data"
}
