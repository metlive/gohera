package gohera

import (
	"github.com/metlive/gohera/mysql"
	"github.com/metlive/gohera/redis"
)

var (
	// Mysql MySQL 连接池映射，Key 为数据库名
	Mysql = make(map[string]*mysql.DB)
	// Redis Redis 客户端实例
	Redis *redis.Client
)
