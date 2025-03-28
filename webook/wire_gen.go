// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"dream/webook/internal/repository"
	"dream/webook/internal/repository/cache"
	"dream/webook/internal/repository/dao"
	"dream/webook/internal/service"
	"dream/webook/internal/web"
	"dream/webook/internal/web/jwt"
	"dream/webook/ioc"
	"github.com/gin-gonic/gin"
)

// Injectors from wire.go:

func initWebServer() *gin.Engine {
	db := ioc.InitDB()
	userDAO := dao.NewUserDAO(db)
	cmdable := ioc.InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	handler := jwt.NewRedisJWTHandler(cmdable)
	v := ioc.InitMiddlewares(cmdable, handler)
	wechatService := ioc.InitWechatService()
	weChatOAuth2Handler := web.NewWeChatOAuth2Handler(wechatService, userService, handler)
	engine := ioc.InitGin(userHandler, v, weChatOAuth2Handler)
	return engine
}
