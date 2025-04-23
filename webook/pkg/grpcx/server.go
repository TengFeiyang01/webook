package grpcx

import (
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
	"github.com/TengFeiyang01/webook/webook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	Port      int
	Name      string
	EtcdAddrs []string
	L         logger.LoggerV1
	kaCancel  func()
	em        endpoints.Manager
	client    *etcdv3.Client
	key       string
}

func NewServer(client etcdv3.Client, port int, name string, etcdAddrs []string, l logger.LoggerV1) *Server {
	return &Server{
		Server:    grpc.NewServer(),
		Port:      port,
		Name:      name,
		EtcdAddrs: etcdAddrs,
		L:         l,
	}
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	// 直接启动，我现在要嵌入服务注册的过程
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: s.EtcdAddrs,
	})
	if err != nil {
		return err
	}
	s.client = client
	// endpoint 以服务为维度，一个服务一个 Manager
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	s.L.Debug(s.EtcdAddrs[0])
	if err != nil {
		return err
	}
	s.em = em

	// key 是这个实例的 key
	// 如果又 instance id， 用 instance id，否则本机的 ip+端口
	// 端口一般是从配置文件里面读
	//addr := "127.0.0.1:8090"
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	s.key = key

	// 这个 ctx 是控制创建租约的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// ttl: 租期 单位：秒
	// 过了 1/3 时间，会自动续期
	var ttl int64 = 30
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	// 注册服务
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))

	kaCtx, kaCancel := context.WithCancel(context.Background())
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		// 定期的去刷新租约
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	s.kaCancel = kaCancel
	return nil
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := s.em.DeleteEndpoint(ctx, s.key)
		if err != nil {
			return err
		}
	}
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return err
		}
	}
	s.Server.GracefulStop()
	return nil
}
