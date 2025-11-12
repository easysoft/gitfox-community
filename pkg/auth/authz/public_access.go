package authz

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

// CheckPublicAccess checks if the requested permission is public for the provided scope and resource.
// CheckPublicAccess 检查所请求的权限是否为所提供的范围和资源的公共权限。
func (a *MembershipAuthorizer) CheckPublicAccess(ctx context.Context, resourceType constant.ResourceType, resourceID uint, permission constant.Permission) (bool, error) {

	// only view permissions allowed for public access.
	// 仅允许公开访问时查看权限。
	if resourceType == constant.ResourceTypeRepo &&
		(permission != constant.PermissionRepoView &&
			permission != constant.PermissionPipelineView) {
		return false, nil
	}

	if resourceType == constant.ResourceTypeSpace &&
		permission != constant.PermissionSpaceView {
		return false, nil
	}

	resourceIsPublic, err := a.PublicAccessCache.Get(ctx, PublicAccessCacheKey{
		Scope: constant.Scope{
			ResourceType: resourceType,
			ResourceID:   resourceID,
		},
	})
	if err != nil {
		return false, errors.Wrapf(err, "failed to check public accessabillity of %s %q", resourceType, resourceID)
	}

	return resourceIsPublic, nil
}
