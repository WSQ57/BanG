package web

import (
	"dream/webook/internal/domain"
	"dream/webook/internal/service"
	"fmt"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// 定义User相关路由
type UserHandler struct {
	svc         *service.UserService
	emailExp    *regexp.Regexp
	passwordExp *regexp.Regexp
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	const (
		emailRegexp    = "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$"
		passwordRegexp = "^(?=.*\\d)(?=.*[A-z])[\\da-zA-Z]{1,72}$"
	)

	emailExp := regexp.MustCompile(emailRegexp, regexp.None)
	passwordExp := regexp.MustCompile(passwordRegexp, regexp.None)

	return &UserHandler{
		svc:         svc,
		emailExp:    emailExp,
		passwordExp: passwordExp,
	}
}

func (u *UserHandler) RegisterRoutesv1(ug *gin.RouterGroup) {
	// 分组路由
	// ug.OPTIONS("signup", )
	ug.POST("signup", u.Signup)
	// ug.POST("login", u.Login)
	// ug.POST("edit", u.Edit)
	// ug.GET("profile", u.Profile)

	ug.POST("login", u.LoginJWT)
	ug.POST("edit", u.EditJWT)
	ug.GET("profile", u.ProfileJWT)

}

func (u *UserHandler) Logout(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	sess.Options(sessions.Options{
		MaxAge: -1,
	})
	sess.Save()
	ctx.String(http.StatusOK, "退出成功")
}

func (u *UserHandler) Signup(ctx *gin.Context) {
	// 内部结构体，不给别人用
	type SignUpReq struct {
		Email           string `json:"email"` // 对应json中的email字段
		ConfirmPassWord string `json:"ConfirmPassWord"`
		Password        string `json:"password"`
	}

	var req SignUpReq
	// Bind 方法 根据Content-type解析数据到req
	// 解析错误则直接写会400的错误
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 使用正则表达式校验正则表达式
	ok, err := u.emailExp.MatchString(req.Email)

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "邮箱格式不正确")
		return
	}

	ok, err = u.passwordExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	if !ok {
		ctx.String(http.StatusOK, "密码小于9位，且由数字字母组成")
		return
	}

	if req.Password != req.ConfirmPassWord {
		ctx.String(http.StatusOK, "两次密码不一致")
		return
	}

	// 调用service方法
	err = u.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	if err == service.ErrUserDuplicateEmail {
		ctx.String(http.StatusOK, "邮箱冲突")
		return
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.String(http.StatusOK, "注册成功")
	fmt.Printf("%v", req)

}

func (u *UserHandler) LoginJWT(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"` // 对应json中的email字段
		Password string `json:"password"`
	}

	var req LoginReq
	// Bind 方法 根据Content-type解析数据到req
	// 解析错误则直接写会400的错误
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	fmt.Println(user)

	// 生成JWTtoken

	claims := UserClaims{
		UserId: user.Id,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte("svpmj5zytsDADRR2YX4ZnrJdT2xQm8BK"))
	if err != nil {
		ctx.String(http.StatusInternalServerError, "系统错误")
		return
	}
	fmt.Println(tokenStr)
	ctx.Header("x-jwt-token", tokenStr)

	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Login(ctx *gin.Context) {
	type LoginReq struct {
		Email    string `json:"email"` // 对应json中的email字段
		Password string `json:"password"`
	}

	var req LoginReq
	// Bind 方法 根据Content-type解析数据到req
	// 解析错误则直接写会400的错误
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	user, err := u.svc.Login(ctx, req.Email, req.Password)
	if err == service.ErrInvalidUserOrPassword {
		ctx.String(http.StatusOK, "用户名或密码错误")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	// 步骤2
	// 登陆成功
	// 获取session
	sess := sessions.Default(ctx)
	// 设置放在session中的值
	sess.Set("userId", user.Id)
	sess.Options(sessions.Options{
		MaxAge: 30 * 60, // 60s
	})
	sess.Save()

	ctx.String(http.StatusOK, "登录成功")
}

func (u *UserHandler) Edit(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		// YYYY-MM-DD
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	sess := sessions.Default(ctx)
	id, ok := sess.Get("userId").(int64)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}
	err = u.svc.EditProfile(ctx, domain.User{
		Id:       id,
		Birthday: birthday,
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
	})
	if err == service.ErrUserNotFound {
		ctx.String(http.StatusOK, "用户不存在")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// ctx.String(http.StatusOK, "修改成功")
	ctx.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "修改成功",
	})
}

func (u *UserHandler) EditJWT(ctx *gin.Context) {
	type EditReq struct {
		Nickname string `json:"nickname"`
		// YYYY-MM-DD
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req EditReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	claims, ok := ctx.MustGet("claims").(*UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}
	err = u.svc.EditProfile(ctx, domain.User{
		Id:       claims.UserId,
		Birthday: birthday,
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
	})
	if err == service.ErrUserNotFound {
		ctx.String(http.StatusOK, "用户不存在")
		return
	}
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	// ctx.String(http.StatusOK, "修改成功")
	ctx.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "修改成功",
	})
}
func (u *UserHandler) ProfileJWT(ctx *gin.Context) {
	// Email: "", Phone: "", Nickname: "", Birthday:"", AboutMe: ""
	type ProfileReq struct {
		Email    string
		Nickname string
		Birthday string
		AboutMe  string
	}

	if claims, ok := ctx.MustGet("claims").(*UserClaims); ok {
		u, err := u.svc.Profile(ctx, claims.UserId)
		if err != nil {
			ctx.String(http.StatusOK, "用户不存在")
			return
		}
		ctx.JSON(http.StatusOK, ProfileReq{
			Email:    u.Email,
			Nickname: u.Nickname,
			Birthday: u.Birthday.Format("2006-01-02"),
			AboutMe:  u.AboutMe,
		})
	} else {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

}

func (u *UserHandler) Profile(ctx *gin.Context) {
	// Email: "", Phone: "", Nickname: "", Birthday:"", AboutMe: ""
	type ProfileReq struct {
		Email    string
		Nickname string
		Birthday string
		AboutMe  string
	}
	sess := sessions.Default(ctx)
	if id, ok := sess.Get("userId").(int64); ok {
		u, err := u.svc.Profile(ctx, id)
		if err != nil {
			ctx.String(http.StatusOK, "用户不存在")
			return
		}
		ctx.JSON(http.StatusOK, ProfileReq{
			Email:    u.Email,
			Nickname: u.Nickname,
			Birthday: u.Birthday.Format("2006-01-02"),
			AboutMe:  u.AboutMe,
		})

	} else {
		ctx.String(http.StatusOK, "非法用户")
		return
	}
}

type UserClaims struct {
	UserId int64
	jwt.RegisteredClaims
}
