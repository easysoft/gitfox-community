package publickey

import (
	"regexp"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

type PublicKey struct {
	ID          int64      `json:"-"`
	Account     string     `json:"account"`     // 关联的账号
	Verified    *time.Time `json:"verified"`    // 公钥是否已验证
	Identifier  string     `json:"identifier"`  // 公钥标识符
	Fingerprint string     `json:"fingerprint"` // 公钥指纹
	Content     string     `json:"-"`           // 公钥内容 eg: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDB...
	Comment     string     `json:"comment"`     // 公钥注释 eg: "user@example.com"
	Type        string     `json:"type"`        // 公钥类型 eg: "ssh-rsa"
}

type AddPublicKeyInput struct {
	Identifier string `json:"identifier" binding:"required"` // 公钥标识符
	Content    string `json:"content" binding:"required"`    // 公钥内容
}

const (
	minIdentifierLength = 1
	MaxIdentifierLength = 100
	identifierRegex     = "^[a-zA-Z0-9-_.]*$"
)

func (p *AddPublicKeyInput) CheckIdentifier() error {
	l := len(p.Identifier)
	if l < minIdentifierLength || l > MaxIdentifierLength {
		return errors.InvalidArgument.Newf("identifier length must be between %d and %d", minIdentifierLength, MaxIdentifierLength)
	}

	if ok, _ := regexp.Match(identifierRegex, []byte(p.Identifier)); !ok {
		return errors.InvalidArgument.Newf("identifier must match regex %s", identifierRegex)
	}

	return nil
}
