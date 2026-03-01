package inpg

import (
	"fmt"

	"github.com/nJannDave/pkg/log"
	"github.com/nJannDave/pkg/log/structure"
	"go.uber.org/zap"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"os"
)

func ProviderConnStr() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta", os.Getenv("PH"), os.Getenv("PU"), os.Getenv("PP"), os.Getenv("PN"), "5434")
}

func Init(connStr string) (*gorm.DB, func()) {
	var logcfg = structure.LogConfig {
		Status: false,
		Service: "connect postgre",
		Error: "",
	}
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logcfg.Error = err.Error()
		log.ZapLog.Sugar().Fatalf("failed connect postgres", zap.Object("log_config", logcfg))
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	return db, func(){
		if err := sqlDB.Close(); err != nil {
			logcfg.Error = err.Error()
			log.ZapLog.Sugar().Fatalf("failed connect postgres", zap.Object("log_config", logcfg))
		}
	}
}