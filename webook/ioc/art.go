package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	artv1 "webook/webook/api/proto/gen/article/v1"
	"webook/webook/article/service"
	"webook/webook/internal/web/client/art"
)

func InitArtGRPCClient(svc service.ArticleService) artv1.ArticleServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.art", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if cfg.Secure {
		// 加载你的证书之类的
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient(cfg.Addr, opts...)
	local := art.NewArticleServiceAdapter(svc)
	remote := artv1.NewArticleServiceClient(cc)
	res := art.NewGrayScaleArticleServiceClient(local, remote)
	viper.OnConfigChange(func(e fsnotify.Event) {
		var cfg Config
		err = viper.UnmarshalKey("grpc.client.art", &cfg)
		if err != nil {
			panic(err)
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
