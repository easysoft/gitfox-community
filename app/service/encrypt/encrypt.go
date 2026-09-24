// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package encrypt

import (
	"errors"
)

// indicates key size is too small.
var errKeySize = errors.New("encryption key must be 32 bytes")

// Encrypter provides field encryption and decryption.
// Encrypted values are currently limited to strings, which is
// reflected in the interface design.
type Encrypter interface {
	Encrypt(plaintext string) ([]byte, error)
	Decrypt(ciphertext []byte) (string, error)
}

// Provide returns an Encrypter implementation based on the key.
// If key is empty, a no-op plaintext encrypter is returned.
func Provide(key string, compat bool) (Encrypter, error) {
	if key == "" {
		return &none{}, nil
	}
	return New(key, compat)
}
