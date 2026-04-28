// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package authorizer

import (
	"context"
	"fmt"

	"errors"

	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac/artifact"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// Authorizer 制品库权限校验接口。
// 支持三级作用域继承链：Repo → Space → Global → 系统默认。
type Authorizer interface {
	Authorize(ctx context.Context, session *auth.Session, spaceID, repoID int64, entity *model.ArtifactEntity, action enum.CasbinAction) error
}

type ArtifactAuthorizer struct {
	rbacManager *artifact.Manager
}

func New(m *artifact.Manager) *ArtifactAuthorizer {
	return &ArtifactAuthorizer{
		rbacManager: m,
	}
}

// Authorize 实现三级作用域权限校验。
//
// 三种权限模式（由外部通过 SetSpacePermission/SetEntityPermission 配置）：
//
//  1. 默认（无任何 space 级策略）：
//     系统全局默认规则 "* * ** reader allow" 中，"*" 匹配所有用户。
//     但默认规则 Sub 是 WildcardIdentify，匿名也匹配——
//     为此在此处对匿名请求提前拦截：已登录才继续，匿名返回 authentication required。
//
//  2. 公开（SetSpacePermission private=false，写入 "* space reader allow" 规则）：
//     匿名 subject 用 WildcardIdentify 代入 Casbin，命中 * 规则，放行。
//
//  3. 白名单（SetSpacePermission private=true + 指定 accounts）：
//     Casbin 中有 deny-all（* deny）+ 允许规则（具体 account allow）。
//     匿名 subject 无具体 account，命中 deny-all，拒绝。
//     已登录但不在白名单中的用户，命中 deny-all，拒绝。
//
// 查找逻辑（策略有无驱动的继承/覆盖）：
//  1. repoID > 0 → 检查 repo 级是否有策略，有则用 repo 级 Enforce（覆盖），无则继续
//  2. spaceID > 0 → 检查 space 级是否有策略，有则用 space 级 Enforce（覆盖），无则继续
//  3. 使用全局级 (space:0) Enforce（包含系统默认规则）
func (a *ArtifactAuthorizer) Authorize(ctx context.Context, session *auth.Session, spaceID, repoID int64, entity *model.ArtifactEntity, action enum.CasbinAction) error {
	// 构建 subject：
	// - 已登录：使用 user:{account}
	// - 匿名：使用 WildcardIdentify（"*"），让 Casbin 根据策略决定是否放行。
	//   默认模式和白名单模式会在全局级拦截匿名；公开模式会放行。
	var sub rbac.RequestIdentity
	if session != nil && session.User != nil {
		sub = rbac.NewUser(session.User.Account)
	} else {
		sub = rbac.WildcardIdentify
	}

	// 1. 仓库级：有策略则覆盖，不回落
	if repoID > 0 {
		repoScope := rbac.NewRepoSpace(spaceID, repoID)
		if a.hasPolicies(repoScope) {
			return a.enforceWithSub(ctx, sub, repoScope, entity, action)
		}
	}

	// 2. 空间级：有策略则覆盖，不回落
	if spaceID > 0 {
		spaceScope := rbac.NewSpace(spaceID)
		if a.hasPolicies(spaceScope) {
			return a.enforceWithSub(ctx, sub, spaceScope, entity, action)
		}
	}

	// 3. 全局级（space:0）
	// 默认模式：系统规则是 "* * ** reader allow"，* 匹配任意 subject 包括匿名。
	// 为实现「默认匿名不可访问」，在全局级无自定义策略时，对匿名单独拦截。
	if sub == rbac.WildcardIdentify {
		return errors.New("authentication required")
	}
	globalScope := rbac.NewSpace(0)
	return a.enforceWithSub(ctx, sub, globalScope, entity, action)
}

// hasPolicies 检查 Casbin 中是否存在指定 scope 的策略记录。
// 通过查询 policy 表中 v1（space 字段）等于 scope.Key() 的记录来判断。
func (a *ArtifactAuthorizer) hasPolicies(scope rbac.RequestIdentity) bool {
	return a.rbacManager.HasPoliciesForScope(scope)
}

// enforceWithSub 执行单次 Casbin Enforce 校验。
// sub 可以是已登录用户（user:{account}）或匿名占位符（WildcardIdentify="*"）。
func (a *ArtifactAuthorizer) enforceWithSub(ctx context.Context, sub rbac.RequestIdentity, scope rbac.RequestIdentity, entity *model.ArtifactEntity, action enum.CasbinAction) error {
	loc := entity.PolicyLocation()
	ok, reason, err := a.rbacManager.Enforce(ctx,
		sub,
		scope,
		rbac.NewObject(loc),
		action,
	)
	if err != nil {
		return fmt.Errorf("rbac enforce failed: %w", err)
	}

	if !ok {
		logger.GetLogger(ctx).Debugf("authorize failed for subject '%s' action '%s' on '%s' scope '%s', reason: '%+v'",
			sub.Key(), action, loc, scope.Key(), reason,
		)
		return errors.New("rbac authorize failed")
	}

	return nil
}
