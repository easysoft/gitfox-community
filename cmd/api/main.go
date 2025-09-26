// cmd/api/main.go
package main

import (
	"context"
	"flag"
	"strconv"
	"strings"

	"github.com/easysoft/dev-gitfox/internal/api/router"
	"github.com/easysoft/dev-gitfox/internal/constants"
	"github.com/easysoft/dev-gitfox/internal/di"
	"github.com/easysoft/dev-gitfox/pkg/config"
	"github.com/sirupsen/logrus"
)

func main() {
	// 1. 解析启动参数（配置文件+版本）
	configPath := flag.String("config", "configs/default.yaml", "配置文件路径")
	versionStr := flag.String("version", constants.VersionBase,
		"版本标识，支持："+strings.Join(constants.SupportedVersions, ","))
	flag.Parse()

	// 校验版本是否合法（利用常量列表）
	isValid := false
	for _, v := range constants.SupportedVersions {
		if *versionStr == v {
			isValid = true
			break
		}
	}
	if !isValid {
		panic("不支持的版本，可选版本：" + strings.Join(constants.SupportedVersions, ","))
	}

	// 根据版本加载配置（使用常量替代硬编码）
	var cfg interface{}
	var version di.Version
	switch *versionStr {
	case constants.VersionEnterprise: // 企业版常量
		version = di.VersionEnterprise
		cfg = config.LoadEnterpriseConfig(*configPath)
	case constants.VersionUltimate: // 旗舰版常量
		version = di.VersionUltimate
		cfg = config.LoadUltimateConfig(*configPath)
	default: // 默认免费版
		version = di.VersionBase
		cfg = config.LoadBaseConfig(*configPath)
	}

	// 3. 初始化依赖容器并按版本启动所有模块
	container := di.NewContainer()
	ctx := context.Background()
	if err := container.InitModulesByVersion(ctx, version, cfg); err != nil {
		panic("模块初始化失败：" + err.Error())
	}

	// 4. 初始化日志（基础配置）
	baseCfg := cfg.(config.BaseConfig)
	// logger.Init(baseCfg.Logger)

	// 5. 初始化路由（从容器获取模块实例，无需关心版本）
	r := router.NewRouter(container)

	// 6. 启动服务
	port := strconv.Itoa(baseCfg.Server.Port)
	logrus.Infof("API服务启动，版本：%s，端口：%s", version, port)
	if err := r.Run(":" + port); err != nil {
		logrus.Fatal("服务启动失败：" + err.Error())
	}
}
