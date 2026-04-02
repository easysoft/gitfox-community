package install

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamStart = "start"
)

func RegisterStart(app *kingpin.Application) {
	app.Command(ParamStart, clii18n.T(i18n.CliInstallStartDescription))
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
		content = urlRegex.ReplaceAllString(content, gitfoxURLLine)
		urlUpdated = true
	}

	// 如果已有 gitfoxPort 行，原地替换
	if portRegex.MatchString(content) {
		content = portRegex.ReplaceAllString(content, gitfoxPortLine)
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
