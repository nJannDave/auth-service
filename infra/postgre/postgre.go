package inpg

import (
	"fmt"
	
	"github.com/nJannDave/pkg/log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"os"
)

func ProviderConnStr() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta", os.Getenv("PH"), os.Getenv("PU"), os.Getenv("PP"), os.Getenv("PN"), "5434")
}

func Init(connStr string) (*gorm.DB, func()) {
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.LogConfig("failed connect db", "connect_db_gorm", err)
		panic(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	return db, func(){
		if err := sqlDB.Close(); err != nil {
			panic(err)
		}
	}
}