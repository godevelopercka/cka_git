// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook_go/webook/internal/repository"
	"webook_go/webook/internal/repository/cache"
	"webook_go/webook/internal/repository/dao"
	"webook_go/webook/internal/service"
	"webook_go/webook/internal/web"
	"webook_go/webook/internal/web/jwt"
	"webook_go/webook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := InitLog()
	v := ioc.InitMiddlewares(cmdable, handler, loggerV1)
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	codeRedisCache := cache.NewCodeCache(cmdable)
	userService := service.NewUserService(userRepository, codeRedisCache, loggerV1)
	codeRepository := repository.NewCodeRepository(codeRedisCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	wechatService := ioc.InitWechatService(loggerV1)
	wechatHandlerConfig := InitWechatHandlerConfig()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler, wechatHandlerConfig)
	articleDAO := dao.NewGORMArticleDAO(gormDB)
	articleRepository := repository.NewArticleRepository(articleDAO)
	articleService := service.NewArticleService(articleRepository)
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler() *web.ArticleHandler {
	gormDB := InitTestDB()
	articleDAO := dao.NewGORMArticleDAO(gormDB)
	articleRepository := repository.NewArticleRepository(articleDAO)
	articleService := service.NewArticleService(articleRepository)
	loggerV1 := InitLog()
	articleHandler := web.NewArticleHandler(articleService, loggerV1)
	return articleHandler
}

func InitUserSvc() service.UserService {
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	codeRedisCache := cache.NewCodeCache(cmdable)
	loggerV1 := InitLog()
	userService := service.NewUserService(userRepository, codeRedisCache, loggerV1)
	return userService
}

func InitJwtHdl() jwt.Handler {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	return handler
}

// wire.go:

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitLog)

var userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewUserCache, cache.NewCodeCache, repository.NewUserRepository, service.NewUserService)