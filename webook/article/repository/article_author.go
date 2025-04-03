package repository

import (
	"golang.org/x/net/context"
	"github.com/TengFeiyang01/webook/webook/article/domain"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}
