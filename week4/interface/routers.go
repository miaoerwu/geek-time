package routers

import (
	"geek-time/week4/interface/v1"
	"github.com/gin-gonic/gin"
)

func NewRouter(hello v1.Hello) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	apiV1 := r.Group("/interface/v1")

	{
		// 路由
		apiV1.GET("/hello", hello.Get)
	}

	return r
}
