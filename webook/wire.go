//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"practice/webook/internal/repository"
	"practice/webook/internal/repository/cache"
	"practice/webook/internal/repository/dao"
	"practice/webook/internal/service"
	"practice/webook/internal/web"
	ijwt "practice/webook/internal/web/jwt"
	"practice/webook/ioc"
)

func InitWebServer5() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,

		// 初始化 DAO
		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		// 直接基于内存实现
		ioc.InitSMSService,
		ioc.InitWechatService,

		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.NewWechatHandlerConfig,
		ijwt.NewRedisJWTHandler,
		// 你中间件呢
		// 你注册路由呢
		// 你这个地方没有用到前面的任何东西
		//gin.Default,
		//ioc.InitGin, // web.go
		//ioc.InitMiddlewares,
		ioc.InitWebServer5,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
