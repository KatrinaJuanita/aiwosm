package model

import (
	"fmt"
	"strings"
	"time"
)

// SysPost 岗位信息表 对应Java后端的SysPost实体
// 严格按照Java后端真实数据库表结构定义（基于ruoyi-java/sql/ry_20250522.sql）：
// post_id, post_code, post_name, post_sort, status, create_by, create_time, update_by, update_time, remark
type SysPost struct {
	PostID     int64      `gorm:"column:post_id;primaryKey;autoIncrement" json:"postId" excel:"name:岗位序号;sort:1;cellType:numeric"` // 岗位ID
	PostCode   string     `gorm:"column:post_code;size:64;not null" json:"postCode" excel:"name:岗位编码;sort:2"`                      // 岗位编码
	PostName   string     `gorm:"column:post_name;size:50;not null" json:"postName" excel:"name:岗位名称;sort:3"`                      // 岗位名称
	PostSort   int        `gorm:"column:post_sort;not null" json:"postSort" excel:"name:岗位排序;sort:4"`                              // 显示顺序
	Status     string     `gorm:"column:status;size:1;default:0" json:"status" excel:"name:状态;sort:5;readConverterExp:0=正常,1=停用"`  // 状态（0正常 1停用）
	CreateBy   string     `gorm:"column:create_by;size:64;default:''" json:"createBy"`                                             // 创建者
	CreateTime *time.Time `gorm:"column:create_time" json:"createTime"`                                                            // 创建时间
	UpdateBy   string     `gorm:"column:update_by;size:64;default:''" json:"updateBy"`                                             // 更新者
	UpdateTime *time.Time `gorm:"column:update_time" json:"updateTime"`                                                            // 更新时间
	Remark     string     `gorm:"column:remark;size:500" json:"remark"`                                                            // 备注

	// 关联字段（不存储在数据库中）
	Flag   bool                   `gorm:"-" json:"flag"` // 用户是否存在此岗位标识 默认不存在
	Params map[string]interface{} `gorm:"-" json:"-"`    // 请求参数（用于数据权限等）
}

// TableName 设置表名
func (SysPost) TableName() string {
	return "sys_post"
}

// IsNormal 判断岗位状态是否正常
func (p *SysPost) IsNormal() bool {
	return p.Status == "0"
}

// IsDisabled 判断岗位状态是否停用
func (p *SysPost) IsDisabled() bool {
	return p.Status == "1"
}

// SetNormal 设置为正常状态
func (p *SysPost) SetNormal() {
	p.Status = "0"
}

// SetDisabled 设置为停用状态
func (p *SysPost) SetDisabled() {
	p.Status = "1"
}

// PostQueryParams 岗位查询参数 对应Java后端的查询条件
type PostQueryParams struct {
	QueryParams        // 继承分页参数
	PostCode    string `form:"postCode" json:"postCode"` // 岗位编码
	PostName    string `form:"postName" json:"postName"` // 岗位名称
	Status      string `form:"status" json:"status"`     // 状态
}

// PostExportData 岗位导出数据结构 对应Java后端的Excel导出
type PostExportData struct {
	PostID   int64  `json:"postId" excel:"岗位序号"`   // 岗位序号
	PostCode string `json:"postCode" excel:"岗位编码"` // 岗位编码
	PostName string `json:"postName" excel:"岗位名称"` // 岗位名称
	PostSort int    `json:"postSort" excel:"岗位排序"` // 岗位排序
	Status   string `json:"status" excel:"状态"`     // 状态
	Remark   string `json:"remark" excel:"备注"`     // 备注
}

// ToExportData 转换为导出数据格式
func (p *SysPost) ToExportData() *PostExportData {
	statusText := "正常"
	if p.Status == "1" {
		statusText = "停用"
	}

	return &PostExportData{
		PostID:   p.PostID,
		PostCode: p.PostCode,
		PostName: p.PostName,
		PostSort: p.PostSort,
		Status:   statusText,
		Remark:   p.Remark,
	}
}

// PostConstants 岗位相关常量 对应Java后端的UserConstants
const (
	// 岗位状态
	PostStatusNormal  = "0" // 正常
	PostStatusDisable = "1" // 停用

	// 岗位唯一性检查结果
	PostNameUnique    = true  // 岗位名称唯一
	PostNameNotUnique = false // 岗位名称不唯一
	PostCodeUnique    = true  // 岗位编码唯一
	PostCodeNotUnique = false // 岗位编码不唯一
)

// ValidatePost 验证岗位信息 对应Java后端的@Validated注解验证
func ValidatePost(post *SysPost, isUpdate bool) error {
	if post == nil {
		return fmt.Errorf("岗位信息不能为空")
	}

	// 对应Java后端的@NotBlank(message = "岗位编码不能为空")
	if strings.TrimSpace(post.PostCode) == "" {
		return fmt.Errorf("岗位编码不能为空")
	}

	// 对应Java后端的@NotBlank(message = "岗位名称不能为空")
	if strings.TrimSpace(post.PostName) == "" {
		return fmt.Errorf("岗位名称不能为空")
	}

	// 对应Java后端的@Size(min = 0, max = 64, message = "岗位编码长度不能超过64个字符")
	if len(post.PostCode) > 64 {
		return fmt.Errorf("岗位编码长度不能超过64个字符")
	}

	// 对应Java后端的@Size(min = 0, max = 50, message = "岗位名称长度不能超过50个字符")
	if len(post.PostName) > 50 {
		return fmt.Errorf("岗位名称长度不能超过50个字符")
	}

	// 对应Java后端的@NotNull(message = "显示顺序不能为空")
	if post.PostSort < 0 {
		return fmt.Errorf("显示顺序不能为负数")
	}

	// 验证状态值
	if post.Status != "" && post.Status != PostStatusNormal && post.Status != PostStatusDisable {
		return fmt.Errorf("岗位状态值无效")
	}

	// 验证备注长度
	if len(post.Remark) > 500 {
		return fmt.Errorf("备注长度不能超过500个字符")
	}

	// 更新时验证ID
	if isUpdate && post.PostID <= 0 {
		return fmt.Errorf("岗位ID不能为空")
	}

	return nil
}

// GetPostStatusText 获取岗位状态文本
func GetPostStatusText(status string) string {
	switch status {
	case PostStatusNormal:
		return "正常"
	case PostStatusDisable:
		return "停用"
	default:
		return "未知"
	}
}

// IsValidPostStatus 验证岗位状态是否有效
func IsValidPostStatus(status string) bool {
	return status == PostStatusNormal || status == PostStatusDisable
}

// SysUserPost 用户和岗位关联表 对应Java后端的SysUserPost实体
type SysUserPost struct {
	UserID int64 `gorm:"column:user_id;primaryKey" json:"userId"` // 用户ID
	PostID int64 `gorm:"column:post_id;primaryKey" json:"postId"` // 岗位ID
}

// TableName 设置表名
func (SysUserPost) TableName() string {
	return "sys_user_post"
}
