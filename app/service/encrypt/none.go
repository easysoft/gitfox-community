package encrypt

// none is an encryption strategy that stores secret
// values in plain text. This is the default strategy
// when no key is specified.
type none struct {
}

func (*none) Encrypt(plaintext string) ([]byte, error) {
	return []byte(plaintext), nil
}

func (*none) Decrypt(ciphertext []byte) (string, error) {
	return string(ciphertext), nil
}
