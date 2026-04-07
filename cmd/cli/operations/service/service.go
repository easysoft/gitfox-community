package service

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamInstall = "install"
)

// lastInitWorkspace 记录 init 成功后的工作目录路径，供 install 命令使用。
var lastInitWorkspace string

// isInstallingMode 标记当前是否处于 install 流程中，用于跳过 init 末尾的「请执行 start」提示。
var isInstallingMode bool

func Register(app *kingpin.Application) {
	RegisterInstall(app)
	RegisterInit(app)
	RegisterStart(app)
	RegisterStop(app)
	RegisterUninstall(app)
	RegisterStatus(app)
	RegisterVersion(app)
}

func RegisterInstall(app *kingpin.Application) {
	app.Command(ParamInstall, clii18n.T(i18n.CliInstallDescription)).
		Action(runInstall)
}

// runInstall 一键执行 init + start 两个步骤。
//
// 流程：
//  1. 执行 init（交互式配置 + 复制资源到工作目录）
//  2. 在工作目录下以子进程方式执行 ./gitfox start
//  3. 等待子进程完成后，快速探测服务健康状态
func runInstall(ctx *kingpin.ParseContext) error {
	t := clii18n.T
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Bold(true)

	printBanner(t(i18n.CliInstallWelcome))

	// ========== 阶段 1：执行 init ==========
	isInstallingMode = true
	if err := runInit(ctx); err != nil {
		return err
	}

	workspace := lastInitWorkspace
	if workspace == "" {
		// init 未正常完成（可能用户主动取消），直接退出
		return nil
	}

	// ========== 阶段 2：在工作目录执行 ./gitfox start ==========
	fmt.Println()
	fmt.Println("🚀 " + t(i18n.CliInstallStepStart))
	fmt.Println()

	binName := "gitfox"
	if runtime.GOOS == "windows" {
		binName = "gitfox.exe"
	}
	binPath := filepath.Join(workspace, binName)

	// 确保二进制文件存在
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		fmt.Println(errStyle.Render("❌ " + fmt.Sprintf(t(i18n.CliInstallErrBinNotFound), binPath)))
		return nil
	}

	cmd := exec.Command(binPath, "start")
	cmd.Dir = workspace
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf("❌ "+t(i18n.CliInstallStartExecFailed), err)))
	}

	// ========== 阶段 3：快速探测服务状态（start 内部已做完整健康检查，此处仅判断最终结果） ==========
	envPath := filepath.Join(workspace, "prod.env")
	envMap, _ := godotenv.Read(envPath)
	host := envMap["GITFOX_SERVER_HOST"]
	port := envMap["GITFOX_HTTP_PORT"]

	if host != "" && port != "" {
		healthURL := fmt.Sprintf("http://%s:%s/public/health", host, port)
		client := http.Client{Timeout: 3 * time.Second}
		if resp, err := client.Get(healthURL); err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			printInstallSuccess(workspace, binName)
			return nil
		}
	}

	return nil
}

// printInstallSuccess 打印安装成功后的引导信息。
func printInstallSuccess(workspace, binName string) {
	t := clii18n.T

	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF82"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

	fmt.Println()
	fmt.Println(doneStyle.Render("✅ " + t(i18n.CliInstallSuccess)))
	fmt.Println()

	// 后续命令指引
	fmt.Println(headerStyle.Render(t(i18n.CliInstallNextSteps)))
	fmt.Println(strings.Repeat("─", 60))

	printRow := func(label, value string) {
		fmt.Printf("  %s %s\n", labelStyle.Render(label), valueStyle.Render(value))
	}

	fmt.Println()
	fmt.Printf("  %s\n", t(i18n.CliInstallCdHint))
	fmt.Printf("  $ cd %s\n", workspace)
	fmt.Println()

	printRow(t(i18n.CliInstallCmdStop), fmt.Sprintf("./%s stop", binName))
	printRow(t(i18n.CliInstallCmdStatus), fmt.Sprintf("./%s status", binName))
	printRow(t(i18n.CliInstallCmdUninstall), fmt.Sprintf("./%s uninstall", binName))
	printRow(t(i18n.CliInstallCmdVersion), fmt.Sprintf("./%s version", binName))

	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
}

// printBanner 打印统一风格的欢迎横幅
func printBanner(text string) {
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 2)

	fmt.Println()
	fmt.Println(bannerStyle.Render(text))
	fmt.Println()
}
