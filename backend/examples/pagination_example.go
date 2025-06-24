// 分页功能示例
// 使用方法：将此文件复制到单独目录并运行 go run pagination_example.go
package examples

import (
	"fmt"
	"net/http"

	"wosm/pkg/response"
	"wosm/pkg/utils"

	"github.com/gin-gonic/gin"
)

// User 示例用户结构
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 模拟用户数据
var users = []User{
	{1, "admin", "admin@example.com"},
	{2, "user1", "user1@example.com"},
	{3, "user2", "user2@example.com"},
	{4, "user3", "user3@example.com"},
	{5, "user4", "user4@example.com"},
	{6, "user5", "user5@example.com"},
	{7, "user6", "user6@example.com"},
	{8, "user7", "user7@example.com"},
	{9, "user8", "user8@example.com"},
	{10, "user9", "user9@example.com"},
}

// RunPaginationExample 运行分页功能示例
// 注意：此函数已重命名以避免与主程序的main函数冲突
// 如需运行此示例，请将此文件复制到单独目录并重命名函数为main
func RunPaginationExample() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 分页查询示例 - 使用修复后的分页功能
	r.GET("/users", getUserList)

	// 分页查询示例 - 兼容Java后端参数名
	r.GET("/users/java-compatible", getUserListJavaCompatible)

	// 排序查询示例
	r.GET("/users/sorted", getUserListSorted)

	fmt.Println("分页功能示例服务器启动在 :8081")
	fmt.Println("测试URL:")
	fmt.Println("1. 基础分页: http://localhost:8081/users?pageNum=1&pageSize=3")
	fmt.Println("2. Java兼容: http://localhost:8081/users/java-compatible?pageNum=1&pageSize=3&orderByColumn=username&isAsc=desc")
	fmt.Println("3. 排序查询: http://localhost:8081/users/sorted?pageNum=1&pageSize=3&orderBy=username&isAsc=asc")

	r.Run(":8081")
}

// getUserList 获取用户列表 - 基础分页功能
func getUserList(c *gin.Context) {
	// 使用修复后的分页功能
	pageDomain := utils.StartPage(c)

	fmt.Printf("分页参数: PageNum=%d, PageSize=%d, OrderBy=%s, IsAsc=%s\n",
		pageDomain.PageNum, pageDomain.PageSize, pageDomain.OrderBy, pageDomain.IsAsc)

	// 模拟分页查询
	offset := pageDomain.GetOffset()
	limit := pageDomain.GetLimit()

	// 获取分页数据
	var pageUsers []User
	total := int64(len(users))

	start := offset
	end := offset + limit
	if start >= len(users) {
		pageUsers = []User{}
	} else {
		if end > len(users) {
			end = len(users)
		}
		pageUsers = users[start:end]
	}

	// 使用修复后的TableDataInfo
	tableData := response.NewTableDataInfoWithData(pageUsers, total)

	// 发送响应
	response.SendTableDataInfo(c, tableData)
}

// getUserListJavaCompatible 获取用户列表 - Java后端兼容模式
func getUserListJavaCompatible(c *gin.Context) {
	// 测试Java后端参数兼容性
	pageDomain := utils.BuildPageRequest(c)

	fmt.Printf("Java兼容模式 - 分页参数: PageNum=%d, PageSize=%d, OrderBy=%s, IsAsc=%s\n",
		pageDomain.PageNum, pageDomain.PageSize, pageDomain.OrderBy, pageDomain.IsAsc)

	// 测试排序SQL构建
	orderBySQL := pageDomain.GetOrderBy()
	if orderBySQL != "" {
		orderBySQL = utils.EscapeOrderBySql(orderBySQL)
		fmt.Printf("排序SQL: %s\n", orderBySQL)
	}

	// 模拟分页查询
	offset := pageDomain.GetOffset()
	limit := pageDomain.GetLimit()

	var pageUsers []User
	total := int64(len(users))

	start := offset
	end := offset + limit
	if start >= len(users) {
		pageUsers = []User{}
	} else {
		if end > len(users) {
			end = len(users)
		}
		pageUsers = users[start:end]
	}

	// 使用PageResult响应
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "查询成功",
		"total": total,
		"rows":  pageUsers,
	})
}

// getUserListSorted 获取用户列表 - 排序功能测试
func getUserListSorted(c *gin.Context) {
	// 测试排序功能
	pageDomain := utils.StartPage(c)

	fmt.Printf("排序测试 - 分页参数: PageNum=%d, PageSize=%d, OrderBy=%s, IsAsc=%s\n",
		pageDomain.PageNum, pageDomain.PageSize, pageDomain.OrderBy, pageDomain.IsAsc)

	// 测试字段名转换
	if pageDomain.OrderBy != "" {
		underscoreField := utils.ToUnderScoreCase(pageDomain.OrderBy)
		fmt.Printf("字段名转换: %s -> %s\n", pageDomain.OrderBy, underscoreField)
	}

	// 测试完整的排序SQL构建
	orderBySQL := pageDomain.GetOrderBy()
	if orderBySQL != "" {
		escapedSQL := utils.EscapeOrderBySql(orderBySQL)
		fmt.Printf("完整排序SQL: %s -> %s\n", orderBySQL, escapedSQL)
	}

	// 测试StartOrderBy函数
	orderBy := utils.StartOrderBy(c)
	if orderBy != "" {
		fmt.Printf("StartOrderBy结果: %s\n", orderBy)
	}

	// 模拟分页查询
	offset := pageDomain.GetOffset()
	limit := pageDomain.GetLimit()

	var pageUsers []User
	total := int64(len(users))

	start := offset
	end := offset + limit
	if start >= len(users) {
		pageUsers = []User{}
	} else {
		if end > len(users) {
			end = len(users)
		}
		pageUsers = users[start:end]
	}

	// 构建完整的PageInfo
	pageInfo := utils.BuildPageInfo(pageUsers, total, pageDomain.PageNum, pageDomain.PageSize)

	// 返回完整的分页信息
	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"msg":      "查询成功",
		"pageInfo": pageInfo,
	})
}
