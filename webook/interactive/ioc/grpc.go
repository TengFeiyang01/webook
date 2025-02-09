package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "webook/webook/interactive/grpc"
	"webook/webook/pkg/grpcx"
)

func NewGRPCxServer(intrServer *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("grpc.server", &cfg); err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	intrServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
