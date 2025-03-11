package middleware

import (
	"encoding/gob"
	"net/http"
	"time"

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
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
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

		// 定期刷新登陆状态，怎么知道登陆时间，尤其是多实例的分布式环境
		// 最简单的方法使用cookie来同步

		// 取上一次时间
		updateTime := sess.Get("update_time") // 问题是每次就要访问redis，访问量很大 因此采用jwt
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 30 * 60,
		})
		now := time.Now().UnixMilli()

		if updateTime == nil {
			// 刚登陆成功
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				panic(err)
			}
			return
		}
		updateTimeVal, ok := updateTime.(int64) // 断言
		if !ok {
			ctx.String(http.StatusInternalServerError, "系统错误")
		}
		if now-updateTimeVal > 60*1000 { // 1000ms
			// 刷新
			sess.Set("update_time", now)
			if err := sess.Save(); err != nil {
				panic(err)
			}
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
