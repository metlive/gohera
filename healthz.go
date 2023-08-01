package gohera

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type healthZ struct {
	Status int    `json:"status"`
	Env    string `json:"env"`
}

func healthCheck(c *gin.Context) {
	h := &healthZ{
		Status: 200,
		Env:    GetEnv(),
	}
	c.JSON(http.StatusOK, h)
}
