package artifact

import (
	"context"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/entity"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/rbac"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// resolveScope 根据 spaceID 和 repoID 构建权限作用域标识。
// repoID > 0 时返回 repo 级 scope (space:{spaceID}/repo:{repoID})，否则返回 space 级 scope (space:{spaceID})。
func resolveScope(spaceID, repoID int64) rbac.RequestIdentity {
	if repoID > 0 {
		return rbac.NewRepoSpace(spaceID, repoID)
	}
	return rbac.NewSpace(spaceID)
}

// SetEntityPermission 设置节点级权限。
func (s *Service) SetEntityPermission(ctx context.Context, spaceID int64, entityId string, permReq *model.PermissionRequest) (interface{}, error) {
	e, err := entity.Parse(ctx, s.artStore, entityId)
	if err != nil {
		return nil, err
	}

	space := rbac.NewSpace(spaceID)
	target := rbac.NewObject(e.PolicyLocation())

	if permReq.Private {
		err = s.perm.PM.DenyAllAccessToNode(space, target)
		if err != nil {
			return nil, err
		}
		var requests []*rbac.Request
		for _, p := range permReq.Rules {
			for _, account := range p.Accounts {
				subject := rbac.NewUser(account)
				requests = append(requests, &rbac.Request{Sub: subject, Role: p.Role.String(), Scope: p.Scope, Effect: rbac.EffectAllow})
			}
		}
		err = s.perm.PM.EnsureRulesOnNode(requests, space, target)
		if err != nil {
			return nil, err
		}
	} else {
		err = s.perm.PM.RemoveAllPolicyFromNode(space, target)
		if err != nil {
			return nil, err
		}
	}

	return s.getEntityPermissions(ctx, space, target)
}

// GetEntityPermissions 获取节点级权限配置。
func (s *Service) GetEntityPermissions(ctx context.Context, spaceID int64, entityId string) (*model.ArtifactEntityPermissions, error) {
	e, err := entity.Parse(ctx, s.artStore, entityId)
	if err != nil {
		return nil, err
	}

	space := rbac.NewSpace(spaceID)
	target := rbac.NewObject(e.PolicyLocation())
	return s.getEntityPermissions(ctx, space, target)
}

func (s *Service) getEntityPermissions(ctx context.Context, space, target rbac.RequestIdentity) (*model.ArtifactEntityPermissions, error) {
	rules, err := s.perm.PM.GetPoliciesOnNode(space, target)
	if err != nil {
		return nil, err
	}

	res := &model.ArtifactEntityPermissions{
		ArtifactPermissions: model.ArtifactPermissions{
			Rules: make([]*model.ArtifactPermRule, 0),
		},
	}
	ruleMap := make(map[string]*model.ArtifactPermRule)
	for _, rule := range rules {
		if rule.Effect == rbac.EffectDeny {
			if rule.Sub.Key() == rbac.WildcardIdentify.Key() {
				res.Private = true
			}
			continue
		}

		// only show allowed roles
		mKey := fmt.Sprintf("%s@%s", rule.Role, rule.Scope)
		r, ok := ruleMap[mKey]
		if !ok {
			r = &model.ArtifactPermRule{
				Role: enum.CasbinRole(rule.Role), Scope: rule.Scope,
			}
			ruleMap[mKey] = r
		}

		subject, e := s.parseSubject(ctx, rule.Sub.Key())
		if e != nil {
			return nil, e
		}
		r.Subjects = append(r.Subjects, subject)
	}

	for _, r := range ruleMap {
		res.Rules = append(res.Rules, r)
	}

	// parse inherit parent
	deniedBy, err := s.perm.PM.GetInheritObject(space, target)
	if err != nil {
		logger.GetLogger(ctx).Errorf("get inherit object failed: %v", err)
		return nil, err
	}

	if deniedBy == "" {
		return res, nil
	}

	var denyEntityId string
	if strings.Contains(deniedBy, "/") {
		denyEntityId = deniedBy[strings.LastIndex(deniedBy, "/")+1:]
	} else {
		denyEntityId = deniedBy
	}
	denyEntityId = strings.ReplaceAll(denyEntityId, ":", ".")

	eid, err := entity.Parse(ctx, s.artStore, denyEntityId)
	if err != nil {
		return nil, err
	}
	res.Inherit.Inherited = true
	res.Inherit.InheritEntityId = denyEntityId
	res.Inherit.InheritFrom = eid.Name
	parents := []string{eid.Name}

	p := eid
	for p.Parent != nil {
		parents = append(parents, p.Parent.Name)
		p = p.Parent
	}
	res.Inherit.InheritParents = parents

	return res, nil
}

// SetSpacePermission 设置空间/仓库级权限。
func (s *Service) SetSpacePermission(ctx context.Context, spaceID, repoID int64, permReq *model.PermissionRequest) (*model.ArtifactPermissions, error) {
	space := resolveScope(spaceID, repoID)

	if permReq.Private {
		// 私有模式：先写 deny-all（拒绝所有人），再写白名单允许规则
		if err := s.perm.PM.DenyAllAccessToSpace(space); err != nil {
			return nil, err
		}
	} else {
		// 公开模式：移除 deny-all 规则（但保留已有的 allow 规则，下面 EnsureRules 会差量更新）
		if err := s.perm.PM.AllowAllAccessOnSpace(space); err != nil {
			return nil, err
		}
	}

	// 无论 private 与否，都写入 rules 中的允许规则（差量 add/remove）
	var requests []*rbac.Request
	for _, p := range permReq.Rules {
		for _, account := range p.Accounts {
			subject := rbac.NewUser(account)
			requests = append(requests, &rbac.Request{Sub: subject, Role: p.Role.String(), Scope: enum.CasbinScopeRecursive, Effect: rbac.EffectAllow})
		}
	}
	if err := s.perm.PM.EnsureRulesOnSpace(requests, space); err != nil {
		return nil, err
	}

	return s.getSpacePermissions(ctx, space)
}

// GetSpacePermissions 获取空间/仓库级权限配置。
func (s *Service) GetSpacePermissions(ctx context.Context, spaceID, repoID int64) (*model.ArtifactPermissions, error) {
	spaceNode := resolveScope(spaceID, repoID)
	return s.getSpacePermissions(ctx, spaceNode)
}

func (s *Service) getSpacePermissions(ctx context.Context, space rbac.RequestIdentity) (*model.ArtifactPermissions, error) {
	rules, err := s.perm.PM.GetPoliciesOnSpace(space)
	if err != nil {
		return nil, err
	}

	res := &model.ArtifactPermissions{
		Rules: make([]*model.ArtifactPermRule, 0),
	}
	ruleMap := make(map[string]*model.ArtifactPermRule)
	for _, rule := range rules {
		if rule.Effect == rbac.EffectDeny {
			if rule.Sub.Key() == rbac.WildcardIdentify.Key() {
				res.Private = true
			}
			continue
		}

		// only show allowed roles
		mKey := fmt.Sprintf("%s@%s", rule.Role, rule.Scope)
		r, ok := ruleMap[mKey]
		if !ok {
			r = &model.ArtifactPermRule{
				Role: enum.CasbinRole(rule.Role), Scope: rule.Scope,
			}
			ruleMap[mKey] = r
		}

		subject, e := s.parseSubject(ctx, rule.Sub.Key())
		if e != nil {
			return nil, e
		}
		r.Subjects = append(r.Subjects, subject)
	}

	for _, r := range ruleMap {
		res.Rules = append(res.Rules, r)
	}
	return res, nil
}

// GetSpaceMembers 获取空间/仓库级制品库成员列表。
func (s *Service) GetSpaceMembers(ctx context.Context, spaceID, repoID int64) ([]*model.CasbinSubject, error) {
	spaceNode := resolveScope(spaceID, repoID)

	res, err := s.perm.RM.GetMembersInSpace(spaceNode)
	if err != nil {
		return []*model.CasbinSubject{}, err
	}

	subjects := make([]*model.CasbinSubject, 0)
	for _, sub := range res {
		if sub == rbac.MatchWildcard || sub == rbac.MatchRecursiveWildcard {
			continue
		}

		subject, e := s.parseSubject(ctx, sub)
		if e != nil {
			return subjects, e
		}
		subjects = append(subjects, subject)
	}
	return subjects, nil
}

// parseSubject 将 casbin 策略中的 subject key（如 "user:admin"）解析为 CasbinSubject。
// subject key 格式为 "{kind}:{account}"，kind 目前只有 "user"。
func (s *Service) parseSubject(ctx context.Context, in string) (*model.CasbinSubject, error) {
	frames := strings.Split(in, ":")
	if len(frames) != 2 {
		return nil, fmt.Errorf("invalid subject format: %s", in)
	}

	kind := frames[0]
	account := frames[1]

	u, err := s.userStore.FindByAccount(ctx, account)
	if err != nil {
		return nil, fmt.Errorf("user not found for account '%s': %w", account, err)
	}

	return &model.CasbinSubject{
		Type:        kind,
		Id:          u.ID,
		UID:         u.Account,
		DisplayName: u.Account,
	}, nil
}
