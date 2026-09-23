// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package repo

import (
	"regexp"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/jobs"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

type ImportInput struct {
	SpaceID  uint     `json:"spaceID" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Desc     string   `json:"desc"`
	ACL      string   `json:"acl" binding:"required,oneof=open private"` // 权限（仅支持open/private）
	Mirror   bool     `json:"mirror"`                                    // 是否镜像导入，默认为false
	Product  string   `json:"product"`                                   // 关联产品
	Provider Provider `json:"provider"`
}

var importIdentifierRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`)

// Validate 校验导入参数中的 Name 与 Description。
func (in *ImportInput) Validate() error {
	if !importIdentifierRegex.MatchString(in.Name) {
		return errors.WithI18nKey(nil, i18n.RepoIdentifierInvalid)
	}
	if len(in.Name) > 64 {
		return errors.WithI18nKey(nil, i18n.RepoIdentifierTooLong)
	}
	if strings.HasSuffix(strings.ToLower(in.Name), ".git") {
		return errors.WithI18nKey(nil, i18n.RepoIdentifierInvalidGit)
	}
	if len(in.Desc) > constant.MaxDescByteCountUTF8MB4 {
		return errors.WithI18nKey(nil, i18n.RepoDescriptionTooLong)
	}
	return nil
}

type RepositoryOutput struct {
	Repo
}

type Provider struct {
	Type      constant.RepoProviderType `json:"type" binding:"required,oneof=GitHub GitLab Gitea Gogs Subversion"`
	ID        uint                      `json:"id" binding:"required"`
	Host      string                    `json:"host" binding:"required"`
	Slug      string                    `json:"slug" binding:"required"`
	ProjectID string                    `json:"projectID"`
	Token     string                    `json:"token"`
	// svn:
	Username string `json:"username"`
	Password string `json:"password"`
}

type Progress struct {
	Status       jobs.State      `json:"status"`
	Result       string          `json:"result,omitempty"`
	Failure      string          `json:"failure,omitempty"`
	LastExecuted *datetime.Local `json:"lastExecuted,omitempty"`
	NextExecuted *datetime.Local `json:"nextExecuted,omitempty"`
}

type GitConnectInfo struct {
	Slug      string `json:"slug"`
	ProjectID string `json:"projectID"`
}

type SVNConnectInfo struct {
	Slug     string `json:"slug"`
	User     string `json:"user"`
	Password string `json:"password"`
}
