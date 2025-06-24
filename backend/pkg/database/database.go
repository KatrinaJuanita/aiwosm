package database

import (
	"context"
	"fmt"
	"log"
	"time"
	"wosm/internal/config"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	cfg := config.AppConfig.Database

	// 构建SQL Server连接字符串
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=disable&trustServerCertificate=true",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// 配置GORM日志级别
	var logLevel logger.LogLevel
	switch config.AppConfig.Log.Level {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	default:
		logLevel = logger.Error
	}

	// 连接数据库
	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 获取底层sql.DB对象配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %v", err)
	}

	// 设置连接池参数 对应Java后端的Druid连接池配置
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)                                    // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second) // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second) // 连接最大空闲时间

	// 测试数据库连接 对应Java后端的validationQuery
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.PingTimeout)*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %v", err)
	}

	DB = db
	log.Printf("数据库连接成功: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)

	// 自动迁移表结构（开发环境）
	// 注释掉自动迁移，因为数据库表已存在且结构正确
	// if err := autoMigrate(); err != nil {
	//	log.Printf("数据库表结构迁移警告: %v", err)
	// }

	return nil
}

// autoMigrate 自动迁移表结构（已禁用，数据库表已存在）
// func autoMigrate() error {
//	return DB.AutoMigrate(
//		&model.SysUser{},
//		&model.SysRole{},
//		&model.SysDept{},
//		&model.SysMenu{},
//	)
// }

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return DB
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
