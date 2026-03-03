package publickey

import (
	"regexp"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type PublicKey struct {
	ID           int64           `json:"-"`
	Account      string          `json:"account"`      // 关联的账号
	VerifiedDate *time.Time      `json:"verifiedDate"` // 公钥是否已验证
	Identifier   string          `json:"identifier"`   // 公钥标识符
	Fingerprint  string          `json:"fingerprint"`  // 公钥指纹
	Content      string          `json:"-"`            // 公钥内容 eg: ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDB...
	Comment      string          `json:"comment"`      // 公钥注释 eg: "user@example.com"
	Type         string          `json:"type"`         // 公钥类型 eg: "ssh-rsa"
	CreatedBy    string          `json:"createdBy"`    // 创建人
	CreatedDate  *datetime.Local `json:"createdDate"`  // 创建时间
	EditedBy     string          `json:"editedBy"`     // 最后编辑人
	EditedDate   *datetime.Local `json:"editedDate"`   // 最后编辑时间
}

type AddPublicKeyInput struct {
	Identifier string `json:"identifier" binding:"required"` // 公钥标识符
	Content    string `json:"content" binding:"required"`    // 公钥内容
}

type UpdatePublicKeyInput struct {
	Identifier string `json:"identifier"` // 公钥标识符
	Content    string `json:"content"`    // 公钥内容
}

const (
	minIdentifierLength = 1
	MaxIdentifierLength = 100
	identifierRegex     = "^[a-zA-Z0-9._$-]+$"
)

func (p *AddPublicKeyInput) Check() error {
	l := len(p.Identifier)
	if l < minIdentifierLength || l > MaxIdentifierLength {
		return errors.InvalidArgument.Newf("identifier length must be between %d and %d", minIdentifierLength, MaxIdentifierLength)
	}
	// 秘钥名称为必填，只能包含字母、数字、'-'、'_'、'.' 和 '$'。
	if ok, _ := regexp.Match(identifierRegex, []byte(p.Identifier)); !ok {
		return errors.InvalidArgument.Newf("identifier must match regex %s", identifierRegex)
	}

	return nil
}

func (p *UpdatePublicKeyInput) Check() error {
	if p.Identifier != "" {
		l := len(p.Identifier)
		if l < minIdentifierLength || l > MaxIdentifierLength {
			return errors.InvalidArgument.Newf("identifier length must be between %d and %d", minIdentifierLength, MaxIdentifierLength)
		}
		if p.Content != "" {
			if ok, _ := regexp.Match(identifierRegex, []byte(p.Identifier)); !ok {
				return errors.InvalidArgument.Newf("identifier must match regex %s", identifierRegex)
			}
		}
	}

	if p.Content == "" && p.Identifier == "" {
		return errors.InvalidArgument.Newf("identifier or content is required")
	}

	return nil
}
