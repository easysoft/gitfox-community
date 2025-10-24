package constant

// PublicKeyUsage represents usage type of public key.
type PublicKeyUsage string

// PublicKeyUsage enumeration.
const (
	PublicKeyUsageAuth PublicKeyUsage = "auth"
	PublicKeyUsageSign PublicKeyUsage = "sign"
)
