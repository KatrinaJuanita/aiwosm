package model

import (
	"time"
)

// SysConfig 参数配置表 对应Java后端的SysConfig实体
// 严格按照Java后端真实数据库表结构定义（基于SqlServer_ry_20250522_COMPLETE.sql）：
// config_id, config_name, config_key, config_value, config_type, create_by, create_time, update_by, update_time, remark
type SysConfig struct {
	ConfigID    int64      `gorm:"column:config_id;primaryKey;autoIncrement" json:"configId" excel:"name:参数主键;sort:1;cellType:numeric"`       // 参数主键
	ConfigName  string     `gorm:"column:config_name;size:100;not null" json:"configName" excel:"name:参数名称;sort:2"`                           // 参数名称
	ConfigKey   string     `gorm:"column:config_key;size:100;not null" json:"configKey" excel:"name:参数键名;sort:3"`                             // 参数键名
	ConfigValue string     `gorm:"column:config_value;size:500;not null;default:''" json:"configValue" excel:"name:参数键值;sort:4"`              // 参数键值
	ConfigType  string     `gorm:"column:config_type;size:1;default:'N'" json:"configType" excel:"name:系统内置;sort:5;readConverterExp:Y=是,N=否"` // 系统内置（Y是 N否）
	CreateBy    string     `gorm:"column:create_by;size:64;default:''" json:"createBy"`                                                       // 创建者
	CreateTime  *time.Time `gorm:"column:create_time" json:"createTime" excel:"name:创建时间;sort:6;type:export;dateFormat:yyyy-MM-dd HH:mm:ss"`  // 创建时间
	UpdateBy    string     `gorm:"column:update_by;size:64;default:''" json:"updateBy"`                                                       // 更新者
	UpdateTime  *time.Time `gorm:"column:update_time" json:"updateTime"`                                                                      // 更新时间
	Remark      string     `gorm:"column:remark;size:500" json:"remark" excel:"name:备注;sort:7"`                                               // 备注
}

// TableName 设置表名
func (SysConfig) TableName() string {
	return "sys_config"
}

// 参数配置类型常量 对应Java后端的常量定义
const (
	ConfigTypeYes = "Y" // 系统内置
	ConfigTypeNo  = "N" // 非系统内置
)

// 系统内置参数键名常量 对应Java后端的系统参数
const (
	// 验证码开关
	SysAccountCaptchaEnabled = "sys.account.captchaEnabled"
	// 用户管理-账号初始密码
	SysUserInitPassword = "sys.user.initPassword"
	// 主框架页-侧边栏主题
	SysIndexSkinName = "sys.index.skinName"
	// 主框架页-默认皮肤样式名称
	SysIndexSidebarTheme = "sys.index.sidebarTheme"
	// 账号自助-验证码开关
	SysAccountRegisterUser = "sys.account.registerUser"
)

// ConfigQueryParams 参数配置查询参数 对应Java后端的查询条件
type ConfigQueryParams struct {
	ConfigName    string `form:"configName" json:"configName"`       // 参数名称（模糊查询）
	ConfigKey     string `form:"configKey" json:"configKey"`         // 参数键名（模糊查询）
	ConfigType    string `form:"configType" json:"configType"`       // 参数类型
	BeginTime     string `form:"beginTime" json:"beginTime"`         // 开始时间
	EndTime       string `form:"endTime" json:"endTime"`             // 结束时间
	PageNum       int    `form:"pageNum" json:"pageNum"`             // 页码
	PageSize      int    `form:"pageSize" json:"pageSize"`           // 每页数量
	OrderByColumn string `form:"orderByColumn" json:"orderByColumn"` // 排序字段
	IsAsc         string `form:"isAsc" json:"isAsc"`                 // 排序方向
}

// IsBuiltIn 判断是否为系统内置参数
func (c *SysConfig) IsBuiltIn() bool {
	return c.ConfigType == ConfigTypeYes
}

// GetCacheKey 获取缓存键名 对应Java后端的getCacheKey方法
func GetConfigCacheKey(configKey string) string {
	return "sys_config:" + configKey
}

// 系统内置参数列表 对应Java后端的内置参数
var BuiltInConfigKeys = []string{
	SysAccountCaptchaEnabled,
	SysUserInitPassword,
	SysIndexSkinName,
	SysIndexSidebarTheme,
	SysAccountRegisterUser,
}

// IsBuiltInConfigKey 判断是否为系统内置参数键名
func IsBuiltInConfigKey(configKey string) bool {
	for _, key := range BuiltInConfigKeys {
		if key == configKey {
			return true
		}
	}
	return false
}

// ConfigExportData 参数配置导出数据结构 对应Java后端的Excel导出
type ConfigExportData struct {
	ConfigID    int64  `json:"configId" excel:"参数主键"`
	ConfigName  string `json:"configName" excel:"参数名称"`
	ConfigKey   string `json:"configKey" excel:"参数键名"`
	ConfigValue string `json:"configValue" excel:"参数键值"`
	ConfigType  string `json:"configType" excel:"系统内置"`
	CreateBy    string `json:"createBy" excel:"创建者"`
	CreateTime  string `json:"createTime" excel:"创建时间"`
	UpdateBy    string `json:"updateBy" excel:"更新者"`
	UpdateTime  string `json:"updateTime" excel:"更新时间"`
	Remark      string `json:"remark" excel:"备注"`
}

// ToExportData 转换为导出数据格式
func (c *SysConfig) ToExportData() *ConfigExportData {
	data := &ConfigExportData{
		ConfigID:    c.ConfigID,
		ConfigName:  c.ConfigName,
		ConfigKey:   c.ConfigKey,
		ConfigValue: c.ConfigValue,
		ConfigType:  c.ConfigType,
		CreateBy:    c.CreateBy,
		UpdateBy:    c.UpdateBy,
		Remark:      c.Remark,
	}

	// 格式化时间
	if c.CreateTime != nil {
		data.CreateTime = c.CreateTime.Format("2006-01-02 15:04:05")
	}
	if c.UpdateTime != nil {
		data.UpdateTime = c.UpdateTime.Format("2006-01-02 15:04:05")
	}

	// 转换配置类型显示
	if c.ConfigType == ConfigTypeYes {
		data.ConfigType = "是"
	} else {
		data.ConfigType = "否"
	}

	return data
}

// GetConfigTypeText 获取配置类型文本
func GetConfigTypeText(configType string) string {
	switch configType {
	case ConfigTypeYes:
		return "系统内置"
	case ConfigTypeNo:
		return "非系统内置"
	default:
		return "未知"
	}
}

// ValidateConfigKey 验证参数键名格式
func ValidateConfigKey(configKey string) bool {
	if configKey == "" {
		return false
	}

	// 参数键名只能包含字母、数字、点号、下划线
	for _, char := range configKey {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '_') {
			return false
		}
	}

	return true
}

// GetDefaultConfigs 获取默认系统参数配置 对应Java后端的初始化数据
func GetDefaultConfigs() []*SysConfig {
	now := time.Now()
	return []*SysConfig{
		{
			ConfigName:  "主框架页-默认皮肤样式名称",
			ConfigKey:   SysIndexSkinName,
			ConfigValue: "skin-blue",
			ConfigType:  ConfigTypeYes,
			CreateBy:    "admin",
			CreateTime:  &now,
			Remark:      "蓝色 skin-blue、绿色 skin-green、紫色 skin-purple、红色 skin-red、黄色 skin-yellow",
		},
		{
			ConfigName:  "用户管理-账号初始密码",
			ConfigKey:   SysUserInitPassword,
			ConfigValue: "123456",
			ConfigType:  ConfigTypeYes,
			CreateBy:    "admin",
			CreateTime:  &now,
			Remark:      "初始化密码 123456",
		},
		{
			ConfigName:  "主框架页-侧边栏主题",
			ConfigKey:   SysIndexSidebarTheme,
			ConfigValue: "theme-dark",
			ConfigType:  ConfigTypeYes,
			CreateBy:    "admin",
			CreateTime:  &now,
			Remark:      "深色主题theme-dark，浅色主题theme-light",
		},
		{
			ConfigName:  "账号自助-验证码开关",
			ConfigKey:   SysAccountCaptchaEnabled,
			ConfigValue: "true",
			ConfigType:  ConfigTypeYes,
			CreateBy:    "admin",
			CreateTime:  &now,
			Remark:      "是否开启验证码功能（true开启，false关闭）",
		},
		{
			ConfigName:  "账号自助-是否开启用户注册功能",
			ConfigKey:   SysAccountRegisterUser,
			ConfigValue: "false",
			ConfigType:  ConfigTypeYes,
			CreateBy:    "admin",
			CreateTime:  &now,
			Remark:      "是否开启注册用户功能（true开启，false关闭）",
		},
	}
}
