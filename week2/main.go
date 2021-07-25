package main

import (
	"database/sql"
	"errors"
	"geek-time/week2/internal/dao"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	// 当dao层返回sql.ErrNoRows错误时，一般出现于查询的数据在数据库中不存在的情况
	// 对于这种情况一般需要单独处理，不需要wrap
	r := gin.Default()
	r.GET("/data", func(c *gin.Context) {
		id := c.DefaultQuery("id", "")
		// 模拟dao查询
		data, err := dao.QuerySomethingFromDB(id)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{
				"message": "success",
				"data":    nil,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "success",
				"data":    data,
			})
		}
	})
	_ = r.Run()
}
