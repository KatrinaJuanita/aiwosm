package model

import (
	"strings"
	"time"
)

// BaseEntity 基础实体 对应Java后端的BaseEntity
type BaseEntity struct {
	SearchValue string                 `gorm:"-" json:"-"`                                          // 搜索值（不存储到数据库，对应Java的@JsonIgnore searchValue）
	CreateBy    string                 `gorm:"column:create_by;size:64;default:''" json:"createBy"` // 创建者
	CreateTime  *time.Time             `gorm:"column:create_time" json:"createTime"`                // 创建时间
	UpdateBy    string                 `gorm:"column:update_by;size:64;default:''" json:"updateBy"` // 更新者
	UpdateTime  *time.Time             `gorm:"column:update_time" json:"updateTime"`                // 更新时间
	Remark      string                 `gorm:"column:remark;size:500" json:"remark"`                // 备注
	Params      map[string]interface{} `gorm:"-" json:"params,omitempty"`                           // 请求参数（不存储到数据库）
}

// GetSearchValue 获取搜索值 对应Java后端的getSearchValue方法
func (b *BaseEntity) GetSearchValue() string {
	return b.SearchValue
}

// SetSearchValue 设置搜索值 对应Java后端的setSearchValue方法
func (b *BaseEntity) SetSearchValue(searchValue string) {
	b.SearchValue = searchValue
}

// GetParams 获取参数映射 对应Java后端的getParams方法
func (b *BaseEntity) GetParams() map[string]interface{} {
	if b.Params == nil {
		b.Params = make(map[string]interface{})
	}
	return b.Params
}

// SetParams 设置参数映射
func (b *BaseEntity) SetParams(params map[string]interface{}) {
	b.Params = params
}

// GetParam 获取指定参数值
func (b *BaseEntity) GetParam(key string) interface{} {
	if b.Params == nil {
		return nil
	}
	return b.Params[key]
}

// SetParam 设置指定参数值
func (b *BaseEntity) SetParam(key string, value interface{}) {
	if b.Params == nil {
		b.Params = make(map[string]interface{})
	}
	b.Params[key] = value
}

// GetDataScope 获取数据权限SQL 对应Java后端的数据权限处理
func (b *BaseEntity) GetDataScope() string {
	if b.Params == nil {
		return ""
	}
	if dataScope, exists := b.Params["dataScope"]; exists {
		if ds, ok := dataScope.(string); ok {
			return ds
		}
	}
	return ""
}

// SetDataScope 设置数据权限SQL
func (b *BaseEntity) SetDataScope(dataScope string) {
	if b.Params == nil {
		b.Params = make(map[string]interface{})
	}
	b.Params["dataScope"] = dataScope
}

// QueryParams 查询参数基础结构 对应Java后端的查询参数
type QueryParams struct {
	BaseEntity
	PageNum       int    `form:"pageNum" json:"pageNum"`             // 页码
	PageSize      int    `form:"pageSize" json:"pageSize"`           // 每页数量
	OrderByColumn string `form:"orderByColumn" json:"orderByColumn"` // 排序字段
	IsAsc         string `form:"isAsc" json:"isAsc"`                 // 排序方向
	BeginTime     string `form:"beginTime" json:"beginTime"`         // 开始时间
	EndTime       string `form:"endTime" json:"endTime"`             // 结束时间
}

// GetOrderBy 获取排序SQL
func (q *QueryParams) GetOrderBy() string {
	if q.OrderByColumn == "" {
		return ""
	}

	direction := "DESC"
	if q.IsAsc == "asc" {
		direction = "ASC"
	}

	return q.OrderByColumn + " " + direction
}

// GetOffset 获取分页偏移量
func (q *QueryParams) GetOffset() int {
	if q.PageNum <= 0 || q.PageSize <= 0 {
		return 0
	}
	return (q.PageNum - 1) * q.PageSize
}

// GetLimit 获取分页限制数量
func (q *QueryParams) GetLimit() int {
	if q.PageSize <= 0 {
		return 10 // 默认每页10条
	}
	return q.PageSize
}

// IsValidPagination 检查分页参数是否有效
func (q *QueryParams) IsValidPagination() bool {
	return q.PageNum > 0 && q.PageSize > 0
}

// SetDefaultPagination 设置默认分页参数
func (q *QueryParams) SetDefaultPagination() {
	if q.PageNum <= 0 {
		q.PageNum = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
}

// TreeEntity 树形实体基础结构
type TreeEntity struct {
	BaseEntity
	ParentID  int    `gorm:"column:parent_id;default:0" json:"parentId"` // 父级ID
	Ancestors string `gorm:"column:ancestors;size:50" json:"ancestors"`  // 祖级列表
	OrderNum  int    `gorm:"column:order_num;default:0" json:"orderNum"` // 显示顺序
}

// IsRoot 判断是否为根节点
func (t *TreeEntity) IsRoot() bool {
	return t.ParentID == 0
}

// GetAncestorList 获取祖先ID列表
func (t *TreeEntity) GetAncestorList() []string {
	if t.Ancestors == "" {
		return []string{}
	}

	ancestors := strings.Split(t.Ancestors, ",")
	var result []string
	for _, ancestor := range ancestors {
		ancestor = strings.TrimSpace(ancestor)
		if ancestor != "" {
			result = append(result, ancestor)
		}
	}
	return result
}

// StatusEntity 状态实体基础结构
type StatusEntity struct {
	BaseEntity
	Status string `gorm:"column:status;size:1;default:0" json:"status"` // 状态（0正常 1停用）
}

// IsNormal 判断状态是否正常
func (s *StatusEntity) IsNormal() bool {
	return s.Status == "0"
}

// IsDisabled 判断状态是否停用
func (s *StatusEntity) IsDisabled() bool {
	return s.Status == "1"
}

// SetNormal 设置为正常状态
func (s *StatusEntity) SetNormal() {
	s.Status = "0"
}

// SetDisabled 设置为停用状态
func (s *StatusEntity) SetDisabled() {
	s.Status = "1"
}

// DelFlagEntity 删除标志实体基础结构
type DelFlagEntity struct {
	BaseEntity
	DelFlag string `gorm:"column:del_flag;size:1;default:0" json:"delFlag"` // 删除标志（0代表存在 2代表删除）
}

// IsDeleted 判断是否已删除
func (d *DelFlagEntity) IsDeleted() bool {
	return d.DelFlag == "2"
}

// IsNotDeleted 判断是否未删除
func (d *DelFlagEntity) IsNotDeleted() bool {
	return d.DelFlag == "0"
}

// SetDeleted 设置为已删除
func (d *DelFlagEntity) SetDeleted() {
	d.DelFlag = "2"
}

// SetNotDeleted 设置为未删除
func (d *DelFlagEntity) SetNotDeleted() {
	d.DelFlag = "0"
}
