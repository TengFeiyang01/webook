package circuitbreaker

import (
	"github.com/go-kratos/aegis/circuitbreaker"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
	// 设置标记为
	// 假如我们使用随机数+阈值的恢复方式
	// 触发熔断的时候，直接将 threshold 置为0
	threshold int
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if b.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			//s, ok := status.FromError(err)
			//if s != nil && s.Code() == codes.Unavailable {
			//	b.breaker.MarkFailed()
			//}
			// 进一步判断，是不是系统错误
			if err != nil {
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
			return resp, err
		}
		b.breaker.MarkFailed()
		return nil, status.Errorf(codes.Unavailable, "熔断")
	}
}
