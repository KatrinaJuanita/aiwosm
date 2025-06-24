package model

import "fmt"

// RegisterBody 用户注册对象 对应Java后端的RegisterBody
type RegisterBody struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
	Code     string `json:"code"`                        // 验证码
	UUID     string `json:"uuid"`                        // 唯一标识
}

// Validate 验证注册参数
func (r *RegisterBody) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if r.Password == "" {
		return fmt.Errorf("用户密码不能为空")
	}
	if len(r.Username) < UserNameMinLength || len(r.Username) > UserNameMaxLength {
		return fmt.Errorf("账户长度必须在%d到%d个字符之间", UserNameMinLength, UserNameMaxLength)
	}
	if len(r.Password) < PasswordMinLength || len(r.Password) > PasswordMaxLength {
		return fmt.Errorf("密码长度必须在%d到%d个字符之间", PasswordMinLength, PasswordMaxLength)
	}
	return nil
}

// 用户注册相关常量 对应Java后端的UserConstants
const (
	UserNameMinLength = 2  // 用户名最小长度
	UserNameMaxLength = 20 // 用户名最大长度
	PasswordMinLength = 5  // 密码最小长度
	PasswordMaxLength = 20 // 密码最大长度
)
