package web

import "github.com/gin-gonic/gin"

// 定义User相关路由
type UserHandler struct {
}

func (u *UserHandler) RegisterRoutesv1(ug *gin.RouterGroup) {
	// 分组路由
	ug.POST("signup", u.Signup)
	ug.POST("login", u.Login)
	ug.POST("edit", u.Edit)
	ug.GET("profil", u.Profile)
}

func (u *UserHandler) RegisterRoutes(server *gin.Engine) {
	// 分组路由
	ug := server.Group("users")
	ug.POST("signup", u.Signup)
	ug.POST("login", u.Login)
	ug.POST("edit", u.Edit)
	ug.GET("profil", u.Profile)
}

func (u *UserHandler) Signup(ctx *gin.Context) {

}

func (u *UserHandler) Login(ctx *gin.Context) {

}

func (u *UserHandler) Edit(ctx *gin.Context) {

}

func (u *UserHandler) Profile(ctx *gin.Context) {

}
