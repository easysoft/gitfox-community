package config

import (
	"time"
)

// BaseConfig 基础配置结构体（所有版本共用的核心配置）
type BaseConfig struct {
	Server   ServerConfig   `yaml:"server"`   // 服务相关配置
	Database DBConfig       `yaml:"database"` // 数据库配置
	Logger   LoggerConfig   `yaml:"logger"`   // 日志配置
	Cache    CacheConfig    `yaml:"cache"`    // 缓存配置（如Redis）
	Security SecurityConfig `yaml:"security"` // 安全相关基础配置
}

// ServerConfig 服务网络配置
type ServerConfig struct {
	Port           int           `yaml:"port"`             // 服务监听端口
	ReadTimeout    time.Duration `yaml:"read_timeout"`     // HTTP读超时
	WriteTimeout   time.Duration `yaml:"write_timeout"`    // HTTP写超时
	IdleTimeout    time.Duration `yaml:"idle_timeout"`     // 连接空闲超时
	MaxHeaderBytes int           `yaml:"max_header_bytes"` // 最大请求头大小（字节）
}

// DBConfig 数据库连接配置
type DBConfig struct {
	Type     string        `yaml:"type"`     // 数据库类型（mysql/postgres/sqlite）
	Host     string        `yaml:"host"`     // 数据库主机地址
	Port     int           `yaml:"port"`     // 数据库端口
	DBName   string        `yaml:"dbname"`   // 数据库名称
	Username string        `yaml:"username"` // 用户名
	Password string        `yaml:"password"` // 密码（支持环境变量占位符）
	SSLMode  string        `yaml:"ssl_mode"` // SSL模式（如"disable"、"require"）
	MaxOpen  int           `yaml:"max_open"` // 最大打开连接数
	MaxIdle  int           `yaml:"max_idle"` // 最大空闲连接数
	Timeout  time.Duration `yaml:"timeout"`  // 连接超时
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level    string `yaml:"level"`    // 日志级别（debug/info/warn/error）
	Output   string `yaml:"output"`   // 输出方式（stdout/file）
	Path     string `yaml:"path"`     // 日志文件路径（output为file时有效）
	MaxSize  int    `yaml:"max_size"` // 单个日志文件最大大小（MB）
	MaxAge   int    `yaml:"max_age"`  // 日志文件保留天数
	Compress bool   `yaml:"compress"` // 是否压缩归档日志
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Enable   bool          `yaml:"enable"`   // 是否启用缓存
	Type     string        `yaml:"type"`     // 缓存类型（redis/memory）
	Address  string        `yaml:"address"`  // 缓存服务地址（如Redis的host:port）
	Password string        `yaml:"password"` // 缓存服务密码
	DB       int           `yaml:"db"`       // 缓存数据库编号（Redis适用）
	TTL      time.Duration `yaml:"ttl"`      // 默认过期时间
}

// SecurityConfig 基础安全配置
type SecurityConfig struct {
	JWT       JWTConfig  `yaml:"jwt"`        // JWT认证配置
	CORS      CORSConfig `yaml:"cors"`       // 跨域配置
	RateLimit int        `yaml:"rate_limit"` // 全局API限流（每秒请求数）
}

// JWTConfig JWT认证配置
type JWTConfig struct {
	Secret string        `yaml:"secret"` // 签名密钥（支持环境变量）
	Expiry time.Duration `yaml:"expiry"` // 令牌过期时间
	Issuer string        `yaml:"issuer"` // 签发者
}

// CORSConfig 跨域资源共享配置
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`     // 允许的源
	AllowMethods     []string `yaml:"allow_methods"`     // 允许的HTTP方法
	AllowHeaders     []string `yaml:"allow_headers"`     // 允许的请求头
	AllowCredentials bool     `yaml:"allow_credentials"` // 是否允许携带凭证
}

// EnterpriseConfig 企业版配置（继承基础配置）
type EnterpriseConfig struct {
	BaseConfig         `yaml:",inline"` // 嵌入基础配置，实现继承
	BranchAutoCreate   AutoCreateConfig `yaml:"branch_auto_create"` // 企业版分支自动创建配置
	PipelineEnterprise PipelineConfig   `yaml:"pipeline"`           // 企业版流水线配置
}

// UltimateConfig 旗舰版配置（继承企业版配置）
type UltimateConfig struct {
	EnterpriseConfig `yaml:",inline"`   // 嵌入企业版配置，实现继承
	Webhook          WebhookConfig      `yaml:"webhook"`       // 旗舰版Webhook配置
	SecurityScan     SecurityScanConfig `yaml:"security_scan"` // 旗舰版安全扫描配置
}

// 以下为版本特有配置结构体（仅示例，根据实际需求扩展）
type AutoCreateConfig struct { /* 企业版分支自动创建规则 */
}
type PipelineConfig struct { /* 企业版流水线高级配置 */
}
type WebhookConfig struct { /* 旗舰版Webhook配置 */
}
type SecurityScanConfig struct { /* 旗舰版安全扫描配置 */
}
