package gohera

import "errors"

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

func ConfigNotFound(config string) error {
	return errors.New("[config] " + config + " not found")
}

func ConfigError(config string) error {
	return errors.New("[config] " + config + " error")
}

// 设置应用的错误信息
func SetMessage(errCode int, message string) error {
	if _, ok := codes[errCode]; ok {
		panic("error code conflict")
	}
	codes[errCode] = message
	return nil
}
