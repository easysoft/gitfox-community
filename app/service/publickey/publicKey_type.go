package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"

	"github.com/gliderlabs/ssh"
)

type Interface interface {
	ValidateKey(ctx context.Context, publicKey ssh.PublicKey) (*principal.User, error)

	CreatePublicKey(ctx context.Context, session *auth.Session, params *publickey.AddPublicKeyInput) error
	ListPublicKeys(ctx context.Context, session *auth.Session, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error)
}
