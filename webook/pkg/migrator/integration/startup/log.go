package startup

import "github.com/TengFeiyang01/webook/webook/pkg/logger"

func InitLog() logger.LoggerV1 {
	return logger.NewNopLogger()
}
