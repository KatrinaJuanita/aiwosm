package tool

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SwaggerController Swagger文档控制器 对应Java后端的SwaggerConfig
type SwaggerController struct{}

// NewSwaggerController 创建Swagger控制器实例
func NewSwaggerController() *SwaggerController {
	return &SwaggerController{}
}

// SwaggerInfo Swagger基本信息 对应Java后端的ApiInfo
type SwaggerInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Contact     struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"contact"`
}

// SwaggerSpec Swagger规范 对应Java后端的Docket
type SwaggerSpec struct {
	Swagger             string                 `json:"swagger"`
	Info                SwaggerInfo            `json:"info"`
	Host                string                 `json:"host"`
	BasePath            string                 `json:"basePath"`
	Schemes             []string               `json:"schemes"`
	Consumes            []string               `json:"consumes"`
	Produces            []string               `json:"produces"`
	Paths               map[string]interface{} `json:"paths"`
	Definitions         map[string]interface{} `json:"definitions"`
	SecurityDefinitions map[string]interface{} `json:"securityDefinitions"`
}

// PathInfo 路径信息 对应Java后端的ApiOperation
type PathInfo struct {
	Tags        []string               `json:"tags"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	OperationID string                 `json:"operationId"`
	Consumes    []string               `json:"consumes"`
	Produces    []string               `json:"produces"`
	Parameters  []ParameterInfo        `json:"parameters"`
	Responses   map[string]interface{} `json:"responses"`
	Security    []map[string][]string  `json:"security"`
}

// ParameterInfo 参数信息 对应Java后端的ApiParam
type ParameterInfo struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Format      string      `json:"format,omitempty"`
	Schema      interface{} `json:"schema,omitempty"`
}

// Index Swagger UI首页 对应Java后端的/swagger-ui/index.html
// @Summary Swagger UI首页
// @Description 显示Swagger API文档界面
// @Tags Swagger文档
// @Accept html
// @Produce html
// @Success 200 {string} string "HTML页面"
// @Router /swagger-ui/index.html [get]
func (c *SwaggerController) Index(ctx *gin.Context) {
	fmt.Printf("SwaggerController.Index: 访问Swagger UI首页\n")

	// 返回Swagger UI页面HTML
	html := c.generateSwaggerUIHTML()
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, html)
}

// Spec Swagger规范文档 对应Java后端的/v3/api-docs
// @Summary Swagger规范文档
// @Description 获取Swagger API规范文档
// @Tags Swagger文档
// @Accept json
// @Produce json
// @Success 200 {object} SwaggerSpec
// @Router /swagger-ui/api-docs [get]
func (c *SwaggerController) Spec(ctx *gin.Context) {
	fmt.Printf("SwaggerController.Spec: 获取Swagger规范文档\n")

	spec := c.generateSwaggerSpec(ctx)
	ctx.JSON(http.StatusOK, spec)
}

// generateSwaggerUIHTML 生成Swagger UI页面HTML 对应Java后端的Swagger UI资源
func (c *SwaggerController) generateSwaggerUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>WOSM API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #337ab7;
        }
        .swagger-ui .topbar .download-url-wrapper .select-label {
            color: #fff;
        }
        .swagger-ui .topbar .download-url-wrapper input[type=text] {
            border: 2px solid #547f00;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger-ui/api-docs',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                requestInterceptor: function(request) {
                    // 添加Authorization头
                    const token = localStorage.getItem('token');
                    if (token) {
                        request.headers['Authorization'] = 'Bearer ' + token;
                    }
                    return request;
                },
                onComplete: function() {
                    // 添加认证按钮
                    const authButton = document.createElement('button');
                    authButton.innerHTML = '设置Token';
                    authButton.style.cssText = 'position: fixed; top: 10px; right: 10px; z-index: 9999; padding: 8px 16px; background: #337ab7; color: white; border: none; border-radius: 4px; cursor: pointer;';
                    authButton.onclick = function() {
                        const token = prompt('请输入Authorization Token (不包含Bearer前缀):');
                        if (token) {
                            localStorage.setItem('token', token);
                            alert('Token已设置，刷新页面后生效');
                        }
                    };
                    document.body.appendChild(authButton);
                }
            });
        };
    </script>
</body>
</html>`
}

// generateSwaggerSpec 生成Swagger规范文档 对应Java后端的Docket配置
func (c *SwaggerController) generateSwaggerSpec(ctx *gin.Context) *SwaggerSpec {

	spec := &SwaggerSpec{
		Swagger: "2.0",
		Info: SwaggerInfo{
			Title:       "WOSM管理系统接口文档",
			Description: "用于管理集团旗下公司的人员信息，具体包括用户管理、角色管理、权限管理、部门管理、岗位管理、字典管理、参数配置、通知公告、操作日志、登录日志、在线用户、服务监控、缓存监控、定时任务、代码生成等模块",
			Version:     "3.9.0",
			Contact: struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			}{
				Name: "WOSM",
				URL:  "http://localhost:8080",
			},
		},
		Host:        ctx.Request.Host,
		BasePath:    "/",
		Schemes:     []string{"http", "https"},
		Consumes:    []string{"application/json", "multipart/form-data"},
		Produces:    []string{"application/json"},
		Paths:       c.generatePaths(),
		Definitions: c.generateDefinitions(),
		SecurityDefinitions: map[string]interface{}{
			"Authorization": map[string]interface{}{
				"type":        "apiKey",
				"name":        "Authorization",
				"in":          "header",
				"description": "JWT token, format: Bearer {token}",
			},
		},
	}

	return spec
}

// generatePaths 生成API路径信息 对应Java后端的RequestMappingHandlerMapping
func (c *SwaggerController) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	// 认证相关API
	paths["/login"] = map[string]interface{}{
		"post": PathInfo{
			Tags:        []string{"认证管理"},
			Summary:     "用户登录",
			Description: "用户登录接口",
			OperationID: "login",
			Consumes:    []string{"application/json"},
			Produces:    []string{"application/json"},
			Parameters: []ParameterInfo{
				{
					Name:        "loginBody",
					In:          "body",
					Description: "登录信息",
					Required:    true,
					Schema: map[string]interface{}{
						"$ref": "#/definitions/LoginBody",
					},
				},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "登录成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/AjaxResult",
					},
				},
			},
		},
	}

	paths["/getInfo"] = map[string]interface{}{
		"get": PathInfo{
			Tags:        []string{"认证管理"},
			Summary:     "获取用户信息",
			Description: "获取当前登录用户信息",
			OperationID: "getInfo",
			Produces:    []string{"application/json"},
			Security: []map[string][]string{
				{"Authorization": {}},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "获取成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/AjaxResult",
					},
				},
			},
		},
	}

	paths["/getRouters"] = map[string]interface{}{
		"get": PathInfo{
			Tags:        []string{"认证管理"},
			Summary:     "获取路由信息",
			Description: "获取当前用户的路由菜单信息",
			OperationID: "getRouters",
			Produces:    []string{"application/json"},
			Security: []map[string][]string{
				{"Authorization": {}},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "获取成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/AjaxResult",
					},
				},
			},
		},
	}

	// 用户管理API
	paths["/system/user/list"] = map[string]interface{}{
		"get": PathInfo{
			Tags:        []string{"用户管理"},
			Summary:     "查询用户列表",
			Description: "分页查询用户列表",
			OperationID: "getUserList",
			Produces:    []string{"application/json"},
			Parameters: []ParameterInfo{
				{Name: "pageNum", In: "query", Description: "页码", Required: false, Type: "integer"},
				{Name: "pageSize", In: "query", Description: "每页数量", Required: false, Type: "integer"},
				{Name: "userName", In: "query", Description: "用户名", Required: false, Type: "string"},
				{Name: "phonenumber", In: "query", Description: "手机号", Required: false, Type: "string"},
				{Name: "status", In: "query", Description: "状态", Required: false, Type: "string"},
			},
			Security: []map[string][]string{
				{"Authorization": {}},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "查询成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/TableDataInfo",
					},
				},
			},
		},
	}

	paths["/system/user"] = map[string]interface{}{
		"post": PathInfo{
			Tags:        []string{"用户管理"},
			Summary:     "新增用户",
			Description: "新增用户信息",
			OperationID: "addUser",
			Consumes:    []string{"application/json"},
			Produces:    []string{"application/json"},
			Parameters: []ParameterInfo{
				{
					Name:        "user",
					In:          "body",
					Description: "用户信息",
					Required:    true,
					Schema: map[string]interface{}{
						"$ref": "#/definitions/SysUser",
					},
				},
			},
			Security: []map[string][]string{
				{"Authorization": {}},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "新增成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/AjaxResult",
					},
				},
			},
		},
		"put": PathInfo{
			Tags:        []string{"用户管理"},
			Summary:     "修改用户",
			Description: "修改用户信息",
			OperationID: "updateUser",
			Consumes:    []string{"application/json"},
			Produces:    []string{"application/json"},
			Parameters: []ParameterInfo{
				{
					Name:        "user",
					In:          "body",
					Description: "用户信息",
					Required:    true,
					Schema: map[string]interface{}{
						"$ref": "#/definitions/SysUser",
					},
				},
			},
			Security: []map[string][]string{
				{"Authorization": {}},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "修改成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/AjaxResult",
					},
				},
			},
		},
	}

	// 角色管理API
	paths["/system/role/list"] = map[string]interface{}{
		"get": PathInfo{
			Tags:        []string{"角色管理"},
			Summary:     "查询角色列表",
			Description: "分页查询角色列表",
			OperationID: "getRoleList",
			Produces:    []string{"application/json"},
			Parameters: []ParameterInfo{
				{Name: "pageNum", In: "query", Description: "页码", Required: false, Type: "integer"},
				{Name: "pageSize", In: "query", Description: "每页数量", Required: false, Type: "integer"},
				{Name: "roleName", In: "query", Description: "角色名称", Required: false, Type: "string"},
				{Name: "roleKey", In: "query", Description: "角色权限", Required: false, Type: "string"},
				{Name: "status", In: "query", Description: "状态", Required: false, Type: "string"},
			},
			Security: []map[string][]string{
				{"Authorization": {}},
			},
			Responses: map[string]interface{}{
				"200": map[string]interface{}{
					"description": "查询成功",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/TableDataInfo",
					},
				},
			},
		},
	}

	return paths
}

// generateDefinitions 生成数据模型定义 对应Java后端的Model类定义
func (c *SwaggerController) generateDefinitions() map[string]interface{} {
	definitions := make(map[string]interface{})

	// AjaxResult 通用返回结果 对应Java后端的AjaxResult
	definitions["AjaxResult"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "integer",
				"description": "状态码",
				"example":     200,
			},
			"msg": map[string]interface{}{
				"type":        "string",
				"description": "返回消息",
				"example":     "操作成功",
			},
			"data": map[string]interface{}{
				"type":        "object",
				"description": "返回数据",
			},
		},
	}

	// TableDataInfo 分页数据 对应Java后端的TableDataInfo
	definitions["TableDataInfo"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"code": map[string]interface{}{
				"type":        "integer",
				"description": "状态码",
				"example":     200,
			},
			"msg": map[string]interface{}{
				"type":        "string",
				"description": "返回消息",
				"example":     "查询成功",
			},
			"rows": map[string]interface{}{
				"type":        "array",
				"description": "数据列表",
				"items": map[string]interface{}{
					"type": "object",
				},
			},
			"total": map[string]interface{}{
				"type":        "integer",
				"description": "总记录数",
				"example":     100,
			},
		},
	}

	// LoginBody 登录请求体 对应Java后端的LoginBody
	definitions["LoginBody"] = map[string]interface{}{
		"type":     "object",
		"required": []string{"username", "password"},
		"properties": map[string]interface{}{
			"username": map[string]interface{}{
				"type":        "string",
				"description": "用户名",
				"example":     "admin",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "密码",
				"example":     "admin123",
			},
			"code": map[string]interface{}{
				"type":        "string",
				"description": "验证码",
				"example":     "1234",
			},
			"uuid": map[string]interface{}{
				"type":        "string",
				"description": "验证码唯一标识",
				"example":     "uuid-1234",
			},
		},
	}

	// SysUser 用户信息 对应Java后端的SysUser
	definitions["SysUser"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"userId": map[string]interface{}{
				"type":        "integer",
				"description": "用户ID",
				"example":     1,
			},
			"userName": map[string]interface{}{
				"type":        "string",
				"description": "用户名",
				"example":     "admin",
			},
			"nickName": map[string]interface{}{
				"type":        "string",
				"description": "用户昵称",
				"example":     "管理员",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"description": "邮箱",
				"example":     "admin@example.com",
			},
			"phonenumber": map[string]interface{}{
				"type":        "string",
				"description": "手机号",
				"example":     "13800138000",
			},
			"sex": map[string]interface{}{
				"type":        "string",
				"description": "性别",
				"example":     "1",
			},
			"avatar": map[string]interface{}{
				"type":        "string",
				"description": "头像",
				"example":     "/profile/avatar/2023/01/01/avatar.jpg",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "状态",
				"example":     "0",
			},
			"deptId": map[string]interface{}{
				"type":        "integer",
				"description": "部门ID",
				"example":     100,
			},
			"postIds": map[string]interface{}{
				"type":        "array",
				"description": "岗位ID列表",
				"items": map[string]interface{}{
					"type": "integer",
				},
			},
			"roleIds": map[string]interface{}{
				"type":        "array",
				"description": "角色ID列表",
				"items": map[string]interface{}{
					"type": "integer",
				},
			},
		},
	}

	// SysRole 角色信息 对应Java后端的SysRole
	definitions["SysRole"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"roleId": map[string]interface{}{
				"type":        "integer",
				"description": "角色ID",
				"example":     1,
			},
			"roleName": map[string]interface{}{
				"type":        "string",
				"description": "角色名称",
				"example":     "管理员",
			},
			"roleKey": map[string]interface{}{
				"type":        "string",
				"description": "角色权限字符串",
				"example":     "admin",
			},
			"roleSort": map[string]interface{}{
				"type":        "integer",
				"description": "显示顺序",
				"example":     1,
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "角色状态",
				"example":     "0",
			},
			"dataScope": map[string]interface{}{
				"type":        "string",
				"description": "数据范围",
				"example":     "1",
			},
			"remark": map[string]interface{}{
				"type":        "string",
				"description": "备注",
				"example":     "管理员角色",
			},
		},
	}

	return definitions
}
