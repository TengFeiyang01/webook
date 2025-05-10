package grpc

import (
	"fmt"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/ratelimit"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
	l       logger.LoggerV1
}

// NewInterceptorBuilder key: user-service
func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.LoggerV1) *InterceptorBuilder {
	return &InterceptorBuilder{
		limiter: limiter,
		key:     key,
		l:       l,
	}
}

func (b *InterceptorBuilder) BuildServerInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if idReq, ok := req.(*GetByIdReq); ok {
			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("%s:%d", b.key, idReq.Id))
			if err != nil {
				// err 不为 nil，你要考虑保守党还是激进的
				b.l.Error("判断限流出现问题", logger.Error(err))
				return nil, status.Errorf(codes.ResourceExhausted, "限流器出错了")
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
			}
			return handler(ctx, req)
		}
		return handler(ctx, req)
	}
}
