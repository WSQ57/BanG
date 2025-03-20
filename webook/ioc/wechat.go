package ioc

import (
	"dream/webook/internal/service/oauth2/wechat"
	"os"
)

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		appId = "123"
		// panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		appKey = "123"
		// panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appKey)
}
