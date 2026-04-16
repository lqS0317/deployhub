package model

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接，logLevel 可选 "debug"/"info"/"warn"/"error"/"silent"
func InitDB(dsn string, logLevel string) (*gorm.DB, error) {
	gormLogLevel := logger.Warn
	switch logLevel {
	case "debug":
		gormLogLevel = logger.Info
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("数据库连接成功")
	return db, nil
}
