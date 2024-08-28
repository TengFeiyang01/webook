package cache

import "context"

// CodeCache 提取为一个接口，所有实现了 Set 和 Verify 的都可以接入
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}
