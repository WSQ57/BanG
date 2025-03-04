package main

import (
	"dream/webook/internal/web"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	// u := &web.UserHandler{}

	u := web.NewUserHandler()

	// 配置 CORS
	server.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost:3000"},         // 允许的跨域源
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},        // 允许的 HTTP 方法
		AllowHeaders:     []string{"Content-Type", "Authorization"}, // 允许的请求头
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

	u.RegisterRoutesv1(server.Group("/users"))
	server.Run(":8080")
}
