package model

import (
	"fmt"
	"time"
)

// SysUserOnline 当前在线会话 对应Java后端的SysUserOnline实体
// 这是一个虚拟实体，不对应数据库表，用于展示在线用户信息
type SysUserOnline struct {
	TokenID       string `json:"tokenId" excel:"name:会话编号;sort:1"`              // 会话编号
	DeptName      string `json:"deptName" excel:"name:部门名称;sort:2"`             // 部门名称
	UserName      string `json:"userName" excel:"name:用户名称;sort:3"`             // 用户名称
	IPAddr        string `json:"ipaddr" excel:"name:登录IP地址;sort:4"`             // 登录IP地址
	LoginLocation string `json:"loginLocation" excel:"name:登录地址;sort:5"`        // 登录地址
	Browser       string `json:"browser" excel:"name:浏览器类型;sort:6"`             // 浏览器类型
	OS            string `json:"os" excel:"name:操作系统;sort:7"`                   // 操作系统
	LoginTime     int64  `json:"loginTime" excel:"name:登录时间;sort:8;dateFormat"` // 登录时间（时间戳）
}

// GetLoginTimeFormatted 获取格式化的登录时间
func (u *SysUserOnline) GetLoginTimeFormatted() string {
	if u.LoginTime == 0 {
		return ""
	}
	// LoginTime是毫秒时间戳，需要除以1000转换为秒
	return time.Unix(u.LoginTime/1000, 0).Format("2006-01-02 15:04:05")
}

// GetOnlineDuration 获取在线时长（分钟）
func (u *SysUserOnline) GetOnlineDuration() int64 {
	if u.LoginTime == 0 {
		return 0
	}
	// LoginTime是毫秒时间戳，需要除以1000转换为秒
	return (time.Now().Unix() - u.LoginTime/1000) / 60
}

// GetOnlineDurationFormatted 获取格式化的在线时长
func (u *SysUserOnline) GetOnlineDurationFormatted() string {
	duration := u.GetOnlineDuration()
	if duration < 60 {
		return fmt.Sprintf("%d分钟", duration)
	}
	hours := duration / 60
	minutes := duration % 60
	return fmt.Sprintf("%d小时%d分钟", hours, minutes)
}
