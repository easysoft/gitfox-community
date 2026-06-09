package repo

import (
	"regexp"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
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
	Type      constant.RepoProviderType `json:"type" binding:"required,oneof=Github Gitlab Gitea Gogs Subversion"`
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
	Status  jobs.State `json:"status"`
	Result  string     `json:"result,omitempty"`
	Failure string     `json:"failure,omitempty"`
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
