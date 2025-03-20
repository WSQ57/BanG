package ioc

import (
	"dream/webook/internal/web"
	ijwt "dream/webook/internal/web/jwt"
	"dream/webook/internal/web/middleware"
	"dream/webook/pkg/ginx/middlewares/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitGin(hdl *web.UserHandler, mdls []gin.HandlerFunc, oauth2WechatHdl *web.WeChatOAuth2Handler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server.Group("/users"))
	oauth2WechatHdl.RegisteRoutes(server)
	return server
}

func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHdl(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).IgnorePaths("/users/login").IgnorePaths("/users/signup").IgnorePaths("/users/login_sms/code/send").IgnorePaths("/users/login_sms").IgnorePaths("/users/refresh_token").IgnorePaths("/oauth2/wechat/authurl").IgnorePaths("/oauth2/wechat/callback").Build(),
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost:3000"},         // 允许的跨域源
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},         // 允许的 HTTP 方法
		AllowHeaders:     []string{"Content-Type", "Authorization"},  // 允许的请求头
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"}, // 可以读这个
		AllowCredentials: true,                                       // 允许携带 Cookie
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
