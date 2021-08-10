package v1

import (
	"geek-time/week4/internal/domain/hello_domain"
	"geek-time/week4/pkg/application/hello_application"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Hello struct {
	hello_application.Application
}

func NewHello() Hello {
	return Hello{}
}

func (t Hello) Get(c *gin.Context) {
	data := c.DefaultQuery("name", "unknown")
	do := t.Handle(hello_domain.DO{Name: data})
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    "hello_" + data,
		"do":      do,
	})
	return
}
