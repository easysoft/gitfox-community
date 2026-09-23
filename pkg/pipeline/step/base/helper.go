// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package base

const (
	// envMavenMirror Maven 仓库镜像地址环境变量名（maven/gradle 共用）
	envMavenMirror = "MAVEN_MIRROR_URL"
	// defaultMavenMirror 默认的 Maven 仓库镜像地址
	defaultMavenMirror = "https://maven.aliyun.com/repository/public"
)

func getOrDefaultTrue(b *bool) bool {
	if b == nil {
		return true
	}

	return *b
}
