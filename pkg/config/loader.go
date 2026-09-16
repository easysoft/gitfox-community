package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/easysoft/gitfox-community/pkg/errors"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

const (
	Debug = "debug"
	Local = "local"
	Dev   = "dev"
	Prod  = "prod"
)

const (
	schemeHTTP    = "http"
	schemeHTTPS   = "https"
	schemeSSH     = "ssh"
	gitfoxHomeDir = "git"
	svnHomeDir    = "svn"
)

// LoadConfig loads the configuration based on the provided mode.
func LoadConfig(mode string) (*Config, error) {
	config := new(Config)
	// 首先从 YAML 文件加载配置

	err := loadFromYaml(mode, config)

	if err != nil {
		return nil, err
	}
	// 从环境变量加载配置,可覆盖yaml中配置
	err = loadFromEnvVars(mode, config)
	if err != nil {
		return nil, err
	}

	// Datasource 为空时从拆分字段组合 DSN
	if config.Database.Datasource == "" {
		config.Database.Datasource = composeDatasource(config)
	}

	// 使用默认值覆盖配置
	config.InstanceID, err = getSanitizedMachineName()
	if err != nil {
		return nil, errors.Wrap(err, "unable to ensure that instance ID is set in config")
	}

	err = backfillURLs(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to backfill urls")
	}

	if config.WorkSpace.Root == "" {
		config.WorkSpace.Root, err = os.UserHomeDir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user home directory")
		}
	}

	if config.SVN.Root == "" {
		config.SVN.Root = filepath.Join(config.WorkSpace.Root, svnHomeDir)
	}

	if config.Git.HookPath == "" {
		config.Git.HookPath = config.WorkSpace.Root
	}

	if config.Git.Root == "" {
		homedir := config.WorkSpace.Root
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user home directory")
		}

		newPath := filepath.Join(homedir, gitfoxHomeDir)
		config.Git.Root = newPath

		oldPath := filepath.Join(homedir, ".gitrpc")
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				config.Git.Root = oldPath
			}
		}
	}

	if config.SSH.HostKeysDir == "" {
		config.SSH.HostKeysDir = filepath.Join(config.WorkSpace.Root, "ssh")
	}
	if config.Logging.LogDir == "" {
		config.Logging.LogDir = filepath.Join(config.WorkSpace.Root, "logs")
	}
	if config.Storage.Filesystem.RootDirectory == "" {
		config.Storage.Filesystem.RootDirectory = filepath.Join(config.WorkSpace.Root, "storage")
	}
	if config.Ci.ContainerVolume == "" {
		config.Ci.ContainerVolume = filepath.Join(config.WorkSpace.Root, "volumes")
	}

	return config, nil
}

// loadFromYaml loads the configuration from a YAML file based on the mode.
func loadFromYaml(mode string, config *Config) error {
	configPath, err := findConfigFilePaths(mode, "yaml")
	if err != nil {
		return err
	}
	err = loadFromFile(configPath, config)
	if err != nil {
		return err
	}
	return nil
}

// loadFromFile loads the configuration from a file.
func loadFromFile(filePath string, config *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

var getenvHookRe = regexp.MustCompile(`Getenv\(("([^"]*)"|'([^']*)')\)`)

// loadFromEnvVars 从gitfox.env文件中加载并设置为环境变量
func loadFromEnvVars(mode string, config *Config) error {
	configPath, err := findConfigFilePaths(mode, "env")
	if err != nil {
		return err
	}

	tempFile, cleanup, err := preprocessEnvFile(configPath)
	if err != nil {
		return errors.SystemErr.Wrapf(err, "failed to preprocess env file")
	}
	defer cleanup()

	//加载.env文件中配置至环境变量，若环境变量已存在则不覆盖
	err = godotenv.Load(tempFile)
	if err != nil {
		return errors.SystemErr.Wrapf(err, "failed to load env file")
	}

	err = envconfig.Process("", config)
	if err != nil {
		return err
	}
	return nil
}

func preprocessEnvFile(filePath string) (string, func(), error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, err
	}

	content := getenvHookRe.ReplaceAllStringFunc(string(data), func(match string) string {
		submatch := getenvHookRe.FindStringSubmatch(match)
		if submatch == nil {
			return match
		}
		varName := submatch[2]
		if varName == "" {
			varName = submatch[3]
		}
		return os.Getenv(varName)
	})

	tmpFile, err := os.CreateTemp("", "gitfox-env-*")
	if err != nil {
		return "", nil, err
	}

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", nil, err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		return "", nil, err
	}

	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }, nil
}

// getSanitizedMachineName gets the name of the machine and returns it in sanitized format.
func getSanitizedMachineName() (string, error) {
	// use the hostname as default id of the instance
	hostName, err := os.Hostname()
	if err != nil {
		return "", err
	}

	// Always cast to lower and remove all unwanted chars
	// NOTE: this could theoretically lead to overlaps, then it should be passed explicitly
	// NOTE: for k8s names/ids below modifications are all noops
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/

	// The following code will:
	// * remove invalid runes
	// * remove diacritical marks (ie "smörgåsbord" to "smorgasbord")
	// * lowercase A-Z to a-z
	// * leave only a-z, 0-9, '-', '.' and replace everything else with '_'
	hostName, _, err = transform.String(
		transform.Chain(
			norm.NFD,
			runes.ReplaceIllFormed(),
			runes.Remove(runes.In(unicode.Mn)),
			runes.Map(
				func(r rune) rune {
					switch {
					case 'A' <= r && r <= 'Z':
						return r + 32
					case 'a' <= r && r <= 'z':
						return r
					case '0' <= r && r <= '9':
						return r
					case r == '-', r == '.':
						return r
					default:
						return '_'
					}
				},
			),
			norm.NFC,
		),
		hostName,
	)
	if err != nil {
		return "", err
	}

	return hostName, nil
}

//nolint:gocognit // refactor if required
func backfillURLs(config *Config) error {
	// default values for HTTP
	scheme, host, port, basePath := schemeHTTP, "localhost", "3000", ""
	if config.Server.Host != "" {
		host = config.Server.Host
	}
	// by default drop scheme's default port
	if config.Server.Port > 0 &&
		(scheme != schemeHTTP || config.Server.Port != 80) &&
		(scheme != schemeHTTPS || config.Server.Port != 443) {
		port = fmt.Sprint(config.Server.Port)
	}

	// default values for SSH
	sshHost, sshPort := "localhost", ""
	if config.SSH.Host != "" {
		sshHost = config.SSH.Host
	}
	// by default drop scheme's default port
	if config.SSH.Port > 0 && config.SSH.Port != 22 {
		sshPort = fmt.Sprint(config.SSH.Port)
	}

	// backfil internal URLS before continuing override with user provided base (which is external facing)
	if config.URL.Internal == "" {
		config.URL.Internal = combineToRawURL(scheme, host, port, "")
	}
	if config.URL.Container == "" {
		config.URL.Container = combineToRawURL(scheme, "host.docker.internal", port, "")
	}

	// override base with whatever user explicit override
	//nolint:nestif // simple conditional override of all elements
	if config.URL.Base != "" {
		u, err := url.Parse(config.URL.Base)
		if err != nil {
			return errors.Newf("failed to parse base url '%s': %v", config.URL.Base, err)
		}
		if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
			return errors.Newf(
				"base url scheme '%s' is not supported (valid values: %v)",
				u.Scheme,
				[]string{
					schemeHTTP,
					schemeHTTPS,
				},
			)
		}
		// url parsing allows empty hostname - we don't want that
		if u.Hostname() == "" {
			return errors.New("a non-empty base url host has to be provided")
		}

		// take everything as is (e.g. if user explicitly adds port 80 for http we take it)
		scheme = u.Scheme
		host = u.Hostname()
		port = u.Port()
		basePath = u.Path

		// overwrite sshhost with base url host, but keep port as is
		sshHost = u.Hostname()
	} else if config.Server.ServerHost != "" {
		scheme = config.Server.Proto
		host = config.Server.ServerHost
		port = fmt.Sprint(config.Server.Port)
		sshHost = config.Server.ServerHost
	}

	clonePort := ""
	if shouldExposeClonePort(scheme, port, config.URL.ExposeRepoPort) {
		clonePort = port
	}
	sshClonePort := ""
	if shouldExposeClonePort(schemeSSH, sshPort, config.URL.ExposeRepoPort) {
		sshClonePort = sshPort
	}

	// backfill external facing URLs
	if config.URL.GitSSH == "" {
		config.URL.GitSSH = combineToRawURL(schemeSSH, sshHost, sshClonePort, "")
	}

	// create base URL object
	baseURLRaw := combineToRawURL(scheme, host, clonePort, basePath)
	baseURL, err := url.Parse(baseURLRaw)
	if err != nil {
		return errors.Newf("failed to parse derived base url '%s': %v", baseURLRaw, err)
	}

	// backfill all external URLs that weren't explicitly overwritten
	if config.URL.Base == "" {
		config.URL.Base = baseURL.String()
	}
	if config.URL.API == "" {
		config.URL.API = baseURL.JoinPath("api").String()
	}
	if config.URL.Git == "" {
		gitURL := *baseURL
		gitURL.Host = formatURLHost(gitURL.Hostname(), clonePort)
		config.URL.Git = gitURL.JoinPath("git").String()
	}
	if config.URL.UI == "" {
		uiURL := *baseURL
		uiURL.Host = formatURLHost(uiURL.Hostname(), clonePort)
		config.URL.UI = uiURL.String()
	}
	if config.URL.Registry == "" {
		config.URL.Registry = combineToRawURL(scheme, "host.docker.internal", port, "")
	}

	return nil
}

func shouldExposeClonePort(scheme string, port string, exposeRepoPort bool) bool {
	if !exposeRepoPort || port == "" {
		return false
	}

	switch scheme {
	case schemeHTTP:
		return port != "80"
	case schemeHTTPS:
		return port != "443"
	case schemeSSH:
		return port != "22"
	default:
		return true
	}
}

func formatURLHost(hostname, port string) string {
	if port == "" {
		if strings.Contains(hostname, ":") {
			return "[" + hostname + "]"
		}
		return hostname
	}

	return net.JoinHostPort(hostname, port)
}

func combineToRawURL(scheme, host, port, path string) string {
	urlRAW := scheme + "://" + formatURLHost(host, port)

	// only add path if it's not empty and non-root
	path = strings.Trim(path, "/")
	if path != "" {
		urlRAW += "/" + path
	}

	return urlRAW
}

var configDir string

// SetConfigDir 设置配置文件所在目录，覆盖默认的二进制同目录行为。
func SetConfigDir(dir string) {
	configDir = dir
}

// findConfigFilePaths finds configuration file paths based on the mode and extension.
func findConfigFilePaths(mode, ext string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", errors.New("获取可执行文件路径失败: " + err.Error())
	}

	exeDir := configDir
	if exeDir == "" {
		exeDir = filepath.Dir(exePath)
	}

	var fileName string
	if mode == Debug {
		if configDir == "" {
			exeDir = filepath.Join(filepath.Dir(filepath.Dir(exeDir)), "configs")
		}
		fileName = fmt.Sprintf("%s.%s", Local, ext)
	} else {
		fileName = fmt.Sprintf("%s.%s", mode, ext)
	}

	configPath := filepath.Join(exeDir, fileName)
	if _, err := os.Stat(configPath); err != nil {
		return "", errors.ConfigNotFoundError.Newf("未找到配置文件: %s", configPath)
	}

	return configPath, nil
}

func composeDatasource(config *Config) string {
	// GaussDB（openGauss 兼容 PostgreSQL 协议）使用 key=value 形式的 DSN
	if config.Database.Driver == "gaussdb" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.Database.Host,
			config.Database.Port,
			config.Database.User,
			config.Database.Password,
			config.Database.Name)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
		url.QueryEscape(config.Database.Timezone))
}
