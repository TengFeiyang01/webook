package grpc

import (
	"context"
	"go.opentelemetry.io/otel"
	"log"
	"time"
)

type Server struct {
	UnimplementedUserServiceServer
	Name string
}

var _ UserServiceServer = &Server{}

func (s Server) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	ctx, span := otel.Tracer("user_server").Start(ctx, "get_by_id_biz")
	defer span.End()
	ddl, ok := ctx.Deadline()
	if ok {
		// 打印剩余超时时间
		log.Println(ddl.Sub(time.Now()).String())
	}
	time.Sleep(time.Millisecond * 50)

	return &GetByIdResp{
		User: &User{
			Id:   123,
			Name: "webook, from " + s.Name,
		},
	}, nil
}
