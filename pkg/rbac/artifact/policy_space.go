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

func (pm *PolicyManager) DenyAllAccessToSpace(space rbac.RequestIdentity) error {
	denyRule, _ := genSpaceRule(rbac.WildcardIdentify, space, rbac.MatchWildcard, rbac.EffectDeny)

	_, err := pm.base.AddPolicies(denyRule)
	return err
}

func (pm *PolicyManager) AllowAllAccessOnSpace(space rbac.RequestIdentity) error {
	denyRule, _ := genSpaceRule(rbac.WildcardIdentify, space, rbac.MatchWildcard, rbac.EffectDeny)

	_, err := pm.base.RemovePolicies(denyRule)
	return err
}

func (pm *PolicyManager) RemoveAllPolicyFromSpace(space rbac.RequestIdentity) error {
	_, err := pm.base.RemoveFilteredPolicy(rbac.PolicyItemSpace, space.Key(), rbac.RecursiveWildcardIdentify.Key())
	if err != nil {
		return err
	}
	return nil
}

func (pm *PolicyManager) GrantRoleToUserOnSpace(subject, space rbac.RequestIdentity, role enum.CasbinRole) error {
	rule, err := genSpaceRule(subject, space, role.String(), rbac.EffectAllow)
	if err != nil {
		return err
	}

	_, err = pm.base.AddPolicies(rule)
	return err
}

func (pm *PolicyManager) RevokeRoleForUserFromSpace(subject, space rbac.RequestIdentity, role enum.CasbinRole) error {
	rule, err := genSpaceRule(subject, space, role.String(), rbac.EffectAllow)
	if err != nil {
		return err
	}

	_, err = pm.base.RemovePolicies(rule)
	return err
}

func (pm *PolicyManager) EnsureRulesOnSpace(want []*rbac.Request, space rbac.RequestIdentity) error {
	current, err := pm.GetPoliciesOnSpace(space)
	if err != nil {
		return err
	}

	addRequests, removeRequests := rbac.CompareRequests(want, current)
	addPolices := make([]rbac.Policy, len(addRequests))
	for idx, req := range addRequests {
		if req.Role == enum.CasbinArtifactRoleAdmin.String() {
			p := genSpaceOwnerRule(req.Sub, space, req.Role)
			addPolices[idx] = p
		} else {
			p, e := genSpaceRule(req.Sub, space, req.Role, req.Effect)
			if e != nil {
				return e
			}
			addPolices[idx] = p
		}
	}

	// remove all policies for subject in space
	removeSubjectsMap := make(map[string]struct{})
	for _, req := range removeRequests {
		sub := req.Sub.Key()
		if sub == rbac.MatchWildcard || sub == rbac.MatchRecursiveWildcard {
			continue
		}
		removeSubjectsMap[sub] = struct{}{}
	}
	var removeSubjects []string
	for sub := range removeSubjectsMap {
		removeSubjects = append(removeSubjects, sub)
	}

	if len(addPolices) > 0 {
		_, err = pm.base.AddPolicies(addPolices...)
		if err != nil {
			return err
		}
	}

	if len(removeSubjects) > 0 {
		for _, subject := range removeSubjects {
			_, err = pm.base.Enf.RemoveFilteredPolicy(0, subject, space.Key())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (pm *PolicyManager) GetPoliciesOnSpace(space rbac.RequestIdentity) ([]*rbac.Request, error) {
	obj := rbac.NewObject(rbac.MatchRecursiveWildcard)

	rules, err := pm.base.Enf.GetFilteredPolicy(1, space.Key(), obj.Key())
	if err != nil {
		return nil, err
	}

	return convertRulesToUserRole(rules)
}

func genSpaceRule(subject, space rbac.RequestIdentity, role string, effect rbac.Effect) (*policy, error) {
	spaceNode := PriorityNode{node: rbac.MatchRecursiveWildcard, priority: PrioritySpace}
	rule := policy{
		Sub:    subject,
		Space:  space,
		Action: role,
		Effect: effect,
	}

	rule.Object = spaceNode.Identity()
	if effect == rbac.EffectDeny {
		rule.Priority = spaceNode.Priority()
	} else {
		rule.Priority = spaceNode.Priority().Sub(PriorityDecreaseAllowRecursiveWildcard)
	}
	return &rule, nil
}

func genSpaceOwnerRule(subject, space rbac.RequestIdentity, role string) *policy {
	rule := policy{
		Sub:      subject,
		Space:    space,
		Object:   rbac.RecursiveWildcardIdentify,
		Action:   role,
		Effect:   rbac.EffectAllow,
		Priority: PriorityOwner,
	}
	return &rule
}
