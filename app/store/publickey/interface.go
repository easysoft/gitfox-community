package publickey

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/publickey_store.go -package=mock
import (
	"context"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/model/publickey"
	"github.com/easysoft/gitfox-community/pkg/database"
)

type Interface interface {
	ListByFingerprint(ctx context.Context, fingerprint string) ([]publickey.PublicKey, error)
	ListByIdentifier(ctx context.Context, identifier string, account string) ([]publickey.PublicKey, error)
	ListByAccount(ctx context.Context, account string, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error)
	ListByAccountWithTotal(ctx context.Context, account string) ([]*publickey.PublicKey, int64, error)
	GetByID(ctx context.Context, account string, id int64) (*publickey.PublicKey, error)
	MarkAsVerified(ctx context.Context, id int64) error

	Create(ctx context.Context, publicKey PublicKey) error
	UpdateByID(ctx context.Context, account string, id int64, publicKey PublicKey) error
	DeleteByID(ctx context.Context, account string, id int64) error
}
