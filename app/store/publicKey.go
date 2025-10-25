package store

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publicKey"
)

type IPublicKeyStore interface {
	ListByFingerprint(ctx context.Context, fingerprint string) ([]model.PublicKey, error)
	MarkAsVerified(ctx context.Context, id int64, verified int64) error
}
