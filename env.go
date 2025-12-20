package gohera

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	DeployEnvDev  = "dev"
	DeployEnvTest = "test"
	DeployEnvPre  = "pre"
	DeployEnvProd = "prod"
)

var (
	appEnv       string
	appMode      string
	appName      string
	appNamespace string
	appVersion   string
	appPodName   string
	appPodIp     string
)

// parseEnv 解析环境变量
func parseEnv(env string) error {
	if env == "" {
		env = DeployEnvDev
	}

	switch env {
	case DeployEnvDev, DeployEnvTest, DeployEnvPre, DeployEnvProd:
		appEnv = env
	default:
		return fmt.Errorf("invalid environment: %s", env)
	}

	appMode = os.Getenv("APP_MODE")
	appName = os.Getenv("APP_NAME")
	appNamespace = os.Getenv("NAMESPACE")
	appVersion = os.Getenv("APP_VERSION")
	appPodName = os.Getenv("HOSTNAME")
	appPodIp = os.Getenv("POD_IP")

	updateGinMode()
	return nil
}

// updateGinMode 根据当前环境更新 Gin 的模式
func updateGinMode() {
	switch appEnv {
	case DeployEnvDev:
		gin.SetMode(gin.DebugMode)
	case DeployEnvTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
}

// GetEnv 获取运行环境
func GetEnv() string {
	return appEnv
}

// IsDev 是否为开发环境
func IsDev() bool {
	return appEnv == DeployEnvDev
}

// IsTest 是否为测试环境
func IsTest() bool {
	return appEnv == DeployEnvTest
}

// IsPre 是否为预发布环境
func IsPre() bool {
	return appEnv == DeployEnvPre
}

// IsProd 是否为生产环境
func IsProd() bool {
	return appEnv == DeployEnvProd
}

// GetAppMode 获取应用运行模式 (OCEAN_MODE)
func GetAppMode() string {
	return appMode
}

// GetAppName 获取应用名称 (OCEAN_APP)
func GetAppName() string {
	return appName
}

// GetAppNamespace 获取应用命名空间 (NAMESPACE)
func GetAppNamespace() string {
	return appNamespace
}

// GetAppVersion 获取应用版本 (OCEAN_VERSION)
func GetAppVersion() string {
	return appVersion
}

// GetAppPodName 获取 Pod 名称 (HOSTNAME)
func GetAppPodName() string {
	return appPodName
}

// GetAppPodIp 获取 Pod IP (POD_IP)
func GetAppPodIp() string {
	return appPodIp
}

// GetEnvWithDefault 获取环境变量，如果不存在则返回默认值
func GetEnvWithDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
