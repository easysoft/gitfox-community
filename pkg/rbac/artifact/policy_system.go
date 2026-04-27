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

func (pm *PolicyManager) ensureSystemDefaultRule() error {
	rule := policy{
		Sub:      rbac.WildcardIdentify,
		Space:    rbac.WildcardIdentify,
		Object:   rbac.NewObject(rbac.MatchRecursiveWildcard),
		Action:   enum.CasbinArtifactRoleReader.String(),
		Effect:   rbac.EffectAllow,
		Priority: PrioritySystem,
	}

	_, err := pm.base.AddPolicies(&rule)
	return err
}


