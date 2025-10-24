package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/file"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

const (
	Debug               = "debug"
	Local               = "local"
	Dev                 = "dev"
	Prod                = "prod"
	LocalConfigFileName = "local.yaml"
	DevConfigFileName   = "dev.yaml"
	ProdConfigFileName  = "prod.yaml"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	schemeSSH   = "ssh"
	// gitfoxHomeDir = ".gitfox"
	// blobDir       = "blob"
)

// LoadConfig loads the configuration based on the provided mode.
func LoadConfig(mode string) (*Config, error) {
	config, err := loadFromYaml(mode)
	if err != nil {
		return nil, err
	}

	err = envconfig.Process("", config)
	if err != nil {
		return nil, err
	}

	config.InstanceID, err = getSanitizedMachineName()
	if err != nil {
		return nil, errors.Wrap(err, "unable to ensure that instance ID is set in config")
	}

	err = backfillURLs(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to backfill urls")
	}

	return config, nil
}

// loadFromYaml loads the configuration from a YAML file based on the mode.
func loadFromYaml(mode string) (*Config, error) {
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
		return nil, errors.Wrapf(err, "查找配置文件失败")
	}
	if len(configPaths) == 0 {
		return nil, errors.Newf("未找到配置文件: %s", fileName)
	}
	config, err := loadFromFile(configPaths[0])
	if err != nil {
		return nil, err
	}
	return config, nil
}

// loadFromFile loads the configuration from a file.
func loadFromFile(filePath string) (*Config, error) {
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
	scheme, host, port, path := schemeHTTP, "localhost", "", ""
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
		config.URL.Internal = combineToRawURL(scheme, "localhost", port, "")
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
		path = u.Path

		// overwrite sshhost with base url host, but keep port as is
		sshHost = u.Hostname()
	}

	// backfill external facing URLs
	if config.URL.GitSSH == "" {
		config.URL.GitSSH = combineToRawURL(schemeSSH, sshHost, sshPort, "")
	}

	// create base URL object
	baseURLRaw := combineToRawURL(scheme, host, port, path)
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
		config.URL.Git = baseURL.JoinPath("git").String()
	}
	if config.URL.UI == "" {
		config.URL.UI = baseURL.String()
	}
	if config.URL.Registry == "" {
		config.URL.Registry = combineToRawURL(scheme, "host.docker.internal", port, "")
	}

	return nil
}

func combineToRawURL(scheme, host, port, path string) string {
	urlRAW := scheme + "://" + host

	// only add port if explicitly provided
	if port != "" {
		urlRAW += ":" + port
	}

	// only add path if it's not empty and non-root
	path = strings.Trim(path, "/")
	if path != "" {
		urlRAW += "/" + path
	}

	return urlRAW
}
