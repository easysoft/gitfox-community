package publickey

import (
	"crypto/sha256"
	"encoding/base64"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
)

var AllowedTypes = []string{
	gossh.KeyAlgoRSA,
	gossh.KeyAlgoECDSA256,
	gossh.KeyAlgoECDSA384,
	gossh.KeyAlgoECDSA521,
	gossh.KeyAlgoED25519,
	gossh.KeyAlgoSKECDSA256,
	gossh.KeyAlgoSKED25519,
}

var DisallowedTypes = []string{
	gossh.KeyAlgoDSA,
}

func From(key gossh.PublicKey) KeyInfo {
	return KeyInfo{
		Key: key,
	}
}

func ParseString(keyData string) (KeyInfo, string, error) {
	return Parse([]byte(keyData))
}

func Parse(keyData []byte) (KeyInfo, string, error) {
	publicKey, comment, _, _, err := gossh.ParseAuthorizedKey(keyData)
	if err != nil {
		return KeyInfo{}, "", err
	}

	keyType := publicKey.Type()

	// explicitly disallowed
	if slices.Contains(DisallowedTypes, keyType) {
		return KeyInfo{}, "", errors.InvalidArgument.Newf("keys of type %s are not allowed", keyType)
	}

	// only allowed
	if !slices.Contains(AllowedTypes, keyType) {
		return KeyInfo{}, "", errors.InvalidArgument.Newf("allowed key types are %v", AllowedTypes)
	}

	return KeyInfo{
		Key: publicKey,
	}, comment, nil
}

type KeyInfo struct {
	Key gossh.PublicKey
}

func (key KeyInfo) Matches(s string) bool {
	otherKey, _, _, _, err := gossh.ParseAuthorizedKey([]byte(s))
	if err != nil {
		return false
	}

	return key.MatchesKey(otherKey)
}

func (key KeyInfo) MatchesKey(otherKey gossh.PublicKey) bool {
	return ssh.KeysEqual(key.Key, otherKey)
}

func (key KeyInfo) Fingerprint() string {
	sum := sha256.New()
	sum.Write(key.Key.Marshal())
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(sum.Sum(nil))
}

func (key KeyInfo) Type() string {
	return key.Key.Type()
}
