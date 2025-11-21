package authz

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
)

type Authorizer interface {
	/*
	 * Checks whether the principal of the current session with the provided metadata
	 * has the permission to execute the action on the resource within the scope.
	 * Returns
	 *		(true, nil)   - the action is permitted
	 *		(false, nil)  - the action is not permitted
	 *		(false, err)  - an error occurred while performing the permission check and the action should be denied
	 */
	Check(ctx context.Context, session *auth.Session, scope constant.Scope, permission constant.Permission) (bool, error)

	/*
	 * Checks whether the principal of the current session with the provided metadata
	 * has the permission to execute ALL the action on the resource within the scope.
	 * Returns
	 *		(true, nil)   - all requested actions are permitted
	 *		(false, nil)  - at least one requested action is not permitted
	 *		(false, err)  - an error occurred while performing the permission check and all actions should be denied
	 */
	CheckAll(ctx context.Context, session *auth.Session, permissionChecks ...constant.PermissionCheck) (bool, error)
}
