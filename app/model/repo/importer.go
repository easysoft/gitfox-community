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
	SpaceID    uint     `json:"spaceID" binding:"required"`
	Identifier string   `json:"name" binding:"required"`
	Desc       string   `json:"desc"`
	ACL        string   `json:"acl" binding:"required,oneof=open private"` // 权限（仅支持open/private）
	Product    string   `json:"product"`                                   // 关联产品
	Provider   Provider `json:"provider"`
}

var importIdentifierRegex = regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`)

// Validate 校验导入参数中的 Identifier 与 Description。
func (in *ImportInput) Validate() error {
	if !importIdentifierRegex.MatchString(in.Identifier) {
		return errors.WithI18nKey(nil, i18n.RepoIdentifierInvalid)
	}
	if len(in.Identifier) > 64 {
		return errors.WithI18nKey(nil, i18n.RepoIdentifierTooLong)
	}
	if strings.HasSuffix(strings.ToLower(in.Identifier), ".git") {
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
	Type     constant.RepoProviderType `json:"type" binding:"required,oneof=github gitlab gitlab.com gitlab.cn gitea gogs"`
	Host     string                    `json:"host"`
	Username string                    `json:"username"`
	Password string                    `json:"password"`
	Repo     string                    `json:"repo" binding:"required"` // 远端仓库地址（如owner/repo）
}

type Progress struct {
	Status  jobs.State `json:"status"`
	Result  string     `json:"result,omitempty"`
	Failure string     `json:"failure,omitempty"`
}
