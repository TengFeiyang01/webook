package integration

import (
	"context"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator/events"
	evtmocks "github.com/TengFeiyang01/webook/webook/pkg/migrator/events/mocks"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator/integration/startup"
	"github.com/TengFeiyang01/webook/webook/pkg/migrator/validator"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"testing"
)

type InteractiveTestSuite struct {
	suite.Suite
	srcDB  *gorm.DB
	intrDB *gorm.DB
}

func (i *InteractiveTestSuite) SetupSuite() {
	i.srcDB = startup.InitSrcDB()
	err := i.srcDB.AutoMigrate(&Interactive{})
	assert.NoError(i.T(), err)
	i.intrDB = startup.InitIntrDB()
	err = i.intrDB.AutoMigrate(&Interactive{})
	assert.NoError(i.T(), err)

}

func (i *InteractiveTestSuite) TearDownTest() {
	i.srcDB.Exec("TRUNCATE TABLE interactives")
	i.intrDB.Exec("TRUNCATE TABLE interactives")
}

func (i *InteractiveTestSuite) TestValidator() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		// 不想真的从 Kafka 里面读取数据，所以mock 一下
		mock func(ctrl *gomock.Controller) events.Producer

		wantErr error
	}{
		{
			name: "src有，但是intr没有",
			before: func(t *testing.T) {
				err := i.srcDB.Create(&Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      123,
					ReadCnt:    111,
					CollectCnt: 222,
					LikeCnt:    333,
					Ctime:      456,
					Utime:      678,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(),
					events.InconsistentEvent{
						Type:      events.InconsistentEventTypeTargetMissing,
						Direction: "SRC",
						ID:        1,
					}).Return(nil)
				return p
			},
		},
		{
			name: "src有，intr也有，数据相同",
			before: func(t *testing.T) {
				intr := &Interactive{
					Id:         2,
					Biz:        "test",
					BizId:      124,
					ReadCnt:    111,
					CollectCnt: 222,
					LikeCnt:    333,
					Ctime:      456,
					Utime:      678,
				}
				err := i.srcDB.Create(intr).Error
				assert.NoError(t, err)
				err = i.intrDB.Create(intr).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				return p
			},
		},
		{
			name: "src有，intr有，但是数据不同",
			before: func(t *testing.T) {
				err := i.srcDB.Create(&Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      123,
					ReadCnt:    111,
					CollectCnt: 222,
					LikeCnt:    333,
					Ctime:      456,
					Utime:      678,
				}).Error
				assert.NoError(t, err)
				err = i.intrDB.Create(&Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      123,
					ReadCnt:    111,
					CollectCnt: 222,
					LikeCnt:    33333333,
					Ctime:      456,
					Utime:      678,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(),
					events.InconsistentEvent{
						Type:      events.InconsistentEventTypeNotEqual,
						Direction: "SRC",
						ID:        1,
					}).Return(nil)
				return p
			},
		},

		{
			name: "src没有，intr有",
			before: func(t *testing.T) {
				err := i.intrDB.Create(&Interactive{
					Id:         1,
					Biz:        "test",
					BizId:      123,
					ReadCnt:    111,
					CollectCnt: 222,
					LikeCnt:    33333333,
					Ctime:      456,
					Utime:      678,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				i.TearDownTest()
			},
			mock: func(ctrl *gomock.Controller) events.Producer {
				p := evtmocks.NewMockProducer(ctrl)
				p.EXPECT().ProduceInconsistentEvent(gomock.Any(),
					events.InconsistentEvent{
						Type:      events.InconsistentEventTypeBaseMissing,
						Direction: "SRC",
						ID:        1,
					}).Return(nil)
				return p
			},
		},
	}
	for _, tc := range testCases {
		i.T().Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			tc.before(t)
			v := validator.NewValidator[Interactive](i.srcDB, i.intrDB,
				"SRC", logger.NewNopLogger(), tc.mock(ctrl))
			err := v.Validate(context.Background())
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestInteractive(t *testing.T) {
	suite.Run(t, &InteractiveTestSuite{})
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	Ctime      int64
	Utime      int64
}

func (i Interactive) ID() int64 {
	return i.Id
}

func (i Interactive) TableName() string {
	return "interactives"
}

func (i Interactive) CompareTo(entity migrator.Entity) bool {
	dst := entity.(migrator.Entity)
	return i == dst
}
