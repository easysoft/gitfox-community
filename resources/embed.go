// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package resources

import "embed"

var (
	//go:embed locales
	EmbeddedLocales embed.FS

	//go:embed pipeline/*.yaml
	EmbeddedPipelines embed.FS

	//go:embed scanrule/*.json
	EmbeddedScanRules embed.FS

	//go:embed license
	Licence embed.FS

	//go:embed gitignore
	Gitignore embed.FS
)
