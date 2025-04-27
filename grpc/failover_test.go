package grpc

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type FailoverTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *FailoverTestSuite) TestFailoverClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	cc, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
  "loadBalancingConfig": [{"round_robin": {}}],
  "methodConfig":  [
    {
      "name": [{"service":  "UserService"}],
      "retryPolicy": {
        "maxAttempts": 4,
        "initialBackoff": "0.01s",
        "maxBackoff": "0.1s",
        "backoffMultiplier": 2.0,
        "retryableStatusCodes": ["UNAVAILABLE"]
      }
    }
  ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &GetByIdReq{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}

func (s *FailoverTestSuite) SetupSuite() {
	cli, err := etcdv3.NewFromURL("localhost:12379")
	// etcdv3.NewFromURLs()
	// etcdv3.New(etcdv3.Config{Endpoints: })
	require.NoError(s.T(), err)
	s.client = cli
}

func TestFailover(t *testing.T) {
	suite.Run(t, &FailoverTestSuite{})
}

func (s *FailoverTestSuite) TestServer() {
	go func() {
		s.startServer(":8091", &AlwaysFailedServer{
			Name: "failed",
		})
	}()
	s.startServer(":8090", &Server{
		Name: "normal",
	})
}

func (s *FailoverTestSuite) startServer(addr string, svc UserServiceServer) {
	t := s.T()
	em, err := endpoints.NewManager(s.client, "service/user")
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
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(t, err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(t, err)
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		_, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
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
