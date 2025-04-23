package connpool

import (
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type Mysql2Mongo struct {
	db      gorm.ConnPool
	mdb     *mongo.Database
	pattern *atomicx.Value[string]
}
