// +build wireinject

package main

import (
	routers "geek-time/week4/interface"
	v1 "geek-time/week4/interface/v1"
	"geek-time/week4/internal/domain/hello_domain"
	"geek-time/week4/internal/infrastructure/hello_infrastructure"
	"geek-time/week4/pkg/application/hello_application"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InjectRouters() *gin.Engine {
	// 这里只是使用wire的示例，实际开发中可能并不应该这么使用
	wire.Build(routers.NewRouter, v1.NewHello)
	return &gin.Engine{}
}

var applicationProviderSet = wire.NewSet(hello_application.NewApplication, hello_domain.NewDomain, hello_infrastructure.NewDB)

func InjectApplication() hello_application.Application {
	//wire.Build(hello_application.NewApplication, hello_domain.NewDomain, hello_infrastructure.NewDB)
	wire.Build(applicationProviderSet)
	return hello_application.Application{}
}

////var domainProviderSet = wire.NewSet(hello_domain.NewDomain, hello_infrastructure.NewDB)
//
//func InjectDomain() hello_domain.Domain {
//	wire.Build(hello_domain.NewDomain, hello_infrastructure.NewDomainDB())
//	return hello_domain.Domain{}
//}
