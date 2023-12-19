//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook_go/webook/internal/repository"
	"webook_go/webook/internal/repository/cache"
	"webook_go/webook/internal/repository/dao"
	"webook_go/webook/internal/service"
	"webook_go/webook/internal/web"
	"webook_go/webook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis,

		// 初始化 DAO
		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewLocalCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		service.NewUserService,
		service.NewCodeService,
		// 直接基于内存的实现
		ioc.InitSMSService,
		web.NewUserHandler,
		// 你中间件呢
		// 你注册路由呢
		// 你这个地方没有用到前面的任何东西
		//gin.Default,

		ioc.InitGin,
		ioc.InitMiddlewares,
	)
	return new(gin.Engine)
}
