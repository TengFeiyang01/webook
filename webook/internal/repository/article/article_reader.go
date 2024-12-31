package article

import (
	"golang.org/x/net/context"
	"webook/webook/internal/domain"
)

type ArticleReaderRepository interface {
	// Save 有就更新、没有就创建
	Save(ctx context.Context, art domain.Article) (int64, error)
}
