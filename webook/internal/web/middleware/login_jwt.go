package middleware

import (
	"dream/webook/internal/web"
	"encoding/gob"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT登陆校验
type LoginJWTMiddleware struct {
	paths []string
}

func NewLoginJWTMiddlewareBuilder() *LoginJWTMiddleware {
	return &LoginJWTMiddleware{}
}

func (l *LoginJWTMiddleware) IgnorePaths(path string) *LoginJWTMiddleware {
	l.paths = append(l.paths, path)
	return l
}
func (l *LoginJWTMiddleware) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		for _, path := range l.paths {
			if path == ctx.Request.URL.Path {
				return
			}
		}

		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(tokenHeader, " ") // 切Bearer
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		claims := &web.UserClaims{} // 因为parse里面会赋值，因此要传入指针

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			return []byte("svpmj5zytsDADRR2YX4ZnrJdT2xQm8BK"), nil
		})
		if err != nil {
			// 没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid || claims.UserId == 0 {
			// 没登陆
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Set("claims", claims)
	}
}
