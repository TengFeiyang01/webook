package dao

import (
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type AuthorDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
}

func NewAuthorDAO(db *gorm.DB) AuthorDao {
	panic("implement me")
}
