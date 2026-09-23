// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package constant

import (
	"fmt"
	"strings"
)

// GitOperationType Git 操作类型
type GitOperationType string

const (
	GitOperationCreate    GitOperationType = "create"
	GitOperationUpdate    GitOperationType = "update"
	GitOperationDelete    GitOperationType = "delete"
	GitOperationForcePush GitOperationType = "force push"
)

// GitServiceType represents the different types of service values send by git's smart http protocol.
// See https://git-scm.com/docs/http-protocol#_smart_clients for more details.
type GitServiceType string

const (
	// GitServiceTypeReceivePack is sent by git push operations (server "receives" data from client).
	GitServiceTypeReceivePack GitServiceType = "receive-pack"
	// GitServiceTypeUploadPack is sent by git pull operations (server "uploads" data to client).
	GitServiceTypeUploadPack GitServiceType = "upload-pack"
)

// ParseGitServiceType parses the git service type string and returns the equivalent enumeration.
// If the value is unknown and doesn't represent a git service type, an error is returned.
func ParseGitServiceType(s string) (GitServiceType, error) {
	switch strings.ToLower(s) {
	case string(GitServiceTypeReceivePack):
		return GitServiceTypeReceivePack, nil
	case string(GitServiceTypeUploadPack):
		return GitServiceTypeUploadPack, nil
	default:
		return "", fmt.Errorf("unknown git service type provided: %q", s)
	}
}
