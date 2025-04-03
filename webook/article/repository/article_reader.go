package repository

import (
	"golang.org/x/net/context"
	"github.com/TengFeiyang01/webook/webook/article/domain"
)

type ArticleReaderRepository interface {
	// Save 有就更新、没有就创建
	Save(ctx context.Context, art domain.Article) (int64, error)
}
