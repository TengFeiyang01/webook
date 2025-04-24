package grpc

import (
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
)

type BalancerTestSuite struct {
	suite.Suite
	cli *etcdv3.Client
}

func (s *BalancerTestSuite) TestRoundRobinClient() {
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	svcCfg := `
{
	"loadBalancingConfig": [
		{
			"weighted_round_robin": {}
		}
	]
}
`
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(svcCfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
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

func (s *BalancerTestSuite) TestServer() {
	go func() {
		s.startServer(":8090", 10, &Server{Name: "8090"})
	}()
	s.startServer(":8091", 20, &Server{Name: "8091"})
}

func (s *BalancerTestSuite) startServer(addr string, weight int, svc UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.cli, "service/user")
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr = "127.0.0.1" + addr
	key := "service/user/" + addr
	l, err := net.Listen("tcp", addr)
	require.NoError(s.T(), err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 租期
	var ttl int64 = 5
	leaseResp, err := s.cli.Grant(ctx, ttl)
	require.NoError(t, err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		// 定位信息，客户端怎么连你
		Addr: addr,
		Metadata: map[string]any{
			"weight": weight,
		},
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err1 := s.cli.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(t, err1)
		//for kaResp := range ch {
		//t.Log(kaResp.String())
		//}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, svc)
	server.Serve(l)
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		t.Log(err)
	}
	server.GracefulStop()
	//s.cli.Close()
}
