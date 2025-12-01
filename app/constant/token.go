package constant

// TokenType represents the type of the JWT token.
type TokenType string

const (
	// TokenTypePAT is a personal access token.
	TokenTypePAT TokenType = "pat"

	TokenTypeApp TokenType = "app"
)
