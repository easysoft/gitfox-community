package paths

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"github.com/gotidy/ptr"
)

/*
 * Concatenate two paths together (takes care of leading / trailing '/')
 * e.g. (space1/, /space2/) -> space1/space2
 *
 * NOTE: All leading, trailing, and consecutive '/' will be trimmed to ensure correct paths.
 */
func Concatenate(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}

	sb := strings.Builder{}
	for i := 0; i < len(paths); i++ {
		// remove all leading, trailing, and consecutive '/'
		var nextRune *rune
		for _, r := range paths[i] {
			// skip '/' if we already have '/' as next rune or we don't have anything other than '/' yet
			if (nextRune == nil || *nextRune == constant.PathSeparator) && r == constant.PathSeparator {
				continue
			}

			// first time we take a rune we have to make sure we add a separator if needed (guaranteed no '/').
			if nextRune == nil && sb.Len() > 0 {
				sb.WriteString(constant.PathSeparatorAsString)
			}

			// flush the previous rune before storing the next rune
			if nextRune != nil {
				sb.WriteRune(*nextRune)
			}

			nextRune = ptr.Of(r)
		}

		// flush the final rune
		if nextRune != nil && *nextRune != constant.PathSeparator {
			sb.WriteRune(*nextRune)
		}
	}

	return sb.String()
}
