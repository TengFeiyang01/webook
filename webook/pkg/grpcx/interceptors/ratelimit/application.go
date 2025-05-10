package ratelimit

import (
	context2 "context"
	"fmt"
	grpc2 "github.com/TengFeiyang01/webook/grpc"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/ratelimit"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
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

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		limited, err := b.limiter.Limit(ctx, b.key)
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
}

func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil || limited {
			ctx = context.WithValue(ctx, "downgrade", "true")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context2.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		limited, err := b.limiter.Limit(ctx, method)
		if err != nil {
			// err 不为 nil，你要考虑保守党还是激进的
			b.l.Error("判断限流出现问题", logger.Error(err))
			return status.Errorf(codes.ResourceExhausted, "限流器出错了")
		}
		if limited {
			return status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// BuildServerInterceptorService 服务级别的限流
func (b *InterceptorBuilder) BuildServerInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.HasPrefix(info.FullMethod, "UserService") {
			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("limiter:service:user:UserService"))
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

func (b *InterceptorBuilder) BuildServerInterceptorBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if idReq, ok := req.(grpc2.GetByIdReq); ok {
			limited, err := b.limiter.Limit(ctx, fmt.Sprintf("limiter:user:%s:%d", info.FullMethod, idReq.Id))
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
