// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package common

func GitImage() string {
	return GetEnv("PIPELINE_GIT_IMAGE", DefaultGitImage)
}

func PlaceholderImage() string {
	return GetEnv("PIPELINE_PLACEHOLDER_IMAGE", DefaultPlaceholderImage)
}

func HelperImage() string {
	return GetEnv("PIPELINE_HELPER_IMAGE", DefaultPipelineHelperImage)
}

func GetImageOrDefault(in, defaults string) string {
	if in == "" {
		return defaults
	}
	return in
}
