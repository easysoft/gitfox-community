// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

const (
	RoleActionGroup = "g2"
)

type RoleManager struct {
	base *rbac.Manager
}

func NewRoleManager(base *rbac.Manager) *RoleManager {
	return &RoleManager{base: base}
}

func (rm *RoleManager) EnsureDefaultRoles() error {
	roles := map[enum.CasbinRole][]enum.CasbinAction{
		enum.CasbinArtifactRoleReader: {enum.CasbinActionArtifactView, enum.CasbinActionArtifactPull},
		enum.CasbinArtifactRoleWriter: {
			enum.CasbinActionArtifactView, enum.CasbinActionArtifactPull,
			enum.CasbinActionArtifactPush,
		},
		enum.CasbinArtifactRoleAdmin: {
			enum.CasbinActionArtifactView, enum.CasbinActionArtifactPull,
			enum.CasbinActionArtifactPush, enum.CasbinActionArtifactOverride,
			enum.CasbinActionArtifactUpdate, enum.CasbinActionArtifactDelete, enum.CasbinActionArtifactCreate,
			enum.CasbinActionPermission,
		},
	}
	for role, actions := range roles {
		rules := make([][]string, len(actions))
		for i := range actions {
			rules[i] = []string{string(actions[i]), string(role)}
		}
		_, err := rm.base.Enf.AddNamedGroupingPoliciesEx(RoleActionGroup, rules)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rm *RoleManager) ListAllRoles() ([]string, error) {
	return rm.base.Enf.GetAllNamedRoles(RoleActionGroup)
}

func (rm *RoleManager) GetMembersInSpace(space rbac.RequestIdentity) ([]string, error) {
	rules, err := rm.base.Enf.GetFilteredPolicy(1, space.Key())
	if err != nil {
		return nil, err
	}

	var subjects []string
	for _, rule := range rules {
		if rule[0] == "*" || rule[0] == "**" {
			continue
		}
		subjects = append(subjects, rule[0])
	}
	return deduplicateArray(subjects), nil
}

func deduplicateArray(arr []string) []string {
	uniqueMap := make(map[string]bool)
	uniqueArr := make([]string, 0)

	for _, v := range arr {
		if !uniqueMap[v] {
			uniqueMap[v] = true
			uniqueArr = append(uniqueArr, v)
		}
	}

	return uniqueArr
}
