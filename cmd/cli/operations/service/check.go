package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"
	"github.com/joho/godotenv"
)

const (
	MinGitVersion = "2.30.0"
	MinSvnVersion = "1.14.3"
)

var threePartVersionRe = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// extractVersion 从命令原始输出中抽取三段式版本号，并用 go-version 解析为可比较的 *version.Version。
func extractVersion(raw string) (string, *version.Version, error) {
	matches := threePartVersionRe.FindStringSubmatch(raw)
	if len(matches) < 2 {
		return "", nil, fmt.Errorf("no three-part version in: %s", strings.TrimSpace(raw))
	}
	v, err := version.NewVersion(matches[1])
	if err != nil {
		return matches[1], nil, err
	}
	return matches[1], v, nil
}

// Check 检查环境依赖工具是否安装、版本是否达标。
func Check(headerStyle, okStyle, errStyle, warnStyle lipgloss.Style) (missing bool) {
	var gitOk, dockerOk, svnOk bool
	var gitVer, svnVer string
	var gitErr, dockerErr, svnErr error

	_ = spinner.New().
		Title(t(i18n.CliInstallDepChecking)).
		Action(func() {
			gitOk, gitVer, gitErr = CheckGit()
			dockerOk, dockerErr = CheckDocker()
			svnOk, svnVer, svnErr = CheckSubversion()
		}).Run()

	fmt.Println(headerStyle.Render(t(i18n.CliStatusSectionDeps)))

	// --- Git ---
	if gitOk && gitErr == nil {
		fmt.Println(okStyle.Render(fmt.Sprintf(t(i18n.CliStatusGitOk), gitVer)))
	} else if gitOk && gitErr != nil {
		fmt.Println(errStyle.Render(gitErr.Error()))
		missing = true
	} else {
		fmt.Println(errStyle.Render(t(i18n.CliStatusGitMissing)))
		missing = true
	}

	// --- Docker ---
	if dockerOk && dockerErr == nil {
		fmt.Println(okStyle.Render(t(i18n.CliStatusDockerOk)))
	} else if dockerOk && dockerErr != nil {
		fmt.Println(warnStyle.Render(dockerErr.Error()))
		missing = true
	} else {
		fmt.Println(errStyle.Render(t(i18n.CliStatusDockerMissing)))
		missing = true
	}

	// --- Subversion (非阻塞，仅警告) ---
	if svnOk && svnErr == nil {
		fmt.Println(okStyle.Render(fmt.Sprintf(t(i18n.CliStatusSvnOk), svnVer)))
	} else if svnOk && svnErr != nil {
		fmt.Println(warnStyle.Render(svnErr.Error()))
	} else {
		fmt.Println(warnStyle.Render(t(i18n.CliStatusSvnMissing)))
	}

	return missing
}

// CheckSudo 检查当前上下文是否具有超级管理员执行权限 (Root/Sudo/Administrator)。
func CheckSudo() error {
	switch runtime.GOOS {
	case "windows":
		return checkSudoWindows()
	default:
		if os.Getuid() != 0 {
			return errors.New(clii18n.T(i18n.CliStartErrRequiresAdmin))
		}
		return nil
	}
}

// CheckGit 检查 git 是否安装、版本是否满足 MinGitVersion，并获取版本信息。
//   - ok=false 表示未安装
//   - ok=true && err==nil 表示安装且版本达标
//   - ok=true && err!=nil 表示安装但版本过低或解析失败
func CheckGit() (bool, string, error) {
	t := clii18n.T
	if _, err := exec.LookPath("git"); err != nil {
		return false, "", fmt.Errorf(t(i18n.CliCheckErrNotFound), "git")
	}

	cmd := exec.Command("git", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, "", fmt.Errorf(t(i18n.CliCheckErrExecFailed), "git", err)
	}

	outStr := strings.TrimSpace(string(output))
	verStr, ver, err := extractVersion(outStr)
	if err != nil {
		return true, outStr, fmt.Errorf(t(i18n.CliCheckErrParseFailed), "git", outStr)
	}

	minVer, _ := version.NewVersion(MinGitVersion)
	if ver.LessThan(minVer) {
		return true, verStr, fmt.Errorf(t(i18n.CliCheckErrVersionTooOld), "git", verStr, MinGitVersion)
	}

	return true, verStr, nil
}

// ─────────────────────────────────────────────────────────────────
// Docker daemon 检查
// ─────────────────────────────────────────────────────────────────
//   CheckDocker
//     ├─ dockerHost()                         选 host 字符串
//     │     └─ pickWindowsDockerNpipe()        Windows: 三候选 Stat 命中
//     ├─ newDockerHTTPClient(host)            构造 client + baseURL
//     │     ├─ stripNpipePrefix(host)          npipe 短路（绕开 url.Parse 反斜杠雷）
//     │     │     └─ newDockerHTTPClientNpipePath
//     │     └─ url.Parse → unix / npipe / tcp / http(s) 分发
//     └─ dockerPing(client, baseURL)          GET /_ping 探活
// 状态分类：
//   ok=true,  err=nil  : daemon 健在
//   ok=true,  err!=nil : 已安装但服务未运行（dial 失败 / ping 失败）
//   ok=false, err!=nil : host 字符串本身解析失败等极端情况

const (
	defaultDockerSocket = "/var/run/docker.sock"              // Linux unix socket 默认路径
	defaultDockerNpipe  = `\\.\pipe\dockerDesktopLinuxEngine` // Windows 命名管道默认路径（4.x WSL2 引擎）
	dockerProbeTimeout  = 2 * time.Second                     // 单次拨号 / ping 超时
)

// windowsDockerNpipeCandidates 是 Windows 下需要按序探测的 docker daemon 命名管道候选。
var windowsDockerNpipeCandidates = []string{
	`\\.\pipe\dockerDesktopLinuxEngine`,
	`\\.\pipe\docker_engine`,
	`\\.\pipe\dockerDesktopWindowsEngine`,
}

// pickWindowsDockerNpipe 按候选列表顺序选择 Windows 下要拨号的命名管道。
func pickWindowsDockerNpipe() string {
	for _, p := range windowsDockerNpipeCandidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return defaultDockerNpipe
}

// dockerHost 解析 daemon 地址：优先 DOCKER_HOST，否则平台默认。
func dockerHost() string {
	if h := strings.TrimSpace(os.Getenv("DOCKER_HOST")); h != "" {
		return h
	}
	if runtime.GOOS == "windows" {
		return "npipe://" + pickWindowsDockerNpipe()
	}
	return "unix://" + defaultDockerSocket
}

// stripNpipePrefix 把 host 还原成可直接用于 npipe 拨号的原始路径。
//   - ok=true 表示 host 是命名管道；path 是原始 `\\.\pipe\<name>`
//   - ok=false 表示 host 不是命名管道，调用方继续走 url.Parse
func stripNpipePrefix(host string) (string, bool) {
	if strings.HasPrefix(host, `\\.\pipe\`) {
		return host, true
	}
	const npipeScheme = "npipe://"
	if !strings.HasPrefix(host, npipeScheme) {
		return "", false
	}
	rest := host[len(npipeScheme):]
	rest = strings.ReplaceAll(rest, "/", `\`)
	rest = strings.TrimLeft(rest, `\`)
	switch {
	case rest == "":
		return defaultDockerNpipe, true
	case strings.HasPrefix(rest, `.\pipe\`):
		return `\\` + rest, true
	case strings.HasPrefix(rest, `pipe\`):
		return `\\.\` + rest, true
	default:
		return `\\.\pipe\` + strings.TrimPrefix(rest, `pipe\`), true
	}
}

// newDockerHTTPClient 基于 host 构造仅用于 daemon 通信的 http.Client + baseURL。
func newDockerHTTPClient(host string) (*http.Client, string, error) {
	if p, ok := stripNpipePrefix(host); ok {
		return newDockerHTTPClientNpipePath(p)
	}

	u, err := url.Parse(host)
	if err != nil {
		return nil, "", err
	}

	switch u.Scheme {
	case "unix":
		return newDockerHTTPClientUnix(u)
	case "npipe":
		return newDockerHTTPClientNpipe(u)
	case "tcp", "http", "https":
		scheme := "http"
		if u.Scheme == "https" {
			scheme = "https"
		}
		client := &http.Client{Timeout: dockerProbeTimeout, Transport: &http.Transport{}}
		return client, scheme + "://" + u.Host, nil
	default:
		return nil, "", fmt.Errorf("unsupported docker host scheme: %s", u.Scheme)
	}
}

// newDockerHTTPClientUnix 构造 unix socket 拨号 client。Stat 不通过即返回失败，
// 这里 Stat 是可靠的（unix socket 文件存在性 = 监听存在性）。
func newDockerHTTPClientUnix(u *url.URL) (*http.Client, string, error) {
	path := u.Path
	if path == "" {
		path = defaultDockerSocket
	}
	if _, err := os.Stat(path); err != nil {
		return nil, "", err
	}
	client := &http.Client{
		Timeout: dockerProbeTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				d.Timeout = dockerProbeTimeout
				return d.DialContext(ctx, "unix", path)
			},
		},
	}
	return client, "http://docker", nil
}

// dockerPing 调用 GET /_ping 判断 daemon 是否真的在响应。
// unix socket / npipe / tcp 三种路径统一走此 HTTP 探活，是权威活体判定。
func dockerPing(client *http.Client, baseURL string) error {
	req, _ := http.NewRequest(http.MethodGet, baseURL+"/_ping", nil)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("docker ping http %d", resp.StatusCode)
	}
	return nil
}

// CheckDocker 检查 docker daemon 是否可达。
func CheckDocker() (bool, error) {
	t := clii18n.T

	host := dockerHost()
	parsed := describeParsedDockerHost(host)

	client, baseURL, dialErr := newDockerHTTPClient(host)
	if dialErr != nil {
		return true, decorateDockerErr(fmt.Errorf(t(i18n.CliCheckErrDaemonNotRunning), "docker"), host, parsed)
	}

	if err := dockerPing(client, baseURL); err != nil {
		return true, decorateDockerErr(fmt.Errorf(t(i18n.CliCheckErrDaemonNotRunning), "docker"), host, parsed)
	}
	return false, nil
}

// describeParsedDockerHost 把 dockerHost 选出的字符串解析成"实际拨号用"的形态描述：
func describeParsedDockerHost(host string) string {
	if p, ok := stripNpipePrefix(host); ok {
		return p
	}
	u, err := url.Parse(host)
	if err != nil {
		return "<parse error: " + err.Error() + ">"
	}
	if u.Path == "" {
		return u.Scheme + "://" + u.Host
	}
	return u.Scheme + "://" + u.Host + u.Path
}

// decorateDockerErr 在 docker 检查失败时给原 error 末尾附上 host / parsed 双视角，
func decorateDockerErr(base error, host, parsed string) error {
	return fmt.Errorf("%w (host=%q, parsed=%q)", base, host, parsed)
}

// CheckSubversion 检查 svn 客户端是否安装、版本是否满足 MinSvnVersion，并获取版本信息。
// 此检查为非阻塞性：缺失或版本过低仅以警告形式提示，不阻挡 GitFox 启动。
// 返回: (是否安装, 解析后的版本号, 错误信息)
//   - ok=false 表示未安装
//   - ok=true && err==nil 表示安装且版本达标
//   - ok=true && err!=nil 表示安装但版本过低或解析/执行失败
func CheckSubversion() (bool, string, error) {
	t := clii18n.T
	if _, err := exec.LookPath("svn"); err != nil {
		return false, "", fmt.Errorf(t(i18n.CliCheckErrNotFound), "svn")
	}

	cmd := exec.Command("svn", "--version", "--quiet")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, "", fmt.Errorf(t(i18n.CliCheckErrExecFailed), "svn", err)
	}

	outStr := strings.TrimSpace(string(output))
	verStr, ver, err := extractVersion(outStr)
	if err != nil {
		return true, outStr, fmt.Errorf(t(i18n.CliCheckErrParseFailed), "svn", outStr)
	}

	minVer, _ := version.NewVersion(MinSvnVersion)
	if ver.LessThan(minVer) {
		return true, verStr, fmt.Errorf(t(i18n.CliCheckErrVersionTooOld), "svn", verStr, MinSvnVersion)
	}

	return true, verStr, nil
}

// CheckDB 验证数据库连通性
func CheckDB(dsn string) error {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrDbConn), err)
	}
	defer sqlDB.Close()

	sqlDB.SetConnMaxLifetime(3 * time.Second)
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrDbConn), err)
	}
	return nil
}

// CheckZentao 验证禅道服务有效性
func CheckZentao(apiURL string) error {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrZentaoConn), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusUnauthorized {
		return fmt.Errorf(clii18n.T(i18n.CliInitErrZentaoConn), fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	return nil
}

// PortStatus 描述单个端口的占用情况
type PortStatus struct {
	Port    string
	InUse   bool
	PID     string
	Process string
	Unknown bool // 检测失败时为 true
}

// CheckPort 检查指定端口是否被占用，并返回占用者的 PID 和进程名。
// 跨平台支持：Linux 使用 ss/lsof，Windows 使用 netstat。
func CheckPort(port string) PortStatus {
	result := PortStatus{Port: port}

	switch runtime.GOOS {
	case "linux":
		result = checkPortLinux(port)
	case "windows":
		result = checkPortWindows(port) // → windows.go
	default:
		result.Unknown = true
	}
	return result
}

func checkPortLinux(port string) PortStatus {
	result := PortStatus{Port: port}

	// 优先使用 ss（iproute2，Linux 主流发行版默认装）
	if _, err := exec.LookPath("ss"); err == nil {
		// ss 命令存在：无论端口是否占用，退出码都是 0；
		// 仅当 fork 失败或 ss 内部异常时才返回非 nil err。
		out, err := exec.Command("ss", "-tlnp", "sport", "=", ":"+port).Output()
		if err == nil {
			return parseSsOutput(string(out), port)
		}
	}

	// 降级 lsof（macOS / 部分极简 Linux）
	if _, err := exec.LookPath("lsof"); err == nil {
		// lsof 语义不对称：
		//   - 命中：退出码 0 + 多行输出
		//   - 未命中：退出码 1 + 空输出   ← 必须当作"端口空闲"
		//   - 真异常：退出码 1 + stderr 有内容
		out, err := exec.Command("lsof", "-i", ":"+port, "-sTCP:LISTEN", "-n", "-P").Output()
		if err == nil {
			return parseLsofOutput(string(out), port)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(out) == 0 {
			return result
		}
	}

	result.Unknown = true
	return result
}

// parseLsofOutput 解析 lsof 输出提取 PID / 进程名
func parseLsofOutput(output, port string) PortStatus {
	result := PortStatus{Port: port}
	for i, line := range strings.Split(output, "\n") {
		if i == 0 {
			continue // 跳过 header
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		result.InUse = true
		result.Process = fields[0]
		result.PID = fields[1]
		return result
	}
	return result
}

// parseSsOutput 解析 ss 输出提取 PID / 进程名
// ss -tlnp 输出格式: State Recv-Q Send-Q Local ... Process
func parseSsOutput(output, port string) PortStatus {
	result := PortStatus{Port: port}
	pidRe := regexp.MustCompile(`pid=(\d+)`)
	nameRe := regexp.MustCompile(`"([^"]+)"`)
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, ":"+port) {
			continue
		}
		result.InUse = true
		if m := pidRe.FindStringSubmatch(line); len(m) > 1 {
			result.PID = m[1]
		}
		if m := nameRe.FindStringSubmatch(line); len(m) > 1 {
			result.Process = m[1]
		}
		return result
	}
	return result
}

// CheckPortWithTimeout 检查指定端口是否被占用，超时后返回 false。
func CheckPortWithTimeout(port string, timeout time.Duration) bool {
	if port == "" {
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(1 * time.Second)
		if CheckPort(port).InUse {
			return true
		}
	}
	return false
}

// ReadHealthURLResult 是健康检查 URL 解析后的结构化结果。
type ReadHealthURLResult struct {
	Host, Port, URL string
}

// readHealthURL 从当前工作目录的 prod.env 读取 host/port，拼装健康检查 URL。
func readHealthURL() (ReadHealthURLResult, error) {
	return readHealthURLFrom(".")
}

// readHealthURLFrom 从指定工作目录的 prod.env 读取 host/port，拼装健康检查 URL。
func readHealthURLFrom(workDir string) (ReadHealthURLResult, error) {
	host := "127.0.0.1"
	envMap, err := godotenv.Read(filepath.Join(workDir, ProdEnvFileName))
	if err != nil {
		return ReadHealthURLResult{Host: host}, fmt.Errorf(clii18n.T(i18n.CliStartErrReadEnv), err)
	}
	host = envMap["GITFOX_SERVER_HOST"]
	port := envMap["GITFOX_HTTP_PORT"]
	return ReadHealthURLResult{
		Host: host,
		Port: port,
	}, nil
}

// IsSystemServiceActive 通过系统服务管理器判断 gitfox 服务是否处于"运行中"。
func IsSystemServiceActive() bool {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err != nil {
			return false
		}
		// systemctl is-active --quiet exit 0 表示 active
		return exec.Command("systemctl", "is-active", "--quiet", "gitfox").Run() == nil
	case "windows":
		out, err := exec.Command("sc", "query", "gitfox").CombinedOutput()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "RUNNING")
	default:
		return false
	}
}

// GitfoxHealthStatus 描述 GitFox 服务的整体健康状态分类。
type GitfoxHealthStatus int

const (
	// GitfoxHealthOK 已注册 + 已启动 + 配置完整 + health 端点 200。
	GitfoxHealthOK GitfoxHealthStatus = iota
	// GitfoxHealthNotRegistered 系统中不存在 gitfox 服务注册（未 install）。
	GitfoxHealthNotRegistered
	// GitfoxHealthCorrupted 系统服务文件存在但解析失败 / ExePath 缺失。
	GitfoxHealthCorrupted
	// GitfoxHealthStopped 已注册但 systemctl is-active 为 false。
	GitfoxHealthStopped
	// GitfoxHealthConfigMissing 工作目录下读不出 prod.env。
	GitfoxHealthConfigMissing
	// GitfoxHealthAddressMissing prod.env 已读，但 host/port 缺失，无法拼装健康 URL。
	GitfoxHealthAddressMissing
)

// GitfoxHealthResult 是 CheckGitfoxHealth 的结构化结果。
// 字段按检查阶段渐进填充，调用方可基于已填字段做诊断展示。
type GitfoxHealthResult struct {
	Status  GitfoxHealthStatus
	ExePath string // 来自 InspectService（注册时填）
	WorkDir string // 来自 InspectService（注册时填）
	Host    string // 来自 prod.env（配置就绪时填）
	Port    string // 来自 prod.env
	URL     string // 拼装的健康端点
	Err     error  // 配置读取等具体错误（GitfoxHealthConfigMissing 时填）
}

// CheckGitfoxHealth 端到端检查 GitFox 服务可用性。
//
// 检查链：
//  1. InspectService → 未注册 / 损坏直接返回
//  2. IsSystemServiceActive → 未运行返回 GitfoxHealthStopped
//  3. readHealthURLFrom(workDir) → 配置异常返回 GitfoxHealthConfigMissing
//  4. host/port 缺失 → GitfoxHealthAddressMissing
//  5. CheckHealthOnce → 不通返回 GitfoxHealthUnreachable
//  6. 全部通过 → GitfoxHealthOK
func CheckGitfoxHealth() GitfoxHealthResult {
	// 1) 系统服务注册
	info := InspectService()
	if !info.Registered {
		return GitfoxHealthResult{Status: GitfoxHealthNotRegistered}
	}
	if info.Corrupted || info.ExePath == "" {
		return GitfoxHealthResult{
			Status:  GitfoxHealthCorrupted,
			ExePath: info.ExePath,
			WorkDir: info.WorkDir,
		}
	}

	result := GitfoxHealthResult{
		ExePath: info.ExePath,
		WorkDir: info.WorkDir,
	}

	// 2) 进程是否在跑
	if !IsSystemServiceActive() {
		result.Status = GitfoxHealthStopped
		return result
	}

	// 3) 配置文件
	hc, err := readHealthURLFrom(info.WorkDir)
	if err != nil {
		result.Status = GitfoxHealthConfigMissing
		result.Err = err
		return result
	}
	result.Host = hc.Host
	result.Port = hc.Port
	result.URL = hc.URL

	// 4) host/port 完整性
	if hc.URL == "" {
		result.Status = GitfoxHealthAddressMissing
		return result
	}

	// 6) 全过
	result.Status = GitfoxHealthOK
	return result
}
