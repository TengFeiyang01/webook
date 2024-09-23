package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg = Config{
		DSN: "root@tcp(localhost:13316)/webook_default",
	}

	// 看起来不支持 key 的分隔
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	//dsn := viper.GetString("db.mysql.dsn")
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
