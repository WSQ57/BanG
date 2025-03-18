package main

import (
	"dream/webook/config"
	"dream/webook/internal/repository"
	"dream/webook/internal/repository/cache"
	"dream/webook/internal/repository/dao"
	"dream/webook/internal/service"
	"dream/webook/internal/service/sms/memory"
	"dream/webook/internal/web"
	"dream/webook/internal/web/middleware"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	db := initDB()
	server := initWebServer()
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	u := initUser(db, redisClient)
	u.RegisterRoutesv1(server.Group("/users"))

	// server := gin.Default()
	// server.GET("/hello", func(ctx *gin.Context) {
	// 	ctx.String(http.StatusOK, "hello go")
	// })

	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	// redisClient := redis.NewClient(&redis.Options{
	// 	Addr: config.Config.Redis.Addr,
	// })
	// server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build()) // 每秒100

	// 配置 CORS
	server.Use(cors.New(cors.Config{
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
	}))

	// 步骤1
	// store := cookie.NewStore([]byte("secret"))

	// 16为最大空闲连接，给将来用，提升速度

	// store, err := redis.NewStore(16, "tcp", "127.0.0.1:6379", "", []byte("5GsytMZWJd6fEHDKyhPH2tPQvbdAUdjp"), []byte("nAdm2JYmuJmyAPKkJpC6sVwsazMNFYuA"))
	// if err != nil {
	// panic(err)
	// }

	// store := memstore.NewStore([]byte("5GsytMZWJd6fEHDKyhPH2tPQvbdAUdjp"), []byte("nAdm2JYmuJmyAPKkJpC6sVwsazMNFYuA"))

	// server.Use(sessions.Sessions("mysession", store))

	// 步骤3
	// server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").Build())

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").IgnorePaths("/users/login_sms/code/send").IgnorePaths("/users/login_sms").Build())

	// v1
	// server.Use(middleware.CheckLogin())
	return server
}

func initUser(db *gorm.DB, rdb redis.Cmdable) *web.UserHandler {
	dao := dao.NewUserDAO(db)
	uc := cache.NewUserCache(rdb)
	repo := repository.NewUserRepository(dao, uc)
	svc := service.NewUserService(repo)

	codeCache := cache.NewCodeCache(rdb)
	codeRepo := repository.NewCodeRepository(codeCache)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc)
	u := web.NewUserHandler(svc, codeSvc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		// 初始化过程出错，应用就不要启动
		panic(err)
	}
	if err := dao.InitTables(db); err != nil {
		panic(err)
	}
	return db
}
