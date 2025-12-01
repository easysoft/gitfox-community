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
	var account string

	// 非用户授权，使用应用名
	if session.User == nil {
		account = session.Principal.App
	} else {
		account = session.User.Account
	}

	// zentao 管理员权限
	if session.Principal != nil && session.Principal.Admin {
		return true, nil
	}

	// 公开资源
	publicAccessAllowed, err := a.CheckPublicAccess(ctx, scope.ResourceType, scope.ResourceID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check public access: %w", err)
	}
	if publicAccessAllowed {
		return true, nil
	}

	log.Debugf(
		"[MembershipAuthorizer]account '%s' requests %s in scope %#v",
		account,
		permission,
		scope,
	)

	return a.permissionCache.Get(
		ctx, PermissionCacheKey{
			Account:    account,
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
