package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

func TestGORMUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name string

		mock func(t *testing.T) *sql.DB

		ctx  context.Context
		user User

		wantErr error
	}{
		{
			name: "插入成功",
			mock: func(t *testing.T) *sql.DB {
				mockDb, mock, err := sqlmock.New()
				res := sqlmock.NewResult(3, 1)
				// 只要是 INSERT 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				require.NoError(t, err)
				return mockDb
			},
			user: User{
				Email: sql.NullString{
					String: "123@qq.com",
				},
			},
		},
		{
			name: "邮箱冲突 or 手机号冲突",
			mock: func(t *testing.T) *sql.DB {
				mockDb, mock, err := sqlmock.New()
				// 只要是 INSERT 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&mysql.MySQLError{
						Number: 1062,
					})
				require.NoError(t, err)
				return mockDb
			},
			user:    User{},
			wantErr: ErrUserDuplicate,
		},
		{
			name: "入库错误",
			mock: func(t *testing.T) *sql.DB {
				mockDb, mock, err := sqlmock.New()
				// 只要是 INSERT 到 users 的语句就行
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("入库错误"))
				require.NoError(t, err)
				return mockDb
			},
			user:    User{},
			wantErr: errors.New("入库错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 你 mock DB 不需要 Ping
				DisableAutomaticPing: true,
				// 这个是什么呢
				SkipDefaultTransaction: true,
			})
			assert.NoError(t, err)
			d := NewUserDAO(db)
			err = d.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
