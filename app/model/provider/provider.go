// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import "github.com/easysoft/gitfox-community/pkg/utils/datetime"

// Provider 外部服务模型。
type Provider struct {
	ID          uint            `json:"id"`
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	URL         string          `json:"url"`
	Token       string          `json:"token"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}
