package dao

import (
	"database/sql"
	"fmt"
)

// QuerySomethingFromDB querySomethingFromDB 模拟查询数据的dao方法
// 查询指定id的数据
func QuerySomethingFromDB(id string) (string, error) {
	if id == "" {
		return "", sql.ErrNoRows
	} else {
		return fmt.Sprintf("{id: %s}", id), nil
	}
}
