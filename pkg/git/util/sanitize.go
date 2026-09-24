// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package util

import (
	"bytes"
	"unicode"
	"unsafe"
)

type sanitizedError struct {
	err error
}

func (err sanitizedError) Error() string {
	return SanitizeCredentialURLs(err.err.Error())
}

func (err sanitizedError) Unwrap() error {
	return err.err
}

// SanitizeErrorCredentialURLs wraps the error and make sure the returned error message doesn't contain sensitive credentials in URLs
func SanitizeErrorCredentialURLs(err error) error {
	return sanitizedError{err: err}
}

const userPlaceholder = "sanitized-credential"

var schemeSep = []byte("://")

// SanitizeCredentialURLs remove all credentials in URLs (starting with "scheme://") for the input string: "https://user:pass@domain.com" => "https://sanitized-credential@domain.com"
func SanitizeCredentialURLs(s string) string {
	bs := UnsafeStringToBytes(s)
	schemeSepPos := bytes.Index(bs, schemeSep)
	if schemeSepPos == -1 || bytes.IndexByte(bs[schemeSepPos:], '@') == -1 {
		return s // fast return if there is no URL scheme or no userinfo
	}
	out := make([]byte, 0, len(bs)+len(userPlaceholder))
	for schemeSepPos != -1 {
		schemeSepPos += 3         // skip the "://"
		sepAtPos := -1            // the possible '@' position: "https://foo@[^here]host"
		sepEndPos := schemeSepPos // the possible end position: "The https://host[^here] in log for test"
	sepLoop:
		for ; sepEndPos < len(bs); sepEndPos++ {
			c := bs[sepEndPos]
			if ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z') || ('0' <= c && c <= '9') {
				continue
			}
			switch c {
			case '@':
				sepAtPos = sepEndPos
			case '-', '.', '_', '~', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '%':
				continue // due to RFC 3986, userinfo can contain - . _ ~ ! $ & ' ( ) * + , ; = : and any percent-encoded chars
			default:
				break sepLoop // if it is an invalid char for URL (eg: space, '/', and others), stop the loop
			}
		}
		// if there is '@', and the string is like "s://u@h", then hide the "u" part
		if sepAtPos != -1 && (schemeSepPos >= 4 && unicode.IsLetter(rune(bs[schemeSepPos-4]))) && sepAtPos-schemeSepPos > 0 && sepEndPos-sepAtPos > 0 {
			out = append(out, bs[:schemeSepPos]...)
			out = append(out, userPlaceholder...)
			out = append(out, bs[sepAtPos:sepEndPos]...)
		} else {
			out = append(out, bs[:sepEndPos]...)
		}
		bs = bs[sepEndPos:]
		schemeSepPos = bytes.Index(bs, schemeSep)
	}
	out = append(out, bs...)
	return UnsafeBytesToString(out)
}

// UnsafeBytesToString uses Go's unsafe package to convert a byte slice to a string.
func UnsafeBytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// UnsafeStringToBytes uses Go's unsafe package to convert a string to a byte slice.
func UnsafeStringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
