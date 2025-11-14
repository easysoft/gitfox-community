package user

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

func GenerateSessionTokenIdentifier() (string, error) {
	r, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	return fmt.Sprintf("login-%d-%04d", time.Now().Unix(), r.Int64()), nil
}
