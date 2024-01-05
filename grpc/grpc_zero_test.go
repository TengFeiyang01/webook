package grpc

//type GoZeroTestSuite struct {
//	suite.Suite
//	client *etcdv3.Client
//}
//
//func (s *GoZeroTestSuite) TestGoZeroClient() {
//	c := zrpc.RpcClientConf{
//		Etcd: discov.EtcdConf{
//			Hosts: []string{"localhost:12379"},
//			Key:   "user",
//		},
//	}
//	zClient := zrpc.MustNewClient(c)
//	cc := zClient.Conn()
//	client := NewUserServiceClient(cc)
//	resp, err := client.GetById(context.Background(), &GetByIdRequest{Id: 123})
//	require.NoError(s.T(), err)
//	s.T().Log(resp.User)
//}
//
//func (s *GoZeroTestSuite) TestGoZeroServer() {
//	c := zrpc.RpcServerConf{
//		// 这个是服务启动的地址
//		ListenOn: ":8090",
//		Etcd: discov.EtcdConf{
//			Hosts: []string{"localhost:12379"},
//			// 你的服务名
//			Key: "user",
//		},
//	}
//	server := zrpc.MustNewServer(c, func(server *grpc.Server) {
//		// 吧你的业务注册到 server 里面
//		RegisterUserServiceServer(server, &Server{})
//	})
//	server.Start()
//}
//
//func TestGoZeroTestSuite(t *testing.T) {
//	suite.Run(t, new(GoZeroTestSuite))
//}
