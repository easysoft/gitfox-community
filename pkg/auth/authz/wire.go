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
