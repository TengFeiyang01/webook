package ioc

import (
	"webook/webook/internal/service/sms"
	"webook/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	return memory.NewService()
}
