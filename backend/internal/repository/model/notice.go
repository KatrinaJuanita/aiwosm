package model

import (
	"time"
)

// SysNotice 通知公告表 对应Java后端的SysNotice实体
// 严格按照Java后端真实数据库表结构定义（基于SqlServer_ry_20250522_COMPLETE.sql）：
// notice_id, notice_title, notice_type, notice_content, status, create_by, create_time, update_by, update_time, remark
type SysNotice struct {
	NoticeID      int64      `gorm:"column:notice_id;primaryKey;autoIncrement" json:"noticeId"`     // 公告ID
	NoticeTitle   string     `gorm:"column:notice_title;size:50;not null" json:"noticeTitle"`       // 公告标题
	NoticeType    string     `gorm:"column:notice_type;size:1;not null" json:"noticeType"`          // 公告类型（1通知 2公告）
	NoticeContent string     `gorm:"column:notice_content;type:nvarchar(max)" json:"noticeContent"` // 公告内容
	Status        string     `gorm:"column:status;size:1;default:0" json:"status"`                  // 公告状态（0正常 1关闭）
	CreateBy      string     `gorm:"column:create_by;size:64;default:''" json:"createBy"`           // 创建者
	CreateTime    *time.Time `gorm:"column:create_time" json:"createTime"`                          // 创建时间
	UpdateBy      string     `gorm:"column:update_by;size:64;default:''" json:"updateBy"`           // 更新者
	UpdateTime    *time.Time `gorm:"column:update_time" json:"updateTime"`                          // 更新时间
	Remark        string     `gorm:"column:remark;size:255" json:"remark"`                          // 备注

	// 扩展字段（不存储在数据库中，用于前端显示）
	NoticeTypeText       string `gorm:"-" json:"noticeTypeText,omitempty"`       // 公告类型文本
	StatusText           string `gorm:"-" json:"statusText,omitempty"`           // 状态文本
	NoticeContentPreview string `gorm:"-" json:"noticeContentPreview,omitempty"` // 内容预览（用于列表显示）
}

// TableName 设置表名
func (SysNotice) TableName() string {
	return "sys_notice"
}

// 通知公告类型常量 对应Java后端的常量定义
const (
	NoticeTypeNotification = "1" // 通知
	NoticeTypeAnnouncement = "2" // 公告
)

// 通知公告状态常量 对应Java后端的常量定义
const (
	NoticeStatusNormal = "0" // 正常
	NoticeStatusClosed = "1" // 关闭
)

// NoticeQueryParams 通知公告查询参数 对应Java后端的查询条件
type NoticeQueryParams struct {
	NoticeTitle   string `form:"noticeTitle" json:"noticeTitle"`     // 公告标题（模糊查询）
	NoticeType    string `form:"noticeType" json:"noticeType"`       // 公告类型
	CreateBy      string `form:"createBy" json:"createBy"`           // 创建者（模糊查询）
	Status        string `form:"status" json:"status"`               // 状态
	PageNum       int    `form:"pageNum" json:"pageNum"`             // 页码
	PageSize      int    `form:"pageSize" json:"pageSize"`           // 每页数量
	OrderByColumn string `form:"orderByColumn" json:"orderByColumn"` // 排序字段
	IsAsc         string `form:"isAsc" json:"isAsc"`                 // 排序方向

	// 数据权限相关字段（不通过表单绑定，由后端设置）
	DataScope       string `form:"-" json:"-"`                 // 数据权限范围
	CurrentUserId   int64  `form:"-" json:"-"`                 // 当前用户ID
	CurrentUserName string `form:"-" json:"-"`                 // 当前用户名
	BeginTime       string `form:"beginTime" json:"beginTime"` // 开始时间
	EndTime         string `form:"endTime" json:"endTime"`     // 结束时间
}

// SysNoticeExport 通知公告导出结构体 对应Java后端的Excel导出格式
type SysNoticeExport struct {
	NoticeID      int64  `excel:"name:公告编号;sort:1"`
	NoticeTitle   string `excel:"name:公告标题;sort:2"`
	NoticeType    string `excel:"name:公告类型;sort:3;readConverterExp:1=通知,2=公告"`
	NoticeContent string `excel:"name:公告内容;sort:4"`
	Status        string `excel:"name:公告状态;sort:5;readConverterExp:0=正常,1=关闭"`
	CreateBy      string `excel:"name:创建者;sort:6"`
	CreateTime    string `excel:"name:创建时间;sort:7"`
	UpdateBy      string `excel:"name:更新者;sort:8"`
	UpdateTime    string `excel:"name:更新时间;sort:9"`
	Remark        string `excel:"name:备注;sort:10"`
}
