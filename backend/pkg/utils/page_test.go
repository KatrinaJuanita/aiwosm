package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestToUnderScoreCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"userName", "user_name"},
		{"createTime", "create_time"},
		{"userId", "user_id"},
		{"ID", "i_d"},
		{"", ""},
		{"name", "name"},
		{"UserName", "user_name"},
	}

	for _, test := range tests {
		result := ToUnderScoreCase(test.input)
		assert.Equal(t, test.expected, result, "ToUnderScoreCase(%s) should return %s", test.input, test.expected)
	}
}

func TestPageDomainGetOrderBy(t *testing.T) {
	tests := []struct {
		orderBy  string
		isAsc    string
		expected string
	}{
		{"userName", "asc", "user_name asc"},
		{"createTime", "desc", "create_time desc"},
		{"userId", "ascending", "user_id asc"},
		{"userName", "descending", "user_name desc"},
		{"", "asc", ""},
		{"userName", "", "user_name asc"},
	}

	for _, test := range tests {
		pageDomain := &PageDomain{
			OrderBy: test.orderBy,
			IsAsc:   test.isAsc,
		}
		result := pageDomain.GetOrderBy()
		assert.Equal(t, test.expected, result, "GetOrderBy() with orderBy=%s, isAsc=%s should return %s", test.orderBy, test.isAsc, test.expected)
	}
}

func TestEscapeOrderBySql(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_name asc", "user_name asc"},
		{"user_name desc", "user_name desc"},
		{"user_name; drop table users", ""}, // 无效表达式，返回空
		{"user_name' or 1=1", ""},           // 无效表达式，返回空
		{"", ""},
		{"invalid-field", ""}, // 包含连字符，无效
		{"user_name, create_time desc", "user_name, create_time desc"},
		{"user_name", "user_name"}, // 只有字段名，有效
	}

	for _, test := range tests {
		result := EscapeOrderBySql(test.input)
		if test.expected == "" {
			assert.Empty(t, result, "EscapeOrderBySql(%s) should return empty string", test.input)
		} else {
			assert.Equal(t, test.expected, result, "EscapeOrderBySql(%s) should return %s", test.input, test.expected)
		}
	}
}

func TestBuildPageRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		queryParams  map[string]string
		expectedPage *PageDomain
	}{
		{
			name:        "默认参数",
			queryParams: map[string]string{},
			expectedPage: &PageDomain{
				PageNum:    1,
				PageSize:   10,
				OrderBy:    "",
				IsAsc:      "",
				Reasonable: true,
			},
		},
		{
			name: "自定义参数",
			queryParams: map[string]string{
				"pageNum":  "2",
				"pageSize": "20",
				"orderBy":  "userName",
				"isAsc":    "desc",
			},
			expectedPage: &PageDomain{
				PageNum:    2,
				PageSize:   20,
				OrderBy:    "userName",
				IsAsc:      "desc",
				Reasonable: true,
			},
		},
		{
			name: "兼容orderByColumn参数",
			queryParams: map[string]string{
				"pageNum":       "1",
				"pageSize":      "15",
				"orderByColumn": "createTime",
				"isAsc":         "asc",
			},
			expectedPage: &PageDomain{
				PageNum:    1,
				PageSize:   15,
				OrderBy:    "createTime",
				IsAsc:      "asc",
				Reasonable: true,
			},
		},
		{
			name: "页大小限制测试",
			queryParams: map[string]string{
				"pageNum":  "1",
				"pageSize": "200", // 超过100的限制
			},
			expectedPage: &PageDomain{
				PageNum:    1,
				PageSize:   100, // 应该被限制为100
				OrderBy:    "",
				IsAsc:      "",
				Reasonable: true,
			},
		},
		{
			name: "页码最小值测试",
			queryParams: map[string]string{
				"pageNum":  "0", // 小于1
				"pageSize": "10",
			},
			expectedPage: &PageDomain{
				PageNum:    1, // 应该被设置为1
				PageSize:   10,
				OrderBy:    "",
				IsAsc:      "",
				Reasonable: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/test", nil)
			q := req.URL.Query()
			for key, value := range test.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// 创建Gin上下文
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// 测试BuildPageRequest
			result := BuildPageRequest(ctx)

			assert.Equal(t, test.expectedPage.PageNum, result.PageNum)
			assert.Equal(t, test.expectedPage.PageSize, result.PageSize)
			assert.Equal(t, test.expectedPage.OrderBy, result.OrderBy)
			assert.Equal(t, test.expectedPage.IsAsc, result.IsAsc)
			assert.Equal(t, test.expectedPage.Reasonable, result.Reasonable)
		})
	}
}

func TestBuildPageInfo(t *testing.T) {
	// 测试数据
	list := []string{"item1", "item2", "item3"}
	total := int64(25)
	pageNum := 2
	pageSize := 10

	pageInfo := BuildPageInfo(list, total, pageNum, pageSize)

	assert.Equal(t, pageNum, pageInfo.PageNum)
	assert.Equal(t, pageSize, pageInfo.PageSize)
	assert.Equal(t, total, pageInfo.Total)
	assert.Equal(t, 3, pageInfo.Pages)                 // 25条记录，每页10条，共3页
	assert.Equal(t, 10, pageInfo.Size)                 // 当前页大小
	assert.Equal(t, 11, pageInfo.StartRow)             // 第二页起始行号
	assert.Equal(t, 20, pageInfo.EndRow)               // 第二页结束行号
	assert.Equal(t, 1, pageInfo.PrePage)               // 前一页
	assert.Equal(t, 3, pageInfo.NextPage)              // 下一页
	assert.False(t, pageInfo.IsFirstPage)              // 不是第一页
	assert.False(t, pageInfo.IsLastPage)               // 不是最后一页
	assert.True(t, pageInfo.HasPrePage)                // 有前一页
	assert.True(t, pageInfo.HasNextPage)               // 有下一页
	assert.Equal(t, 8, pageInfo.NavigatePages)         // 导航页数
	assert.True(t, len(pageInfo.NavigatePageNums) > 0) // 导航页码数组不为空
}

func TestIsValidOrderBy(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"user_name", true},
		{"user_name asc", true},
		{"user_name desc", true},
		{"user_name asc, create_time desc", true},
		{"", false},
		{"user-name", false},             // 包含连字符
		{"user_name; drop table", false}, // 包含分号
		{"user_name or 1=1", false},      // 包含SQL关键词
		{"123field", false},              // 以数字开头
		{"_field", true},                 // 以下划线开头（允许）
		{"field123", true},               // 包含数字（允许）
	}

	for _, test := range tests {
		result := isValidOrderBy(test.input)
		assert.Equal(t, test.expected, result, "isValidOrderBy(%s) should return %t", test.input, test.expected)
	}
}
