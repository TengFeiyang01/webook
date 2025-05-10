package logging

import (
	"fmt"
	"github.com/TengFeiyang01/webook/webook/pkg/grpcx/interceptors"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime"
	"time"
)

type InterceptorBuilder struct {
	l logger.LoggerV1
	//fn func(msg string, fields ...logger.Field)
	interceptors.Builder
}

func NewInterceptorBuilder(l logger.LoggerV1) *InterceptorBuilder {
	return &InterceptorBuilder{l: l}
}

func (b *InterceptorBuilder) BuildServerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		event := "normal"
		defer func() {
			// 最终输出日志
			cost := time.Since(start)

			// 发生了 panic
			if rec := recover(); rec != nil {
				switch re := rec.(type) {
				case error:
					err = re
				default:
					err = fmt.Errorf("%v", rec)
				}
				event = "recover"
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				err = status.New(codes.Internal, "panic, err "+err.Error()).Err()
			}

			fields := []logger.Field{
				// unary stream 是 grpc 的两种调用形态
				logger.String("type", "unary"),
				logger.Int64("cost", cost.Milliseconds()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
				// 客户端的信息
				logger.String("peer", b.PeerName(ctx)),
				logger.String("peer_ip", b.PeerIP(ctx)),
			}
			st, _ := status.FromError(err)
			if st != nil {
				// 错误码
				fields = append(fields, logger.String("code", st.Code().String()))
				fields = append(fields, logger.String("code_msg", st.Message()))
			}

			b.l.Info("RPC调用", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (i *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		start := time.Now()
		defer func() {
			duration := time.Since(start)
			fields := []logger.Field{
				logger.String("type", "unary"),
				logger.Int64("duration", duration.Milliseconds()),
				logger.String("method", info.FullMethod),
				logger.String("peer", i.PeerName(ctx)),
				logger.String("peer_ip", i.PeerIP(ctx)),
			}
			if err != nil {
				st, _ := status.FromError(err)
				fields = append(fields, logger.String("code", st.Code().String()), logger.String("code_message", st.Message()))
			}

			i.l.Info("RPC请求", fields...)
		}()
		return resp, err
	}
}
