package system

import (
	"context"
	"encoding/json"
	"fmt"
	"wosm/internal/repository/model"
	"wosm/internal/utils"
	"wosm/pkg/redis"
)

// UserOnlineService 在线用户服务 对应Java后端的ISysUserOnlineService
type UserOnlineService struct{}

// NewUserOnlineService 创建在线用户服务实例
func NewUserOnlineService() *UserOnlineService {
	return &UserOnlineService{}
}

// SelectOnlineByIpaddr 通过登录地址查询信息 对应Java后端的selectOnlineByIpaddr
func (s *UserOnlineService) SelectOnlineByIpaddr(ipaddr string, user *model.LoginUser) *model.SysUserOnline {
	if user == nil || user.IPAddr != ipaddr {
		return nil
	}
	return s.LoginUserToUserOnline(user)
}

// SelectOnlineByUserName 通过用户名称查询信息 对应Java后端的selectOnlineByUserName
func (s *UserOnlineService) SelectOnlineByUserName(userName string, user *model.LoginUser) *model.SysUserOnline {
	// 对应Java后端的 StringUtils.equals(userName, user.getUsername())
	if user == nil || user.User == nil {
		return nil
	}
	// 使用user.User.UserName而不是user.Username，保持与Java后端一致
	if userName == user.User.UserName {
		return s.LoginUserToUserOnline(user)
	}
	return nil
}

// SelectOnlineByInfo 通过登录地址/用户名称查询信息 对应Java后端的selectOnlineByInfo
func (s *UserOnlineService) SelectOnlineByInfo(ipaddr, userName string, user *model.LoginUser) *model.SysUserOnline {
	// 对应Java后端的 StringUtils.equals(ipaddr, user.getIpaddr()) && StringUtils.equals(userName, user.getUsername())
	if user == nil || user.User == nil {
		return nil
	}
	if ipaddr == user.IPAddr && userName == user.User.UserName {
		return s.LoginUserToUserOnline(user)
	}
	return nil
}

// LoginUserToUserOnline 设置在线用户信息 对应Java后端的loginUserToUserOnline
func (s *UserOnlineService) LoginUserToUserOnline(user *model.LoginUser) *model.SysUserOnline {
	if user == nil || user.User == nil {
		return nil
	}

	// 获取部门名称
	deptName := ""
	if user.User.Dept != nil {
		deptName = user.User.Dept.DeptName
	}

	sysUserOnline := &model.SysUserOnline{
		TokenID:       user.Token,
		UserName:      user.User.UserName,
		IPAddr:        user.IPAddr,
		LoginLocation: user.LoginLocation,
		Browser:       user.Browser,
		OS:            user.OS,
		LoginTime:     user.LoginTime, // 直接返回毫秒时间戳，与Java后端完全一致
		DeptName:      deptName,
	}

	return sysUserOnline
}

// SelectOnlineUsers 查询在线用户列表 对应Java后端的list方法逻辑
func (s *UserOnlineService) SelectOnlineUsers(ipaddr, userName string) ([]model.SysUserOnline, error) {
	fmt.Printf("UserOnlineService.SelectOnlineUsers: 查询在线用户列表, IPAddr=%s, UserName=%s\n", ipaddr, userName)

	var userOnlineList []model.SysUserOnline

	// 获取所有登录token的key
	ctx := context.Background()
	keys, err := redis.GetRedis().Keys(ctx, "login_tokens:*").Result()
	if err != nil {
		fmt.Printf("SelectOnlineUsers: 获取Redis keys失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("SelectOnlineUsers: 找到%d个在线会话\n", len(keys))

	// 遍历所有key，获取登录用户信息
	for _, key := range keys {
		// 从Redis获取登录用户信息
		loginUserData, err := redis.Get(key)
		if err != nil {
			fmt.Printf("SelectOnlineUsers: 获取用户信息失败, key=%s, error=%v\n", key, err)
			continue
		}

		if loginUserData == "" {
			continue
		}

		// 解析登录用户信息
		var loginUser model.LoginUser
		if err := json.Unmarshal([]byte(loginUserData), &loginUser); err != nil {
			fmt.Printf("SelectOnlineUsers: 解析用户信息失败, key=%s, error=%v\n", key, err)
			continue
		}

		// 根据查询条件过滤 - 对应Java后端的过滤逻辑
		var userOnline *model.SysUserOnline
		if utils.IsNotEmpty(ipaddr) && utils.IsNotEmpty(userName) {
			userOnline = s.SelectOnlineByInfo(ipaddr, userName, &loginUser)
		} else if utils.IsNotEmpty(ipaddr) {
			userOnline = s.SelectOnlineByIpaddr(ipaddr, &loginUser)
		} else if utils.IsNotEmpty(userName) && loginUser.User != nil {
			// 对应Java后端的 StringUtils.isNotEmpty(userName) && StringUtils.isNotNull(user.getUser())
			userOnline = s.SelectOnlineByUserName(userName, &loginUser)
		} else {
			userOnline = s.LoginUserToUserOnline(&loginUser)
		}

		if userOnline != nil {
			userOnlineList = append(userOnlineList, *userOnline)
		}
	}

	// 对应Java后端的Collections.reverse(userOnlineList) - 反转列表，最新登录的在前
	for i, j := 0, len(userOnlineList)-1; i < j; i, j = i+1, j-1 {
		userOnlineList[i], userOnlineList[j] = userOnlineList[j], userOnlineList[i]
	}

	// 对应Java后端的userOnlineList.removeAll(Collections.singleton(null)) - 移除空值
	// Go中我们在添加时已经检查了nil，所以这里不需要额外处理

	fmt.Printf("SelectOnlineUsers: 查询到在线用户数量=%d\n", len(userOnlineList))
	return userOnlineList, nil
}

// ForceLogout 强制用户下线 对应Java后端的forceLogout方法逻辑
func (s *UserOnlineService) ForceLogout(tokenId string) error {
	fmt.Printf("UserOnlineService.ForceLogout: 强制用户下线, TokenID=%s\n", tokenId)

	// 构建Redis key
	key := "login_tokens:" + tokenId

	// 从Redis删除用户会话
	err := redis.Del(key)
	if err != nil {
		fmt.Printf("ForceLogout: 删除用户会话失败: %v\n", err)
		return err
	}

	fmt.Printf("ForceLogout: 强制用户下线成功, TokenID=%s\n", tokenId)
	return nil
}

// GetOnlineUserCount 获取在线用户数量
func (s *UserOnlineService) GetOnlineUserCount() (int, error) {
	ctx := context.Background()
	keys, err := redis.GetRedis().Keys(ctx, "login_tokens:*").Result()
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}

// CleanExpiredUsers 清理过期用户会话
func (s *UserOnlineService) CleanExpiredUsers() error {
	fmt.Printf("UserOnlineService.CleanExpiredUsers: 清理过期用户会话\n")

	ctx := context.Background()
	keys, err := redis.GetRedis().Keys(ctx, "login_tokens:*").Result()
	if err != nil {
		return err
	}

	expiredCount := 0
	for _, key := range keys {
		// 检查key是否过期
		ttl, err := redis.GetRedis().TTL(ctx, key).Result()
		if err != nil {
			continue
		}

		// 如果TTL为-2，表示key已过期或不存在
		if ttl.Seconds() < 0 {
			redis.Del(key)
			expiredCount++
		}
	}

	fmt.Printf("CleanExpiredUsers: 清理过期会话数量=%d\n", expiredCount)
	return nil
}
