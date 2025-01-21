package gormx

import (
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"time"
)

type Callbacks struct {
	vector *prometheus.SummaryVec
}

func (c *Callbacks) Name() string {
	return "prometheus"
}

func (c *Callbacks) Initialize(db *gorm.DB) error {
	err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", c.Before())
	if err != nil {
		return err
	}

	err = db.Callback().Create().After("*").
		Register("prometheus_create_after", c.After("CREATE"))
	if err != nil {
		return err
	}

	err = db.Callback().Query().Before("*").
		Register("prometheus_query_before", c.Before())
	if err != nil {
		return err
	}

	err = db.Callback().Query().After("*").
		Register("prometheus_query_after", c.After("QUERY"))
	if err != nil {
		return err
	}

	err = db.Callback().Query().Before("*").
		Register("prometheus_raw_before", c.Before())
	if err != nil {
		return err
	}

	err = db.Callback().Raw().After("*").
		Register("prometheus_raw_after", c.After("RAW"))
	if err != nil {
		return err
	}

	err = db.Callback().Update().Before("*").
		Register("prometheus_update_before", c.Before())
	if err != nil {
		return err
	}

	err = db.Callback().Update().After("*").
		Register("prometheus_update_after", c.After("UPDATE"))
	if err != nil {
		return err
	}

	err = db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", c.Before())
	if err != nil {
		return err
	}

	err = db.Callback().Update().After("*").
		Register("prometheus_delete_after", c.After("DELETE"))
	if err != nil {
		return err
	}

	err = db.Callback().Row().Before("*").
		Register("prometheus_row_before", c.Before())
	if err != nil {
		return err
	}

	err = db.Callback().Update().After("*").
		Register("prometheus_row_after", c.After("ROW"))
	return err
}

func NewCallbacks(opts prometheus.SummaryOpts) *Callbacks {
	vector := prometheus.NewSummaryVec(opts,
		[]string{"type", "table"})
	prometheus.MustRegister(vector)
	return &Callbacks{
		vector: vector,
	}
}

func (c *Callbacks) Before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()
		db.Set("start_time", start)
	}
}

func (c *Callbacks) After(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		start, ok := val.(time.Time)
		if ok {
			duration := time.Since(start).Milliseconds()
			c.vector.WithLabelValues(typ, db.Statement.Table).
				Observe(float64(duration))
		}
	}
}
