package grpc

import (
	"context"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

var _ UserServiceServer = &Server{}

func (s Server) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		User: &User{
			Id:   123,
			Name: "webook, from " + s.Name,
		},
	}, nil
}
