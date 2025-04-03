package ioc

import (
	promsdk "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"time"
	"github.com/TengFeiyang01/webook/webook/internal/repository/dao"
	gormx "github.com/TengFeiyang01/webook/webook/pkg/gormx"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg = Config{
		DSN: "root:root@tcp(localhost:13316)/webook",
	}

	// 看起来不支持 key 的分隔
	if err := viper.UnmarshalKey("db", &cfg); err != nil {
		panic(err)
	}
	//dsn := viper.GetString("db.mysql.dsn")
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 慢查询阈值，超过这个阈值，才会使用
			// 50ms 100ms
			// SQL 查询必须要求命中索引，最好就是走一次磁盘 IO
			// 一次磁盘 iO 是不到 10ms
			SlowThreshold:             time.Millisecond * 10,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
			ParameterizedQueries:      true,
			LogLevel:                  glogger.Error,
		}),
	})
	if err != nil {
		panic(err)
	}
	if err := db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	})); err != nil {
		panic(err)
	}

	cb := gormx.NewCallbacks(promsdk.SummaryOpts{
		Namespace: "ytf",
		Subsystem: "webook",
		Name:      "gorm_db",
		Help:      "统计 GORM 的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "my_instance",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	if err := db.Use(tracing.NewPlugin(tracing.WithDBName("webook"),
		//tracing.WithQueryFormatter(func(query string) string {
		//	l.Debug("query", logger.String("query", query))
		//	return query
		//}),
		// 不要记录 metrics
		tracing.WithoutMetrics(),
		// 不用记录查询参数
		tracing.WithoutQueryVariables())); err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		return nil
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
