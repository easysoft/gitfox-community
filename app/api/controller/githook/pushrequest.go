// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package githook

import (
	"context"

	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/utils"
	"github.com/easysoft/gitfox-community/pkg/git/hook"
	"github.com/easysoft/gitfox-community/pkg/git/sha"
)

// PushRequestFlow encapsulates the paid "code push request" (AGit refs/for) flow.
//
// The open core calls into this interface at the git-hook decision points. The
// community edition binds NoopPushRequestFlow; the paid edition provides the
// Zentao/ztflow implementation that turns a push to refs/for/<branch> into a
// review pull request.
type PushRequestFlow interface {
	// IsPushRequestRef reports whether ref is a push-request reference (refs/for/...).
	IsPushRequestRef(ref string) bool

	// ResolveReport returns the ref name to report and the effective old SHA for a
	// push-request reference update (used by post-receive event reporting).
	ResolveReport(
		ctx context.Context,
		rgit RestrictedGIT,
		repo *repoModel.Repo,
		env hook.Environment,
		refUpdate hook.ReferenceUpdate,
	) (ref string, effectiveOld sha.SHA)

	// Handle creates the review pull request for a push-request ref. It returns true
	// when it handled the event (whether or not a pull request was created).
	Handle(
		ctx context.Context,
		repo *repoModel.Repo,
		principalAccount string,
		in hook.PostReceiveInput,
		out *hook.Output,
	) bool

	// GroupRef maps a push-request ref to the internal branch name used for
	// pre-receive grouping. ok is false when the ref is not handled.
	GroupRef(ref string) (branchName string, ok bool)

	// CheckForceReview reports whether a direct push to branchName must be
	// rejected because the paid forced-review flow is required. It returns the
	// rejection message when blocked.
	CheckForceReview(
		ctx context.Context,
		repo *repoModel.Repo,
		branchName string,
		rule *utils.BranchPermissionResult,
	) (blocked bool, message string)
}

// NoopPushRequestFlow is the community-edition implementation of PushRequestFlow.
// The code push request feature is a paid capability; this implementation makes
// the git-hook call sites inert without affecting normal branch/tag pushes.
type NoopPushRequestFlow struct{}

var _ PushRequestFlow = (*NoopPushRequestFlow)(nil)

// NewNoopPushRequestFlow creates the community-edition PushRequestFlow implementation.
func NewNoopPushRequestFlow() *NoopPushRequestFlow {
	return &NoopPushRequestFlow{}
}

// IsPushRequestRef always reports false in the community edition.
func (NoopPushRequestFlow) IsPushRequestRef(string) bool { return false }

// ResolveReport returns an empty result; it is never reached in the community edition.
func (NoopPushRequestFlow) ResolveReport(
	_ context.Context,
	_ RestrictedGIT,
	_ *repoModel.Repo,
	_ hook.Environment,
	_ hook.ReferenceUpdate,
) (string, sha.SHA) {
	return "", sha.None
}

// Handle is a no-op in the community edition.
func (NoopPushRequestFlow) Handle(
	_ context.Context,
	_ *repoModel.Repo,
	_ string,
	_ hook.PostReceiveInput,
	_ *hook.Output,
) bool {
	return false
}

// GroupRef is a no-op in the community edition.
func (NoopPushRequestFlow) GroupRef(string) (string, bool) { return "", false }

// CheckForceReview never blocks in the community edition.
func (NoopPushRequestFlow) CheckForceReview(
	_ context.Context,
	_ *repoModel.Repo,
	_ string,
	_ *utils.BranchPermissionResult,
) (bool, string) {
	return false, ""
}
