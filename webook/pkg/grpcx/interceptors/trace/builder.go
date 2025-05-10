package trace

import (
	"context"
	"github.com/TengFeiyang01/webook/webook/pkg/grpcx/interceptors"
	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type InterceptorBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	interceptors.Builder
}

func NewInterceptorBuilder(tracer trace.Tracer, propagator propagation.TextMapPropagator) *InterceptorBuilder {
	return &InterceptorBuilder{tracer: tracer, propagator: propagator}
}

func (b *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	propagator := b.propagator
	if propagator == nil {
		// 这个是全局
		propagator = otel.GetTextMapPropagator()
	}
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/TengFeiyang01/webook/webook/pkg/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("client"),
	}
	return func(ctx context.Context, method string,
		req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ctx, span := tracer.Start(ctx, method,
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()
		// inject 过程
		// 要把跟 trace 有关的链路元数据，传递到服务端
		ctx = inject(ctx, propagator)
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}

func (b *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	propagator := b.propagator
	if propagator == nil {
		// 这个是全局
		propagator = otel.GetTextMapPropagator()
	}
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/TengFeiyang01/webook/webook/pkg/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context,
		req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		ctx = extract(ctx, propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(attrs...))
		defer span.End()
		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
			semconv.NetPeerNameKey.String(b.PeerName(ctx)),
			attribute.Key("net.peer.ip").String(b.PeerIP(ctx)),
		)
		defer func() {
			// 就要结束了
			if err != nil {
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	// 先看 ctx 里面有没有元数据
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	// 把元数据放回去 ctx，具体怎么放，放什么内容，由 propagator 决定
	propagators.Inject(ctx, GrpcHeaderCarrier(md))
	// 为什么还要把这个搞回去？
	return metadata.NewOutgoingContext(ctx, md)
}

func extract(ctx context.Context, p propagation.TextMapPropagator) context.Context {
	// 拿到客户端过来的链路元数据
	// "md": map[string]string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	// 把这个 md 注入到 ctx 里面
	// 根据你采用 zipkin 或者 jeager，它的注入方式不同
	return p.Extract(ctx, GrpcHeaderCarrier(md))
}

type GrpcHeaderCarrier metadata.MD

// Get returns the value associated with the passed key.
func (mc GrpcHeaderCarrier) Get(key string) string {
	vals := metadata.MD(mc).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Set stores the key-value pair.
func (mc GrpcHeaderCarrier) Set(key string, value string) {
	metadata.MD(mc).Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (mc GrpcHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(mc))
	for k := range metadata.MD(mc) {
		keys = append(keys, k)
	}
	return keys
}
