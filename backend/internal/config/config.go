package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构 对应Java后端的application.yml
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Log      LogConfig      `yaml:"log"`
	Captcha  CaptchaConfig  `yaml:"captcha"`
	File     FileConfig     `yaml:"file"`
	User     UserConfig     `yaml:"user"` // 用户配置 对应Java后端的user配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `yaml:"driver"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Database        string `yaml:"database"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`  // 连接最大生存时间（秒）
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time"` // 连接最大空闲时间（秒）
	PingTimeout     int    `yaml:"ping_timeout"`       // 连接测试超时时间（秒）
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	Database     int    `yaml:"database"`
	PoolSize     int    `yaml:"pool_size"`
	DialTimeout  int    `yaml:"dial_timeout"`  // 连接超时时间（秒）
	ReadTimeout  int    `yaml:"read_timeout"`  // 读取超时时间（秒）
	WriteTimeout int    `yaml:"write_timeout"` // 写入超时时间（秒）
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireTime  int64  `yaml:"expire_time"`
	RefreshTime int64  `yaml:"refresh_time"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
}

// CaptchaConfig 验证码配置
type CaptchaConfig struct {
	Enabled    bool `yaml:"enabled"`
	Length     int  `yaml:"length"`
	Width      int  `yaml:"width"`
	Height     int  `yaml:"height"`
	ExpireTime int  `yaml:"expire_time"`
}

// FileConfig 文件上传配置 对应Java后端的RuoYiConfig
type FileConfig struct {
	UploadPath        string   `yaml:"upload_path"`        // 文件上传路径
	ResourcePrefix    string   `yaml:"resource_prefix"`    // 资源访问路径前缀
	MaxSize           int64    `yaml:"max_size"`           // 最大文件大小
	MaxNameLength     int      `yaml:"max_name_length"`    // 最大文件名长度
	AllowedExtensions []string `yaml:"allowed_extensions"` // 允许的文件扩展名
}

// UserConfig 用户配置 对应Java后端的user配置
type UserConfig struct {
	Password PasswordConfig `yaml:"password"` // 密码配置
}

// PasswordConfig 密码配置 对应Java后端的user.password配置
type PasswordConfig struct {
	MaxRetryCount int `yaml:"max_retry_count"` // 密码最大错误次数
	LockTime      int `yaml:"lock_time"`       // 密码锁定时间（分钟）
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}

	AppConfig = config
	log.Printf("配置加载成功: %s", configPath)
	return nil
}
