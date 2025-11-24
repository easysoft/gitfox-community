package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type Interface interface {
	ListByFingerprint(ctx context.Context, fingerprint string) ([]publickey.PublicKey, error)
	ListByAccount(ctx context.Context, account string, params req.ListFilter) (*database.PageResult[*publickey.PublicKey], error)
	MarkAsVerified(ctx context.Context, id int64, verified int64) error

	Create(ctx context.Context, publicKey PublicKey) error
}
