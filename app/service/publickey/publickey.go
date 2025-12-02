package publickey

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	PublicKeyModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/publickey"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/publickey"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/gliderlabs/ssh"
)

func NewPublicKeyService(
	publicKeyStore publickey.Interface,
	userStore user.Interface,
) Interface {
	return &Service{
		publicKeyStore: publicKeyStore,
		userStore:      userStore,
	}
}

type Service struct {
	publicKeyStore publickey.Interface
	userStore      user.Interface
}

// ValidateKey tries to match the provided key to one of the keys in the database.
// ValidateKey尝试将提供的键与数据库中的一个键匹配。
func (s Service) ValidateKey(ctx context.Context, publicKey ssh.PublicKey) (*principal.User, int64, error) {
	key := From(publicKey)
	fingerprint := key.Fingerprint()

	existingKeys, err := s.publicKeyStore.ListByFingerprint(ctx, fingerprint)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read keys by fingerprint: %w", err)
	}

	var keyID int64
	var account string

	for _, existingKey := range existingKeys {
		if !key.Matches(existingKey.Content) {
			continue
		}

		keyID = existingKey.ID
		account = existingKey.Account
		break
	}

	if keyID == 0 {
		return nil, 0, errors.NotFound.New("Unrecognized key")
	}

	user, err := s.userStore.FindByAccount(ctx, account)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to pull user info by public key's account: %w", err)
	}

	return user, keyID, nil
}

// MarkKeyAsVerified updates the verified timestamp of the matched key to mark it as used.
// MarkKeyAsVerified 更新匹配密钥的验证时间戳,标记为已使用。
func (s Service) MarkKeyAsVerified(ctx context.Context, keyID int64) error {
	err := s.publicKeyStore.MarkAsVerified(ctx, keyID)
	if err != nil {
		return fmt.Errorf("failed mark key as verified: %w", err)
	}
	return nil
}

func (s Service) CreatePublicKey(ctx context.Context, session *auth.Session, params *PublicKeyModel.AddPublicKeyInput) error {
	key, comment, err := ParseString(params.Content)
	if err != nil {
		err = errors.Wrap(err, "could not parse public key")
		return err
	}

	input := publickey.PublicKey{
		Account:     session.User.Account,
		Identifier:  params.Identifier,
		Fingerprint: key.Fingerprint(), // 根据公钥内容计算的指纹（稳定摘要），用于快速查重与匹配。
		Content:     params.Content,    // 公钥原始内容（如 ssh-ed25519 AAAA... user@host ），用于精确匹配与验证
		Comment:     comment,           // 公钥末尾的注释部分（通常是 user@host ），便于展示和识别。
		Type:        key.Type(),        // 公钥算法类型（如 ssh-ed25519 、 ssh-rsa ），用于展示与后续策略控制。
	}

	err = s.publicKeyStore.Create(ctx, input)
	if err != nil {
		return errors.Wrap(err, "failed to create public key")
	}
	return nil
}

func (s Service) ListPublicKeys(ctx context.Context, session *auth.Session, params req.ListFilter) (*database.PageResult[*PublicKeyModel.PublicKey], error) {
	keys, err := s.publicKeyStore.ListByAccount(ctx, session.User.Account, params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list public keys")
	}
	return keys, nil
}
