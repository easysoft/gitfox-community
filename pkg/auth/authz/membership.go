package authz

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	"github.com/easysoft/gitfox-community/app/store/token"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/httpclient"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

var _ Authorizer = (*MembershipAuthorizer)(nil)

type MembershipAuthorizer struct {
	permissionCache   PermissionCache
	PublicAccessCache PublicAccessCache
	tokenStore        token.Interface
	devopsConfig      config.DevopsConfig
}

func NewMembershipAuthorizer(permissionCache PermissionCache, publicAccessCache PublicAccessCache, tokenStore token.Interface, devopsConfig config.DevopsConfig) *MembershipAuthorizer {
	return &MembershipAuthorizer{
		permissionCache:   permissionCache,
		PublicAccessCache: publicAccessCache,
		tokenStore:        tokenStore,
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

	// pipeline 权限
	if account == "pipeline" {
		return true, nil
	}

	// 当前只对仓库和制品库资源进行成员权限检查
	if scope.ResourceType == constant.ResourceTypeRepo || scope.ResourceType == constant.ResourceTypeArtifact {
		// 从数据库获取 DevOps Token
		entry, err := a.tokenStore.GetDevopsToken(ctx)
		if err != nil {
			log.Errorf("failed to get devops token: %v", err)
			return false, fmt.Errorf("failed to get devops token: %w", err)
		}

		privs, err := a.getMemberPriv(ctx, scope, entry.Key)
		if err != nil {
			log.Errorf("failed to get %s member priv: %v", scope.ResourceType, err)
			return false, err
		}

		// 查找用户权限
		userPriv, ok := privs[account]
		if !ok {
			log.Infof("MemberPriv: resourceType=%s, resourceID=%d, account=%s not found in privs", scope.ResourceType, scope.ResourceID, account)
			return false, nil
		}

		return memberPrivAllows(permission, userPriv), nil
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

// memberPriv 成员权限
type memberPriv struct {
	Pull bool `json:"pull"`
	Push bool `json:"push"`
}

func (a *MembershipAuthorizer) getMemberPriv(ctx context.Context, scope constant.Scope, authorization string) (map[string]memberPriv, error) {
	switch scope.ResourceType {
	case constant.ResourceTypeRepo:
		return getRepoMemberPriv(ctx, a.devopsConfig.URL, scope.ResourceID, authorization)
	case constant.ResourceTypeArtifact:
		return getArtifactMemberPriv(ctx, a.devopsConfig.URL, scope.ResourceID, authorization)
	default:
		return nil, fmt.Errorf("unsupported membership resource type: %s", scope.ResourceType)
	}
}

func memberPrivAllows(permission constant.Permission, priv memberPriv) bool {
	switch permission {
	case constant.PermissionRepoView,
		constant.PermissionArtifactView,
		constant.PermissionArtifactPull,
		constant.PermissionArtifactsDownload:
		return priv.Pull
	default:
		// 其他权限默认允许 push 权限的用户访问
		return priv.Push
	}
}

// getRepoMemberPriv 获取仓库成员权限
func getRepoMemberPriv(ctx context.Context, baseURL string, repoID uint, authorization string) (map[string]memberPriv, error) {
	url := fmt.Sprintf("%s/repomemberspriv?id=%d", baseURL, repoID)
	headers := map[string]string{
		"Authorization": authorization,
	}
	result := make(map[string]memberPriv)
	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get repo member priv: %w", err)
	}
	return result, nil
}

// getArtifactMemberPriv 获取制品库成员权限
func getArtifactMemberPriv(ctx context.Context, baseURL string, artifactID uint, authorization string) (map[string]memberPriv, error) {
	url := fmt.Sprintf("%s/artifactmemberspriv?id=%d", baseURL, artifactID)
	headers := map[string]string{
		"Authorization": authorization,
	}
	result := make(map[string]memberPriv)
	if err := httpclient.Request(ctx, "GET", url, headers, nil, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to get artifact member priv: %w", err)
	}
	return result, nil
}
