package test

import (
	"testing"
	"wosm/internal/repository/model"

	"github.com/go-playground/validator/v10"
)

// TestUserModelValidation 测试用户模型验证功能
func TestUserModelValidation(t *testing.T) {
	// 初始化验证器
	validate := validator.New()
	// 注册自定义验证器（XSS验证等）
	validate.RegisterValidation("xss", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		// 简单的XSS检测
		return !contains(value, "<script") && !contains(value, "javascript:") && !contains(value, "onclick=")
	})

	tests := []struct {
		name      string
		user      model.SysUser
		expectErr bool
		errField  string
		errTag    string
	}{
		{
			name: "正常用户数据",
			user: model.SysUser{
				UserName: "testuser",
				NickName: "测试用户",
				Email:    "test@example.com",
			},
			expectErr: false,
		},
		{
			name: "用户名包含XSS脚本",
			user: model.SysUser{
				UserName: "<script>alert('xss')</script>",
				NickName: "测试用户",
				Email:    "test@example.com",
			},
			expectErr: true,
			errField:  "UserName",
			errTag:    "xss",
		},
		{
			name: "昵称包含HTML标签",
			user: model.SysUser{
				UserName: "testuser",
				NickName: "<div>测试用户</div>",
				Email:    "test@example.com",
			},
			expectErr: true,
			errField:  "NickName",
			errTag:    "xss",
		},
		{
			name: "邮箱格式错误",
			user: model.SysUser{
				UserName: "testuser",
				NickName: "测试用户",
				Email:    "invalid-email",
			},
			expectErr: true,
			errField:  "Email",
			errTag:    "email",
		},
		{
			name: "用户名为空",
			user: model.SysUser{
				UserName: "",
				NickName: "测试用户",
				Email:    "test@example.com",
			},
			expectErr: true,
			errField:  "UserName",
			errTag:    "required",
		},
		{
			name: "用户名包含JavaScript事件",
			user: model.SysUser{
				UserName: "onclick=alert('xss')",
				NickName: "测试用户",
				Email:    "test@example.com",
			},
			expectErr: true,
			errField:  "UserName",
			errTag:    "xss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.user)

			if tt.expectErr {
				if err == nil {
					t.Errorf("期望验证失败，但验证通过了")
					return
				}

				// 检查是否是预期的字段错误
				if validationErrors, ok := err.(validator.ValidationErrors); ok {
					found := false
					for _, fieldErr := range validationErrors {
						if fieldErr.Field() == tt.errField && fieldErr.Tag() == tt.errTag {
							found = true
							t.Logf("✅ 正确拦截了 %s 字段的 %s 验证错误", tt.errField, tt.errTag)
							break
						}
					}
					if !found {
						t.Errorf("期望字段 %s 的 %s 验证失败，但没有找到相关错误。实际错误: %v", tt.errField, tt.errTag, err)
					}
				} else {
					t.Errorf("期望ValidationErrors类型的错误，但得到: %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("期望验证通过，但验证失败: %v", err)
				} else {
					t.Logf("✅ 正常数据验证通过")
				}
			}
		})
	}
}

// TestUserModelString 测试用户模型String方法
func TestUserModelString(t *testing.T) {
	user := &model.SysUser{
		UserID:   1,
		UserName: "admin",
		NickName: "管理员",
		Email:    "admin@example.com",
		Status:   "0",
	}

	result := user.String()

	// 检查String方法是否包含关键信息
	expectedContains := []string{
		"SysUser{",
		"userId=1",
		"userName='admin'",
		"nickName='管理员'",
		"email='admin@example.com'",
		"status='0'",
		"}",
	}

	for _, expected := range expectedContains {
		if !contains(result, expected) {
			t.Errorf("String()方法结果应该包含 '%s'，但实际结果为: %s", expected, result)
		}
	}

	t.Logf("✅ String()方法测试通过: %s", result)
}

// TestUserModelSearchValue 测试searchValue字段
func TestUserModelSearchValue(t *testing.T) {
	user := &model.SysUser{
		UserName:    "testuser",
		SearchValue: "search_test",
	}

	// 验证searchValue字段存在且可以设置
	if user.SearchValue != "search_test" {
		t.Errorf("SearchValue字段设置失败，期望: search_test，实际: %s", user.SearchValue)
	} else {
		t.Logf("✅ SearchValue字段测试通过")
	}
}

// TestUserModelIsAdmin 测试IsAdmin方法
func TestUserModelIsAdmin(t *testing.T) {
	adminUser := &model.SysUser{UserID: 1}
	normalUser := &model.SysUser{UserID: 2}

	if !adminUser.IsAdmin() {
		t.Error("UserID=1的用户应该是管理员")
	} else {
		t.Logf("✅ 管理员判断测试通过")
	}

	if normalUser.IsAdmin() {
		t.Error("UserID=2的用户不应该是管理员")
	} else {
		t.Logf("✅ 普通用户判断测试通过")
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
