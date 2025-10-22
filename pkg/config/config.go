package config

import "time"

// Config 顶级配置结构体，对应整个YAML文件
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	SQLite     SQLiteConfig     `yaml:"sqlite"`
	Docker     DockerConfig     `yaml:"docker"`
	Git        GitConfig        `yaml:"git"`
	Webhook    WebhookConfig    `yaml:"webhook"`
	Ssh        SshConfig        `yaml:"ssh"`
	Metric     MetricConfig     `yaml:"metric"`
	Ci         CiConfig         `yaml:"ci"`
	Gitspace   GitspaceConfig   `yaml:"gitspace"`
	Registry   RegistryConfig   `yaml:"registry"`
	Logging    LoggingConfig    `yaml:"logging"`
	Middleware MiddlewareConfig `yaml:"middleware"`
}

// ServerConfig HTTP服务配置
type ServerConfig struct {
	Port int `yaml:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Datasource      string        `yaml:"datasource"`
	PrintSQL        bool          `yaml:"printSQL"`
	MaxIdleConn     int           `yaml:"maxIdleConn"`
	MaxOpenConn     int           `yaml:"maxOpenConn"`
	ConnMaxLifeTime time.Duration `yaml:"connMaxLifeTime"`
}

// SQLiteConfig 缓存数据库配置
type SQLiteConfig struct {
	Path string `yaml:"path"`
}

// DockerConfig Docker配置
type DockerConfig struct {
	ApiVersion string `yaml:"apiVersion"` // 用string类型兼容版本号格式
}

// GitConfig Git配置
type GitConfig struct {
	Trace         bool   `yaml:"trace"`
	DefaultBranch string `yaml:"defaultBranch"`
	Root          string `yaml:"root"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	AllowLoopback       bool `yaml:"allowLoopback"`
	AllowPrivateNetwork bool `yaml:"allowPrivateNetwork"`
}

// SshConfig SSH服务配置
type SshConfig struct {
	Enable      bool   `yaml:"enable"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	HostKeysDir string `yaml:"hostKeysDir"`
}

// MetricConfig 监控配置
type MetricConfig struct {
	Enabled bool `yaml:"enabled"`
}

// CiConfig CI配置
type CiConfig struct {
	PluginsZipUrl string `yaml:"pluginsZipUrl"`
}

// GitspaceConfig Gitspace配置
type GitspaceConfig struct {
	Enable bool `yaml:"enable"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Http    RegistryHttpConfig    `yaml:"http"`
	Storage RegistryStorageConfig `yaml:"storage"`
}

// RegistryHttpConfig 注册中心HTTP配置
type RegistryHttpConfig struct {
	Secret string `yaml:"secret"`
}

// RegistryStorageConfig 注册中心存储配置
type RegistryStorageConfig struct {
	StorageType       string                  `yaml:"storageType"`
	FilesystemStorage FilesystemStorageConfig `yaml:"filesystemStorage"`
}

// FilesystemStorageConfig 文件系统存储配置
type FilesystemStorageConfig struct {
	RootDirectory string `yaml:"rootDirectory"`
}

// LoggingiConfig 日志配置
type LoggingConfig struct {
	Level            int                `yaml:"level"`
	Stdout           bool               `yaml:"stdout"`
	Caller           bool               `yaml:"caller"`
	Encoder          string             `yaml:"encoder"`
	MaxAge           int                `yaml:"maxAge"`
	MaxCount         int                `yaml:"maxCount"`
	MaxSize          int                `yaml:"maxSize"`
	LevelColor       bool               `yaml:"levelColor"`
	TimeFormat       string             `yaml:"timeFormat"`
	StackTraceLevel  int                `yaml:"stackTraceLevel"`
	LogAPIServerConf LogAPIServerConfig `yaml:"logAPIServerConf"`
	LogRunnerConf    LogRunnerConfig    `yaml:"logRunnerConf"`
	LogSSHConf       LogSSHConfig       `yaml:"logSSHConf"`
	LogScanConf      LogScanConfig      `yaml:"logScanConf"`
	LogSystemConf    LogSystemConfig    `yaml:"logSystemConf"`
}

// LogAPIServerConfig
type LogAPIServerConfig struct {
	Path string `yaml:"path"`
}

// LogRunnerConfig
type LogRunnerConfig struct {
	Path string `yaml:"path"`
}

// LogSSHConfig
type LogSSHConfig struct {
	Path string `yaml:"path"`
}

// LogScanConfig
type LogScanConfig struct {
	Path string `yaml:"path"`
}

// LogSystemConfig
type LogSystemConfig struct {
	Path string `yaml:"path"`
}

// Middleware Config 中间件配置
type MiddlewareConfig struct {
	LogRequestBody     bool `yaml:"logRequestBody"`     // 如需记录请求体，改为 true
	LogResponseBody    bool `yaml:"logResponseBody"`    // 如需记录响应体，改为 true
	MaxLoggedBodyBytes int  `yaml:"maxLoggedBodyBytes"` // 体积上限，超过将截断
}
