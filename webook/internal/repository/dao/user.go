package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound
)

type GORMUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GORMUserDAO{db: db}
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConflictErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConflictErrNo {
			// 邮箱冲突 or 手机号码冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GORMUserDAO) FindByID(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) UpdateByID(ctx context.Context, id int64, u User) error {
	err := dao.db.WithContext(ctx).Where("id = ?", id).Updates(&u).Error
	if err != nil {
		return err
	}
	return nil
}

// User 对应数据库表
type User struct {
	ID       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique;comment:邮箱"`
	Password string

	// 唯一索引允许有多个空值
	// 但是不能有多个 ""
	Phone sql.NullString `gorm:"unique"`

	// 创建时间
	Ctime int64
	// 更新时间
	Utime int64
}
