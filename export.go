package gohera

import (
	"github.com/metlive/gohera/mysql"
	"github.com/metlive/gohera/redis"
)

var (
	Mysql = make(map[string]*mysql.DB)
	Redis *redis.Client
)
