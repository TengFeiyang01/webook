package ioc

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	intrv1 "webook/webook/api/proto/gen/intr/v1"
	"webook/webook/interactive/service"
	"webook/webook/internal/web/client/intr"
)

func InitIntrGRPCClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
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
	local := intr.NewInteractiveServiceAdapter(svc)
	remote := intrv1.NewInteractiveServiceClient(cc)
	res := intr.NewGrayScaleInteractiveServiceClient(local, remote)
	viper.OnConfigChange(func(e fsnotify.Event) {
		var cfg Config
		err = viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			panic(err)
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
