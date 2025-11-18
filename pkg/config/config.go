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
	DefaultUser string `yaml:"defaultUser"`
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
	RootDirectory string `envconfig:"GITFOX_REGISTRY_FS_ROOT_DIR" default:"/data/gitfox/registry"`
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

// Middleware Config 中间件配置
type MiddlewareConfig struct {
	LogRequestBody     bool `yaml:"logRequestBody"`     // 如需记录请求体，改为 true
	LogResponseBody    bool `yaml:"logResponseBody"`    // 如需记录响应体，改为 true
	MaxLoggedBodyBytes int  `yaml:"maxLoggedBodyBytes"` // 体积上限，超过将截断
}

type URL struct {
	// Base is used to generate external facing URLs in case they aren't provided explicitly.
	// Value is derived from Server.HTTP Config unless explicitly specified (e.g. http://localhost:3000).
	Base string

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
		UID         string `yaml:"uid"           default:"admin"`
		DisplayName string `yaml:"name"  default:"Administrator"`
		Email       string `yaml:"email"`    // No default email
		Password    string `yaml:"password"` // No default password
	}
}
