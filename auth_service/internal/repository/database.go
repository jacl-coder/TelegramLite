package repository

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	applogger "github.com/jacl-coder/TelegramLite/common/go/logger"
	"github.com/jacl-coder/telegramlite/auth_service/internal/config"
	"github.com/jacl-coder/telegramlite/auth_service/internal/model"
)

// DB 数据库实例
var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) error {
	// 构建连接字符串
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)

	// 配置GORM
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db

	// 使用统一日志系统
	log := applogger.GetDefault()
	if log != nil {
		log.Info("Database connected successfully", applogger.Fields{
			"host":   cfg.Host,
			"dbname": cfg.DBName,
			"port":   cfg.Port,
		})
	}

	return nil
}

// AutoMigrate 自动迁移数据表
func AutoMigrate() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// 迁移表结构
	err := DB.AutoMigrate(
		&model.User{},
		&model.Device{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// 使用统一日志系统
	log := applogger.GetDefault()
	if log != nil {
		log.Info("Database migration completed successfully")
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
