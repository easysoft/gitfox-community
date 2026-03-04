package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"

	"github.com/gin-gonic/gin"
	"github.com/gliderlabs/ssh"
)

type Interface interface {
	ValidateKey(ctx context.Context, publicKey ssh.PublicKey) (*principal.User, int64, error)
	MarkKeyAsVerified(ctx context.Context, keyID int64) error

	CreatePublicKey(ctx context.Context, session *auth.Session, params *publickey.AddPublicKeyInput) error
	UpdatePublicKey(ctx context.Context, session *auth.Session, id int64, params *publickey.UpdatePublicKeyInput) error
	DeletePublicKey(ctx context.Context, session *auth.Session, id int64) error
	ListPublicKeys(ctx context.Context, session *auth.Session, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error)
	ListPublicKeysWithTotal(ctx *gin.Context, session *auth.Session) ([]*publickey.PublicKey, int64, error)
	GetPublicKeyByID(ctx *gin.Context, session *auth.Session, id int64) (*publickey.PublicKey, error)
}
