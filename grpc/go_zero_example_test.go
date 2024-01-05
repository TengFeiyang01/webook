package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func init() {

}

type GoZeroTestSuite struct {
	suite.Suite
}

// TestGoZeroClient 启动 grpc 客户端
func (s *GoZeroTestSuite) TestGoZeroClient() {
	zClient := zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:12379"},
			Key:   "user",
		},
	},
	// 强制覆盖，使用我们指定的负载均衡算法
	//zrpc.WithDialOption(grpc.WithDefaultServiceConfig()),
	)
	client := NewUserServiceClient(zClient.Conn())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}

// TestGoZeroServer 启动 grpc 服务端
func (s *GoZeroTestSuite) TestGoZeroServer() {
	// 正常来说，这个都是从配置文件中读取的
	//var c zrpc.RpcServerConf
	// 类似与 main 函数那样，从命令行接收配置文件的路径
	//conf.MustLoad(*configFile, &c)
	c := zrpc.RpcServerConf{
		ListenOn: ":8090",
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:12379"},
			Key:   "user",
		},
	}
	// 创建一个服务器，并且注册服务实例
	server := zrpc.MustNewServer(c, func(grpcServer *grpc.Server) {
		RegisterUserServiceServer(grpcServer, &Server{})
	})

	// 这个是往 gRPC 里面增加拦截器（也可以叫做插件）
	// server.AddUnaryInterceptors(interceptor)
	server.Start()
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}
