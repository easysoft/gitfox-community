package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"
	"github.com/joho/godotenv"
)

// 依赖最低版本约束。
const (
	// MinGitVersion 是 GitFox 支持的最低 git 客户端版本。
	MinGitVersion = "2.30.0"
	// MinSvnVersion 是 GitFox 关联 Subversion 代码库功能要求的最低 svn 客户端版本。
	MinSvnVersion = "1.14.3"
)

// threePartVersionRe 匹配输出中的三段式版本号 (X.Y.Z)，用于从 `xxx --version` 文本中抽取版本。
var threePartVersionRe = regexp.MustCompile(`(\d+\.\d+\.\d+)`)

// extractVersion 从命令原始输出中抽取三段式版本号，并用 go-version 解析为可比较的 *version.Version。
// 返回原始捕获字符串与解析对象，便于上层既能展示又能比较。
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

// CheckSudo 检查当前上下文是否具有超级管理员执行权限 (Root/Sudo/Administrator)。
// 主要由于许多 CLI 的核心命令 (如挂载、停止后台守护进程) 需要穿透系统底层安全限制。
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
// 返回: (是否安装, 解析后的版本号, 错误信息)
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

// CheckDocker 检查 docker 是否安装、daemon 是否正在运行，并获取版本信息。
func CheckDocker() (bool, string, error) {
	t := clii18n.T
	if _, err := exec.LookPath("docker"); err != nil {
		return false, "", fmt.Errorf(t(i18n.CliCheckErrNotFound), "docker")
	}

	cmd := exec.Command("docker", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return true, "", fmt.Errorf(t(i18n.CliCheckErrExecFailed), "docker", err)
	}

	outStr := strings.TrimSpace(string(output))
	verStr, _, err := extractVersion(outStr)
	if err != nil {
		// 解析失败但二进制存在，返回原始输出便于排查
		verStr = outStr
	}

	// 验证 Docker daemon 是否正在运行
	// 注意：不能用 docker info，它会连接 docker.socket 触发 systemd socket activation 把 Docker 拉起来
	if !isDockerDaemonRunning() {
		return true, verStr, fmt.Errorf(t(i18n.CliCheckErrDaemonNotRunning), "docker")
	}

	return true, verStr, nil
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

// isDockerDaemonRunning 检测 Docker daemon 是否正在运行，且不会触发 systemd socket activation。
// Linux 上优先使用 systemctl is-active，若 systemctl 不存在则回退到 pgrep 检测 dockerd 进程。
// Windows 上通过 sc query 检测服务状态。
func isDockerDaemonRunning() bool {
	switch runtime.GOOS {
	case "linux":
		// systemctl is-active 只读查询，不触发 socket activation
		if err := exec.Command("systemctl", "is-active", "--quiet", "docker").Run(); err == nil {
			return true
		}
		// 非 systemd 环境回退：检测 dockerd 进程是否存在
		if err := exec.Command("pgrep", "-x", "dockerd").Run(); err == nil {
			return true
		}
		// 最后手段：尝试 docker info，如果能成功执行说明 daemon 也在运行（但可能会触发 socket activation）
		if err := exec.Command("docker", "info").Run(); err == nil {
			return true
		}
		return false
	case "windows":
		out, err := exec.Command("sc", "query", "docker").CombinedOutput()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "RUNNING")
	default:
		return false
	}
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

// CheckPortDown 反向健康检查：在服务停止后确认端口不再响应。
//
// 在 timeout 时间窗口内每秒探测一次健康端点，端点连接被拒绝即视为已下线返回 false；
// 整个窗口都仍能 200 则返回 true（仍在线）。
//
// healthURL 为空时无法探测，sleep 一个轮询周期后视为已下线（避免误报"仍在线"）。
func CheckPortDown(healthURL string, timeout time.Duration) bool {
	if healthURL == "" {
		time.Sleep(timeout)
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(1 * time.Second)
		// 健康端点没有响应 = 服务确实已停止
		if !CheckHealthOnce(healthURL, 5*time.Second) {
			return false
		}
	}
	return true
}

// =====================================================================
// 健康探针（公共工具，被 install / start / stop / upgrade 等命令复用）
// =====================================================================

// healthPath 是 GitFox 服务暴露的健康检查端点路径。
const healthPath = "/public/health"

// BuildHealthURL 拼装健康检查 URL。host 或 port 任一为空时返回空字符串，
// 调用方据此判断是否具备探测条件。
func BuildHealthURL(host, port string) string {
	if host == "" || port == "" {
		return ""
	}
	return fmt.Sprintf("http://%s:%s%s", host, port, healthPath)
}

// CheckHealthOnce 单次同步健康探测：URL 非空且 GET 返回 200 视为存活。
// timeout 为空时使用 5s 默认值。
func CheckHealthOnce(healthURL string, timeout time.Duration) bool {
	if healthURL == "" {
		return false
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// WaitForServiceUp 轮询等待服务存活。最多 maxRetries 次，每次间隔 1s。
// 返回 true=已存活，false=超过重试上限仍未存活。
func WaitForServiceUp(healthURL string, maxRetries int) bool {
	if healthURL == "" {
		return false
	}
	for range maxRetries {
		if CheckHealthOnce(healthURL, 5*time.Second) {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// ReadHealthURLResult 是健康检查 URL 解析后的结构化结果。
type ReadHealthURLResult struct {
	Host, Port, URL string
}

// readHealthURL 从当前工作目录的 prod.env 读取 host/port，拼装健康检查 URL。
// 这是 stop / uninstall / status 等"在工作目录里运行"的命令的便捷封装。
//
// 失败时 result.Host="127.0.0.1"、其他字段为空，error 非 nil；
// 调用方自行决定是否中止主流程或仅以警告形式提示。
func readHealthURL() (ReadHealthURLResult, error) {
	return readHealthURLFrom(".")
}

// readHealthURLFrom 从指定工作目录的 prod.env 读取 host/port，拼装健康检查 URL。
// upgrade 等命令在任意 cwd 运行，需显式传入 systemd 注册的工作目录。
//
// 优先使用 GITFOX_SERVER_HOST，兜底 127.0.0.1。port 缺失时 url 为空字符串。
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
		URL:  BuildHealthURL(host, port),
	}, nil
}

// IsSystemServiceActive 通过系统服务管理器判断 gitfox 服务是否处于"运行中"。
//
// 比 health 探针更准：直接读 systemd / Windows SCM 状态，不受网络/端口监听延迟干扰。
// 不支持的平台返回 false。
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
	// GitfoxHealthUnreachable 进程在跑但 health 端点未响应或非 200（启动中 / panic / DB 断连）。
	GitfoxHealthUnreachable
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

	// 5) 健康端点
	if !CheckHealthOnce(hc.URL, 3*time.Second) {
		result.Status = GitfoxHealthUnreachable
		return result
	}

	// 6) 全过
	result.Status = GitfoxHealthOK
	return result
}
