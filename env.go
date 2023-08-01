package gohera

import (
	"errors"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	DeployEnvDev  = "Dev"
	DeployEnvTest = "Test"
	DeployEnvPre  = "Pre"
	DeployEnvProd = "Prod"
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

func init() {
	if GetEnv() == DeployEnvDev {
		gin.SetMode(gin.DebugMode)
	} else if GetEnv() == DeployEnvTest {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// 解析环境变量
func parseEnv(env string) error {
	appEnv = env
	if env == "" {
		appEnv = DeployEnvDev
	}

	appMode = os.Getenv("OCEAN_MODE")
	appName = os.Getenv("OCEAN_APP")
	appNamespace = os.Getenv("NAMESPACE")
	appVersion = os.Getenv("OCEAN_VERSION")
	appPodName = os.Getenv("HOSTNAME")
	appPodIp = os.Getenv("POD_IP")

	switch env {
	case DeployEnvProd:
		return nil
	case DeployEnvPre:
		return nil
	case DeployEnvTest:
		return nil
	case DeployEnvDev:
		return nil
	}
	return errors.New("parse env error")
}

// 获取运行环境
func GetEnv() string {
	return appEnv
}

func GetAppMode() string {
	return appMode
}

func GetAppName() string {
	return appName
}

func GetAppNamespace() string {
	return appNamespace
}

func GetAppVersion() string {
	return appVersion
}

func GetAppPodName() string {
	return appPodName
}

func GetAppPodIp() string {
	return appPodIp
}
