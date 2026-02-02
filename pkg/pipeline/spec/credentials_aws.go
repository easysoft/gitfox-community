package yaml

// Configures AWS credentials.
type CredentialsAWS struct {
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}
