//go:build wireinject

package Integration

import (
	"dream/webook/internal/repository"
	"dream/webook/internal/repository/cache"
	"dream/webook/internal/repository/dao"
	"dream/webook/internal/service"
	"dream/webook/internal/web"
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
		web.NewUserHandler,

		ioc.InitMiddlewares,
		ioc.InitGin,
	)
	return new(gin.Engine)
}
