package ioc

import (
	artv1 "github.com/TengFeiyang01/webook/webook/api/proto/gen/article/v1"
	"github.com/TengFeiyang01/webook/webook/article/service"
	"github.com/TengFeiyang01/webook/webook/internal/web/client/art"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitArtGRPCClientV1(client *clientv3.Client) artv1.ArticleServiceClient {
	type Config struct {
		Name   string `yaml:"name"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.art", &cfg)
	if err != nil {
		panic(err)
	}
	bd, err := resolver.NewBuilder(client)
	if err != nil {
		panic(err)
	}
	var opts = []grpc.DialOption{grpc.WithResolvers(bd)}
	if cfg.Secure {

	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.NewClient("etcd:///service/"+cfg.Name, opts...)
	if err != nil {
		panic(err)
	}
	return artv1.NewArticleServiceClient(cc)
}

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
