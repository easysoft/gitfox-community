package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	_ "github.com/go-sql-driver/mysql"
)

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

// CheckGit 检查 git 是否安装并获取版本信息。
// 返回: (是否安装, 解析道德版本号, 错误信息)
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

	// 标准输出形如: "git version 2.39.2.windows.1"
	outStr := strings.TrimSpace(string(output))
	version := strings.TrimPrefix(outStr, "git version ")
	if version == outStr {
		// 没有匹配到 "git version " 前缀，尝试直接用正则提取数字版本
		re := regexp.MustCompile(`(\d+\.\d+\.\d+(?:\.[-a-zA-Z0-9]+)*)`)
		if matches := re.FindStringSubmatch(outStr); len(matches) > 1 {
			return true, matches[1], nil
		}
		// 如果正则都没提取出来，退化到返回整段字符串
		return true, outStr, nil
	}

	return true, version, nil
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

	// 标准输出形如: "Docker version 24.0.5, build ced0996"
	outStr := strings.TrimSpace(string(output))
	re := regexp.MustCompile(`[Vv]ersion\s+(\d+\.\d+\.\d+(?:[-a-zA-Z0-9.]+)*)`)
	matches := re.FindStringSubmatch(outStr)
	version := outStr
	if len(matches) > 1 {
		version = matches[1]
	}

	// 验证 Docker daemon 是否正在运行
	// 注意：不能用 docker info，它会连接 docker.socket 触发 systemd socket activation 把 Docker 拉起来
	if !isDockerDaemonRunning() {
		return true, version, fmt.Errorf(t(i18n.CliCheckErrDaemonNotRunning), "docker")
	}

	return true, version, nil
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
// 最多轮询 maxRetries 次（每次间隔 1s），若端口始终有响应则返回 true（仍在线）。
func CheckPortDown(healthURL string, maxRetries int) bool {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < maxRetries; i++ {
		time.Sleep(1 * time.Second)
		resp, err := client.Get(healthURL)
		if err != nil {
			// 连接被拒绝 = 服务确实已停止
			return false
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return true
}
