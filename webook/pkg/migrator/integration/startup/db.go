package startup

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

// InitSrcDB 初始化源表
func InitSrcDB() *gorm.DB {
	return initDB("webook")
}

func InitIntrDB() *gorm.DB {
	return initDB("webook_intr")
}

func initDB(dbName string) *gorm.DB {
	dsn := fmt.Sprintf("root:root@tcp(localhost:13316)/%s", dbName)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err = sqlDB.PingContext(ctx)
		cancel()
		if err == nil {
			break
		}
		log.Println("等待连接 MySQL", err)
	}
	db, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		panic(err)
	}
	return db
}
