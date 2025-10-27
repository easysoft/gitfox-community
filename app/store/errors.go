package store

import "errors"

var (
	ErrUserPass                   = errors.New("invalid username or password")
	ErrUserPassUpdateNotAllowed   = errors.New("password update is not allowed for zentao or ldap users")
	ErrResourceNotFound           = errors.New("resource not found")
	ErrDuplicate                  = errors.New("resource is a duplicate")
	ErrForeignKeyViolation        = errors.New("foreign resource does not exists")
	ErrVersionConflict            = errors.New("resource version conflict")
	ErrPathTooLong                = errors.New("the path is too long")
	ErrPrimaryPathAlreadyExists   = errors.New("primary path already exists for resource")
	ErrPrimaryPathRequired        = errors.New("path has to be primary")
	ErrAliasPathRequired          = errors.New("path has to be an alias")
	ErrPrimaryPathCantBeDeleted   = errors.New("primary path can't be deleted")
	ErrNoChangeInRequestedMove    = errors.New("the requested move doesn't change anything")
	ErrIllegalMoveCyclicHierarchy = errors.New("the requested move is not permitted as it would cause a " +
		"cyclic depdency")
	ErrSpaceWithChildsCantBeDeleted = errors.New("the space can't be deleted as it still contains " +
		"spaces or repos")
	ErrPreConditionFailed          = errors.New("precondition failed")
	ErrReadOnlyMirrorRepo          = errors.New("repository is a read-only mirror")
	ErrRequireFitClient            = errors.New("require fit client")
	ErrBuildInRuleNotAllowDeleted  = errors.New("built-in rule not allow delete operation")
	ErrRepoConfigAINotFoundProvied = errors.New("require space ai provider configure")
	ErrUserAdminNotAllowBlocked    = errors.New("user admin not allow blocked")
	ErrScanRuleNotFound            = errors.New("scan rule not found")
	ErrScanRuleSetNotFound         = errors.New("scan rule set not found")
	ErrScanPlanNotFound            = errors.New("scan plan not found")
	ErrScanRuleSetInUse            = errors.New("scan rule set is being used by scan solution")
	ErrCodOwnersNotFound           = errors.New("CODEOWNERS record not found")
)
