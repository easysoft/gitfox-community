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
// 查找逻辑（策略有无驱动的继承/覆盖）：
//  1. repoID > 0 → 检查 repo 级是否有策略，有则用 repo 级 Enforce（覆盖），无则继续
//  2. spaceID > 0 → 检查 space 级是否有策略，有则用 space 级 Enforce（覆盖），无则继续
//  3. 使用全局级 (space:0) Enforce（包含系统默认规则）
func (a *ArtifactAuthorizer) Authorize(ctx context.Context, session *auth.Session, spaceID, repoID int64, entity *model.ArtifactEntity, action enum.CasbinAction) error {
	// For anonymous pull access
	if action == enum.CasbinActionArtifactPull && (session == nil || session.User == nil) {
		return nil // Allow anonymous pull by default (configurable later)
	}

	if session == nil || session.User == nil {
		return errors.New("authentication required")
	}

	// 1. 仓库级：有策略则覆盖，不回落
	if repoID > 0 {
		repoScope := rbac.NewRepoSpace(spaceID, repoID)
		if a.hasPolicies(repoScope) {
			return a.enforce(ctx, session, repoScope, entity, action)
		}
	}

	// 2. 空间级：有策略则覆盖，不回落
	if spaceID > 0 {
		spaceScope := rbac.NewSpace(spaceID)
		if a.hasPolicies(spaceScope) {
			return a.enforce(ctx, session, spaceScope, entity, action)
		}
	}

	// 3. 全局级（space:0），系统默认规则 (* * ** reader allow 999) 也会在此匹配
	globalScope := rbac.NewSpace(0)
	return a.enforce(ctx, session, globalScope, entity, action)
}

// hasPolicies 检查 Casbin 中是否存在指定 scope 的策略记录。
// 通过查询 policy 表中 v1（space 字段）等于 scope.Key() 的记录来判断。
func (a *ArtifactAuthorizer) hasPolicies(scope rbac.RequestIdentity) bool {
	return a.rbacManager.HasPoliciesForScope(scope)
}

// enforce 执行单次 Casbin Enforce 校验。
func (a *ArtifactAuthorizer) enforce(ctx context.Context, session *auth.Session, scope rbac.RequestIdentity, entity *model.ArtifactEntity, action enum.CasbinAction) error {
	loc := entity.PolicyLocation()
	ok, reason, err := a.rbacManager.Enforce(ctx,
		rbac.NewUser(session.User.ID),
		scope,
		rbac.NewObject(loc),
		action,
	)
	if err != nil {
		return fmt.Errorf("rbac enforce failed: %w", err)
	}

	if !ok {
		logger.GetLogger(ctx).Debugf("authorize failed for user '%d' action '%s' on '%s' scope '%s', reason: '%+v'",
			session.User.ID, action, loc, scope.Key(), reason,
		)
		return errors.New("rbac authorize failed")
	}

	return nil
}
