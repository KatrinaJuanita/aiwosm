package model

// LoginUser 登录用户身份权限 对应Java后端的LoginUser
type LoginUser struct {
	UserID        int64    `json:"userId"`        // 用户ID
	DeptID        *int64   `json:"deptId"`        // 部门ID
	Token         string   `json:"token"`         // 用户唯一标识
	LoginTime     int64    `json:"loginTime"`     // 登录时间
	ExpireTime    int64    `json:"expireTime"`    // 过期时间
	IPAddr        string   `json:"ipaddr"`        // 登录IP地址
	LoginLocation string   `json:"loginLocation"` // 登录地点
	Browser       string   `json:"browser"`       // 浏览器类型
	OS            string   `json:"os"`            // 操作系统
	Permissions   []string `json:"permissions"`   // 权限列表
	User          *SysUser `json:"user"`          // 用户信息
}

// LoginBody 登录请求体 对应Java后端的LoginBody
type LoginBody struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
	Code     string `json:"code"`                        // 验证码
	UUID     string `json:"uuid"`                        // 唯一标识
}

// CaptchaResponse 验证码响应 对应Java后端的验证码返回格式
type CaptchaResponse struct {
	CaptchaEnabled bool   `json:"captchaEnabled"` // 验证码开关
	UUID           string `json:"uuid,omitempty"` // 唯一标识
	Img            string `json:"img,omitempty"`  // 验证码图片base64
}

// UserInfoResponse 用户信息响应 对应Java后端的getInfo返回格式
type UserInfoResponse struct {
	User               *SysUser `json:"user"`               // 用户信息
	Roles              []string `json:"roles"`              // 角色集合
	Permissions        []string `json:"permissions"`        // 权限集合
	IsDefaultModifyPwd bool     `json:"isDefaultModifyPwd"` // 是否需要修改初始密码
	IsPasswordExpired  bool     `json:"isPasswordExpired"`  // 密码是否过期
}

// RouterResponse 路由响应 对应Java后端的getRouters返回格式
type RouterResponse struct {
	Name       string           `json:"name"`                 // 路由名字
	Path       string           `json:"path"`                 // 路由地址
	Hidden     bool             `json:"hidden"`               // 是否隐藏路由，当设置 true 的时候该路由不会再侧边栏出现
	Redirect   string           `json:"redirect,omitempty"`   // 重定向地址，当设置 noRedirect 的时候该路由在面包屑导航中不可被点击
	Component  string           `json:"component"`            // 组件地址
	Query      string           `json:"query,omitempty"`      // 路由参数：如 {"id": 1, "name": "ry"}
	AlwaysShow bool             `json:"alwaysShow,omitempty"` // 当你一个路由下面的 children 声明的路由大于1个时，自动会变成嵌套的模式--如组件页面
	Meta       RouterMeta       `json:"meta"`                 // 其他元素
	Children   []RouterResponse `json:"children,omitempty"`   // 子路由
}

// RouterMeta 路由元信息
type RouterMeta struct {
	Title   string `json:"title"`   // 设置该路由在侧边栏和面包屑中展示的名字
	Icon    string `json:"icon"`    // 设置该路由的图标，对应路径src/assets/icons/svg
	NoCache bool   `json:"noCache"` // 如果设置为true，则不会被 <keep-alive> 缓存
	Link    string `json:"link"`    // 内链地址（http(s)://开头）
}
