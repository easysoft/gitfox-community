package config

import (
	"os"
	"path/filepath"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/file"
	"gopkg.in/yaml.v3"
)

const (
	Debug string = "debug"
	Local string = "local"
	Dev   string = "dev"
	Prod  string = "prod"
	LocalConfigFileName = "local.yaml"
	DevConfigFileName   = "dev.yaml"
	ProdConfigFileName  = "prod.yaml"
)

// LoadConfig loads the configuration based on the provided mode.
func LoadConfig(mode string) (*Config, error) {
	// 获取当前可执行文件的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		return nil, errors.New("获取可执行文件路径失败: " + err.Error())
	}

	// 获取可执行文件所在目录（配置文件与二进制同目录）
	exeDir := filepath.Dir(exePath)

	// 根据模式确定配置文件名
	var fileName string
	switch mode {
		case Debug:
			exeDir = filepath.Join(filepath.Dir(filepath.Dir(exeDir)), "configs")
			fileName = LocalConfigFileName
		case Local:
			fileName = LocalConfigFileName
		case Dev:
			fileName = DevConfigFileName
		case Prod:
			fileName = ProdConfigFileName
		default:
			fileName = DevConfigFileName
	}

	configPaths, err := file.FindFilesRecursive(exeDir, fileName)
	if err != nil {
		return nil, errors.Wrapf(err,"查找配置文件失败")
	}
	if len(configPaths) == 0 {
		return nil, errors.Newf("未找到配置文件: %s", fileName)
	}
	config, err := loadConfigFromFile(configPaths[0])
	if err != nil {
		return nil, err
	}
	return config, nil
}

// loadConfigFromFile loads the configuration from a YAML file.
func loadConfigFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}