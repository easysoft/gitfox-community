package publickey

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

type PublicKey struct {
	ID          int64                   `json:"-"` // frontend doesn't need it
	PrincipalID int64                   `json:"-"` // API always returns keys for the same user
	Created     int64                   `json:"created"`
	Verified    *int64                  `json:"verified"`
	Identifier  string                  `json:"identifier"`
	Usage       constant.PublicKeyUsage `json:"usage"`
	Fingerprint string                  `json:"fingerprint"`
	Content     string                  `json:"-"`
	Comment     string                  `json:"comment"`
	Type        string                  `json:"type"`
}
