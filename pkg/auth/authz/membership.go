package authz

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/httpclient"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

var _ Authorizer = (*MembershipAuthorizer)(nil)

type MembershipAuthorizer struct {
	permissionCache   PermissionCache
	PublicAccessCache PublicAccessCache
	devopsConfig      config.DevopsConfig
}

func NewMembershipAuthorizer(permissionCache PermissionCache, publicAccessCache PublicAccessCache, devopsConfig config.DevopsConfig) *MembershipAuthorizer {
	return &MembershipAuthorizer{
		permissionCache:   permissionCache,
		PublicAccessCache: publicAccessCache,
		devopsConfig:      devopsConfig,
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

	// 当前只对仓库资源进行权限检查
	if scope.ResourceType == constant.ResourceTypeRepo {
		privs, err := getRepoMemberPriv(ctx, a.devopsConfig.URL, scope.ResourceID, a.devopsConfig.Authorization)
		if err != nil {
			log.Errorf("failed to get repo member priv: %v", err)
			return false, fmt.Errorf("failed to get repo member priv: %w", err)
		}

		// 查找用户权限
		userPriv, ok := privs[account]
		if !ok {
			log.Infof("RepoMemberPriv: repoID=%d, account=%s not found in privs", scope.ResourceID, account)
			return false, nil
		}

		// 根据请求的权限类型检查
		switch permission {
		case constant.PermissionRepoView:
			return userPriv.Pull, nil
		default:
			// 其他权限默认允许 push 权限的用户访问
			return userPriv.Push, nil
		}
	}

	return true, nil
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

// repoMemberPriv 仓库成员权限
type repoMemberPriv struct {
	Pull bool `json:"pull"`
	Push bool `json:"push"`
}

// getRepoMemberPriv 获取仓库成员权限
func getRepoMemberPriv(ctx context.Context, baseURL string, repoID uint, authorization string) (map[string]repoMemberPriv, error) {
	url := fmt.Sprintf("%s/repomemberspriv?id=%d", baseURL, repoID)
	headers := map[string]string{
		"Authorization": authorization,
	}
	result := make(map[string]repoMemberPriv)
	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get repo member priv: %w", err)
	}
	return result, nil
}
