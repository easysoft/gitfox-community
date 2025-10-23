package enum

// PrincipalType defines the supported types of principals.
type PrincipalType string

const (
	// PrincipalTypeUser represents a user.
	PrincipalTypeUser PrincipalType = "user"
	// PrincipalTypeServiceAccount represents a service account.
	PrincipalTypeServiceAccount PrincipalType = "serviceaccount"
	// PrincipalTypeService represents a service.
	PrincipalTypeService PrincipalType = "service"
)

type PrincipalSource string

const (
	PrincipalSourceInternal PrincipalSource = "internal"
	PrincipalSourceLDAP     PrincipalSource = "ldap"
	PrincipalSourceZentao   PrincipalSource = "zentao"
)
