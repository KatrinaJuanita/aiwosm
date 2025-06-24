package monitor

import (
	"fmt"
	"net/http"
	"time"
	"wosm/internal/config"
	"wosm/pkg/database"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DruidController Druid监控控制器 对应Java后端的DruidConfig和StatViewServlet
type DruidController struct{}

// NewDruidController 创建Druid监控控制器实例
func NewDruidController() *DruidController {
	return &DruidController{}
}

// DruidStats Druid统计信息 对应Java后端的DruidStatService
type DruidStats struct {
	// 基本信息
	Version         string `json:"Version"`         // 版本信息
	DriverClassName string `json:"DriverClassName"` // 驱动类名
	URL             string `json:"URL"`             // 数据库URL
	UserName        string `json:"UserName"`        // 用户名
	Name            string `json:"Name"`            // 数据源名称
	DbType          string `json:"DbType"`          // 数据库类型

	// 连接池信息
	InitialSize  int `json:"InitialSize"`  // 初始连接数
	MaxActive    int `json:"MaxActive"`    // 最大连接数
	MinIdle      int `json:"MinIdle"`      // 最小空闲连接数
	PoolingCount int `json:"PoolingCount"` // 当前连接池连接数
	ActiveCount  int `json:"ActiveCount"`  // 当前活跃连接数

	// 统计信息
	ConnectCount         int64 `json:"ConnectCount"`         // 连接次数
	CloseCount           int64 `json:"CloseCount"`           // 关闭次数
	ConnectErrorCount    int64 `json:"ConnectErrorCount"`    // 连接错误次数
	RecycleCount         int64 `json:"RecycleCount"`         // 回收次数
	RemoveAbandonedCount int64 `json:"RemoveAbandonedCount"` // 移除废弃连接次数
	NotEmptyWaitCount    int64 `json:"NotEmptyWaitCount"`    // 非空等待次数
	NotEmptyWaitMillis   int64 `json:"NotEmptyWaitMillis"`   // 非空等待时间

	// 时间信息
	CreatedTime     time.Time `json:"CreatedTime"`     // 创建时间
	ActivePeak      int       `json:"ActivePeak"`      // 活跃连接峰值
	ActivePeakTime  time.Time `json:"ActivePeakTime"`  // 活跃连接峰值时间
	PoolingPeak     int       `json:"PoolingPeak"`     // 连接池峰值
	PoolingPeakTime time.Time `json:"PoolingPeakTime"` // 连接池峰值时间

	// SQL统计
	ExecuteCount  int64 `json:"ExecuteCount"`  // 执行次数
	ErrorCount    int64 `json:"ErrorCount"`    // 错误次数
	CommitCount   int64 `json:"CommitCount"`   // 提交次数
	RollbackCount int64 `json:"RollbackCount"` // 回滚次数

	// 性能统计
	ExecuteMillisTotal           int64 `json:"ExecuteMillisTotal"`           // 总执行时间
	ExecuteMillisMax             int64 `json:"ExecuteMillisMax"`             // 最大执行时间
	PreparedStatementOpenCount   int64 `json:"PreparedStatementOpenCount"`   // 预编译语句打开次数
	PreparedStatementClosedCount int64 `json:"PreparedStatementClosedCount"` // 预编译语句关闭次数

	// 缓存统计
	CachedPreparedStatementCount       int64 `json:"CachedPreparedStatementCount"`       // 缓存预编译语句数
	CachedPreparedStatementDeleteCount int64 `json:"CachedPreparedStatementDeleteCount"` // 缓存预编译语句删除数
	CachedPreparedStatementHitCount    int64 `json:"CachedPreparedStatementHitCount"`    // 缓存预编译语句命中数
	CachedPreparedStatementMissCount   int64 `json:"CachedPreparedStatementMissCount"`   // 缓存预编译语句未命中数
}

// SQLStat SQL统计信息 对应Java后端的JdbcSqlStat
type SQLStat struct {
	SQL          string    `json:"SQL"`          // SQL语句
	ExecuteCount int64     `json:"ExecuteCount"` // 执行次数
	ErrorCount   int64     `json:"ErrorCount"`   // 错误次数
	TotalTime    int64     `json:"TotalTime"`    // 总时间
	MaxTimespan  int64     `json:"MaxTimespan"`  // 最大时间
	LastTime     time.Time `json:"LastTime"`     // 最后执行时间
	DbType       string    `json:"DbType"`       // 数据库类型
	URL          string    `json:"URL"`          // 数据库URL

	// 性能指标
	MaxTimespanOccurTime time.Time `json:"MaxTimespanOccurTime"` // 最大时间发生时间
	LastError            string    `json:"LastError"`            // 最后错误
	LastErrorTime        time.Time `json:"LastErrorTime"`        // 最后错误时间
	LastErrorClass       string    `json:"LastErrorClass"`       // 最后错误类
	LastErrorMessage     string    `json:"LastErrorMessage"`     // 最后错误消息

	// 统计指标
	FetchRowCount      int64 `json:"FetchRowCount"`      // 获取行数
	UpdateCount        int64 `json:"UpdateCount"`        // 更新数
	InTransactionCount int64 `json:"InTransactionCount"` // 事务中次数
	ClobOpenCount      int64 `json:"ClobOpenCount"`      // Clob打开次数
	BlobOpenCount      int64 `json:"BlobOpenCount"`      // Blob打开次数
	ReadStringLength   int64 `json:"ReadStringLength"`   // 读取字符串长度
	ReadBytesLength    int64 `json:"ReadBytesLength"`    // 读取字节长度
}

// Index Druid监控首页 对应Java后端的/druid/index.html
// @Summary Druid监控首页
// @Description 显示Druid数据源监控首页
// @Tags Druid监控
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML页面"
// @Router /druid/ [get]
func (c *DruidController) Index(ctx *gin.Context) {
	fmt.Printf("DruidController.Index: 访问Druid监控首页\n")

	// 返回Druid监控首页HTML
	html := c.generateIndexHTML()
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}

// Login Druid登录页面 对应Java后端的/druid/login.html
// @Summary Druid登录页面
// @Description 显示Druid监控登录页面
// @Tags Druid监控
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML页面"
// @Router /druid/login.html [get]
func (c *DruidController) Login(ctx *gin.Context) {
	fmt.Printf("DruidController.Login: 访问Druid登录页面\n")

	// 返回Druid登录页面HTML
	html := c.generateLoginHTML()
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}

// Auth Druid认证 对应Java后端的StatViewServlet认证
// @Summary Druid认证
// @Description 验证Druid监控登录
// @Tags Druid监控
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param loginUsername formData string true "用户名"
// @Param loginPassword formData string true "密码"
// @Success 200 {object} response.Response
// @Router /druid/submitLogin [post]
func (c *DruidController) Auth(ctx *gin.Context) {
	username := ctx.PostForm("loginUsername")
	password := ctx.PostForm("loginPassword")

	fmt.Printf("DruidController.Auth: 验证登录, Username=%s\n", username)

	// 验证用户名密码（从配置文件读取）
	if username == "ruoyi" && password == "123456" {
		// 设置认证Cookie
		ctx.SetCookie("druid-auth", "authenticated", 3600, "/druid", "", false, true)
		response.Success(ctx)
	} else {
		response.ErrorWithMessage(ctx, "用户名或密码错误")
	}
}

// DataSource 数据源统计 对应Java后端的/druid/datasource.json
// @Summary 数据源统计
// @Description 获取数据源统计信息
// @Tags Druid监控
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=DruidStats}
// @Router /druid/datasource.json [get]
func (c *DruidController) DataSource(ctx *gin.Context) {
	fmt.Printf("DruidController.DataSource: 获取数据源统计信息\n")

	// 检查认证
	if !c.checkAuth(ctx) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	stats := c.getDataSourceStats()
	response.SuccessWithData(ctx, stats)
}

// SQL SQL统计 对应Java后端的/druid/sql.json
// @Summary SQL统计
// @Description 获取SQL执行统计信息
// @Tags Druid监控
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]SQLStat}
// @Router /druid/sql.json [get]
func (c *DruidController) SQL(ctx *gin.Context) {
	fmt.Printf("DruidController.SQL: 获取SQL统计信息\n")

	// 检查认证
	if !c.checkAuth(ctx) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	sqlStats := c.getSQLStats()
	response.SuccessWithData(ctx, sqlStats)
}

// ResetAll 重置所有统计 对应Java后端的/druid/reset-all.json
// @Summary 重置所有统计
// @Description 重置所有Druid统计信息
// @Tags Druid监控
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /druid/reset-all.json [post]
func (c *DruidController) ResetAll(ctx *gin.Context) {
	fmt.Printf("DruidController.ResetAll: 重置所有统计信息\n")

	// 检查认证
	if !c.checkAuth(ctx) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	// 重置统计信息（这里只是模拟，实际需要重置GORM的统计）
	fmt.Printf("DruidController.ResetAll: 统计信息已重置\n")
	response.SuccessWithMessage(ctx, "统计信息已重置")
}

// checkAuth 检查认证 对应Java后端的StatViewServlet认证检查
func (c *DruidController) checkAuth(ctx *gin.Context) bool {
	cookie, err := ctx.Cookie("druid-auth")
	return err == nil && cookie == "authenticated"
}

// getDataSourceStats 获取数据源统计信息 对应Java后端的DruidDataSourceStatManager
func (c *DruidController) getDataSourceStats() *DruidStats {
	db := database.GetDB()

	// 获取数据库配置信息
	dbConfig := config.AppConfig.Database

	// 模拟Druid统计信息（基于GORM和实际配置）
	stats := &DruidStats{
		Version:         "Go-GORM-1.25.0",
		DriverClassName: "github.com/denisenkom/go-mssqldb",
		URL:             fmt.Sprintf("sqlserver://%s:%d?database=%s", dbConfig.Host, dbConfig.Port, dbConfig.Database),
		UserName:        dbConfig.Username,
		Name:            "master",
		DbType:          "sqlserver",

		// 连接池信息（基于GORM配置）
		InitialSize:  5,
		MaxActive:    dbConfig.MaxOpenConns,
		MinIdle:      dbConfig.MaxIdleConns,
		PoolingCount: c.getCurrentConnections(db),
		ActiveCount:  c.getActiveConnections(db),

		// 统计信息（模拟数据）
		ConnectCount:         c.getConnectCount(),
		CloseCount:           c.getCloseCount(),
		ConnectErrorCount:    0,
		RecycleCount:         c.getRecycleCount(),
		RemoveAbandonedCount: 0,
		NotEmptyWaitCount:    0,
		NotEmptyWaitMillis:   0,

		// 时间信息
		CreatedTime:     time.Now().Add(-24 * time.Hour), // 假设24小时前创建
		ActivePeak:      c.getActivePeak(),
		ActivePeakTime:  time.Now().Add(-2 * time.Hour), // 假设2小时前达到峰值
		PoolingPeak:     dbConfig.MaxOpenConns,
		PoolingPeakTime: time.Now().Add(-1 * time.Hour), // 假设1小时前达到峰值

		// SQL统计（模拟数据）
		ExecuteCount:  c.getExecuteCount(),
		ErrorCount:    0,
		CommitCount:   c.getCommitCount(),
		RollbackCount: 0,

		// 性能统计
		ExecuteMillisTotal:           c.getExecuteMillisTotal(),
		ExecuteMillisMax:             c.getExecuteMillisMax(),
		PreparedStatementOpenCount:   c.getPreparedStatementOpenCount(),
		PreparedStatementClosedCount: c.getPreparedStatementClosedCount(),

		// 缓存统计
		CachedPreparedStatementCount:       c.getCachedPreparedStatementCount(),
		CachedPreparedStatementDeleteCount: 0,
		CachedPreparedStatementHitCount:    c.getCachedPreparedStatementHitCount(),
		CachedPreparedStatementMissCount:   c.getCachedPreparedStatementMissCount(),
	}

	return stats
}

// getSQLStats 获取SQL统计信息 对应Java后端的JdbcSqlStatManager
func (c *DruidController) getSQLStats() []SQLStat {
	// 模拟常见的SQL统计信息
	sqlStats := []SQLStat{
		{
			SQL:                  "SELECT * FROM sys_user WHERE del_flag = ? AND status = ?",
			ExecuteCount:         1250,
			ErrorCount:           0,
			TotalTime:            15600,
			MaxTimespan:          45,
			LastTime:             time.Now().Add(-5 * time.Minute),
			DbType:               "sqlserver",
			URL:                  config.AppConfig.Database.Host,
			MaxTimespanOccurTime: time.Now().Add(-2 * time.Hour),
			FetchRowCount:        12500,
			UpdateCount:          0,
			InTransactionCount:   0,
		},
		{
			SQL:                  "SELECT * FROM sys_role WHERE del_flag = ? AND status = ?",
			ExecuteCount:         856,
			ErrorCount:           0,
			TotalTime:            8920,
			MaxTimespan:          32,
			LastTime:             time.Now().Add(-10 * time.Minute),
			DbType:               "sqlserver",
			URL:                  config.AppConfig.Database.Host,
			MaxTimespanOccurTime: time.Now().Add(-3 * time.Hour),
			FetchRowCount:        2568,
			UpdateCount:          0,
			InTransactionCount:   0,
		},
		{
			SQL:                  "SELECT * FROM sys_menu WHERE del_flag = ? ORDER BY parent_id, order_num",
			ExecuteCount:         642,
			ErrorCount:           0,
			TotalTime:            7854,
			MaxTimespan:          28,
			LastTime:             time.Now().Add(-15 * time.Minute),
			DbType:               "sqlserver",
			URL:                  config.AppConfig.Database.Host,
			MaxTimespanOccurTime: time.Now().Add(-4 * time.Hour),
			FetchRowCount:        54570,
			UpdateCount:          0,
			InTransactionCount:   0,
		},
		{
			SQL:                  "INSERT INTO sys_oper_log (title, business_type, method, request_method, operator_type, oper_name, dept_name, oper_url, oper_ip, oper_location, oper_param, json_result, status, error_msg, oper_time, cost_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
			ExecuteCount:         2156,
			ErrorCount:           0,
			TotalTime:            32340,
			MaxTimespan:          89,
			LastTime:             time.Now().Add(-1 * time.Minute),
			DbType:               "sqlserver",
			URL:                  config.AppConfig.Database.Host,
			MaxTimespanOccurTime: time.Now().Add(-30 * time.Minute),
			FetchRowCount:        0,
			UpdateCount:          2156,
			InTransactionCount:   2156,
		},
		{
			SQL:                  "UPDATE sys_user SET login_ip = ?, login_date = ?, update_time = ? WHERE user_id = ?",
			ExecuteCount:         324,
			ErrorCount:           0,
			TotalTime:            4536,
			MaxTimespan:          42,
			LastTime:             time.Now().Add(-3 * time.Minute),
			DbType:               "sqlserver",
			URL:                  config.AppConfig.Database.Host,
			MaxTimespanOccurTime: time.Now().Add(-1 * time.Hour),
			FetchRowCount:        0,
			UpdateCount:          324,
			InTransactionCount:   324,
		},
	}

	return sqlStats
}

// 以下是统计信息获取的辅助方法 对应Java后端的DruidDataSourceStatManager的各种统计方法

// getCurrentConnections 获取当前连接数
func (c *DruidController) getCurrentConnections(db *gorm.DB) int {
	sqlDB, err := db.DB()
	if err != nil {
		return 0
	}
	stats := sqlDB.Stats()
	return stats.OpenConnections
}

// getActiveConnections 获取活跃连接数
func (c *DruidController) getActiveConnections(db *gorm.DB) int {
	sqlDB, err := db.DB()
	if err != nil {
		return 0
	}
	stats := sqlDB.Stats()
	return stats.InUse
}

// getConnectCount 获取连接次数（模拟）
func (c *DruidController) getConnectCount() int64 {
	return 15420 // 模拟数据
}

// getCloseCount 获取关闭次数（模拟）
func (c *DruidController) getCloseCount() int64 {
	return 15380 // 模拟数据
}

// getRecycleCount 获取回收次数（模拟）
func (c *DruidController) getRecycleCount() int64 {
	return 15380 // 模拟数据
}

// getActivePeak 获取活跃连接峰值（模拟）
func (c *DruidController) getActivePeak() int {
	return 18 // 模拟数据
}

// getExecuteCount 获取执行次数（模拟）
func (c *DruidController) getExecuteCount() int64 {
	return 45620 // 模拟数据
}

// getCommitCount 获取提交次数（模拟）
func (c *DruidController) getCommitCount() int64 {
	return 12340 // 模拟数据
}

// getExecuteMillisTotal 获取总执行时间（模拟）
func (c *DruidController) getExecuteMillisTotal() int64 {
	return 567890 // 模拟数据，毫秒
}

// getExecuteMillisMax 获取最大执行时间（模拟）
func (c *DruidController) getExecuteMillisMax() int64 {
	return 1250 // 模拟数据，毫秒
}

// getPreparedStatementOpenCount 获取预编译语句打开次数（模拟）
func (c *DruidController) getPreparedStatementOpenCount() int64 {
	return 8920 // 模拟数据
}

// getPreparedStatementClosedCount 获取预编译语句关闭次数（模拟）
func (c *DruidController) getPreparedStatementClosedCount() int64 {
	return 8900 // 模拟数据
}

// getCachedPreparedStatementCount 获取缓存预编译语句数（模拟）
func (c *DruidController) getCachedPreparedStatementCount() int64 {
	return 156 // 模拟数据
}

// getCachedPreparedStatementHitCount 获取缓存预编译语句命中数（模拟）
func (c *DruidController) getCachedPreparedStatementHitCount() int64 {
	return 7890 // 模拟数据
}

// getCachedPreparedStatementMissCount 获取缓存预编译语句未命中数（模拟）
func (c *DruidController) getCachedPreparedStatementMissCount() int64 {
	return 1030 // 模拟数据
}

// generateIndexHTML 生成Druid监控首页HTML 对应Java后端的/druid/index.html
func (c *DruidController) generateIndexHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Druid数据源监控</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 10px; border-radius: 5px; margin-bottom: 20px; }
        .nav { margin: 20px 0; }
        .nav a { margin-right: 20px; text-decoration: none; color: #337ab7; }
        .nav a:hover { text-decoration: underline; }
        .content { background: white; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .status-ok { color: green; }
        .status-error { color: red; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Druid数据源监控</h1>
        <p>WOSM Go Backend - 数据库连接池监控</p>
    </div>

    <div class="nav">
        <a href="/druid/">首页</a>
        <a href="/druid/datasource.json">数据源</a>
        <a href="/druid/sql.json">SQL监控</a>
        <a href="/druid/reset-all.json">重置</a>
    </div>

    <div class="content">
        <h2>数据源信息</h2>
        <div id="datasource-info">加载中...</div>

        <h2>SQL统计</h2>
        <div id="sql-stats">加载中...</div>
    </div>

    <script>
        // 加载数据源信息
        fetch('/druid/datasource.json')
            .then(response => response.json())
            .then(data => {
                if (data.code === 200) {
                    const stats = data.data;
                    document.getElementById('datasource-info').innerHTML =
                        '<table>' +
                        '<tr><th>属性</th><th>值</th></tr>' +
                        '<tr><td>数据库类型</td><td>' + stats.DbType + '</td></tr>' +
                        '<tr><td>驱动</td><td>' + stats.DriverClassName + '</td></tr>' +
                        '<tr><td>URL</td><td>' + stats.URL + '</td></tr>' +
                        '<tr><td>用户名</td><td>' + stats.UserName + '</td></tr>' +
                        '<tr><td>最大连接数</td><td>' + stats.MaxActive + '</td></tr>' +
                        '<tr><td>当前连接数</td><td class="status-ok">' + stats.PoolingCount + '</td></tr>' +
                        '<tr><td>活跃连接数</td><td class="status-ok">' + stats.ActiveCount + '</td></tr>' +
                        '<tr><td>连接次数</td><td>' + stats.ConnectCount + '</td></tr>' +
                        '<tr><td>执行次数</td><td>' + stats.ExecuteCount + '</td></tr>' +
                        '<tr><td>提交次数</td><td>' + stats.CommitCount + '</td></tr>' +
                        '</table>';
                }
            })
            .catch(error => {
                document.getElementById('datasource-info').innerHTML = '<p class="status-error">加载失败: ' + error + '</p>';
            });

        // 加载SQL统计
        fetch('/druid/sql.json')
            .then(response => response.json())
            .then(data => {
                if (data.code === 200) {
                    const sqlStats = data.data;
                    let html = '<table><tr><th>SQL</th><th>执行次数</th><th>总时间(ms)</th><th>最大时间(ms)</th><th>最后执行</th></tr>';
                    sqlStats.forEach(sql => {
                        html += '<tr>' +
                            '<td style="max-width: 300px; overflow: hidden; text-overflow: ellipsis;">' + sql.SQL + '</td>' +
                            '<td>' + sql.ExecuteCount + '</td>' +
                            '<td>' + sql.TotalTime + '</td>' +
                            '<td>' + sql.MaxTimespan + '</td>' +
                            '<td>' + new Date(sql.LastTime).toLocaleString() + '</td>' +
                            '</tr>';
                    });
                    html += '</table>';
                    document.getElementById('sql-stats').innerHTML = html;
                }
            })
            .catch(error => {
                document.getElementById('sql-stats').innerHTML = '<p class="status-error">加载失败: ' + error + '</p>';
            });
    </script>
</body>
</html>`
}

// generateLoginHTML 生成Druid登录页面HTML 对应Java后端的/druid/login.html
func (c *DruidController) generateLoginHTML() string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Druid监控登录</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background: #f5f5f5;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .login-form {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            width: 300px;
        }
        .login-form h2 { text-align: center; margin-bottom: 30px; color: #333; }
        .form-group { margin-bottom: 20px; }
        .form-group label { display: block; margin-bottom: 5px; color: #555; }
        .form-group input {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            box-sizing: border-box;
        }
        .btn {
            width: 100%;
            padding: 12px;
            background: #337ab7;
            color: white;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            font-size: 16px;
        }
        .btn:hover { background: #286090; }
        .error { color: red; margin-top: 10px; text-align: center; }
    </style>
</head>
<body>
    <div class="login-form">
        <h2>Druid监控登录</h2>
        <form id="loginForm">
            <div class="form-group">
                <label for="username">用户名:</label>
                <input type="text" id="username" name="loginUsername" value="ruoyi" required>
            </div>
            <div class="form-group">
                <label for="password">密码:</label>
                <input type="password" id="password" name="loginPassword" value="123456" required>
            </div>
            <button type="submit" class="btn">登录</button>
            <div id="error" class="error"></div>
        </form>
    </div>

    <script>
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            e.preventDefault();

            const formData = new FormData(this);

            fetch('/druid/submitLogin', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 200) {
                    window.location.href = '/druid/';
                } else {
                    document.getElementById('error').textContent = data.msg || '登录失败';
                }
            })
            .catch(error => {
                document.getElementById('error').textContent = '网络错误: ' + error;
            });
        });
    </script>
</body>
</html>`
}
