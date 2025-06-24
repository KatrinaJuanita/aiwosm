package datascope

import (
	"testing"
	"wosm/internal/repository/model"
)

// TestDataScopeSelfWithCreator 测试基于创建者的"仅本人数据"权限
func TestDataScopeSelfWithCreator(t *testing.T) {
	// 创建测试用户
	user := &model.SysUser{
		UserID:   123,
		UserName: "testuser",
		DeptID:   func() *int64 { id := int64(100); return &id }(),
		Roles: []model.SysRole{
			{
				RoleID:    2,
				DataScope: DataScopeSelf, // 仅本人数据权限
				Status:    RoleStatusNormal,
				Permissions: []string{"system:notice:list"},
			},
		},
	}

	tests := []struct {
		name         string
		config       *DataScopeConfig
		expectedSQL  string
		description  string
	}{
		{
			name: "仅用户ID权限",
			config: &DataScopeConfig{
				DeptAlias:  "d",
				UserAlias:  "u",
				Permission: "system:notice:list",
			},
			expectedSQL: "u.user_id = 123",
			description: "只配置UserAlias时，使用user_id字段",
		},
		{
			name: "仅创建者权限",
			config: &DataScopeConfig{
				DeptAlias:    "d",
				CreatorAlias: "c",
				Permission:   "system:notice:list",
			},
			expectedSQL: "c.create_by = 'testuser'",
			description: "只配置CreatorAlias时，使用create_by字段",
		},
		{
			name: "混合权限",
			config: &DataScopeConfig{
				DeptAlias:    "d",
				UserAlias:    "u",
				CreatorAlias: "c",
				Permission:   "system:notice:list",
			},
			expectedSQL: "u.user_id = 123 OR c.create_by = 'testuser'",
			description: "同时配置UserAlias和CreatorAlias时，使用OR连接",
		},
		{
			name: "无权限别名",
			config: &DataScopeConfig{
				DeptAlias:  "d",
				Permission: "system:notice:list",
			},
			expectedSQL: "d.dept_id = 0",
			description: "没有配置任何权限别名时，返回限制性条件",
		},
	}

	processor := NewDataScopeProcessor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualSQL := processor.generateDataScopeSQL(user, tt.config)
			
			if actualSQL != tt.expectedSQL {
				t.Errorf("测试 %s 失败:\n期望SQL: %s\n实际SQL: %s\n描述: %s", 
					tt.name, tt.expectedSQL, actualSQL, tt.description)
			} else {
				t.Logf("测试 %s 成功: %s", tt.name, tt.description)
			}
		})
	}
}

// TestDataScopeWithDifferentRoles 测试不同角色的数据权限
func TestDataScopeWithDifferentRoles(t *testing.T) {
	user := &model.SysUser{
		UserID:   123,
		UserName: "testuser",
		DeptID:   func() *int64 { id := int64(100); return &id }(),
	}

	config := &DataScopeConfig{
		DeptAlias:    "d",
		CreatorAlias: "c",
		Permission:   "system:notice:list",
	}

	tests := []struct {
		name        string
		roles       []model.SysRole
		expectedSQL string
		description string
	}{
		{
			name: "全部数据权限",
			roles: []model.SysRole{
				{
					RoleID:      1,
					DataScope:   DataScopeAll,
					Status:      RoleStatusNormal,
					Permissions: []string{"system:notice:list"},
				},
			},
			expectedSQL: "",
			description: "全部数据权限不添加任何条件",
		},
		{
			name: "本部门数据权限",
			roles: []model.SysRole{
				{
					RoleID:      2,
					DataScope:   DataScopeDept,
					Status:      RoleStatusNormal,
					Permissions: []string{"system:notice:list"},
				},
			},
			expectedSQL: "d.dept_id = 100",
			description: "本部门数据权限基于用户部门ID",
		},
		{
			name: "仅本人数据权限（创建者）",
			roles: []model.SysRole{
				{
					RoleID:      3,
					DataScope:   DataScopeSelf,
					Status:      RoleStatusNormal,
					Permissions: []string{"system:notice:list"},
				},
			},
			expectedSQL: "c.create_by = 'testuser'",
			description: "仅本人数据权限基于创建者字段",
		},
	}

	processor := NewDataScopeProcessor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user.Roles = tt.roles
			actualSQL := processor.generateDataScopeSQL(user, config)
			
			if actualSQL != tt.expectedSQL {
				t.Errorf("测试 %s 失败:\n期望SQL: %s\n实际SQL: %s\n描述: %s", 
					tt.name, tt.expectedSQL, actualSQL, tt.description)
			} else {
				t.Logf("测试 %s 成功: %s", tt.name, tt.description)
			}
		})
	}
}

// TestApplyDataScopeWithCreator 测试ApplyDataScopeWithCreator函数
func TestApplyDataScopeWithCreator(t *testing.T) {
	user := &model.SysUser{
		UserID:   123,
		UserName: "testuser",
		DeptID:   func() *int64 { id := int64(100); return &id }(),
		Roles: []model.SysRole{
			{
				RoleID:      2,
				DataScope:   DataScopeSelf,
				Status:      RoleStatusNormal,
				Permissions: []string{"system:notice:list"},
			},
		},
	}

	params := make(map[string]interface{})

	err := ApplyDataScopeWithCreator(user, "d", "", "c", "system:notice:list", params)
	if err != nil {
		t.Fatalf("ApplyDataScopeWithCreator 失败: %v", err)
	}

	dataScope, exists := params["dataScope"]
	if !exists {
		t.Fatal("数据权限参数未设置")
	}

	expectedSQL := " AND (c.create_by = 'testuser')"
	if dataScope != expectedSQL {
		t.Errorf("数据权限SQL不正确:\n期望: %s\n实际: %s", expectedSQL, dataScope)
	} else {
		t.Logf("ApplyDataScopeWithCreator 测试成功: %s", dataScope)
	}
}
