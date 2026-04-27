// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	objectKind  = "artifact"
	modelDefine = `
[request_definition]
r = sub, space, obj, act

[policy_definition]
p = sub, space, obj, act, eft, priority

[role_definition]
g = _, _
g2 = _, _

[policy_effect]
e = priority(p.eft) || deny

[matchers]
m = (g(r.sub, p.sub) || keyMatch(r.sub, p.sub)) && keyMatch(r.space, p.space) && wMatch(r.obj, p.obj) && (keyMatch(r.act, p.act) || g2(r.act, p.act))
`
)

type Manager struct {
	base *rbac.Manager
	enf  casbin.IEnforcer
	PM   *PolicyManager
	RM   *RoleManager
}

func NewManager(db *gorm.DB) (*Manager, error) {
	m, err := model.NewModelFromString(modelDefine)
	if err != nil {
		return nil, err
	}

	base, err := rbac.NewManager(db, objectKind, m)
	if err != nil {
		return nil, err
	}
	base.Enf.AddFunction("wMatch", rbac.WildcardMatchFunc)

	return &Manager{
		enf: base.Enf,
		PM:  NewPolicyManager(base),
		RM:  NewRoleManager(base),
	}, nil
}

func (m *Manager) Enforce(ctx context.Context, sub, space, obj rbac.RequestIdentity, act enum.CasbinAction) (bool, []string, error) {
	log.Ctx(ctx).Debug().Msgf("sub: %s, space: %s, obj: %s, act: %s", sub.Key(), space.Key(), obj.Key(), act)
	ok, reason, err := m.enf.EnforceEx(sub.Key(), space.Key(), obj.Key(), act.String())
	if err != nil {
		return false, nil, err
	}
	if !ok {
		log.Ctx(ctx).Warn().Msgf("enforce failed, %+v", reason)
		return false, nil, rbac.ErrNotAuthorized
	}
	return ok, reason, err
}

// HasPoliciesForScope 检查指定作用域（space key）是否存在任何策略记录。
// 用于三级权限继承链的判断：有策略则覆盖（不回落），无策略则向上继承。
func (m *Manager) HasPoliciesForScope(scope rbac.RequestIdentity) bool {
	policies, err := m.enf.GetFilteredPolicy(1, scope.Key())
	if err != nil {
		return false
	}
	return len(policies) > 0
}

