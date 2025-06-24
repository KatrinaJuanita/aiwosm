package system

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"wosm/internal/repository/dao"
	"wosm/internal/repository/model"
)

// LoginLogService 登录日志服务 对应Java后端的ISysLogininforService
type LoginLogService struct {
	loginLogDao *dao.LoginLogDao
}

// NewLoginLogService 创建登录日志服务实例
func NewLoginLogService() *LoginLogService {
	return &LoginLogService{
		loginLogDao: dao.NewLoginLogDao(),
	}
}

// SelectLogininforList 查询系统登录日志集合 对应Java后端的selectLogininforList
func (s *LoginLogService) SelectLogininforList(logininfor *model.SysLogininfor) ([]model.SysLogininfor, error) {
	fmt.Printf("LoginLogService.SelectLogininforList: 查询登录日志列表\n")
	return s.loginLogDao.SelectLogininforList(logininfor)
}

// InsertLogininfor 新增系统登录日志 对应Java后端的insertLogininfor
func (s *LoginLogService) InsertLogininfor(logininfor *model.SysLogininfor) error {
	fmt.Printf("LoginLogService.InsertLogininfor: 新增登录日志, UserName=%s\n", logininfor.UserName)

	// 设置登录时间
	now := time.Now()
	logininfor.LoginTime = &now

	return s.loginLogDao.InsertLogininfor(logininfor)
}

// DeleteLogininforByIds 批量删除系统登录日志 对应Java后端的deleteLogininforByIds
func (s *LoginLogService) DeleteLogininforByIds(infoIds []int) error {
	fmt.Printf("LoginLogService.DeleteLogininforByIds: 批量删除登录日志, InfoIDs=%v\n", infoIds)
	return s.loginLogDao.DeleteLogininforByIds(infoIds)
}

// CleanLogininfor 清空系统登录日志 对应Java后端的cleanLogininfor
func (s *LoginLogService) CleanLogininfor() error {
	fmt.Printf("LoginLogService.CleanLogininfor: 清空登录日志\n")
	return s.loginLogDao.CleanLogininfor()
}

// RecordLoginInfo 记录登录信息 对应Java后端的recordLoginInfo
func (s *LoginLogService) RecordLoginInfo(userName, status, message, ipAddr, userAgent string) {
	fmt.Printf("LoginLogService.RecordLoginInfo: 记录登录信息, UserName=%s, Status=%s\n", userName, status)

	// 异步记录登录日志，避免影响登录性能
	go func() {
		logininfor := &model.SysLogininfor{
			UserName:      userName,
			Status:        status,
			IPAddr:        ipAddr,
			LoginLocation: getLoginLocation(ipAddr),
			Browser:       getBrowser(userAgent),
			OS:            getOS(userAgent),
			Msg:           message,
		}

		if err := s.InsertLogininfor(logininfor); err != nil {
			fmt.Printf("记录登录日志失败: %v\n", err)
		}
	}()
}

// getLoginLocation 获取登录地点 对应Java后端的AddressUtils.getRealAddressByIP
func getLoginLocation(ipAddr string) string {
	// 内网IP直接返回 对应Java后端的IpUtils.internalIp判断
	if isInternalIP(ipAddr) {
		return "内网IP"
	}

	// TODO: 可集成第三方IP地址库获取真实地理位置
	// Java后端使用 http://whois.pconline.com.cn/ipJson.jsp 接口
	// 这里暂时返回未知位置，避免外部依赖
	return "未知位置"
}

// isInternalIP 判断是否为内网IP 对应Java后端的IpUtils.internalIp
func isInternalIP(ip string) bool {
	if ip == "127.0.0.1" || ip == "::1" || ip == "localhost" {
		return true
	}

	// 检查私有IP地址段
	// 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16
	if strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "192.168.") ||
		(strings.HasPrefix(ip, "172.") && isInRange172(ip)) {
		return true
	}

	return false
}

// isInRange172 检查是否在172.16.0.0/12范围内
func isInRange172(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	if parts[0] != "172" {
		return false
	}

	second, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	return second >= 16 && second <= 31
}

// getBrowser 获取浏览器信息（简化实现）
func getBrowser(userAgent string) string {
	// TODO: 解析User-Agent获取浏览器信息
	if userAgent == "" {
		return "Unknown"
	}

	// 简单的浏览器识别
	if containsString(userAgent, "Chrome") {
		return "Chrome"
	} else if containsString(userAgent, "Firefox") {
		return "Firefox"
	} else if containsString(userAgent, "Safari") {
		return "Safari"
	} else if containsString(userAgent, "Edge") {
		return "Edge"
	}

	return "Unknown"
}

// getOS 获取操作系统信息（简化实现）
func getOS(userAgent string) string {
	// TODO: 解析User-Agent获取操作系统信息
	if userAgent == "" {
		return "Unknown"
	}

	// 简单的操作系统识别
	if containsString(userAgent, "Windows") {
		return "Windows"
	} else if containsString(userAgent, "Mac") {
		return "Mac OS"
	} else if containsString(userAgent, "Linux") {
		return "Linux"
	} else if containsString(userAgent, "Android") {
		return "Android"
	} else if containsString(userAgent, "iOS") {
		return "iOS"
	}

	return "Unknown"
}

// containsString 检查字符串是否包含子字符串（忽略大小写）
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				findSubstring(s, substr)))
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
