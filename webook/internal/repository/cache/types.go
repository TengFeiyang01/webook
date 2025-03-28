package cache

import (
	"github.com/TengFeiyang01/webook/webook/article/domain"
	"golang.org/x/net/context"
)

type Cache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}
