package auth

type LoginInput struct {
	LoginIdentifier string `json:"login_identifier"`
	Password        string `json:"password"`
}
