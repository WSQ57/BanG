package main

import (
	"dream/webook/internal/web"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	u := &web.UserHandler{}
	// u.RegisterRoutes(server)
	u.RegisterRoutesv1(server.Group("/users"))
	server.Run(":8000")
}
