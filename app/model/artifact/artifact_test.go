// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package artifact

import (
	"os"
	"testing"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/stretchr/testify/assert"
)

// TestMain 加载真实翻译文件。全局 i18n manager 只能初始化一次，放在 TestMain 里
// 保证早于其他测试触发的懒初始化（否则配置会被忽略）。
func TestMain(m *testing.M) {
	i18n.InitGlobal(&i18n.ManagerConfig{
		AutoLoad:        true,
		TranslationsDir: "../../../resources/locales",
	})
	os.Exit(m.Run())
}

// 普通建仓与导入建仓必须得到同一个默认制品库名：按语言取本地化名称；
// locale 缺失或翻译不可用时回退到常量，不得把 i18n key 或 "[locale:key]" 占位串当库名落库。
func TestDefaultArtifactLibViewName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "中文环境", locale: "zh-CN", want: "默认制品库"},
		{name: "英文环境", locale: "en-US", want: "Default"},
		{name: "locale 为空", locale: "", want: DefaultArtifactLibName},
		{name: "未加载的语言", locale: "fr-FR", want: DefaultArtifactLibName},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, DefaultArtifactLibViewName(tt.locale))
		})
	}
}
