// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

type StepAssetDownloadItem struct {
	SpaceRef          string `json:"spaceRef,omitempty"`
	Format            string `json:"format,omitempty"`
	Type              string `json:"type,omitempty"`
	FilePath          string `json:"filePath,omitempty"`
	DisableUnCompress bool   `json:"disableUnCompress,omitempty"`
	GroupId           int64  `json:"groupId,omitempty"`
	DestPath          string `json:"destPath,omitempty"`
	Location          string `json:"location,omitempty"`
}
