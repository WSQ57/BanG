package ioc

import (
	"dream/webook/internal/service/sms"
	"dream/webook/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 换内存，还是换别的
	return memory.NewService()
}
