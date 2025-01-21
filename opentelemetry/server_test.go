package opentelemetry

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"net/http"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	res, err := newResource("webook", "v0.0.1")
	require.NoError(t, err)

	prop := newPropagator()
	// 在客户端和服务端之间传递 tracing 的相关信息
	otel.SetTextMapPropagator(prop)

	// 初始化 trace provider
	// 这个 provider 就是用来在打点的时候构建 trace 的
	tp, err := newTraceProvider(res)
	require.NoError(t, err)
	defer tp.Shutdown(context.Background())
	otel.SetTracerProvider(tp)

	server := gin.Default()
	server.GET("/test", func(ginCtx *gin.Context) {
		// 名字唯一
		tracer := otel.Tracer("opentelemetry")
		var ctx context.Context = ginCtx
		ctx, span := tracer.Start(ctx, "top-span")
		defer span.End()
		span.AddEvent("发生了某事")
		time.Sleep(time.Second)
		ctx, subSpan := tracer.Start(ctx, "sub-span")
		defer subSpan.End()
		subSpan.SetAttributes(attribute.String("key", "value"))
		time.Sleep(time.Millisecond * 300)
		ginCtx.String(http.StatusOK, "测试 span")
	})
	_ = server.Run(":8082")
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New(
		"http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
