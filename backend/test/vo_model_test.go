package test

import (
	"testing"
	"wosm/internal/repository/model"
	"wosm/internal/service/system"
)

// TestVOModelComplete 测试VO模型的完整功能
func TestVOModelComplete(t *testing.T) {
	// 创建菜单服务
	menuService := system.NewMenuService()

	// 测试数据：模拟不同类型的菜单
	testMenus := []model.SysMenu{
		{
			MenuID:    1,
			MenuName:  "系统管理",
			ParentID:  0,
			OrderNum:  1,
			Path:      "system",
			Component: "",
			MenuType:  "M", // 目录
			Visible:   "0",
			Status:    "0",
			IsFrame:   "1",
			IsCache:   "0",
			Icon:      "system",
			RouteName: "System",
		},
		{
			MenuID:    2,
			MenuName:  "用户管理",
			ParentID:  1,
			OrderNum:  1,
			Path:      "user",
			Component: "system/user/index",
			MenuType:  "C", // 菜单
			Visible:   "0",
			Status:    "0",
			IsFrame:   "1",
			IsCache:   "0",
			Icon:      "user",
			RouteName: "User",
		},
		{
			MenuID:    3,
			MenuName:  "内链测试",
			ParentID:  0,
			OrderNum:  2,
			Path:      "http://www.example.com",
			Component: "",
			MenuType:  "C", // 菜单
			Visible:   "0",
			Status:    "0",
			IsFrame:   "1", // 内链
			IsCache:   "0",
			Icon:      "link",
			RouteName: "InnerLinkTest",
		},
		{
			MenuID:    4,
			MenuName:  "外链测试",
			ParentID:  0,
			OrderNum:  3,
			Path:      "https://www.baidu.com",
			Component: "",
			MenuType:  "C", // 菜单
			Visible:   "0",
			Status:    "0",
			IsFrame:   "0", // 外链
			IsCache:   "0",
			Icon:      "link",
			RouteName: "OuterLinkTest",
		},
		{
			MenuID:    5,
			MenuName:  "菜单框架测试",
			ParentID:  0,
			OrderNum:  4,
			Path:      "frame",
			Component: "system/frame/index",
			MenuType:  "C", // 菜单
			Visible:   "0",
			Status:    "0",
			IsFrame:   "1",
			IsCache:   "0",
			Icon:      "frame",
			RouteName: "Frame",
		},
	}

	// 构建子菜单关系
	for i := range testMenus {
		if testMenus[i].ParentID == 1 {
			testMenus[0].Children = append(testMenus[0].Children, testMenus[i])
		}
	}

	t.Run("测试基础路由构建", func(t *testing.T) {
		routers := menuService.BuildMenus(testMenus)

		if len(routers) == 0 {
			t.Error("路由构建失败，返回空数组")
			return
		}

		// 验证第一个路由（系统管理目录）
		systemRouter := routers[0]
		if systemRouter.Name != "System" {
			t.Errorf("系统管理路由名称错误，期望: System, 实际: %s", systemRouter.Name)
		}

		if systemRouter.Path != "/system" {
			t.Errorf("系统管理路由路径错误，期望: /system, 实际: %s", systemRouter.Path)
		}

		if !systemRouter.AlwaysShow {
			t.Error("系统管理路由应该设置AlwaysShow为true")
		}

		if len(systemRouter.Children) == 0 {
			t.Error("系统管理路由应该有子路由")
		}
	})

	t.Run("测试内链处理", func(t *testing.T) {
		routers := menuService.BuildMenus(testMenus)

		// 打印所有路由信息用于调试
		t.Logf("构建的路由数量: %d", len(routers))
		for i, router := range routers {
			t.Logf("路由%d: Name=%s, Path=%s, Component=%s", i, router.Name, router.Path, router.Component)
		}

		// 查找内链测试路由
		var innerLinkRouter *model.RouterVo
		for _, router := range routers {
			if router.Name == "InnerLinkTest" {
				innerLinkRouter = &router
				break
			}
		}

		if innerLinkRouter == nil {
			t.Error("未找到内链测试路由")
			return
		}

		// 验证内链处理逻辑
		// 根据Java后端逻辑，内链应该设置Path为"/"并创建子路由
		if innerLinkRouter.Path != "/" {
			t.Errorf("内链路由的Path应该为'/'，实际为: %s", innerLinkRouter.Path)
		}

		if len(innerLinkRouter.Children) == 0 {
			t.Error("内链路由应该有子路由")
		}
	})

	t.Run("测试外链处理", func(t *testing.T) {
		routers := menuService.BuildMenus(testMenus)

		// 查找外链测试路由
		var outerLinkRouter *model.RouterVo
		for _, router := range routers {
			if router.Name == "OuterLinkTest" {
				outerLinkRouter = &router
				break
			}
		}

		if outerLinkRouter == nil {
			t.Error("未找到外链测试路由")
			return
		}

		// 验证外链处理逻辑
		if outerLinkRouter.Meta == nil {
			t.Error("外链路由应该有Meta信息")
			return
		}

		if outerLinkRouter.Meta.Link != "https://www.baidu.com" {
			t.Errorf("外链路由的Link错误，期望: https://www.baidu.com, 实际: %s", outerLinkRouter.Meta.Link)
		}
	})

	t.Run("测试MetaVo字段", func(t *testing.T) {
		routers := menuService.BuildMenus(testMenus)

		if len(routers) == 0 {
			t.Error("路由构建失败")
			return
		}

		router := routers[0]
		if router.Meta == nil {
			t.Error("路由Meta信息不能为空")
			return
		}

		meta := router.Meta
		if meta.Title != "系统管理" {
			t.Errorf("Meta标题错误，期望: 系统管理, 实际: %s", meta.Title)
		}

		if meta.Icon != "system" {
			t.Errorf("Meta图标错误，期望: system, 实际: %s", meta.Icon)
		}

		if meta.NoCache != false {
			t.Errorf("Meta缓存设置错误，期望: false, 实际: %t", meta.NoCache)
		}
	})

	t.Log("VO模型测试完成")
}

// TestInnerLinkReplaceEach 测试内链域名特殊字符替换
func TestInnerLinkReplaceEach(t *testing.T) {
	menuService := system.NewMenuService()

	testCases := []struct {
		input    string
		expected string
	}{
		{"http://www.example.com", "example/com"},
		{"https://www.baidu.com", "baidu/com"},
		{"http://localhost:8080", "localhost/8080"},
		{"https://api.github.com/users", "api/github/com/users"},
	}

	for _, tc := range testCases {
		result := menuService.InnerLinkReplaceEach(tc.input)
		if result != tc.expected {
			t.Errorf("内链替换错误，输入: %s, 期望: %s, 实际: %s", tc.input, tc.expected, result)
		}
	}
}
