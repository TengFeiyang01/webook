package grpc

import (
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	server := grpc.NewServer()
	// 我们的业务的 server
	RegisterUserServiceServer(server, &Server{})
	// 创建一个监听器，监听 tcp 协议，8090 端口
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	err = server.Serve(l)
	t.Log(err)
}
