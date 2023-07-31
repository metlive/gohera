/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

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
