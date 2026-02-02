package yaml

// Configures registry credentials.
type Credentials struct {
	Username string          `json:"username,omitempty"`
	Password string          `json:"password,omitempty"`
	Aws      *CredentialsAWS `json:"aws,omitempty"`
}
