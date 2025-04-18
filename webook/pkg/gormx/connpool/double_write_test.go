package connpool

import (
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestConnPool(t *testing.T) {
	webook, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	require.NoError(t, err)
	err = webook.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	intr, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook_intr"))
	require.NoError(t, err)
	err = intr.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DoubleWritePool{
			src:     webook.ConnPool,
			dst:     intr.ConnPool,
			pattern: atomicx.NewValueOf(patternSrcFirst),
		},
	}))
	require.NoError(t, err)
	t.Log(db)
	err = db.Create(&Interactive{
		BizId: 123,
		Biz:   "test",
	}).Error
	require.NoError(t, err)
	err1 := db.Transaction(func(tx *gorm.DB) error {
		err1 := db.Create(&Interactive{
			BizId: 123,
			Biz:   "test1",
		}).Error
		return err1
	})
	err = db.Model(&Interactive{}).Where("id > ?", 0).Updates(map[string]any{
		"biz_id": 789,
	}).Error
	require.NoError(t, err1)
}

type Interactive struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// <bizid, biz>
	// 联合索引的列的顺序：查询条件、区分度
	BizId int64 `gorm:"uniqueIndex:biz_type_id"`
	// WHERE biz = ?
	Biz string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`

	ReadCnt    int64
	LikeCnt    int64
	CollectCnt int64
	Utime      int64
	Ctime      int64
}
