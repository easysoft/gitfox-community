// Copyright (c) 2023-2025 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package artifact

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

type policy struct {
	Sub      rbac.RequestIdentity
	Space    rbac.RequestIdentity
	Object   rbac.RequestIdentity
	Action   string
	Effect   rbac.Effect
	Priority Priority
}

func (p *policy) Values() []string {
	return []string{p.Sub.Key(), p.Space.Key(), p.Object.Key(), p.Action, string(p.Effect), p.Priority.String()}
}

type UserRole struct {
	Sub    rbac.RequestIdentity
	Role   string
	Scope  enum.CasbinScope
	Effect rbac.Effect
}

type PolicyManager struct {
	base *rbac.Manager
}

func NewPolicyManager(base *rbac.Manager) *PolicyManager {
	return &PolicyManager{base: base}
}

func (pm *PolicyManager) DenyAllAccessToNode(space, node rbac.RequestIdentity) error {
	denyRule1, err := genNodeRule(rbac.WildcardIdentify, space, node, rbac.MatchWildcard, rbac.EffectDeny, enum.CasbinScopeRecursive)
	if err != nil {
		return err
	}

	denyRule2, _ := genNodeRule(rbac.WildcardIdentify, space, node, rbac.MatchWildcard, rbac.EffectDeny, enum.CasbinScopeSub)

	_, err = pm.base.AddPolicies(denyRule1, denyRule2)
	return err
}

func (pm *PolicyManager) AllowAllAccessOnNode(space, node rbac.RequestIdentity) error {
	denyRule1, err := genNodeRule(rbac.WildcardIdentify, space, node, rbac.MatchWildcard, rbac.EffectDeny, enum.CasbinScopeRecursive)
	if err != nil {
		return err
	}

	denyRule2, _ := genNodeRule(rbac.WildcardIdentify, space, node, rbac.MatchWildcard, rbac.EffectDeny, enum.CasbinScopeSub)

	_, err = pm.base.RemovePolicies(denyRule1, denyRule2)
	return err
}

func (pm *PolicyManager) RemoveAllPolicyFromNode(space, node rbac.RequestIdentity) error {
	obj, err := NewPriorityNode(node.Key())
	if err != nil {
		return err
	}
	_, err = pm.base.RemoveFilteredPolicy(rbac.PolicyItemSpace, space.Key(), node.Key())
	if err != nil {
		return err
	}

	_, err = pm.base.RemoveFilteredPolicy(rbac.PolicyItemSpace, space.Key(), obj.Wildcard().Key())
	if err != nil {
		return err
	}

	_, err = pm.base.RemoveFilteredPolicy(rbac.PolicyItemSpace, space.Key(), obj.RecursiveWildcard().Key())
	if err != nil {
		return err
	}

	return nil
}

func (pm *PolicyManager) GrantRoleToUserOnNode(subject, space, node rbac.RequestIdentity, role enum.CasbinRole, scope enum.CasbinScope) error {
	rule, err := genNodeRule(subject, space, node, role.String(), rbac.EffectAllow, scope)
	if err != nil {
		return err
	}

	_, err = pm.base.AddPolicies(rule)
	return err
}

func (pm *PolicyManager) RevokeRoleForUserFromNode(subject, space, node rbac.RequestIdentity, role enum.CasbinRole, scope enum.CasbinScope) error {
	rule, err := genNodeRule(subject, space, node, role.String(), rbac.EffectAllow, scope)
	if err != nil {
		return err
	}

	_, err = pm.base.RemovePolicies(rule)
	return err
}

func (pm *PolicyManager) EnsureRulesOnNode(want []*rbac.Request, space, node rbac.RequestIdentity) error {
	current, err := pm.GetPoliciesOnNode(space, node)
	if err != nil {
		return err
	}

	addRequests, removeRequests := rbac.CompareRequests(want, current)
	addPolices := make([]rbac.Policy, len(addRequests))
	removePolices := make([]rbac.Policy, len(removeRequests))
	for idx, req := range addRequests {
		p, e := genNodeRule(req.Sub, space, node, req.Role, req.Effect, req.Scope)
		if e != nil {
			return e
		}
		addPolices[idx] = p
	}

	for idx, req := range removeRequests {
		p, e := genNodeRule(req.Sub, space, node, req.Role, req.Effect, req.Scope)
		if e != nil {
			return e
		}
		removePolices[idx] = p
	}

	if len(addPolices) > 0 {
		_, err = pm.base.AddPolicies(addPolices...)
		if err != nil {
			return err
		}
	}

	if len(removePolices) > 0 {
		_, err = pm.base.RemovePolicies(removePolices...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PolicyManager) GetPoliciesOnNode(space, node rbac.RequestIdentity) ([]*rbac.Request, error) {
	obj, err := NewPriorityNode(node.Key())
	if err != nil {
		return nil, err
	}

	rules1, err := pm.base.Enf.GetFilteredPolicy(1, space.Key(), obj.Wildcard().Key())
	if err != nil {
		return nil, err
	}

	rules2, err := pm.base.Enf.GetFilteredPolicy(1, space.Key(), obj.RecursiveWildcard().Key())
	if err != nil {
		return nil, err
	}

	rules := append(rules1, rules2...)
	return convertRulesToUserRole(rules)
}

/*
	GetInheritObject mock an impossible request on special node, should be rejected by parent policy , return the parent object
	example:
    - rules:
      - * space:1 raw/group:1/** * deny 100
      - * space:1 raw/group:1/group:2/** deny 99
    - request:
	   sub1, space:1, raw/group:1/group:2, act1

	request denied by rule1, return raw/group:1 as parent
*/

func (pm *PolicyManager) GetInheritObject(space, node rbac.RequestIdentity) (string, error) {
	mockSub := "_inner_subject"
	mockAction := "_impossible"
	mockSpace := space.Key()
	mockObject := node.Key() + ":_impossible"

	ok, reason, err := pm.base.Enf.EnforceEx(mockSub, mockSpace, mockObject, mockAction)
	if err != nil {
		return "", err
	}

	if ok {
		return "", nil
	}

	objStr := reason[2]
	if objStr == rbac.MatchRecursiveWildcard || objStr == rbac.MatchWildcard {
		return "", nil
	}
	//log.Debug().Msgf("rejected by mock policy sub: %s, space: %s, object: %s, action: %s,reason: %s",
	//	mockSub, mockSpace, mockObject, mockAction, reason)

	return strings.TrimSuffix(strings.TrimSuffix(objStr, "/**"), "/*"), nil
}

func genNodeRule(subject, space, node rbac.RequestIdentity, role string, effect rbac.Effect, scope enum.CasbinScope) (*policy, error) {
	obj, err := NewPriorityNode(node.Key())
	if err != nil {
		return nil, err
	}

	rule := policy{
		Sub:    subject,
		Space:  space,
		Action: role,
		Effect: effect,
	}

	if scope == enum.CasbinScopeRecursive {
		rule.Object = obj.RecursiveWildcard()
		if effect == rbac.EffectDeny {
			rule.Priority = obj.Priority()
		} else {
			rule.Priority = obj.Priority().Sub(PriorityDecreaseAllowRecursiveWildcard)
		}
	} else {
		rule.Object = obj.Wildcard()
		if effect == rbac.EffectDeny {
			rule.Priority = obj.Priority().Sub(PriorityDecreaseDenyWildcard)
		} else {
			rule.Priority = obj.Priority().Sub(PriorityDecreaseAllowWildcard)
		}
	}

	return &rule, nil
}

func convertRulesToUserRole(in [][]string) ([]*rbac.Request, error) {
	result := make([]*rbac.Request, 0)
	for _, rule := range in {
		r := &rbac.Request{
			// maybe an act
			Role:   rule[3],
			Effect: rbac.Effect(rule[4]),
		}
		result = append(result, r)

		sub, err := rbac.ParseIdentity(rule[0])
		if err != nil {
			return nil, err
		}
		r.Sub = sub

		var scope enum.CasbinScope
		obj := rule[2]
		if strings.HasSuffix(obj, "**") {
			scope = enum.CasbinScopeRecursive
		} else if strings.HasSuffix(obj, "*") {
			scope = enum.CasbinScopeSub
		} else {
			scope = enum.CasbinScopeSingle
		}

		r.Scope = scope
	}
	return result, nil
}
