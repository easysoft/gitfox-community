package publickey

import (
	"context"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/model/auth"
	principal "github.com/easysoft/gitfox-community/app/model/user"
	"github.com/easysoft/gitfox-community/pkg/database"

	"github.com/easysoft/gitfox-community/app/model/publickey"

	"github.com/gin-gonic/gin"
	"github.com/gliderlabs/ssh"
)

type Interface interface {
	ValidateKey(ctx context.Context, publicKey ssh.PublicKey) (*principal.User, int64, error)
	MarkKeyAsVerified(ctx context.Context, keyID int64) error

	Create(ctx context.Context, session *auth.Session, params *publickey.AddPublicKeyInput) error
	Update(ctx context.Context, session *auth.Session, id int64, params *publickey.UpdatePublicKeyInput) error
	Delete(ctx context.Context, session *auth.Session, id int64) error
	List(ctx context.Context, session *auth.Session, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error)
	ListWithTotal(ctx *gin.Context, session *auth.Session) ([]*publickey.PublicKey, int64, error)
	GetByID(ctx *gin.Context, session *auth.Session, id int64) (*publickey.PublicKey, error)
}
