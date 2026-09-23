// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package helm

import (
	"encoding/json"

	"github.com/easysoft/gitfox-community/pkg/utils/jsonutil"

	"helm.sh/helm/v3/pkg/chart"
)

type AssetMetadata struct {
	data *chart.Metadata
}

func (am AssetMetadata) ToJSON() (json.RawMessage, error) {
	return jsonutil.ToJSON(am.data)
}
