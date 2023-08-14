package gohera

import (
	"github.com/gin-gonic/gin"
)

var (
	Mysql  = make(map[string]*DB)
	Engine *gin.Engine
	//Redis  *redis.Client
)
