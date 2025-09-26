package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadBaseConfig 加载免费版基础配置
func LoadBaseConfig(path string) BaseConfig {
	data, err := readConfigFile(path)
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}

	var cfg BaseConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic("解析基础配置失败: " + err.Error())
	}

	// 替换环境变量占位符（如 ${DB_PASSWORD:123456}）
	resolveEnvVars(&cfg)
	return cfg
}

// LoadEnterpriseConfig 加载企业版配置（继承基础配置）
func LoadEnterpriseConfig(path string) EnterpriseConfig {
	// 1. 先加载基础配置
	baseCfg := LoadBaseConfig(path)

	// 2. 解析企业版扩展配置
	data, err := readConfigFile(path)
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}

	var versionedCfg struct {
		Versions struct {
			Enterprise EnterpriseConfig `yaml:"enterprise"`
		} `yaml:"versions"`
	}
	if err := yaml.Unmarshal(data, &versionedCfg); err != nil {
		panic("解析企业版配置失败: " + err.Error())
	}

	// 3. 合并基础配置与企业版配置（企业版配置覆盖基础配置同名字段）
	enterpriseCfg := versionedCfg.Versions.Enterprise
	enterpriseCfg.BaseConfig = baseCfg

	// 4. 替换环境变量占位符
	resolveEnvVars(&enterpriseCfg)
	return enterpriseCfg
}

// LoadUltimateConfig 加载旗舰版配置（继承企业版配置）
func LoadUltimateConfig(path string) UltimateConfig {
	// 1. 先加载企业版配置（包含基础配置）
	enterpriseCfg := LoadEnterpriseConfig(path)

	// 2. 解析旗舰版扩展配置
	data, err := readConfigFile(path)
	if err != nil {
		panic("加载配置文件失败: " + err.Error())
	}

	var versionedCfg struct {
		Versions struct {
			Ultimate UltimateConfig `yaml:"ultimate"`
		} `yaml:"versions"`
	}
	if err := yaml.Unmarshal(data, &versionedCfg); err != nil {
		panic("解析旗舰版配置失败: " + err.Error())
	}

	// 3. 合并企业版配置与旗舰版配置
	ultimateCfg := versionedCfg.Versions.Ultimate
	ultimateCfg.EnterpriseConfig = enterpriseCfg

	// 4. 替换环境变量占位符
	resolveEnvVars(&ultimateCfg)
	return ultimateCfg
}

// 读取配置文件内容
func readConfigFile(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	return os.ReadFile(path)
}

// 解析配置中的环境变量占位符（如 ${VAR_NAME:default}）
func resolveEnvVars(cfg interface{}) {
	// 实际实现会递归遍历配置结构体，替换所有字符串字段中的环境变量占位符
	// 简化示例：具体实现需使用反射处理结构体字段
}
