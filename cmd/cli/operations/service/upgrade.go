package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/common"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/go-version"
	"github.com/minio/selfupdate"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamUpgrade = "upgrade"

	// upgradeStopTimeout 等待服务停止的最长时间。
	upgradeStopTimeout = 5 * time.Second

	// devVersionFallback 与 common.GetVersionInfo 中 fallback 一致；
	// 用于识别"未经 task build 注入"的开发构建。
	devVersionFallback = "0.0.1"
)

// RegisterUpgrade 注册 `gitfox upgrade` 子命令。
//
// 新流程：upgrade 由"新版二进制自身"调起，不再接收任何参数。
// 通过 InspectService() 自动定位 systemd 注册的旧二进制路径，对其执行替换。
func RegisterUpgrade(app *kingpin.Application) {
	app.Command(ParamUpgrade, clii18n.T(i18n.CliUpgradeDescription)).
		Action(runUpgrade)
}

// runUpgrade 是 `gitfox upgrade` 的核心执行入口。
//
// 流程：
//  1. CheckSudo + dev-build 自检
//  2. InspectService() 获取系统注册的旧二进制路径与工作目录
//  3. 路径自检：若当前进程已是注册位置，报错（无意义升级）
//  4. 探测旧二进制版本（容忍旧版本无 --json）
//  5. 版本对比：相同跳过 / 降级拒绝 / 升级继续
//  6. 备份 backups/<oldVer>/<oldBuildDate>/{gitfox, prod.env, prod.yaml}
//  7. 服务在跑则 stop，30s 限时确认下线
//  8. selfupdate.Apply 用本进程二进制覆盖旧二进制（自带 .old 兜底 + 失败回滚）
//  9. 重启 + 健康探针
func runUpgrade(_ *kingpin.ParseContext) error {
	t := clii18n.T

	errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4500"))
	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8C00"))
	headerStyle := lipgloss.NewStyle().Bold(true)
	doneStyle := lipgloss.NewStyle()

	// ========== 阶段 0：权限 + dev-build 自检 ==========
	if err := CheckSudo(); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	newInfo := common.GetVersionInfo()
	if common.Version == "" || common.Version == devVersionFallback {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrDevBuild), newInfo.Version)))
		return nil
	}

	// ========== 阶段 1：定位系统注册的旧二进制 ==========
	existing := InspectService()
	if !existing.Registered {
		fmt.Println(errStyle.Render(t(i18n.CliUpgradeErrNotRegistered)))
		return nil
	}
	if existing.Corrupted || existing.ExePath == "" {
		fmt.Println(errStyle.Render(t(i18n.CliUpgradeErrCorrupted)))
		return nil
	}
	oldExePath := existing.ExePath
	oldWorkDir := existing.WorkDir

	// ========== 阶段 2：路径自检 ==========
	curBinPath, err := os.Executable()
	if err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}
	if resolved, rerr := filepath.EvalSymlinks(curBinPath); rerr == nil {
		curBinPath = resolved
	}
	if curBinPath == oldExePath {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrSelfOverwrite), curBinPath)))
		return nil
	}

	// ========== 阶段 3：探测旧二进制版本 ==========
	fmt.Println(t(i18n.CliUpgradeStepCompare))

	skipCompare := false
	oldInfo, err := readBinaryVersion(oldExePath)
	if err != nil {
		// 旧版本不支持 `version --json`，降级为"对方一定是旧版"语义，跳过对比。
		fmt.Println(warnStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeWarnUnknownVersion), err)))
		oldInfo = common.VersionInfo{
			Version:   "unknown",
			BuildDate: "unknown",
		}
		skipCompare = true
	}

	fmt.Println(fmt.Sprintf(t(i18n.CliUpgradeInfoOld), oldInfo.Version, oldInfo.BuildDate))
	fmt.Println(fmt.Sprintf(t(i18n.CliUpgradeInfoNew), newInfo.Version, newInfo.BuildDate))
	fmt.Println()

	if !skipCompare {
		cmp := compareVersionAndDate(oldInfo.Version, oldInfo.BuildDate, newInfo.Version, newInfo.BuildDate)
		switch cmp {
		case 0:
			fmt.Println(doneStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeSameVersion), oldInfo.Version, oldInfo.BuildDate)))
			return nil
		case 1:
			// 旧版本号 > 新版本号 → 拒绝降级
			fmt.Println(warnStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeDowngradeNotSupported),
				newInfo.Version, newInfo.BuildDate, oldInfo.Version, oldInfo.BuildDate)))
			return nil
		}
	}

	// ========== 阶段 4：备份 ==========
	// backups/<oldVersion>/<sanitizedOldBuildDate>/{gitfox, prod.env, prod.yaml}
	backupDir := filepath.Join(oldWorkDir, "backups", oldInfo.Version, sanitizePathSegment(oldInfo.BuildDate))
	fmt.Println(fmt.Sprintf(t(i18n.CliUpgradeStepBackup), backupDir))
	if err := backupCurrent(oldExePath, oldWorkDir, backupDir); err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrBackup), err)))
		return nil
	}
	fmt.Println(doneStyle.Render("✓ " + fmt.Sprintf(t(i18n.CliUpgradeBackupSuccess), backupDir)))
	fmt.Println()

	// ========== 阶段 5：若服务正在运行则停止 ==========
	// prod.env 是端口下线确认与启动后健康探针的必需输入：
	// 缺失则后续 selfupdate 与 systemctl restart 仍会执行，但无法验证应用层是否真停 / 真起。
	// 为避免"看似升级成功实则旧端口僵尸"的误判，此处直接阻断并提示用户修复。
	hc, hcErr := readHealthURLFrom(oldWorkDir)
	if hcErr != nil {
		fmt.Println(errStyle.Render(hcErr.Error()))
		return nil
	}
	wasRunning := IsSystemServiceActive()

	if wasRunning {
		fmt.Println(t(i18n.CliUpgradeStepStop))
		if err := StopService(); err != nil {
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrStop), err)))
			return nil
		}
		// 端口下线确认（用 health 端点反向探测，更彻底）
		if hc.Port != "" && !CheckPortWithTimeout(hc.Port, upgradeStopTimeout) {
			fmt.Println(errStyle.Render(t(i18n.CliUpgradeErrStopTimeout)))
			return nil
		}
		fmt.Println(doneStyle.Render("✓ " + t(i18n.CliUpgradeStopSuccess)))
	} else {
		fmt.Println(warnStyle.Render(t(i18n.CliUpgradeServiceNotRunning)))
	}
	fmt.Println()

	// ========== 阶段 6：替换二进制 ==========
	fmt.Println(t(i18n.CliUpgradeStepApply))
	src, err := os.Open(curBinPath)
	if err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrApply), err)))
		return nil
	}
	defer src.Close()

	// OldSavePath 让 selfupdate 把旧二进制保留到 <oldExe>.old，作为兜底
	oldSavePath := oldExePath + ".old"
	applyOpts := selfupdate.Options{
		TargetPath:  oldExePath,
		TargetMode:  0755,
		OldSavePath: oldSavePath,
	}
	if err := selfupdate.Apply(src, applyOpts); err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrApplyRollback), err, rerr)))
			return nil
		}
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrApply), err)))
		return nil
	}

	// 双保险：显式 chmod +x，防御异常 umask
	if runtime.GOOS != "windows" {
		if err := os.Chmod(oldExePath, 0755); err != nil {
			fmt.Println(warnStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrChmod), err)))
		}
	}

	fmt.Println(doneStyle.Render("✓ " + fmt.Sprintf(t(i18n.CliUpgradeApplySuccess), oldSavePath)))
	fmt.Println()

	// ========== 阶段 7：启动新服务 + 健康检查 ==========
	if wasRunning {
		fmt.Println(t(i18n.CliUpgradeStepStart))
		if err := startServiceForUpgrade(); err != nil {
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUpgradeErrStart), err)))
			return nil
		}
		_ = CheckPortWithTimeout(hc.Port, 15*time.Second)
	}

	// ========== 阶段 8：摘要 ==========
	fmt.Println(headerStyle.Render(t(i18n.CliUpgradeSummaryTitle)))
	fmt.Println(strings.Repeat("─", 60))
	printUpgradeRow(t(i18n.CliUpgradeSummaryOldVersion), fmt.Sprintf("%s (%s)", oldInfo.Version, oldInfo.BuildDate))
	printUpgradeRow(t(i18n.CliUpgradeSummaryNewVersion), fmt.Sprintf("%s (%s)", newInfo.Version, newInfo.BuildDate))
	printUpgradeRow(t(i18n.CliUpgradeSummaryBackupDir), backupDir)
	printUpgradeRow(t(i18n.CliUpgradeSummaryOldBinary), oldSavePath)
	fmt.Println(strings.Repeat("─", 60))

	fmt.Println()
	fmt.Println(doneStyle.Render("✓ " + fmt.Sprintf(t(i18n.CliUpgradeSuccess), oldInfo.Version, newInfo.Version)))
	return nil
}

// printUpgradeRow 升级摘要的统一行渲染。
func printUpgradeRow(label, value string) {
	labelStyle := lipgloss.NewStyle().Width(20)
	valueStyle := lipgloss.NewStyle()
	fmt.Printf("  %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
}

// sanitizePathSegment 把 BuildDate 这类含空格/冒号的字符串清洗成跨平台合法的目录名。
//
//	"2026-06-11 14:30:00" → "2026-06-11_14-30-00"
//	"unknown"             → "unknown"
//	""                    → "unknown"（防御空目录名）
func sanitizePathSegment(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	r := strings.NewReplacer(
		" ", "_",
		":", "-",
		"/", "-",
		"\\", "-",
	)
	return r.Replace(s)
}

// readBinaryVersion 调用 `<bin> version --json` 子进程并解析其输出。
//
// 通过 JSON 协议而非文本解析，避免 i18n 文案变化或 lipgloss 渲染干扰，
// 并与 `gitfox version` 命令共用 common.VersionInfo 结构体。
//
// 旧版本 gitfox 不识别 --json flag，会从 stderr 输出 "unknown long flag" 并 exit 1，
// 此时返回非 nil error，由调用方决定是否容忍。
func readBinaryVersion(binPath string) (common.VersionInfo, error) {
	cmd := exec.Command(binPath, "version", "--json")
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(exitErr.Stderr))
		}
		if stderr != "" {
			return common.VersionInfo{}, fmt.Errorf("%w: %s", err, stderr)
		}
		return common.VersionInfo{}, err
	}

	var info common.VersionInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return common.VersionInfo{}, fmt.Errorf("%s: %w", clii18n.T(i18n.CliUpgradeErrParseVersion), err)
	}
	if info.Version == "" {
		return common.VersionInfo{}, fmt.Errorf(clii18n.T(i18n.CliUpgradeErrParseVersion), "empty version field")
	}
	return info, nil
}

// compareVersionAndDate 先按 SemVer 比较 version，再按 BuildDate 时间比较。
// 返回值：-1 表示 a < b（应升级）；0 完全相同；1 表示 a > b（降级）。
func compareVersionAndDate(aVer, aDate, bVer, bDate string) int {
	if c := compareSemver(aVer, bVer); c != 0 {
		return c
	}
	return compareBuildDate(aDate, bDate)
}

// compareSemver 先剥掉项目惯用的 "cli-" 前缀，再交给 hashicorp/go-version。
// 解析失败时回退到字符串字典序，保证不会因为奇异 tag 直接 panic。
func compareSemver(a, b string) int {
	va, errA := version.NewVersion(stripVersionPrefix(a))
	vb, errB := version.NewVersion(stripVersionPrefix(b))
	if errA == nil && errB == nil {
		return va.Compare(vb)
	}
	return strings.Compare(a, b)
}

// stripVersionPrefix 去掉 gitfox 编译注入的 "cli-" 前缀，
// 让 "cli-2.0.1" 这样的字符串能被 go-version 识别。
func stripVersionPrefix(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "cli-")
	return v
}

// buildDateFormat 与 Taskfile.yml 中 `date '+%Y-%m-%d %H:%M:%S'` 注入格式一致。
const buildDateFormat = "yyyy-mm-dd hh:mm:ss"

// compareBuildDate 用 datetime 包做真实时间比较，解析失败回退字典序。
func compareBuildDate(a, b string) int {
	ta, errA := datetime.FormatStrToTime(a, buildDateFormat)
	tb, errB := datetime.FormatStrToTime(b, buildDateFormat)
	if errA == nil && errB == nil {
		switch {
		case ta.Before(tb):
			return -1
		case ta.After(tb):
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(a, b)
}

// backupCurrent 把当前二进制和两份关键配置文件归档到 backupDir。
// 配置文件不存在时跳过（不视为错误）。
func backupCurrent(curBin, workDir, backupDir string) error {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("mkdir backup dir: %w", err)
	}

	// 二进制（保留可执行权限）
	if err := copyFileWithMode(curBin, filepath.Join(backupDir, filepath.Base(curBin)), 0755); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}

	// 配置文件
	for _, name := range []string{ProdEnvFileName, ProdYamlFileName} {
		srcPath := filepath.Join(workDir, name)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}
		if err := copyFileWithMode(srcPath, filepath.Join(backupDir, name), 0644); err != nil {
			return fmt.Errorf("copy %s: %w", name, err)
		}
	}
	return nil
}

// copyFileWithMode 通过临时文件原子复制，避免半成品备份。
func copyFileWithMode(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}

// startServiceForUpgrade 升级后拉起服务：
//   - Linux：systemctl start gitfox（unit 已存在，路径未变）
//   - Windows：复用 RestartService → SCM
//
// 不复用 runStart：runStart 自带交互式确认与多重健康探针，升级流程希望尽可能静默 + 自动。
func startServiceForUpgrade() error {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err == nil {
			return exec.Command("systemctl", "start", "gitfox").Run()
		}
	case "windows":
		return RestartService() // → windows.go
	}
	return fmt.Errorf("service not registered, please run `gitfox start` manually")
}

// sudo systemctl stop gitfox
//   sudo mv /opt/gitfox/gitfox.old /opt/gitfox/gitfox
//   sudo systemctl start gitfox
