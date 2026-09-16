package util

// pushRequestHookCase is an optional shell `case` arm injected into the update
// hook. It implements the paid "code push request" (AGit refs/for) flow which
// derives a temporary ztflow-* branch from a refs/for/* push.
//
// The community edition leaves it empty, so refs/for pushes have no special
// handling. The paid edition installs its implementation via
// SetPushRequestHookCase during startup.
var pushRequestHookCase string

// SetPushRequestHookCase installs the update-hook case arm for the code push
// request flow. It is intended to be called by the paid edition during startup.
func SetPushRequestHookCase(caseArm string) {
	pushRequestHookCase = caseArm
}

// PushRequestHookCase returns the currently installed update-hook case arm.
func PushRequestHookCase() string {
	return pushRequestHookCase
}
