/*
 * Copyright (c) 2023. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
 * Morbi non lorem porttitor neque feugiat blandit. Ut vitae ipsum eget quam lacinia accumsan.
 * Etiam sed turpis ac ipsum condimentum fringilla. Maecenas magna.
 * Proin dapibus sapien vel ante. Aliquam erat volutpat. Pellentesque sagittis ligula eget metus.
 * Vestibulum commodo. Ut rhoncus gravida arcu.
 */

package gohera

import (
	"errors"
)

func ConfigNotFound(config string) error {
	return errors.New("[config] " + config + " not found")
}

func ConfigError(config string) error {
	return errors.New("[config] " + config + " error")
}

const (
	Success = 0

	ErrSystem       = 1000000 // 系统错误
	ErrUnknown      = 9999999 // 未知错误
	ErrInternal     = 1010101 // 内部错误
	ErrMysql        = 1010102 // Mysql错误
	ErrRedis        = 1010103 // Redis错误
	ErrAccessToken  = 1010201 // token错误
	ErrParam        = 1010301 // 参数错误
	DefaultErrorMsg = 1000001
)

// 错误码列表
var codes = make(map[int]string)

// 初始化错误码列表
func init() {
	codes[Success] = "操作成功"
	codes[ErrSystem] = "系统错误"
	codes[ErrUnknown] = "未知错误"
	codes[ErrInternal] = "内部错误"
	codes[ErrMysql] = "Mysql错误"
	codes[ErrRedis] = "Redis错误"
	codes[ErrAccessToken] = "token缺失或错误"
	codes[ErrParam] = "参数错误"
	codes[DefaultErrorMsg] = ""
}

// 获取应用设置的错误信息
func GetMessage(errCode int) string {
	v, ok := codes[errCode]
	if ok {
		return v
	}
	return codes[ErrUnknown]
}

// 设置应用的错误信息
func SetMessage(errCode int, message string) error {
	if _, ok := codes[errCode]; ok {
		panic("error code conflict")
	}
	codes[errCode] = message
	return nil
}
