package web

import (
	"dream/webook/internal/service"
	"dream/webook/internal/service/oauth2/wechat"
	ijwt "dream/webook/internal/web/jwt"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/satori/go.uuid"
)

// WeChatOAuth2Handler 处理WeChat oauth2 请求的web handler
type WeChatOAuth2Handler struct {
	svc      wechat.Service
	userSvc  service.UserService
	stateKey []byte
	ijwt.Handler
}

func NewWeChatOAuth2Handler(svc wechat.Service, userService service.UserService, jwtHdl ijwt.Handler) *WeChatOAuth2Handler {
	return &WeChatOAuth2Handler{
		svc:      svc,
		userSvc:  userService,
		stateKey: []byte("svpmj5zytsDADRR2YX4ZnrJdT2xQm8BK"),
		Handler:  jwtHdl,
	}
}

// RegisteRoutes 注册视图函数
func (h *WeChatOAuth2Handler) RegisteRoutes(server *gin.Engine) {
	group := server.Group("/oauth2/wechat")
	group.GET("/authurl", h.AuthUrl)
	group.Any("/callback", h.CallBack)
}

// AuthUrl 返回获取微信登录code的 url
func (h *WeChatOAuth2Handler) AuthUrl(ctx *gin.Context) {
	state := uuid.NewV4().String()
	url, err := h.svc.AuthUrl(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造扫码登录URL失败",
		})
		return
	}

	// 上面url构造成功后
	// 我们需要在这里把这次会话的state存起来
	err = h.SetStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

// SetStateCookie 在构造完跳转微信的URL后，将state值以token的形式传入存到cookie
func (h *WeChatOAuth2Handler) SetStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, stateClaims{
		state,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr, 600,
		"/oauth2/wechat/callback", "", false, true)
	return nil
}

type stateClaims struct {
	State string
	jwt.RegisteredClaims
}

// CallBack 微信返回code的回调函数
func (h *WeChatOAuth2Handler) CallBack(ctx *gin.Context) {
	// 获取code 和 state
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "登录失败",
		})
		return
	}
	code := ctx.Query("code")

	info, err := h.svc.VerifyCode(ctx, code)
	fmt.Println("111", err)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 拿到微信的OpenID 之后，即登录成功，返回用户id
	// 但数据库中用户不一定存在，通过FindOrCreate写入，拿到uid, 交给jwtToken并返回
	u, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	fmt.Println("222", err)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	// 设置jwttoken
	err = h.SetLoginToken(ctx, u.Id)
	fmt.Println("333", err)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 200,
	})
}

func (h *WeChatOAuth2Handler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	// 校验state
	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		// 有人搞
		// 做监控
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return fmt.Errorf("拿不到 state 的 cookie, %w", err)
	}
	var sc stateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		// 做监控
		return fmt.Errorf("token 已经过期了, %w", err)
	}

	if sc.State != state {
		// 做监控
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "登陆失败",
		})
		return errors.New("state 不相等")
	}
	return nil
}
