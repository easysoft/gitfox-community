package config

import "time"

// Config 顶级配置结构体，对应整个YAML文件
type Config struct {
	InstanceID string           `yaml:"-"`
	URL        URL              `yaml:"-"`
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	SQLite     SQLiteConfig     `yaml:"sqlite"`
	Docker     DockerConfig     `yaml:"docker"`
	Git        GitConfig        `yaml:"git"`
	Job        JobConfig        `yaml:"job"`
	Webhook    WebhookConfig    `yaml:"webhook"`
	SSH        SSHConfig        `yaml:"ssh"`
	Metric     MetricConfig     `yaml:"metric"`
	Ci         CiConfig         `yaml:"ci"`
	Gitspace   GitspaceConfig   `yaml:"gitspace"`
	Registry   RegistryConfig   `yaml:"registry"`
	Logging    LoggingConfig    `yaml:"logging"`
	Middleware MiddlewareConfig `yaml:"middleware"`
	I18n       I18nConfig       `yaml:"i18n"`
	Principal  PrincipalConfig  `yaml:"principal"`
	Devops     DevopsConfig     `yaml:"devops"`
	Storage    StorageConfig    `yaml:"storage"`
}

// ServerConfig HTTP服务配置
type ServerConfig struct {
	Port            int           `yaml:"port"`
	Host            string        `yaml:"host"`
	Proto           string        `yaml:"proto"`
	ShutdownTimeout time.Duration `yaml:"shutdownTimeout"` // 优雅关闭超时时间，默认 5s
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string        `envconfig:"GITFOX_DB_DRIVER" default:"mysql"`
	Datasource      string        `envconfig:"GITFOX_DB_DATASOURCE" required:"true"`
	PrintSQL        bool          `yaml:"printSQL"`
	MaxIdleConn     int           `yaml:"maxIdleConn"`
	MaxOpenConn     int           `yaml:"maxOpenConn"`
	ConnMaxLifeTime time.Duration `yaml:"connMaxLifeTime"`
}

// SQLiteConfig 缓存数据库配置
type SQLiteConfig struct {
	Path string `envconfig:"GITFOX_SQLITE_FILE" default:"/data/gitfox/sqlite/sqlite.db"`
}

// DockerConfig Docker配置
type DockerConfig struct {
	ApiVersion string `yaml:"apiVersion"` // 用string类型兼容版本号格式
}

// GitConfig Git配置
type GitConfig struct {
	Trace         bool   `yaml:"trace"`
	TmpDir        string `yaml:"tmpDir"`
	DefaultBranch string `yaml:"defaultBranch"`
	Root          string `envconfig:"GITFOX_GIT_ROOT_DIR" default:"/data/gitfox/git"`
	HookPath      string `yaml:"hookPath"`
}

// JobConfig holds configuration for job scheduler.
type JobConfig struct {
	InstanceID    string        `yaml:"instanceID"`
	MaxRunning    int           `yaml:"maxRunning"`
	RetentionTime time.Duration `yaml:"retentionTime"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	AllowLoopback       bool `yaml:"allowLoopback"`
	AllowPrivateNetwork bool `yaml:"allowPrivateNetwork"`
}

// SSHConfig SSH服务配置
type SSHConfig struct {
	Enable      bool   `yaml:"enable"`
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	DefaultUser string `yaml:"defaultUser" default:"git"`
	HostKeysDir string `envconfig:"GITFOX_SSH_HOST_KEYS_DIR" default:"/data/gitfox/ssh"`
}

type I18nConfig struct {
	DefaultLocale    string   `yaml:"defaultLocale"`
	FallbackLocale   string   `yaml:"fallbackLocale"`
	SupportedLocales []string `yaml:"supportedLocales"`
	TranslationsDir  string   `yaml:"translationsDir"`
	QueryParam       string   `yaml:"queryParam,omitempty"`
	CookieName       string   `yaml:"cookieName,omitempty"`
	HeaderName       string   `yaml:"headerName,omitempty"`
	SessionKey       string   `yaml:"sessionKey,omitempty"`
}

// MetricConfig 监控配置
type MetricConfig struct {
	Enabled bool `yaml:"enabled"`
}

// CiConfig CI配置
type CiConfig struct {
	PluginsZipUrl     string   `envconfig:"GITFOX_CI_PLUGINS_ZIP_URL"`
	Runner            string   `envconfig:"GITFOX_CI_RUNNER"`
	ParallelWorkers   uint     `envconfig:"GITFOX_CI_ParallelWorkers" default:"2"`
	ContainerNetworks []string `envconfig:"GITFOX_CI_CONTAINER_NETWORKS"`
}

// GitspaceConfig Gitspace配置
type GitspaceConfig struct {
	Enable bool `yaml:"enable"`
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	Http RegistryHttpConfig `yaml:"http"`
}

// RegistryHttpConfig 注册中心HTTP配置
type RegistryHttpConfig struct {
	Secret string `yaml:"secret"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level            int                `envconfig:"GITFOX_LOG_LEVEL" default:"-1"`
	LogDir           string             `envconfig:"GITFOX_LOG_DIR" default:"/data/gitfox/logs"`
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
	LogJobConf       LogJobConfig       `yaml:"logJobConf"`
}

// LogAPIServerConfig
type LogAPIServerConfig struct {
	Filename string `yaml:"filename"`
}

// LogRunnerConfig
type LogRunnerConfig struct {
	Filename string `yaml:"filename"`
}

// LogSSHConfig
type LogSSHConfig struct {
	Filename string `yaml:"filename"`
}

// LogScanConfig
type LogScanConfig struct {
	Filename string `yaml:"filename"`
}

// LogSystemConfig
type LogSystemConfig struct {
	Filename string `yaml:"filename"`
}

// LogJobConfig
type LogJobConfig struct {
	Filename string `yaml:"filename"`
}

// Middleware Config 中间件配置
type MiddlewareConfig struct {
	LogRequestBody     bool `yaml:"logRequestBody"`     // 如需记录请求体，改为 true
	LogResponseBody    bool `yaml:"logResponseBody"`    // 如需记录响应体，改为 true
	MaxLoggedBodyBytes int  `yaml:"maxLoggedBodyBytes"` // 体积上限，超过将截断
}

type URL struct {
	// Base is used to generate external facing URLs in case they aren't provided explicitly.
	// Value is derived from Server.HTTP Config unless explicitly specified (e.g. http://localhost:3000).
	Base string `envconfig:"GITFOX_URL_BASE"`

	// Git defines the external URL via which the GIT API is reachable.
	// NOTE: for routing to work properly, the request path & hostname reaching gitfox
	// have to statisfy at least one of the following two conditions:
	// - Path ends with `/git`
	// - Hostname is different to API hostname
	// (this could be after proxy path / header rewrite).
	// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000/git).
	Git string

	// GitSSH defines the external URL via which the GIT SSH server is reachable.
	// Value is derived from Base or SSH Config unless explicitly specified (e.g. ssh://localhost).
	GitSSH string

	// API defines the external URL via which the rest API is reachable.
	// NOTE: for routing to work properly, the request path reaching gitfox has to end with `/api`
	// (this could be after proxy path rewrite).
	// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000/api).
	API string
	// UI defines the external URL via which the UI is reachable.
	// Value is derived from Base unless explicitly specified (e.g. http://localhost:3000).
	UI string

	// Internal defines the internal URL via which the service is reachable.
	// Value is derived from HTTP.Server unless explicitly specified (e.g. http://localhost:3000).
	Internal string

	// Container is the endpoint that can be used by running container builds to communicate
	// with Gitfox (for example while performing a clone on a local repo).
	// host.docker.internal allows a running container to talk to services exposed on the host
	// (either running directly or via a port exposed in a docker container).
	// Value is derived from HTTP.Server unless explicitly specified (e.g. http://host.docker.internal:3000).
	Container string

	// Registry is used as a base to generate external facing URLs.
	// Value is derived from HTTP.Server unless explicitly specified (e.g. http://host.docker.internal:3000).
	Registry string
}

type PrincipalConfig struct {

	// Admin defines the principal information used to create the admin user. 管理员定义了创建管理员用户所使用的主体信息。
	// NOTE: The admin user is only auto-created in case a password and an email is provided.
	// 备注：只有在提供密码和电子邮件的情况下，才会自动创建管理员用户。
	Admin struct {
		UID         string `yaml:"uid"`
		DisplayName string `yaml:"name"`
		Email       string `yaml:"email"`    // No default email
		Password    string `yaml:"password"` // No default password
	}
}

// DevopsConfig DevOps 平台配置
type DevopsConfig struct {
	URL           string `yaml:"url"`           // DevOps 平台 API 地址
	Authorization string `yaml:"authorization"` // DevOps 平台 API 授权密钥
}

// StorageConfig 通用存储配置
type StorageConfig struct {
	StorageType string           `yaml:"storageType" envconfig:"GITFOX_STORAGE_TYPE"` // 存储驱动类型: filesystem, s3
	Filesystem  FilesystemConfig `yaml:"filesystem"`                                  // 文件系统存储配置
	S3          S3Config         `yaml:"s3"`                                          // S3 存储配置
}

type FilesystemConfig struct {
	RootDirectory string `yaml:"rootDirectory" envconfig:"GITFOX_FILESYSTEM_ROOT_DIRECTORY"` // 存储根目录
	MaxThreads    uint64 `yaml:"maxThreads"`                                                 // 最大并发线程数
	Prefix        string `yaml:"Prefix"`                                                     // 路径前缀
}
type S3Config struct {
	Enabled   bool   `yaml:"enabled"`                   // 开启 S3 存储
	Bucket    string `yaml:"bucket"`                    // S3 存储桶名称
	Prefix    string `yaml:"prefix"`                    // S3 存储路径前缀
	Endpoint  string `yaml:"endpoint"`                  // S3 服务端地址
	Region    string `yaml:"region"`                    // S3 区域，如 cn-qingdao
	PathStyle bool   `yaml:"pathStyle"`                 // 是否使用路径样式访问
	AccessKey string `envconfig:"GITFOX_S3_ACCESS_KEY"` // S3 访问密钥
	SecretKey string `envconfig:"GITFOX_S3_SECRET_KEY"` // S3 访问密钥
}
