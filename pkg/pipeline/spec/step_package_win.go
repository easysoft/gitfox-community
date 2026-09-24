// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepWinPackage struct {
	Tool          string `json:"tool,omitempty"`
	AppName       string `json:"appName,omitempty"`
	AppVersion    string `json:"appVersion,omitempty"`
	AppPublisher  string `json:"appPublisher,omitempty"`
	AppBaseDir    string `json:"appBaseDir,omitempty"`
	AppInstallDir string `json:"appInstallDir,omitempty"`
	AppMainFile   string `json:"appMainFile,omitempty"`
	AssetDir      string `json:"assetDir,omitempty"`
	FileName      string `json:"fileName,omitempty"`
}
