package main

import (
	"dream/webook/internal/repository"
	"dream/webook/internal/repository/dao"
	"dream/webook/internal/service"
	"dream/webook/internal/web"
	"dream/webook/internal/web/middleware"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {

	db := initDB()
	server := initWebServer()

	u := initUser(db)
	u.RegisterRoutesv1(server.Group("/users"))
	server.Run(":8080")
}

func initWebServer() *gin.Engine {
	server := gin.Default()
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
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// 步骤1
	// store := cookie.NewStore([]byte("secret"))

	// 16为最大空闲连接，给将来用，提升速度
	store, err := redis.NewStore(16, "tcp", "127.0.0.1:6379", "", []byte("5GsytMZWJd6fEHDKyhPH2tPQvbdAUdjp"), []byte("nAdm2JYmuJmyAPKkJpC6sVwsazMNFYuA"))
	if err != nil {
		panic(err)
	}

	server.Use(sessions.Sessions("mysession", store))

	// 步骤3
	// server.Use(middleware.NewLoginMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").Build())

	server.Use(middleware.NewLoginJWTMiddlewareBuilder().IgnorePaths("/users/login").IgnorePaths("/users/signup").Build())

	// v1
	// server.Use(middleware.CheckLogin())
	return server
}

func initUser(db *gorm.DB) *web.UserHandler {
	dao := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(127.0.0.1:13316)/webook"))
	if err != nil {
		// 初始化过程出错，应用就不要启动
		panic(err)
	}
	if err := dao.InitTables(db); err != nil {
		panic(err)
	}
	return db
}
