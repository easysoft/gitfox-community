package publickey

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
)

type Interface interface {
	ListByFingerprint(ctx context.Context, fingerprint string) ([]publickey.PublicKey, error)
	MarkAsVerified(ctx context.Context, id int64, verified int64) error

	Create(ctx context.Context, publicKey PublicKey) error
}
