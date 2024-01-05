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
	)
	require.NoError(s.T(), err)
	defer cc.Close()

	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdRequest{
			Id: 123,
		})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *KratosTestSuite) TestClientLoadBalancer() {
	// 指定一个算法
	selector.SetGlobalSelector(random.NewBuilder())
	r := etcd.New(s.etcdClient)
	cc, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///user"),
		grpc.WithDiscovery(r),
		//可以在这里传入节点的筛选过滤器
		grpc.WithNodeFilter(func(ctx context.Context, nodes []selector.Node) []selector.Node {
			res := make([]selector.Node, 0, len(nodes))
			for _, n := range nodes {
				// 我只用 vip 节点
				if n.Metadata()["vip"] == "true" {
					res = append(res, n)
				}

				//if n.Metadata()["vip"] == ctx.Value("is_vip") {
				//
				//}
			}
			if len(res) == 0 {
				// VIP 节点都崩溃了
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
		// 通过查询用户，来设置这个标记位
		ctx = context.WithValue(ctx, "is_vip", true)
		resp, err := client.GetById(ctx, &GetByIdRequest{
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
		s.startServer(":8090", "true")
	}()
	s.startServer(":8091", "false")
}

func (s *KratosTestSuite) startServer(addr string, vip string) {
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
		kratos.Metadata(map[string]string{"vip": vip, "region": "shanghai"}),
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
