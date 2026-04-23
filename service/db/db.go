package db

import (
	"SService/config"
	"SService/model"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// CustomLogger 屏蔽 ErrRecordNotFound 的噪音日志
type CustomLogger struct {
	logger.Interface
}

func (l *CustomLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		l.Interface.Trace(ctx, begin, fc, nil)
		return
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

func InitDB() error {
	mysqlCfg := config.AppConfig.MySQL
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		mysqlCfg.Username,
		mysqlCfg.Password,
		mysqlCfg.Path,
		mysqlCfg.Port,
		mysqlCfg.DBName,
		mysqlCfg.Config,
	)
	defaultLogger := logger.Default.LogMode(logger.Error)
	customLogger := &CustomLogger{Interface: defaultLogger}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         customLogger,
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	if err := DB.AutoMigrate(&model.User{}, &model.Example{}); err != nil {
		return fmt.Errorf("模型迁移失败: %w", err)
	}

	log.Println("数据库初始化成功")
	return nil
}
