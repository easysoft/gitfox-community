package publickey

import (
	"regexp"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type PublicKey struct {
	ID          int64      `json:"-"`
	PrincipalID int64      `json:"-"`
	Verified    *time.Time `json:"verified"`
	Identifier  string     `json:"identifier"`
	Fingerprint string     `json:"fingerprint"`
	Content     string     `json:"-"`
	Comment     string     `json:"comment"`
	Type        string     `json:"type"`
}

type CreatePublicKeyInput struct {
	Identifier string `json:"identifier" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

const (
	minIdentifierLength = 1
	MaxIdentifierLength = 100
	identifierRegex     = "^[a-zA-Z0-9-_.]*$"
)

func (p *CreatePublicKeyInput) CheckIdentifier() error {
	l := len(p.Identifier)
	if l < minIdentifierLength || l > MaxIdentifierLength {
		return errors.InvalidArgument.Newf("identifier length must be between %d and %d", minIdentifierLength, MaxIdentifierLength)
	}

	if ok, _ := regexp.Match(identifierRegex, []byte(p.Identifier)); !ok {
		return errors.InvalidArgument.Newf("identifier must match regex %s", identifierRegex)
	}

	return nil
}
