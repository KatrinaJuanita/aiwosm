package model

import (
	"time"
)

// SysLogininfor 系统访问记录表 对应Java后端的SysLogininfor实体
// 严格按照真实数据库表结构定义（基于SqlServer_ry_20250522_COMPLETE.sql）：
// 数据库表只有9个字段：info_id, user_name, ipaddr, login_location, browser, os, status, msg, login_time
// 注意：虽然Java后端继承BaseEntity，但数据库表实际没有BaseEntity的字段
type SysLogininfor struct {
	InfoID        int64      `gorm:"column:info_id;primaryKey;autoIncrement" json:"infoId" excel:"name:访问编号;sort:1;cellType:numeric"`     // 访问ID
	UserName      string     `gorm:"column:user_name;size:50" json:"userName" excel:"name:用户名称;sort:2"`                                   // 用户账号
	IPAddr        string     `gorm:"column:ipaddr;size:128" json:"ipaddr" excel:"name:登录地址;sort:3"`                                       // 登录IP地址
	LoginLocation string     `gorm:"column:login_location;size:255" json:"loginLocation" excel:"name:登录地点;sort:4"`                        // 登录地点
	Browser       string     `gorm:"column:browser;size:50" json:"browser" excel:"name:浏览器;sort:5"`                                       // 浏览器类型
	OS            string     `gorm:"column:os;size:50" json:"os" excel:"name:操作系统;sort:6"`                                                // 操作系统
	Status        string     `gorm:"column:status;size:1;default:0" json:"status" excel:"name:登录状态;sort:7;readConverterExp:0=成功,1=失败"`    // 登录状态（0成功 1失败）
	Msg           string     `gorm:"column:msg;size:255" json:"msg" excel:"name:提示消息;sort:8;width:30"`                                    // 提示消息
	LoginTime     *time.Time `gorm:"column:login_time" json:"loginTime" excel:"name:访问时间;sort:9;width:30;dateFormat:yyyy-MM-dd HH:mm:ss"` // 访问时间

	// 查询条件字段（不映射到数据库）
	BeginTime string `gorm:"-" json:"beginTime"` // 开始时间
	EndTime   string `gorm:"-" json:"endTime"`   // 结束时间
}

// TableName 指定表名
func (SysLogininfor) TableName() string {
	return "sys_logininfor"
}

// 登录状态常量
const (
	LoginStatusSuccess = "0" // 成功
	LoginStatusFail    = "1" // 失败
)

// GetStatusName 获取状态名称
func GetLoginStatusName(status string) string {
	switch status {
	case LoginStatusSuccess:
		return "成功"
	case LoginStatusFail:
		return "失败"
	default:
		return "未知"
	}
}

// 登录消息常量
const (
	LoginMsgSuccess         = "登录成功"
	LoginMsgLoginSuccess    = "登录成功" // 对应Java后端的登录成功消息
	LoginMsgUserNotExists   = "用户不存在/密码错误"
	LoginMsgPasswordError   = "用户不存在/密码错误"
	LoginMsgUserDisabled    = "用户已停用，请联系管理员"
	LoginMsgPasswordRetry   = "密码输入错误{0}次"
	LoginMsgUserLocked      = "用户账户已锁定"
	LoginMsgCaptchaError    = "验证码错误"
	LoginMsgCaptchaExpire   = "验证码已过期"
	LoginMsgLogoutSuccess   = "退出成功"
	LoginMsgRegisterSuccess = "注册成功"
	LoginMsgRegisterError   = "注册失败"
	LoginMsgUnknownError    = "未知错误"
)
