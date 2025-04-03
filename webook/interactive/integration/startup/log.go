package startup

import (
	"go.uber.org/zap"
	"github.com/TengFeiyang01/webook/webook/pkg/logger"
)

func InitLogger() logger.LoggerV1 {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
