package config

const (
	Local string = "local"
	Dev   string = "dev"
	Prod  string = "prod"
)

// LoadConfig loads the configuration based on the provided mode.
func LoadConfig(mode string) {
	// 根据不同的模式加载相应的配置文件
	var configPath string
	switch mode {
	case Local:
		configPath = "../configs/local.yaml"
	case Dev:
		configPath = "../configs/dev.yaml"
	case Prod:
		configPath = "../configs/prod.yaml"
	default:
		configPath = "../configs/dev.yaml"
	}
	loadConfigFromFile(configPath)

}

func loadConfigFromFile(filePath string) {
	// 这里实现从指定文件路径加载配置的逻辑
	// 例如，使用 Viper 或其他配置库来读取 YAML 文件并解析配置
	// 示例代码（伪代码）：
	// viper.SetConfigFile(filePath)
}
