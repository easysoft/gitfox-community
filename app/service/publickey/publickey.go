package publickey

import (
	"context"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	PublicKeyModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/publickey"
	"github.com/easysoft/gitfox/errors"

	"github.com/gliderlabs/ssh"
)

func NewPublicKeyService(
	publicKeyStore publickey.Store,
	pCache store.PrincipalInfoCache,
) Interface {
	return &Service{
		publicKeyStore: publicKeyStore,
		pCache:         pCache,
	}
}

type Service struct {
	publicKeyStore publickey.Store
	pCache         store.PrincipalInfoCache
}

// ValidateKey tries to match the provided key to one of the keys in the database.
// It updates the verified timestamp of the matched key to mark it as used.
func (s Service) ValidateKey(
	ctx context.Context,
	publicKey ssh.PublicKey,
) (*principal.PrincipalInfo, error) {
	key := From(publicKey)
	fingerprint := key.Fingerprint()

	existingKeys, err := s.publicKeyStore.ListByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("failed to read keys by fingerprint: %w", err)
	}

	var keyID int64
	var principalID int64

	for _, existingKey := range existingKeys {
		if !key.Matches(existingKey.Content) {
			continue
		}

		keyID = existingKey.ID
		principalID = existingKey.PrincipalID
	}

	if keyID == 0 {
		return nil, errors.NotFound("Unrecognized key")
	}

	pInfo, err := s.pCache.Get(ctx, principalID)
	if err != nil {
		return nil, fmt.Errorf("failed to pull principal info by public key's principal ID: %w", err)
	}

	err = s.publicKeyStore.MarkAsVerified(ctx, keyID, time.Now().UnixMilli())
	if err != nil {
		return nil, fmt.Errorf("failed mark key as verified: %w", err)
	}

	return pInfo, nil
}

func (s Service) CreatePublicKey(
	ctx context.Context,
	session *auth.Session,
	params *PublicKeyModel.CreatePublicKeyInput,
) error {
	return nil
}
