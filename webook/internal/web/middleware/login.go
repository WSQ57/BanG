package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddleware struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddleware {
	return &LoginMiddleware{}
}

func (l *LoginMiddleware) IgnorePaths(path string) *LoginMiddleware {
	l.paths = append(l.paths, path)
	return l
}
func (l *LoginMiddleware) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}

		// 无需登陆校验
		// if ctx.Request.URL.Path == "/users/login" ||
		// 	ctx.Request.URL.Path == "/users/signup" {
		// 	return
		// }

		sess := sessions.Default(ctx)
		// 验证一下
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 无需登陆校验
		if ctx.Request.URL.Path == "/users/login" ||
			ctx.Request.URL.Path == "/users/signup" {
			return
		}

		sess := sessions.Default(ctx)
		// 验证一下
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
