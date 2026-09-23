// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package authz

import (
	"time"

	"github.com/google/wire"
)

// AuthzWireSet provides authorization related dependencies
var AuthzWireSet = wire.NewSet(
	ProvideCacheDuration,
	NewPermissionCache,
	NewPublicAccessCache,
	NewMembershipAuthorizer,
	wire.Bind(new(Authorizer), new(*MembershipAuthorizer)),
)

// ProvideCacheDuration 提供缓存时间配置
func ProvideCacheDuration() time.Duration {
	// 缓存时间可以根据需要调整，这里设置为 5 分钟
	return 5 * time.Minute
}
