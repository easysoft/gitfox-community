package service

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/quicklyon/kingpin/v2"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/configs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
)

// t 是 clii18n.T 的简写，用于在代码中简洁地调用翻译函数。
var t = clii18n.T

const defaultPort = "3052"

// dbConfig 保存从禅道 my.php 中解析出的数据库连接信息。
type dbConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// toDSN 将数据库配置格式化为 MySQL DSN 连接字符串。
// 例如: root:password@tcp(127.0.0.1:3306)/zentao?parseTime=true&loc=Local
func (d *dbConfig) toDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// runInit 是交互式 init 命令的入口函数。
//
// 整体流程：
//  1. 显示欢迎横幅
//  2. 询问是否与禅道安装在同一机器
//  3. 根据选择进入不同的交互流程（同机/异机），收集数据库、禅道IP、安装目录等信息
//  4. 自动获取本机 IP 地址
//  5. 将配置保存到用户指定安装目录下的 local.env
//  6. 将当前二进制复制到安装目录
//  7. 打印配置摘要
func runInit(_ *kingpin.ParseContext) error {
	// ========== 第零步：启动前置安全拦截 (必须具备 Root 提权环境) ==========
	if err := CheckSudo(); err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	// ========== 第一步：显示欢迎横幅 ==========
	printBanner(t(i18n.CliInitWelcome))

	// ========== 第二步：询问是否与禅道在同一机器 ==========
	var sameMachine bool
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(t(i18n.CliInitSameMachine)).
				Affirmative(t(i18n.CliInitYes)).
				Negative(t(i18n.CliInitNo)).
				Value(&sameMachine),
		),
	).Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	var db dbConfig
	var zentaoIP string
	var workspace string
	var zentaoPath string // 禅道配置目录路径（仅同机模式下有值）

	// 尝试加载自定义的 prod.env
	existingEnv := loadExistingEnv()
	// 尝试加载内置的 prod.env
	templateEnv := loadTemplateEnv()

	defaultWorkspace := existingEnv["GITFOX_WORKSPACE_ROOT_DIR"]
	if defaultWorkspace == "" {
		defaultWorkspace = templateEnv["GITFOX_WORKSPACE_ROOT_DIR"]
	}
	if defaultWorkspace == "" {
		defaultWorkspace = t(i18n.CliInitWorkspacePlaceholder)
	}

	// ========== 第三步：根据用户选择进入不同交互流程 ==========
	if sameMachine {
		// 同机模式：从禅道配置目录中读取 my.php 获取数据库信息
		if err := runSameMachineFlow(&db, &zentaoIP, &workspace, &zentaoPath, defaultWorkspace); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
	} else {
		// 异机模式：需要用户手动输入数据库连接信息

		if err := runDifferentMachineFlow(&db, &zentaoIP, &workspace, defaultWorkspace); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
	}

	// ========== 第四步：自动获取本机 IP 并拼接服务地址 ==========
	localIP := getLocalIP()

	zentaoBaseURL := normalizeZentaoURL(zentaoIP)
	devopsURL := fmt.Sprintf("%s/api.php/v1", zentaoBaseURL)

	gitfoxPort := existingEnv["GITFOX_HTTP_PORT"]
	if gitfoxPort == "" {
		gitfoxPort = templateEnv["GITFOX_HTTP_PORT"]
	}
	if gitfoxPort == "" {
		gitfoxPort = defaultPort
	}

	urlBase := fmt.Sprintf("http://%s:%s", localIP, gitfoxPort)

	// ========== 验证禅道配置信息 ==========
	var verifyErr error
	_ = spinner.New().
		Title(t(i18n.CliInitVerifying)).
		Action(func() {
			verifyErr = CheckZentao(devopsURL)
		}).Run()

	if verifyErr != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println()
		fmt.Println(errStyle.Render("✗ 验证失败: " + verifyErr.Error()))
		fmt.Println()
		// 直接返回避免 kingpin 吃掉错误
		return verifyErr
	}

	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)
	// fmt.Println()
	// fmt.Println(successStyle.Render("✓ ") + t(i18n.CliInitSuccessZentaoConn))
	// fmt.Println()

	// （之前打印警告的逻辑已移动到输入的 Validator 中直接进行硬性拦截）

	// ========== 检查文件是否已存在并询问是否覆盖 ==========
	targetEnvPath := filepath.Join(workspace, "prod.env")
	targetYamlPath := filepath.Join(workspace, "prod.yaml")
	targetBinPath := filepath.Join(workspace, "gitfox")
	if runtime.GOOS == "windows" {
		targetBinPath += ".exe"
	}

	// --- 检查 prod.env 是否已存在 ---
	if info, err := os.Stat(targetEnvPath); err == nil && !info.IsDir() {
		var overwrite bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf(t(i18n.CliInitOverwriteEnv), targetEnvPath)).
			Affirmative(t(i18n.CliInitYes)).
			Negative(t(i18n.CliInitNo)).
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println()
			fmt.Println(t(i18n.CliInitOverwriteCancel))
			return nil
		}
		if bakPath, err := backupFile(targetEnvPath); err == nil {
			fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliInitBackupCreated), bakPath))
		}
	}

	// --- 检查 prod.yaml 是否已存在 ---
	if info, err := os.Stat(targetYamlPath); err == nil && !info.IsDir() {
		var overwrite bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf(t(i18n.CliInitOverwriteYaml), targetYamlPath)).
			Affirmative(t(i18n.CliInitYes)).
			Negative(t(i18n.CliInitNo)).
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println()
			fmt.Println(t(i18n.CliInitOverwriteCancel))
			return nil
		}
		if bakPath, err := backupFile(targetYamlPath); err == nil {
			fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliInitBackupCreated), bakPath))
		}
	}

	// --- 检查 gitfox 二进制是否已存在 ---
	if info, err := os.Stat(targetBinPath); err == nil && !info.IsDir() {
		var overwrite bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf(t(i18n.CliInitOverwriteBin), targetBinPath)).
			Affirmative(t(i18n.CliInitYes)).
			Negative(t(i18n.CliInitNo)).
			Value(&overwrite).
			Run()
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Println()
			fmt.Println(t(i18n.CliInitOverwriteCancel))
			return nil
		}
		if bakPath, err := backupFile(targetBinPath); err == nil {
			fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliInitBackupCreated), bakPath))
		}
	}

	// ========== 第五步：保存配置到安装目录的 prod.env ==========
	// fmt.Println(successStyle.Render("✓ ") + t(i18n.CliInitSavingConfig))

	envPath, err := saveEnvFile(existingEnv, &db, workspace, localIP, devopsURL, urlBase)
	if err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println(errStyle.Render("✗ " + fmt.Sprintf(t(i18n.CliInitErrSaveConfig), err)))
		return fmt.Errorf(t(i18n.CliInitErrSaveConfig), err)
	}

	// ========== 第五步（续）：释放内置 prod.yaml 到安装目录 ==========
	if err := copyEmbeddedYaml(workspace); err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println(errStyle.Render("✗ prod.yaml: " + fmt.Sprintf("%v", err)))
		return err
	}
	// fmt.Println(successStyle.Render("✓ ") + "prod.yaml → " + filepath.Join(workspace, "prod.yaml"))

	// ========== 第六步：复制当前二进制到安装目录 ==========
	if _, err := copyBinary(workspace); err != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println(errStyle.Render("✗ " + fmt.Sprintf(t(i18n.CliInitErrCopyBin), err)))
		return fmt.Errorf(t(i18n.CliInitErrCopyBin), err)
	}

	// ========== 第七步：处理禅道 my.php 的联动信息配置 ==========
	if sameMachine && zentaoPath != "" {
		myPHPPath := filepath.Join(zentaoPath, "my.php")
		portInt, _ := strconv.Atoi(gitfoxPort)
		if err := updateZentaoPHPConfig(myPHPPath, localIP, portInt); err != nil {
			errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliInitWarnMyphpWrite), err)))
		} else {
			fmt.Println(successStyle.Render("✓ ") + t(i18n.CliInitSuccessMyphpWrite))
		}
	} else {
		// 异机部署，打印向导性指引文本
		fmt.Println()
		fmt.Println(t(i18n.CliInitManualMyphpHint))
		fmt.Printf("   $config->devops->gitfoxURL = 'http://%s';\n", localIP)
		fmt.Printf("   $config->devops->gitfoxPort = %s;\n", gitfoxPort)
	}

	// ========== 第八步：打印配置摘要 ==========
	printSummary(envPath, &db, workspace, localIP, devopsURL)

	// 记录 workspace 路径，供 install 命令使用
	lastInitWorkspace = workspace

	return nil
}

// runSameMachineFlow 处理与禅道在同一机器上的交互流程。
//
// 流程：
//  1. 让用户输入禅道配置文件目录（包含 my.php 的目录）
//  2. 校验目录存在且包含 my.php 文件
//  3. 解析 my.php 提取数据库连接信息
//  4. 计算默认安装目录（禅道目录的父目录/gitfox）
//  5. 让用户输入禅道服务 IP 和确认安装目录
//  6. 返回 zentaoPath，供后续写回 my.php 使用
func runSameMachineFlow(db *dbConfig, zentaoIP *string, workspace *string, zentaoPath *string, defaultWorkspace string) error {
	// 第一步：让用户输入禅道配置目录，并进行实时校验
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t(i18n.CliInitZentaoPath)).
				Description(t(i18n.CliInitZentaoPathDesc)).
				Placeholder(t(i18n.CliInitZentaoPathPlaceholder)).
				Value(zentaoPath).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New(t(i18n.CliInitErrRequired))
					}
					// 检查目录是否存在
					info, err := os.Stat(s)
					if err != nil || !info.IsDir() {
						return errors.New(t(i18n.CliInitErrDirNotFound))
					}
					// 检查目录下是否包含 my.php 文件
					myPHPPath := filepath.Join(s, "my.php")
					if _, err := os.Stat(myPHPPath); err != nil {
						return errors.New(t(i18n.CliInitErrMyphpNotFound))
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return err
	}

	// 第二步：解析 my.php 提取数据库连接信息
	myPHPPath := filepath.Join(*zentaoPath, "my.php")
	parsed, err := parsePHPConfig(myPHPPath)
	if err != nil {
		return fmt.Errorf("%s: %w", t(i18n.CliInitErrMyphpParse), err)
	}
	*db = *parsed

	// 第二步其二：立即测试提取的数据库信息
	var verifyErr error
	_ = spinner.New().Title(t(i18n.CliInitTestingDbConn)).Action(func() {
		verifyErr = CheckDB(db.toDSN())
	}).Run()

	if verifyErr != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println(errStyle.Render("✗ " + fmt.Sprintf(t(i18n.CliInitErrDbConn), verifyErr.Error())))
		return errors.New(t(i18n.CliInitErrDbExtractedFail))
	}

	// successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)
	// fmt.Println(successStyle.Render("✓ ") + t(i18n.CliInitSuccessDbConn))

	// 第三步：回显默认安装目录
	*workspace = defaultWorkspace

	// 第四步：让用户输入禅道 IP 和确认/修改安装目录
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t(i18n.CliInitZentaoIp)).
				Placeholder(t(i18n.CliInitZentaoIpPlaceholder)).
				Value(zentaoIP).
				Validate(requiredURLOrHostValidator),
			huh.NewInput().
				Title(t(i18n.CliInitWorkspace)).
				Placeholder(t(i18n.CliInitWorkspacePlaceholder)).
				Value(workspace).
				Validate(requiredAbsPathValidator),
		),
	).Run()

	return err
}

// runDifferentMachineFlow 处理与禅道不在同一机器上的交互流程。
//
// 流程：
//  1. 让用户手动输入数据库连接信息（主机、端口、库名、用户名、密码）
//  2. 让用户输入禅道服务 IP 和安装目录
func runDifferentMachineFlow(db *dbConfig, zentaoIP *string, workspace *string, defaultWorkspace string) error {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	// 使用翻译后的 placeholder 作为默认值预填充
	db.Host = t(i18n.CliInitDbHostPlaceholder)
	db.Port = t(i18n.CliInitDbPortPlaceholder)
	db.Name = t(i18n.CliInitDbNamePlaceholder)
	db.User = t(i18n.CliInitDbUserPlaceholder)
	*workspace = defaultWorkspace

	// 第一组表单：数据库连接信息（5 个字段）
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t(i18n.CliInitDbHost)).
				Description(t(i18n.CliInitDbHostDesc)).
				Placeholder(t(i18n.CliInitDbHostPlaceholder)).
				Value(&db.Host).
				Validate(dbHostValidator),
			huh.NewInput().
				Title(t(i18n.CliInitDbPort)).
				Placeholder(t(i18n.CliInitDbPortPlaceholder)).
				Value(&db.Port).
				Validate(requiredValidator),
			huh.NewInput().
				Title(t(i18n.CliInitDbName)).
				Placeholder(t(i18n.CliInitDbNamePlaceholder)).
				Value(&db.Name).
				Validate(requiredValidator),
			huh.NewInput().
				Title(t(i18n.CliInitDbUser)).
				Placeholder(t(i18n.CliInitDbUserPlaceholder)).
				Value(&db.User).
				Validate(requiredValidator),
			huh.NewInput().
				Title(t(i18n.CliInitDbPassword)).
				EchoMode(huh.EchoModePassword).
				Value(&db.Password).
				Validate(requiredValidator),
		),
	).Run()
	if err != nil {
		return err
	}

	// 打印用户输入的信息（隐藏密码）
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999")).
		Width(20)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000"))
	printRow := func(label, value string) {
		fmt.Printf("  %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	fmt.Println(headerStyle.Render(t(i18n.CliInitConnDbInfoTitle)))
	fmt.Println(strings.Repeat("─", 60))
	printRow(t(i18n.CliInitConnDbHost), db.Host)
	printRow(t(i18n.CliInitConnDbPort), db.Port)
	printRow(t(i18n.CliInitConnDbName), db.Name)
	printRow(t(i18n.CliInitConnDbUser), db.User)
	printRow(t(i18n.CliInitConnDbPass), "***")
	fmt.Println(strings.Repeat("─", 60))

	var verifyErr error
	_ = spinner.New().Title(t(i18n.CliInitTestingDbConn)).Action(func() {
		verifyErr = CheckDB(db.toDSN())
	}).Run()

	if verifyErr != nil {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
		fmt.Println(errStyle.Render("✗ " + fmt.Sprintf(t(i18n.CliInitErrDbConn), verifyErr.Error())))
		return errors.New(t(i18n.CliInitErrDbConnFailAbort))
	}

	// successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)
	// fmt.Println(successStyle.Render("✓ ") + t(i18n.CliInitSuccessDbConn))

	// 第二组表单：禅道 IP 和安装目录
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(t(i18n.CliInitZentaoIp)).
				Placeholder(t(i18n.CliInitZentaoIpPlaceholder)).
				Value(zentaoIP).
				Validate(requiredValidator),
			huh.NewInput().
				Title(t(i18n.CliInitWorkspace)).
				Placeholder(t(i18n.CliInitWorkspacePlaceholder)).
				Value(workspace).
				Validate(requiredAbsPathValidator),
		),
	).Run()

	return err
}

// requiredValidator 校验输入不能为空。
func requiredValidator(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New(t(i18n.CliInitErrRequired))
	}
	return nil
}

// requiredAbsPathValidator 校验输入非空且必须为绝对路径
func requiredAbsPathValidator(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New(t(i18n.CliInitErrRequired))
	}
	if !filepath.IsAbs(s) {
		return errors.New(t(i18n.CliInitErrAbsPath))
	}
	return nil
}

// requiredURLOrHostValidator 校验输入非空且必须为合法的 URL 或主机地址。
func requiredURLOrHostValidator(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New(t(i18n.CliInitErrRequired))
	}

	testStr := s
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		testStr = "http://" + s
	}

	u, err := url.ParseRequestURI(testStr)
	if err != nil || u.Host == "" {
		return errors.New(t(i18n.CliInitErrUrlFormat))
	}
	return nil
}

// normalizeZentaoURL 格式化用户输入的禅道地址为标准的 HTTP/HTTPS BaseURL。
func normalizeZentaoURL(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}
	return strings.TrimRight(s, "/")
}

// dbHostValidator 校验数据库主机输入非空即可，连通性由后续测试保证。
func dbHostValidator(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New(t(i18n.CliInitErrRequired))
	}
	return nil
}

// isIPAddress 检查字符串是否为有效的 IPv4/IPv6 地址
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

// (This line is intentionally left blank since old verify functions are removed)

// parsePHPConfig 解析禅道 my.php 配置文件，提取数据库连接信息。
//
// 禅道 my.php 的典型格式：
//
//	<?php
//	$config->installed       = true;
//	$config->db->driver      = 'mysql';
//	$config->db->host        = 'dev-mysql';
//	$config->db->port        = '3306';
//	$config->db->name        = 'zentaoipd';
//	$config->db->user        = 'root';
//	$config->db->encoding    = 'utf8mb4';
//	$config->db->password    = '123456';
//	$config->db->prefix      = 'zt_';
//	$config->devops->gitfoxURL  = 'http://10.0.4.2';
//	$config->devops->gitfoxPort = 3000;
//
// 使用正则逐行匹配提取 host、port、name、user、password 字段。
func parsePHPConfig(path string) (*dbConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	db := &dbConfig{
		Host: "127.0.0.1",
		Port: "3306",
	}

	// 定义正则模式：匹配 $config->db->字段名 = '值'; 格式的行
	patterns := map[string]*string{
		`\$config->db->host\s*=\s*['"]([^'"]+)['"]`:      &db.Host,
		`\$config->db->port\s*=\s*['"]?(\d+)['"]?`:       &db.Port,
		`\$config->db->name\s*=\s*['"]([^'"]+)['"]`:      &db.Name,
		`\$config->db->user\s*=\s*['"]([^'"]+)['"]`:      &db.User,
		`\$config->db->password\s*=\s*['"]([^'"]*?)['"]`: &db.Password,
	}

	// 预编译所有正则表达式
	compiled := make(map[*regexp.Regexp]*string)
	for pattern, target := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf(t(i18n.CliInitErrMyphpRegex), err)
		}
		compiled[re] = target
	}

	// 逐行扫描文件，匹配并提取数据库配置值
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for re, target := range compiled {
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				*target = matches[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 校验必要字段：数据库名和用户名不能为空
	if db.Name == "" || db.User == "" {
		return nil, errors.New(t(i18n.CliInitErrMyphpMissingFields))
	}

	return db, nil
}

// getLocalIP 获取本机非回环 IPv4 地址。
// 遍历所有网络接口，返回第一个非 loopback 的 IPv4 地址。
// 如果获取失败，返回 127.0.0.1 作为兜底。
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}

// loadExistingEnv 从当前二进制所在目录加载已有的 local.env 文件。
// 返回环境变量键值对 map，用于回显默认值给用户。
// 如果文件不存在或读取失败，返回空 map（不影响后续流程）。
func loadExistingEnv() map[string]string {
	exePath, err := os.Executable()
	if err != nil {
		return make(map[string]string)
	}

	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, "prod.env")

	envMap, err := godotenv.Read(envPath)
	if err != nil {
		return make(map[string]string)
	}

	return envMap
}

// loadTemplateEnv 从内嵌的 prod.env 模板中读取所有的默认配置键值对
func loadTemplateEnv() map[string]string {
	envMap := make(map[string]string)
	templateBytes, err := configs.FS.ReadFile("prod.env")
	if err == nil {
		if parsed, err := godotenv.Unmarshal(string(templateBytes)); err == nil {
			return parsed
		}
	}
	return envMap
}

// saveEnvFile 将配置保存到用户指定安装目录下的 prod.env 文件。
//
// 替换策略：
//  1. 读取 configs/prod.env 的内嵌模板
//  2. 针对用户交互产生的关键值，利用正则保留注释并替换配置项
//  3. 若模板中没有对应的 key 但流程中有值，则追加到文件尾部
//  4. 如果 existing 中有独立修改，也会被保留。
func saveEnvFile(existing map[string]string, db *dbConfig, workspace, serverHost, devopsURL, urlBase string) (string, error) {
	// 确保安装目录存在
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrMkdir), err)
	}

	envPath := filepath.Join(workspace, "prod.env")

	// 读取内嵌的配置模板
	templateBytes, err := configs.FS.ReadFile("prod.env")
	if err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrReadTemplate), err)
	}
	content := string(templateBytes)

	// 需要覆盖的关键键值对
	overrides := map[string]string{
		"GITFOX_DB_DATASOURCE":      db.toDSN(),
		"GITFOX_DB_DRIVER":          "mysql",
		"GITFOX_WORKSPACE_ROOT_DIR": workspace,
		"GITFOX_SERVER_HOST":        serverHost,
		"GITFOX_DEVOPS_URL":         devopsURL,
	}

	// 合并用户已有其他配置项
	for k, v := range existing {
		if _, ok := overrides[k]; !ok {
			overrides[k] = v
		}
	}

	// 执行正则文本替换，保留注释结构
	for key, val := range overrides {
		// 匹配该键所在的一行
		re := regexp.MustCompile(fmt.Sprintf(`(?m)^%s=.*$`, regexp.QuoteMeta(key)))
		if re.MatchString(content) {
			content = re.ReplaceAllString(content, fmt.Sprintf("%s=%s", key, val))
		} else {
			content += fmt.Sprintf("\n%s=%s", key, val)
		}
	}

	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		return "", err
	}

	return envPath, nil
}

// copyBinary 将当前运行的可执行文件复制到安装目录。
//
// 注意事项：
//   - 会解析符号链接以获取真实的二进制路径
//   - 如果源路径和目标路径相同，自动跳过（避免自我覆盖）
//   - 目标文件名固定为 "gitfox"，权限 0755
func copyBinary(workspace string) (string, error) {
	// 获取当前可执行文件路径
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrGetExec), err)
	}

	// 解析符号链接，获取真实路径
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrEvalSymlink), err)
	}

	dstPath := filepath.Join(workspace, "gitfox")
	if runtime.GOOS == "windows" {
		dstPath += ".exe"
	}

	// 如果源和目标是同一个文件，跳过复制
	if absSrc, err1 := filepath.Abs(exePath); err1 == nil {
		if absDst, err2 := filepath.Abs(dstPath); err2 == nil {
			if absSrc == absDst {
				return dstPath, nil
			}
		}
	}

	// 确保安装目录存在
	if err := os.MkdirAll(workspace, 0755); err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrMkdir), err)
	}

	// 打开源文件
	src, err := os.Open(exePath)
	if err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrOpenSrcBin), err)
	}
	defer src.Close()

	// 创建目标文件（可执行权限 0755）
	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrCreateDestBin), err)
	}
	defer dst.Close()

	// 流式复制文件内容
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf(t(i18n.CliInitErrCopyBin), err)
	}

	return dstPath, nil
}

// backupFile 将指定路径的文件重命名为 xxx_bak，用于覆盖前的安全备份。
// 返回备份后的文件路径。如果源文件不存在则跳过。
func backupFile(filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", nil
	}
	bakPath := filePath + "_bak"
	if err := os.Rename(filePath, bakPath); err != nil {
		return "", fmt.Errorf("backup %s: %w", filePath, err)
	}
	return bakPath, nil
}

// copyEmbeddedYaml 将内嵌的 prod.yaml 配置文件释放到目标工作目录。
func copyEmbeddedYaml(workspace string) error {
	dstPath := filepath.Join(workspace, "prod.yaml")

	yamlBytes, err := configs.FS.ReadFile("prod.yaml")
	if err != nil {
		return fmt.Errorf("read embedded prod.yaml: %w", err)
	}

	if err := os.MkdirAll(workspace, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", workspace, err)
	}

	if err := os.WriteFile(dstPath, yamlBytes, 0644); err != nil {
		return fmt.Errorf("write prod.yaml: %w", err)
	}

	return nil
}

// printSummary 以美化的表格形式打印最终配置摘要。
// 密码在 DSN 中以 *** 替代，避免泄露。
func printSummary(envPath string, db *dbConfig, workspace, serverHost, devopsURL string) {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)

	fmt.Println()
	fmt.Println(successStyle.Render("✓ ") + t(i18n.CliInitSuccess))
	fmt.Println()

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999999")).
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000"))

	fmt.Println(headerStyle.Render(t(i18n.CliInitConfigSummary)))
	fmt.Println(strings.Repeat("─", 60))

	printRow := func(label, value string) {
		fmt.Printf("  %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	// 密码脱敏显示
	dsnDisplay := fmt.Sprintf("%s:***@tcp(%s:%s)/%s", db.User, db.Host, db.Port, db.Name)

	printRow(t(i18n.CliInitConfigFilePath), envPath)
	printRow(t(i18n.CliInitConfigDb), dsnDisplay)
	printRow(t(i18n.CliInitConfigWorkspace), workspace)
	printRow(t(i18n.CliInitConfigServerHost), serverHost)
	printRow(t(i18n.CliInitConfigDevopsUrl), devopsURL)
	fmt.Println(strings.Repeat("─", 60))

	// install 模式下跳过启动提示（install 会自动执行 start）
	if !isInstallingMode {
		fmt.Println(t(i18n.CliInitNextStepHint))
		fmt.Printf("   cd %s\n", workspace)

		cmdPrefix := "./gitfox"
		if runtime.GOOS == "windows" {
			cmdPrefix = ".\\gitfox.exe"
		}
		fmt.Printf("   %s start\n\n", cmdPrefix)
	}
}

// updateZentaoPHPConfig 将 GitFox 的服务地址写回禅道的 my.php 配置文件。
//
// 写入/更新以下两行：
//
//	$config->devops->gitfoxURL = 'http://{ip}';
//	$config->devops->gitfoxPort = {port};
//
// 如果 my.php 中已存在这两行，则原地更新；否则在文件末尾（?>之前或文件尾部）追加。
func updateZentaoPHPConfig(myPHPPath string, gitfoxIP string, gitfoxPort int) error {
	data, err := os.ReadFile(myPHPPath)
	if err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrMyphpRead), err)
	}

	content := string(data)
	gitfoxURLLine := fmt.Sprintf("$config->devops->gitfoxURL = 'http://%s';", gitfoxIP)
	gitfoxPortLine := fmt.Sprintf("$config->devops->gitfoxPort = %d;", gitfoxPort)

	// 使用正则匹配已有的 gitfoxURL 和 gitfoxPort 行
	urlRegex := regexp.MustCompile(`(?m)^\s*\$config->devops->gitfoxURL\s*=.*$`)
	portRegex := regexp.MustCompile(`(?m)^\s*\$config->devops->gitfoxPort\s*=.*$`)

	urlUpdated := false
	portUpdated := false

	// 如果已有 gitfoxURL 行，原地替换
	if urlRegex.MatchString(content) {
		content = urlRegex.ReplaceAllLiteralString(content, gitfoxURLLine)
		urlUpdated = true
	}

	// 如果已有 gitfoxPort 行，原地替换
	if portRegex.MatchString(content) {
		content = portRegex.ReplaceAllLiteralString(content, gitfoxPortLine)
		portUpdated = true
	}

	// 如果有未更新的行，追加到文件末尾
	if !urlUpdated || !portUpdated {
		var appendLines []string
		if !urlUpdated {
			appendLines = append(appendLines, gitfoxURLLine)
		}
		if !portUpdated {
			appendLines = append(appendLines, gitfoxPortLine)
		}
		appendContent := "\n" + strings.Join(appendLines, "\n") + "\n"

		// 如果文件包含 ?>，在其前面插入；否则直接追加到末尾
		if strings.Contains(content, "?>") {
			content = strings.Replace(content, "?>", appendContent+"?>", 1)
		} else {
			content += appendContent
		}
	}

	// 写回文件
	if err := os.WriteFile(myPHPPath, []byte(content), 0644); err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrMyphpWrite), err)
	}

	return nil
}
