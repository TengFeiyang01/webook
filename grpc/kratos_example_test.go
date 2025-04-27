package grpc

import (
	"context"
	etcd "github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/random"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	_ "google.golang.org/grpc"
	"testing"
	"time"
)

type KratosTestSuite struct {
	suite.Suite
	etcdClient *etcdv3.Client
}

func (s *KratosTestSuite) SetupSuite() {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.etcdClient = cli
}

func (s *KratosTestSuite) TestClient() {
	// 默认是 WRR 负载均衡算法
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
		grpc.WithNodeFilter(func(ctx context.Context, nodes []selector.Node) []selector.Node {
			// 你可以在这里过滤一些东西
			res := make([]selector.Node, 0, len(nodes))
			for _, n := range nodes {
				if n.Metadata()["vip"] == "true" {
					res = append(res, n)
				}
			}
			if len(res) == 0 {
				return nodes
			}
			return res
		}),
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdReq{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *KratosTestSuite) TestClientLoadBalancer() {
	selector.SetGlobalSelector(random.NewBuilder())
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdReq{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

// TestServer 启动服务器
func (s *KratosTestSuite) TestServer() {
	go func() {
		s.startServer(":8090")
	}()
	s.startServer(":8091")
}

func (s *KratosTestSuite) startServer(addr string) {
	grpcSrv := grpc.NewServer(
		grpc.Address(addr),
		grpc.Middleware(recovery.Recovery()),
	)
	RegisterUserServiceServer(grpcSrv, &Server{
		Name: addr,
	})
	// etcd 注册中心
	r := etcd.New(s.etcdClient)
	app := kratos.New(
		kratos.Name("user"),
		kratos.Server(
			grpcSrv,
		),
		kratos.Registrar(r),
	)
	app.Run()
}

func TestKratos(t *testing.T) {
	suite.Run(t, new(KratosTestSuite))
}
