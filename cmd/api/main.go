// cmd/api/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/demo/gitfox-pms/app/api/router"
	"gitfox.zcorp.cc/git/demo/gitfox-pms/app/constants"
	"gitfox.zcorp.cc/git/demo/gitfox-pms/app/di"
	"gitfox.zcorp.cc/git/demo/gitfox-pms/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 解析启动参数（配置文件+版本）
	configPath := flag.String("config", "../../configs/default.yaml", "配置文件路径")
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
	version := di.VersionBase
	cfg = config.LoadBaseConfig(*configPath)

	db, err := gorm.Open(mysql.Open("root:123456@tcp(localhost:3306)/gitfox_free?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})


	// 3. 使用Wire初始化所有服务
	ctx := context.Background()
	branchSvc, err := InitPmsServices(ctx, db)
	if err != nil {
		fmt.Printf("服务初始化失败: %v", err)
	}

	// 4. 初始化日志（基础配置）
	baseCfg := cfg.(config.BaseConfig)
	// logger.Init(baseCfg.Logger)

	// 5. 创建服务容器（统一管理服务实例）
	serviceContainer := router.NewServiceContainer(branchSvc)

	// 4. 注册路由
	r := router.NewRouter(serviceContainer)

	// 6. 启动服务
	port := strconv.Itoa(baseCfg.Server.Port)
	fmt.Printf("API服务启动，版本：%s，端口：%s", version, port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("服务启动失败：" + err.Error())
	}
}
