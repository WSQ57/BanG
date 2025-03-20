//go:build wireinject

package main

import (
	"dream/webook/internal/repository"
	"dream/webook/internal/repository/cache"
	"dream/webook/internal/repository/dao"
	"dream/webook/internal/service"
	"dream/webook/internal/web"
	ijwt "dream/webook/internal/web/jwt"
	"dream/webook/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func initWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewCodeRepository,
		repository.NewUserRepository,

		service.NewCodeService,
		service.NewUserService,

		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewWeChatOAuth2Handler,

		ijwt.NewRedisJWTHandler,

		ioc.InitMiddlewares,
		ioc.InitGin,
	)
	return new(gin.Engine)
}
