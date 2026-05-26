package container

import (
	"context"

	requtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	ctnAdapter "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// authorize 校验当前 session 对指定制品库的某项权限。
// 返回 nil 表示通过；否则返回 OCI Registry 错误（带 i18n 文案）供 docker 客户端展示。
//   - 未登录或 session 缺失 → ErrUnauthorized + container_auth_required。
//   - 校验失败或拒绝 → ErrDenied + 对应 pull/push/delete i18n key。
func (c *Controller) authorize(ctx context.Context, artifactID int64, perm constant.Permission, deniedKey i18n.I18nKey) error {
	if c.authorizer == nil {
		return nil
	}
	session, ok := requtil.GetAuthSession(ctx)
	if !ok || session == nil {
		return ctnAdapter.ErrUnauthorized.WithDetail(translate(ctx, i18n.ArtifactContainerAuthRequired))
	}
	allowed, err := c.authorizer.Check(ctx, session, constant.Scope{
		ResourceType: constant.ResourceTypeArtifact,
		ResourceID:   uint(artifactID),
	}, perm)
	if err != nil {
		logger.GetLogger(ctx).Warnf("【container authz】check failed: artifactID=%d perm=%s err=%v", artifactID, perm, err)
		return ctnAdapter.ErrDenied.WithDetail(translate(ctx, deniedKey))
	}
	if !allowed {
		return ctnAdapter.ErrDenied.WithDetail(translate(ctx, deniedKey))
	}
	return nil
}

// translate 取当前 ctx 语言对应文案；缺省回退到默认语言。
func translate(ctx context.Context, key i18n.I18nKey) string {
	locale := i18n.GetLocaleFromContext(ctx)
	if locale == "" {
		return i18n.T(string(key))
	}
	return i18n.TWithLocale(locale, string(key))
}
