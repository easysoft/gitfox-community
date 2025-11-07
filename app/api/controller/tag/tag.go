package tag

// CreateCommitTag creates a new tag for a repo.
// func (c *Controller) CreateCommitTag(ctx context.Context,
// 	session *auth.Session,
// 	repoRef string,
// 	in *CreateCommitTagInput,
// ) (*CommitTag, []types.RuleViolations, error) {
// 	repo, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoPush)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	if repo.Mirror {
// 		return nil, nil, store.ErrReadOnlyMirrorRepo
// 	}
// 	// set target to default branch in case no branch or commit was provided
// 	if in.Target == "" {
// 		in.Target = repo.DefaultBranch
// 	}

// 	rules, isRepoOwner, err := c.fetchRules(ctx, session, repo)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	violations, err := rules.RefChangeVerify(ctx, protection.RefChangeVerifyInput{
// 		Actor:       &session.Principal,
// 		AllowBypass: in.BypassRules,
// 		IsRepoOwner: isRepoOwner,
// 		Repo:        repo,
// 		RefAction:   protection.RefActionCreate,
// 		RefType:     protection.RefTypeTag,
// 		RefNames:    []string{in.Name},
// 	})
// 	if err != nil {
// 		return nil, nil, fmt.Errorf("failed to verify protection rules: %w", err)
// 	}
// 	if protection.IsCritical(violations) {
// 		return nil, violations, nil
// 	}

// }
