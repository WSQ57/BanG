package ioc

import (
	"dream/webook/internal/web"
	"dream/webook/internal/web/middleware"
	"dream/webook/pkg/ginx/middlewares/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitGin(hdl *web.UserHandler, mdls []gin.HandlerFunc) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server.Group("/users"))
	return server
}

func InitMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").IgnorePaths("/users/login_sms/code/send").IgnorePaths("/users/login_sms").Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost:3000"},         // 允许的跨域源
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},        // 允许的 HTTP 方法
		AllowHeaders:     []string{"Content-Type", "Authorization"}, // 允许的请求头
		ExposeHeaders:    []string{"x-jwt-token"},                   // 可以读这个
		AllowCredentials: true,                                      // 允许携带 Cookie
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "live.webook.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
