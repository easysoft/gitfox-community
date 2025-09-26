// internal/di/module.go - 版本化模块接口定义
package di

import (
	"context"
	"fmt"
)

// Version 版本枚举（规范版本标识）
type Version string

const (
	VersionBase       Version = "base"       // 免费版
	VersionEnterprise Version = "enterprise" // 企业版
	VersionUltimate   Version = "ultimate"   // 旗舰版
)

// Module 版本化模块接口：所有需要版本差异化的模块必须实现此接口
type Module interface {
	// ModuleName 返回模块名（如 "branch"、"pipeline"），确保唯一
	ModuleName() string
	// SupportVersions 返回该模块支持的版本（如 core/branch 仅支持 base）
	SupportVersions() []Version
	// Init 初始化模块实例，返回模块的核心接口（如 branch.Service）
	Init(ctx context.Context, cfg interface{}) (interface{}, error)
}

// Container 依赖注入容器：管理所有模块的注册与实例化
type Container struct {
	modules   map[string][]Module    // 模块名 → 该模块的所有版本实现
	instances map[string]interface{} // 模块名 → 已初始化的模块实例
}

// NewContainer 创建容器实例
func NewContainer() *Container {
	return &Container{
		modules:   make(map[string][]Module),
		instances: make(map[string]interface{}),
	}
}

// Register 注册模块：各版本模块在初始化时自动调用此方法注册
func (c *Container) Register(module Module) {
	moduleName := module.ModuleName()
	c.modules[moduleName] = append(c.modules[moduleName], module)
}

// InitModulesByVersion 按版本初始化所有模块
func (c *Container) InitModulesByVersion(ctx context.Context, version Version, cfg interface{}) error {
	// 遍历所有已注册的模块，为指定版本初始化实例
	for moduleName, modules := range c.modules {
		var targetModule Module
		// 筛选当前版本支持的模块实现
		for _, m := range modules {
			for _, v := range m.SupportVersions() {
				if v == version {
					targetModule = m
					break
				}
			}
			if targetModule != nil {
				break
			}
		}
		if targetModule == nil {
			return fmt.Errorf("模块 %s 无版本 %s 的实现", moduleName, version)
		}
		// 初始化模块实例并存入容器
		instance, err := targetModule.Init(ctx, cfg)
		if err != nil {
			return fmt.Errorf("模块 %s 初始化失败：%v", moduleName, err)
		}
		c.instances[moduleName] = instance
	}
	return nil
}

// Get 获取模块实例（API层从容器中获取实例，无需关心版本）
func (c *Container) Get(moduleName string) (interface{}, error) {
	instance, ok := c.instances[moduleName]
	if !ok {
		return nil, fmt.Errorf("模块 %s 未初始化", moduleName)
	}
	return instance, nil
}
