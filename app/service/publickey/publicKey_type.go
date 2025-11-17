package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"github.com/gliderlabs/ssh"
)

type Interface interface {
	ValidateKey(ctx context.Context, publicKey ssh.PublicKey) (*principal.PrincipalInfo, error)

	CreatePublicKey(ctx context.Context, session *auth.Session, params *publickey.CreatePublicKeyInput) error
}
