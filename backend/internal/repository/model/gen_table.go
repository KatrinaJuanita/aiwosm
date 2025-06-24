package model

import (
	"time"
)

// GenTable 代码生成业务表 对应Java后端的GenTable实体
type GenTable struct {
	TableID        int64      `gorm:"column:table_id;primaryKey;autoIncrement" json:"tableId" binding:"required"` // 编号
	Name           string     `gorm:"column:table_name" json:"tableName" binding:"required"`                      // 表名称
	TableComment   string     `gorm:"column:table_comment" json:"tableComment" binding:"required"`                // 表描述
	SubTableName   string     `gorm:"column:sub_table_name" json:"subTableName"`                                  // 关联子表的表名
	SubTableFkName string     `gorm:"column:sub_table_fk_name" json:"subTableFkName"`                             // 子表关联的外键名
	ClassName      string     `gorm:"column:class_name" json:"className" binding:"required"`                      // 实体类名称
	TplCategory    string     `gorm:"column:tpl_category" json:"tplCategory"`                                     // 使用的模板（crud单表操作 tree树表操作）
	TplWebType     string     `gorm:"column:tpl_web_type" json:"tplWebType"`                                      // 前端模板类型（element-ui模版）
	PackageName    string     `gorm:"column:package_name" json:"packageName" binding:"required"`                  // 生成包路径
	ModuleName     string     `gorm:"column:module_name" json:"moduleName" binding:"required"`                    // 生成模块名
	BusinessName   string     `gorm:"column:business_name" json:"businessName" binding:"required"`                // 生成业务名
	FunctionName   string     `gorm:"column:function_name" json:"functionName" binding:"required"`                // 生成功能名
	FunctionAuthor string     `gorm:"column:function_author" json:"functionAuthor" binding:"required"`            // 生成功能作者
	GenType        string     `gorm:"column:gen_type" json:"genType"`                                             // 生成代码方式（0zip压缩包 1自定义路径）
	GenPath        string     `gorm:"column:gen_path" json:"genPath"`                                             // 生成路径（不填默认项目路径）
	Options        string     `gorm:"column:options" json:"options"`                                              // 其它生成选项
	CreateBy       string     `gorm:"column:create_by" json:"createBy"`                                           // 创建者
	CreateTime     *time.Time `gorm:"column:create_time" json:"createTime"`                                       // 创建时间
	UpdateBy       string     `gorm:"column:update_by" json:"updateBy"`                                           // 更新者
	UpdateTime     *time.Time `gorm:"column:update_time" json:"updateTime"`                                       // 更新时间
	Remark         string     `gorm:"column:remark" json:"remark"`                                                // 备注

	// 树表相关字段（对应Java后端的树表功能）
	TreeCode       string `gorm:"-" json:"treeCode"`       // 树编码字段
	TreeParentCode string `gorm:"-" json:"treeParentCode"` // 树父编码字段
	TreeName       string `gorm:"-" json:"treeName"`       // 树名称字段
	ParentMenuId   int64  `gorm:"-" json:"parentMenuId"`   // 上级菜单ID字段
	ParentMenuName string `gorm:"-" json:"parentMenuName"` // 上级菜单名称字段

	// 扩展字段（不映射到数据库）
	Columns    []GenTableColumn `gorm:"-" json:"columns" binding:"dive"` // 表列信息
	PkColumn   *GenTableColumn  `gorm:"-" json:"pkColumn"`               // 主键信息
	SubTable   *GenTable        `gorm:"-" json:"subTable"`               // 子表信息
	ParentMenu *SysMenu         `gorm:"-" json:"parentMenu"`             // 上级菜单
}

// TableName 设置表名
func (GenTable) TableName() string {
	return "gen_table"
}

// GenTableColumn 代码生成业务表字段 对应Java后端的GenTableColumn实体
type GenTableColumn struct {
	ColumnID      int64      `gorm:"column:column_id;primaryKey;autoIncrement" json:"columnId"` // 编号
	TableID       int64      `gorm:"column:table_id" json:"tableId"`                            // 归属表编号
	ColumnName    string     `gorm:"column:column_name" json:"columnName"`                      // 列名称
	ColumnComment string     `gorm:"column:column_comment" json:"columnComment"`                // 列描述
	ColumnType    string     `gorm:"column:column_type" json:"columnType"`                      // 列类型
	JavaType      string     `gorm:"column:java_type" json:"javaType"`                          // JAVA类型
	JavaField     string     `gorm:"column:java_field" json:"javaField" binding:"required"`     // JAVA字段名
	IsPk          string     `gorm:"column:is_pk" json:"isPk"`                                  // 是否主键（1是）
	IsIncrement   string     `gorm:"column:is_increment" json:"isIncrement"`                    // 是否自增（1是）
	IsRequired    string     `gorm:"column:is_required" json:"isRequired"`                      // 是否必填（1是）
	IsInsert      string     `gorm:"column:is_insert" json:"isInsert"`                          // 是否为插入字段（1是）
	IsEdit        string     `gorm:"column:is_edit" json:"isEdit"`                              // 是否编辑字段（1是）
	IsList        string     `gorm:"column:is_list" json:"isList"`                              // 是否列表字段（1是）
	IsQuery       string     `gorm:"column:is_query" json:"isQuery"`                            // 是否查询字段（1是）
	QueryType     string     `gorm:"column:query_type" json:"queryType"`                        // 查询方式（等于、不等于、大于、小于、范围）
	HtmlType      string     `gorm:"column:html_type" json:"htmlType"`                          // 显示类型（文本框、文本域、下拉框、复选框、单选框、日期控件）
	DictType      string     `gorm:"column:dict_type" json:"dictType"`                          // 字典类型
	Sort          int        `gorm:"column:sort" json:"sort"`                                   // 排序
	CreateBy      string     `gorm:"column:create_by" json:"createBy"`                          // 创建者
	CreateTime    *time.Time `gorm:"column:create_time" json:"createTime"`                      // 创建时间
	UpdateBy      string     `gorm:"column:update_by" json:"updateBy"`                          // 更新者
	UpdateTime    *time.Time `gorm:"column:update_time" json:"updateTime"`                      // 更新时间
}

// TableName 设置表名
func (GenTableColumn) TableName() string {
	return "gen_table_column"
}

// DbTable 数据库表信息
type DbTable struct {
	TableName    string `json:"tableName"`    // 表名称
	TableComment string `json:"tableComment"` // 表描述
	CreateTime   string `json:"createTime"`   // 创建时间
	UpdateTime   string `json:"updateTime"`   // 更新时间
}

// DbTableColumn 数据库表字段信息
type DbTableColumn struct {
	ColumnName    string `json:"columnName"`    // 列名称
	ColumnComment string `json:"columnComment"` // 列描述
	ColumnType    string `json:"columnType"`    // 列类型
	DataType      string `json:"dataType"`      // 数据类型
	ColumnKey     string `json:"columnKey"`     // 主键类型
	Extra         string `json:"extra"`         // 额外参数
}

// 模板类型常量
const (
	TplCategoryCrud = "crud" // 单表（增删改查）
	TplCategoryTree = "tree" // 树表（增删改查）
	TplCategorySub  = "sub"  // 主子表（增删改查）
)

// 前端模板类型常量
const (
	TplWebTypeElementUI   = "element-ui"   // Element UI模版
	TplWebTypeElementPlus = "element-plus" // Element Plus模版
)

// 生成代码方式常量
const (
	GenTypeZip  = "0" // zip压缩包
	GenTypePath = "1" // 自定义路径
)

// 查询方式常量
const (
	QueryTypeEQ      = "EQ"      // 等于
	QueryTypeNE      = "NE"      // 不等于
	QueryTypeGT      = "GT"      // 大于
	QueryTypeGTE     = "GTE"     // 大于等于
	QueryTypeLT      = "LT"      // 小于
	QueryTypeLTE     = "LTE"     // 小于等于
	QueryTypeLike    = "LIKE"    // 模糊
	QueryTypeBetween = "BETWEEN" // 范围
)

// 显示类型常量
const (
	HtmlTypeInput       = "input"       // 文本框
	HtmlTypeTextarea    = "textarea"    // 文本域
	HtmlTypeSelect      = "select"      // 下拉框
	HtmlTypeRadio       = "radio"       // 单选框
	HtmlTypeCheckbox    = "checkbox"    // 复选框
	HtmlTypeDatetime    = "datetime"    // 日期控件
	HtmlTypeImageUpload = "imageUpload" // 图片上传
	HtmlTypeFileUpload  = "fileUpload"  // 文件上传
	HtmlTypeEditor      = "editor"      // 富文本控件
)

// Java类型常量
const (
	JavaTypeLong       = "Long"
	JavaTypeString     = "String"
	JavaTypeInteger    = "Integer"
	JavaTypeDouble     = "Double"
	JavaTypeBigDecimal = "BigDecimal"
	JavaTypeDate       = "Date"
	JavaTypeBoolean    = "Boolean"
)

// GetJavaType 根据数据库类型获取Java类型
func GetJavaType(columnType string) string {
	if columnType == "" {
		return JavaTypeString
	}

	// 简化的类型映射
	switch {
	case contains(columnType, "bigint"):
		return JavaTypeLong
	case contains(columnType, "int"):
		return JavaTypeInteger
	case contains(columnType, "float"), contains(columnType, "double"):
		return JavaTypeDouble
	case contains(columnType, "decimal"), contains(columnType, "numeric"):
		return JavaTypeBigDecimal
	case contains(columnType, "date"), contains(columnType, "time"):
		return JavaTypeDate
	case contains(columnType, "bit"), contains(columnType, "boolean"):
		return JavaTypeBoolean
	default:
		return JavaTypeString
	}
}

// GetHtmlType 根据数据库类型获取HTML类型
func GetHtmlType(columnType string) string {
	if columnType == "" {
		return HtmlTypeInput
	}

	// 简化的类型映射
	switch {
	case contains(columnType, "text"):
		return HtmlTypeTextarea
	case contains(columnType, "date"), contains(columnType, "time"):
		return HtmlTypeDatetime
	default:
		return HtmlTypeInput
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring 查找子字符串
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// IsSuperColumn 判断是否为基础字段（对应Java后端的isSuperColumn方法）
func (c *GenTableColumn) IsSuperColumn() bool {
	superColumns := []string{
		"createBy", "createTime", "updateBy", "updateTime", "remark",
		"parentName", "parentId", "orderNum", "ancestors",
	}

	for _, col := range superColumns {
		if c.JavaField == col {
			return true
		}
	}
	return false
}

// IsUsableColumn 判断是否为可用字段（对应Java后端的isUsableColumn方法）
func (c *GenTableColumn) IsUsableColumn() bool {
	usableColumns := []string{"parentId", "orderNum", "remark"}

	for _, col := range usableColumns {
		if c.JavaField == col {
			return true
		}
	}
	return false
}

// IsPrimaryKey 判断是否为主键
func (c *GenTableColumn) IsPrimaryKey() bool {
	return c.IsPk == "1"
}

// IsAutoIncrement 判断是否自增
func (c *GenTableColumn) IsAutoIncrement() bool {
	return c.IsIncrement == "1"
}

// IsRequiredField 判断是否必填
func (c *GenTableColumn) IsRequiredField() bool {
	return c.IsRequired == "1"
}

// IsInsertField 判断是否为插入字段
func (c *GenTableColumn) IsInsertField() bool {
	return c.IsInsert == "1"
}

// IsEditField 判断是否为编辑字段
func (c *GenTableColumn) IsEditField() bool {
	return c.IsEdit == "1"
}

// IsListField 判断是否为列表字段
func (c *GenTableColumn) IsListField() bool {
	return c.IsList == "1"
}

// IsQueryField 判断是否为查询字段
func (c *GenTableColumn) IsQueryField() bool {
	return c.IsQuery == "1"
}

// GetPkColumn 获取主键字段
func (t *GenTable) GetPkColumn() *GenTableColumn {
	for i := range t.Columns {
		if t.Columns[i].IsPrimaryKey() {
			return &t.Columns[i]
		}
	}
	return nil
}

// IsCrud 判断是否为单表操作
func (t *GenTable) IsCrud() bool {
	return t.TplCategory == TplCategoryCrud
}

// IsTree 判断是否为树表操作
func (t *GenTable) IsTree() bool {
	return t.TplCategory == TplCategoryTree
}

// IsSub 判断是否为主子表操作
func (t *GenTable) IsSub() bool {
	return t.TplCategory == TplCategorySub
}

// IsZipGenType 判断是否为zip压缩包生成方式
func (t *GenTable) IsZipGenType() bool {
	return t.GenType == GenTypeZip
}

// IsPathGenType 判断是否为自定义路径生成方式
func (t *GenTable) IsPathGenType() bool {
	return t.GenType == GenTypePath
}
