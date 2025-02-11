package dao

import (
	"golang.org/x/net/context"
	"gorm.io/gorm"
)

type ReaderDAO interface {
	Upsert(ctx context.Context, art Article) error
	UpsertV2(ctx context.Context, art PublishedArticleV1) error
}

type PublishedArticle Article

// PublishedArticleV1 这个代表的是线上表
type PublishedArticleV1 struct {
	Article
}

func NewReaderDAO(db *gorm.DB) ReaderDAO {
	panic("implement me")
}
