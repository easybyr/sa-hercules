package repository

import (
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Store 封装所有数据库访问。
type Store struct{}

// New 创建数据仓库。
func New() *Store {
	return &Store{}
}

// DB 返回默认数据库连接。
func (s *Store) DB() gdb.DB {
	return g.DB()
}
