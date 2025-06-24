package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// TableSupport 常量 对应Java后端的TableSupport
const (
	PAGE_NUM        = "pageNum"       // 当前记录起始索引
	PAGE_SIZE       = "pageSize"      // 每页显示记录数
	ORDER_BY_COLUMN = "orderByColumn" // 排序列
	IS_ASC          = "isAsc"         // 排序的方向 "desc" 或者 "asc"
	REASONABLE      = "reasonable"    // 分页参数合理化
)

// PageDomain 分页数据 对应Java后端的PageDomain
type PageDomain struct {
	PageNum    int    `json:"pageNum"`    // 当前记录起始索引
	PageSize   int    `json:"pageSize"`   // 每页显示记录数
	OrderBy    string `json:"orderBy"`    // 排序列
	IsAsc      string `json:"isAsc"`      // 排序的方向desc或者asc
	Reasonable bool   `json:"reasonable"` // 分页合理化参数，默认false禁用
}

// PageInfo 分页信息 对应Java后端的PageInfo
type PageInfo struct {
	PageNum           int   `json:"pageNum"`           // 当前页
	PageSize          int   `json:"pageSize"`          // 每页的数量
	Size              int   `json:"size"`              // 当前页的数量
	StartRow          int   `json:"startRow"`          // 当前页面第一个元素在数据库中的行号
	EndRow            int   `json:"endRow"`            // 当前页面最后一个元素在数据库中的行号
	Total             int64 `json:"total"`             // 总记录数
	Pages             int   `json:"pages"`             // 总页数
	PrePage           int   `json:"prePage"`           // 前一页
	NextPage          int   `json:"nextPage"`          // 下一页
	IsFirstPage       bool  `json:"isFirstPage"`       // 是否为第一页
	IsLastPage        bool  `json:"isLastPage"`        // 是否为最后一页
	HasPrePage        bool  `json:"hasPrePage"`        // 是否有前一页
	HasNextPage       bool  `json:"hasNextPage"`       // 是否有下一页
	NavigatePages     int   `json:"navigatePages"`     // 导航页码数
	NavigatePageNums  []int `json:"navigatePageNums"`  // 所有导航页号
	NavigateFirstPage int   `json:"navigateFirstPage"` // 导航条上的第一页
	NavigateLastPage  int   `json:"navigateLastPage"`  // 导航条上的最后一页
}

// BuildPageRequest 构建分页请求 对应Java后端的TableSupport.buildPageRequest()
func BuildPageRequest(ctx *gin.Context) *PageDomain {
	pageNum, _ := strconv.Atoi(ctx.DefaultQuery(PAGE_NUM, "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery(PAGE_SIZE, "10"))

	// 兼容Java后端的orderByColumn参数名
	orderBy := ctx.Query("orderBy")
	if orderBy == "" {
		orderBy = ctx.Query(ORDER_BY_COLUMN) // 兼容Java后端参数名
	}

	isAsc := ctx.Query(IS_ASC)
	reasonable := ctx.DefaultQuery(REASONABLE, "true") == "true"

	// 限制分页大小
	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageNum < 1 {
		pageNum = 1
	}

	return &PageDomain{
		PageNum:    pageNum,
		PageSize:   pageSize,
		OrderBy:    orderBy,
		IsAsc:      isAsc,
		Reasonable: reasonable,
	}
}

// GetPageDomain 封装分页对象 对应Java后端的TableSupport.getPageDomain()
func GetPageDomain(ctx *gin.Context) *PageDomain {
	return BuildPageRequest(ctx)
}

// StartPage 设置请求分页数据 对应Java后端的PageUtils.startPage()
func StartPage(ctx *gin.Context) *PageDomain {
	return BuildPageRequest(ctx)
}

// StartOrderBy 设置请求排序数据 对应Java后端的BaseController.startOrderBy()
func StartOrderBy(ctx *gin.Context) string {
	pageDomain := BuildPageRequest(ctx)
	orderBy := pageDomain.GetOrderBy()
	if orderBy != "" {
		return EscapeOrderBySql(orderBy)
	}
	return ""
}

// ClearPage 清理分页的线程变量 对应Java后端的PageUtils.clearPage()
// Go语言无需清理线程变量，此函数为兼容性保留
func ClearPage() {
	// Go语言无需清理线程变量，此函数为兼容性保留
}

// GetOffset 获取偏移量
func (p *PageDomain) GetOffset() int {
	return (p.PageNum - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *PageDomain) GetLimit() int {
	return p.PageSize
}

// GetOrderBy 获取排序SQL 对应Java后端的PageDomain.getOrderBy()
func (p *PageDomain) GetOrderBy() string {
	if p.OrderBy == "" {
		return ""
	}

	// 转换驼峰命名为下划线命名
	orderByColumn := ToUnderScoreCase(p.OrderBy)

	// 处理排序方向兼容性
	isAsc := p.IsAsc
	if isAsc == "ascending" {
		isAsc = "asc"
	} else if isAsc == "descending" {
		isAsc = "desc"
	}

	// 默认升序
	if isAsc == "" {
		isAsc = "asc"
	}

	return orderByColumn + " " + isAsc
}

// SetOrderByColumn 设置排序字段 对应Java后端的setOrderByColumn
func (p *PageDomain) SetOrderByColumn(orderByColumn string) {
	p.OrderBy = orderByColumn
}

// SetIsAsc 设置排序方向 对应Java后端的setIsAsc
func (p *PageDomain) SetIsAsc(isAsc string) {
	if isAsc != "" {
		// 兼容前端排序类型
		if isAsc == "ascending" {
			isAsc = "asc"
		} else if isAsc == "descending" {
			isAsc = "desc"
		}
		p.IsAsc = isAsc
	}
}

// BuildPageInfo 构建分页信息 对应Java后端的PageInfo构造函数
func BuildPageInfo(list interface{}, total int64, pageNum, pageSize int) *PageInfo {
	pages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 计算当前页的数量
	size := pageSize
	if pageNum == pages && total%int64(pageSize) != 0 {
		size = int(total % int64(pageSize))
	}

	// 计算起始和结束行号
	startRow := (pageNum-1)*pageSize + 1
	endRow := startRow + size - 1

	// 计算前一页和下一页
	prePage := 0
	if pageNum > 1 {
		prePage = pageNum - 1
	}

	nextPage := 0
	if pageNum < pages {
		nextPage = pageNum + 1
	}

	// 计算导航页码
	navigatePages := 8 // 默认导航页码数
	navigatePageNums := make([]int, 0)
	navigateFirstPage := 1
	navigateLastPage := pages

	// 计算导航页码范围
	if pages <= navigatePages {
		for i := 1; i <= pages; i++ {
			navigatePageNums = append(navigatePageNums, i)
		}
	} else {
		startNum := pageNum - navigatePages/2
		endNum := pageNum + navigatePages/2

		if startNum < 1 {
			startNum = 1
			endNum = navigatePages
		}
		if endNum > pages {
			endNum = pages
			startNum = pages - navigatePages + 1
		}

		navigateFirstPage = startNum
		navigateLastPage = endNum

		for i := startNum; i <= endNum; i++ {
			navigatePageNums = append(navigatePageNums, i)
		}
	}

	return &PageInfo{
		PageNum:           pageNum,
		PageSize:          pageSize,
		Size:              size,
		StartRow:          startRow,
		EndRow:            endRow,
		Total:             total,
		Pages:             pages,
		PrePage:           prePage,
		NextPage:          nextPage,
		IsFirstPage:       pageNum == 1,
		IsLastPage:        pageNum == pages,
		HasPrePage:        pageNum > 1,
		HasNextPage:       pageNum < pages,
		NavigatePages:     navigatePages,
		NavigatePageNums:  navigatePageNums,
		NavigateFirstPage: navigateFirstPage,
		NavigateLastPage:  navigateLastPage,
	}
}

// ToUnderScoreCase 驼峰转下划线 对应Java后端的StringUtils.toUnderScoreCase
func ToUnderScoreCase(str string) string {
	if str == "" {
		return ""
	}

	// 使用正则表达式将驼峰命名转换为下划线命名
	// 在大写字母前插入下划线，然后转为小写
	result := ""
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += "_"
		}
		result += strings.ToLower(string(r))
	}

	return result
}

// EscapeOrderBySql 转义排序SQL 对应Java后端的SqlUtil.escapeOrderBySql
func EscapeOrderBySql(orderBy string) string {
	if orderBy == "" {
		return ""
	}

	// 先验证是否为有效的排序表达式
	if !isValidOrderBy(orderBy) {
		return ""
	}

	// 检查是否包含危险字符和关键词（完整单词匹配）
	dangerousChars := []string{";", "--", "/*", "*/", "'", "\"", "\\", "<", ">"}
	dangerousKeywords := []string{
		"exec", "execute", "insert", "update", "delete",
		"drop", "alter", "union", "select", "xp_", "sp_",
	}

	result := orderBy

	// 移除危险字符
	for _, danger := range dangerousChars {
		result = strings.ReplaceAll(result, danger, "")
	}

	// 检查危险关键词（作为独立单词）
	lowerResult := strings.ToLower(result)
	for _, keyword := range dangerousKeywords {
		// 使用正则表达式确保是完整单词匹配
		pattern := `\b` + regexp.QuoteMeta(keyword) + `\b`
		if matched, _ := regexp.MatchString(pattern, lowerResult); matched {
			return "" // 如果包含危险关键词，返回空字符串
		}
	}

	return result
}

// isValidOrderBy 验证排序表达式是否有效
func isValidOrderBy(orderBy string) bool {
	if orderBy == "" {
		return false
	}

	// 基本的排序表达式验证
	// 允许格式: field_name asc/desc, field_name2 asc/desc
	parts := strings.Split(orderBy, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查每个排序字段
		tokens := strings.Fields(part)
		if len(tokens) < 1 || len(tokens) > 2 {
			return false
		}

		// 检查字段名（只允许字母、数字、下划线）
		fieldName := tokens[0]
		if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(fieldName) {
			return false
		}

		// 检查排序方向
		if len(tokens) == 2 {
			direction := strings.ToLower(tokens[1])
			if direction != "asc" && direction != "desc" {
				return false
			}
		}
	}

	return true
}
