package authz

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

var _ Authorizer = (*MembershipAuthorizer)(nil)

type MembershipAuthorizer struct {
	permissionCache   PermissionCache
	PublicAccessCache PublicAccessCache
}

func NewMembershipAuthorizer(permissionCache PermissionCache, publicAccessCache PublicAccessCache) *MembershipAuthorizer {
	return &MembershipAuthorizer{
		permissionCache:   permissionCache,
		PublicAccessCache: publicAccessCache,
	}
}

func (a *MembershipAuthorizer) Check(ctx context.Context, session *auth.Session, scope constant.Scope, permission constant.Permission) (bool, error) {
	log := logger.GetLogger(ctx)

	// 公开资源
	publicAccessAllowed, err := a.CheckPublicAccess(ctx, scope.ResourceType, scope.ResourceID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check public access: %w", err)
	}
	if publicAccessAllowed {
		return true, nil
	}

	log.Debugf(
		"[MembershipAuthorizer]id '%d' requests %s in scope %#v",
		session.Principal.ID,
		permission,
		scope,
	)

	if scope.ResourceType == constant.ResourceTypeRepo &&
		(permission != constant.PermissionRepoView &&
			permission != constant.PermissionPipelineView) {
		return false, nil
	}

	if scope.ResourceType == constant.ResourceTypeSpace &&
		permission != constant.PermissionSpaceView {
		return false, nil
	}

	return a.permissionCache.Get(
		ctx, PermissionCacheKey{
			Account:    session.User.Account,
			Scope:      scope,
			Permission: permission,
		},
	)
}

func (a *MembershipAuthorizer) CheckAll(ctx context.Context, session *auth.Session, permissionChecks ...constant.PermissionCheck,
) (bool, error) {
	for i := range permissionChecks {
		p := permissionChecks[i]
		if _, err := a.Check(ctx, session, p.Scope, p.Permission); err != nil {
			return false, err
		}
	}

	return true, nil
}
