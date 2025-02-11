package cache

import (
	"golang.org/x/net/context"
	"webook/webook/article/domain"
)

type Cache interface {
	Set(ctx context.Context, arts []domain.Article) error
	Get(ctx context.Context) ([]domain.Article, error)
}
