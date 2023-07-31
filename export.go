/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package gohera

import (
	"1v1.group/mysql"
	"1v1.group/redis"
	"1v1.group/zlog"
	"github.com/gin-gonic/gin"
)

var (
	Logger *zlog.Logger
	Mysql  *mysql.Engine
	Engine *gin.Engine
	Redis  *redis.Client
)
