package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AlwaysFailedServer struct {
	UnimplementedUserServiceServer
	Name string
}

var _ UserServiceServer = &Server{}

func (s AlwaysFailedServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	fmt.Printf("failover 收到请求: %v\n", req)
	return &GetByIdResp{
		User: &User{
			Id:   123,
			Name: "来自永远失败的服务端结点, from " + s.Name,
		},
	}, status.Errorf(codes.Unavailable, "模拟服务端异常")
}
