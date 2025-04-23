package grpc

import (
	artv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/article/v1"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/TengFeiyang01/webook/webook/pkg/netx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *EtcdTestSuite) TestClient() {
	bd, err := resolver.NewBuilder(s.client)

	cc, err := grpc.Dial("etcd:///service/article",
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	client := artv1.NewArticleServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetPubById(ctx, &artv1.GetPubByIdRequest{
		Id: 1,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp)
}

func (s *EtcdTestSuite) TestServer() {
	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)

	// endpoint 以服务为维度，一个服务一个 Manager
	em, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(s.T(), err)

	// key 是这个实例的 key
	// 如果又 instance id， 用 instance id，否则本机的 ip+端口
	// 端口一般是从配置文件里面读
	//addr := "127.0.0.1:8090"
	addr := netx.GetOutboundIP() + ":8091"
	key := "service/article/" + addr

	// 这个 ctx 是控制创建租约的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// ttl: 租期 单位：秒
	// 过了 1/3 时间，会自动续期
	var ttl int64 = 30
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(s.T(), err)

	// 注册服务
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		// 定期的去刷新租约
		ch, err := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err)
		for kaResp := range ch {
			s.T().Log(kaResp.String(), time.Now().String())
		}
	}()

	require.NoError(s.T(), err)

	// 万一我的注册信息有变动
	go func() {
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			// AddEndpoint 覆盖语义
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err = em.AddEndpoint(ctx1, key, endpoints.Endpoint{
				Addr: addr,
				// 分组信息，权重信息，机房信息
				// 以及动态判定负载的时候，可以把你的负载信息也写到这里
				Metadata: now.String(),
			}, etcdv3.WithLease(leaseResp.ID))
			if err != nil {
				s.T().Log(err)
			}
			// Update
			//em.Update(ctx, []*endpoints.UpdateWithOpts{
			//	{
			//		Update: endpoints.Update{
			//			Op: endpoints.Add,
			//			Endpoint: endpoints.Endpoint{
			//				Addr:     addr,
			//				Metadata: now.String(),
			//			},
			//			Key: key,
			//		},
			//	},
			//})
			//require.NoError(s.T(), err)
			cancel1()
			require.NoError(s.T(), err)
		}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	err = server.Serve(l)
	s.T().Log(err)
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 先取消续约
	kaCancel()
	// 退出阶段，删除注册信息
	err = em.DeleteEndpoint(ctx, key)
	require.NoError(s.T(), err)
	server.GracefulStop()
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
