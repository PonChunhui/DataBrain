package initialize

import (
	"fmt"

	"devops-backend/global"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() {
	m := global.GVA_CONFIG.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		m.Username,
		m.Password,
		m.Path,
		m.Port,
		m.DBName,
		m.Config,
	)

	var gormLogger logger.Interface
	if m.LogMode {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		global.GVA_LOG.Error("MySQL启动失败", zap.Error(err))
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		global.GVA_LOG.Error("MySQL获取连接池失败", zap.Error(err))
		panic(err)
	}

	sqlDB.SetMaxIdleConns(m.MaxIdleConns)
	sqlDB.SetMaxOpenConns(m.MaxOpenConns)

	global.GVA_DB = db
	global.GVA_LOG.Info("MySQL初始化成功", zap.String("dsn", dsn))
}
