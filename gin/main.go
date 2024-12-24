package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	// 路由注册 静态路由 此外还有参数路由和通配符路由
	server.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello go")
	})

	server.POST("/post", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "post success")
	})

	// 参数路由
	server.GET("/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		ctx.String(http.StatusOK, "传输名字为：%s", name)
	})

	// 通配符匹配
	server.GET("/views/*.html", func(ctx *gin.Context) {
		name := ctx.Param(".html")
		ctx.String(http.StatusOK, "匹配的值为：%s", name)
	})

	// 查询参数
	server.GET("/order", func(ctx *gin.Context) {
		oid := ctx.Query("id")
		ctx.String(http.StatusOK, "hello,查询参数"+oid)
	})

	server.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}
