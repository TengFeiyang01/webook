package ioc

import (
	promsdk "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
	"time"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/logger"
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
			LogLevel:                  glogger.Info,
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

	pcb := newCallbacks()
	pcb.registerAll(db)

	err = dao.InitTables(db)
	if err != nil {
		return nil
	}

	return db
}

func (pcb *Callbacks) registerAll(db *gorm.DB) {
	if err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", pcb.before()); err != nil {
		panic(err)
	}
	if err := db.Callback().Create().After("*").
		Register("prometheus_create_after", pcb.after("create")); err != nil {
		panic(err)
	}
	if err := db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", pcb.before()); err != nil {
		panic(err)
	}
	if err := db.Callback().Delete().After("*").
		Register("prometheus_delete_after", pcb.after("delete")); err != nil {
		panic(err)
	}
	if err := db.Callback().Update().Before("*").
		Register("prometheus_update_before", pcb.before()); err != nil {
		panic(err)
	}
	if err := db.Callback().Update().After("*").
		Register("prometheus_update_after", pcb.after("update")); err != nil {
		panic(err)
	}
	if err := db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", pcb.before()); err != nil {
		panic(err)
	}
	if err := db.Callback().Raw().After("*").
		Register("prometheus_raw_after", pcb.after("raw")); err != nil {
		panic(err)
	}
	if err := db.Callback().Row().Before("*").
		Register("prometheus_row_before", pcb.before()); err != nil {
		panic(err)
	}
	if err := db.Callback().Row().After("*").
		Register("prometheus_row_after", pcb.after("row")); err != nil {
		panic(err)
	}
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}

type Callbacks struct {
	vector *promsdk.SummaryVec
}

func newCallbacks() *Callbacks {
	vector := promsdk.NewSummaryVec(promsdk.SummaryOpts{
		// 在这边, 你要考虑设置各种 namespace
		Namespace: "ytf",
		Subsystem: "webook",
		Name:      "gorm_query_time",
		Help:      "统计 GORM 的执行时间",
		ConstLabels: map[string]string{
			"db": "webook",
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.001,
		},
	}, []string{"type", "table"})
	pcb := &Callbacks{
		vector: vector,
	}
	promsdk.MustRegister(vector)
	return pcb
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		starTime := time.Now()
		db.Set("star_time", starTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("star_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		// 准备上报给 prometheus
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).
			Observe(float64(time.Since(startTime).Milliseconds()))
	}
}
