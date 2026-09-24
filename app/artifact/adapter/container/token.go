// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package container

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type hmacKey string

var errInvalidSecret = fmt.Errorf("invalid secret")

// unpackUploadState unpacks and validates the blob upload state from the
// token, using the hmacKey secret.
func (secret hmacKey) unpackUploadState(token string) (blobUploadState, error) {
	var state blobUploadState

	tokenBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return state, err
	}
	mac := hmac.New(sha256.New, []byte(secret))

	if len(tokenBytes) < mac.Size() {
		return state, errInvalidSecret
	}

	macBytes := tokenBytes[:mac.Size()]
	messageBytes := tokenBytes[mac.Size():]

	mac.Write(messageBytes)
	if !hmac.Equal(mac.Sum(nil), macBytes) {
		return state, errInvalidSecret
	}

	if err := json.Unmarshal(messageBytes, &state); err != nil {
		return state, err
	}

	return state, nil
}

// packUploadState packs the upload state signed with and hmac digest using
// the hmacKey secret, encoding to url safe base64. The resulting token can be
// used to share data with minimized risk of external tampering.
func (secret hmacKey) packUploadState(lus blobUploadState) (string, error) {
	mac := hmac.New(sha256.New, []byte(secret))
	p, err := json.Marshal(lus)
	if err != nil {
		return "", err
	}

	mac.Write(p)

	return base64.URLEncoding.EncodeToString(append(mac.Sum(nil), p...)), nil
}
