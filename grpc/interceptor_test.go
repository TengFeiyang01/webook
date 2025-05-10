package grpc

import (
	"context"
	"github.com/TengFeiyang01/webook/webook/pkg/grpcx/interceptors/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type InterceptorTestSuite struct {
	suite.Suite
}

func (s *InterceptorTestSuite) SetupSuite() {
	initZipkin()
}

func (s *InterceptorTestSuite) TestServer() {
	t := s.T()
	//tracer := otel.Tracer("interceptor-test")
	//propagator := otel.GetTextMapPropagator()
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			NewLogInterceptor(t),
			trace.NewInterceptorBuilder(nil, nil).BuildServer(),
		))
	// 这个是生成的代码
	RegisterUserServiceServer(server, &Server{})
	l, err := net.Listen("tcp", ":8090")
	assert.NoError(t, err)
	// 启动
	if err = server.Serve(l); err != nil {
		// 启动失败，或者退出了服务器
		t.Log("退出 gRPC 服务", err)
	}
}

func (s *InterceptorTestSuite) TestClient() {
	t := s.T()
	conn, err := grpc.Dial(":8090",
		grpc.WithChainUnaryInterceptor(
			trace.NewInterceptorBuilder(nil, nil).BuildClient(),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	client := NewUserServiceClient(conn)
	for i := 0; i < 1; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		spanCtx, span := otel.Tracer("interceptor_test").Start(ctx, "client_getbyid")
		resp, err := client.GetById(spanCtx, &GetByIdReq{
			Id: 123,
		})
		cancel()
		// 模拟复杂的业务
		time.Sleep(time.Millisecond * 100)
		span.End()
		assert.NoError(t, err)
		t.Log(resp.User)
	}
	time.Sleep(time.Second * 3)
}

func NewLogInterceptor(t *testing.T) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		t.Log("请求", req, info)
		res, err := handler(ctx, req)
		t.Log("响应", resp, err)
		return res, err
	}
}

func TestInterceptor(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}
