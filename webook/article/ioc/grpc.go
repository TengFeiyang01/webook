package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	grpc2 "github.com/TengFeiyang01/webook/webook/article/grpc"
	"github.com/TengFeiyang01/webook/webook/pkg/grpcx"
)

func NewGRPCxServer(artServer *grpc2.ArticleServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}

	var cfg Config
	if err := viper.UnmarshalKey("grpc.server", &cfg); err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	artServer.Register(server)

	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
