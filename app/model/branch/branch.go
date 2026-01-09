package branch

import (
	"regexp"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

var (
	// alphanumStartRegex 以字母开头,可包含字母、数字、/-_.
	alphanumStartRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9/_.-]*$`)
	// validPrefixRegex 前缀只能包含字母和数字,必须以分隔符(/-_.)结尾
	validPrefixRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*[/._-]$`)
)

type ReqForCreate struct {
	Name   string `json:"name" binding:"required,max=100" example:"dev"`
	Source string `json:"source"`
}

type ReqForDelete struct {
	Name string `json:"name" binding:"required,max=100" example:"dev"`
}

// ReqFilter stores branch query parameters.
type ReqFilter struct {
	Query           string `json:"query" form:"query"`
	Prefix          string `json:"prefix" form:"prefix"`
	IncludeArchived bool   `json:"includeArchived" form:"includeArchived"`
	IncludeCommit   bool   `json:"includeCommit" form:"includeCommit"`
	MaxDivergence   uint   `json:"maxDivergence" form:"maxDivergence"`
	Sort            string `json:"sort" form:"sort" default:"name"`
	Order           string `json:"order" form:"order" default:"asc"`
	Page            int    `json:"page" form:"page" default:"1"`
	PageSize        int    `json:"pageSize" form:"pageSize" default:"20"`
}

type Prefix struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

type Branch struct {
	Name             string                   `json:"name"`
	SHA              string                   `json:"sha"`
	Commit           *commit.Commit           `json:"commit,omitempty"`
	IsDefault        bool                     `json:"isDefault"`
	CommitDivergence *commit.CommitDivergence `json:"commitDivergence,omitempty"`
}

// BranchType 分支类型请求/响应模型
type BranchType struct {
	ID          uint            `json:"id"`
	RepoID      uint            `json:"repo"`
	Name        string          `json:"name"`
	Key         string          `json:"key"`
	Prefixes    []string        `json:"prefixes"`
	Desc        string          `json:"desc"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

// ReqForCreateBranchType 创建分支类型请求
type ReqForCreateBranchType struct {
	Name     string   `json:"name" binding:"required,max=100" example:"功能分支"`
	Key      string   `json:"key" binding:"required,max=50" example:"feature"`
	Prefixes []string `json:"prefixes" binding:"required,min=1,max=5,unique"`
	Desc     string   `json:"desc" binding:"max=500" example:"用于开发新功能的分支"`
}

func (r *ReqForCreateBranchType) Check() error {
	var err error
	if len(r.Name) > 100 {
		err = errors.New("name length must be less than 100")
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckNameLength)
		return err
	}

	// Key必填,以字母开头,仅能输入英文字母、数字与符号（/-_.），最多50个字符
	if len(r.Key) > 50 {
		err = errors.New("key length must be less than 50")
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckKeyLength)
		return err
	}
	if !alphanumStartRegex.MatchString(r.Key) {
		err = errors.New("key must start with a letter and can only contain letters, numbers, and symbols (/-_.)")
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckKeyFormat)
		return err
	}

	// 前缀必填,1-5条,每条最多50个字符,仅能输入英文字母和数字,必须以分隔符(/-_.)结尾
	if len(r.Prefixes) < 1 || len(r.Prefixes) > 5 {
		err = errors.New("prefixes length must be between 1 and 5")
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixNum)
		return err
	}
	for _, prefix := range r.Prefixes {
		if len(prefix) > 50 {
			err = errors.New("prefix length must be less than 50")
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixLength)
			return err
		}
		if !validPrefixRegex.MatchString(prefix) {
			err = errors.New("prefix can only contain letters and numbers, and must end with a separator (/-_.)")
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixFormat)
			return err
		}
	}

	// 描述选填,最多500个字符
	if len(r.Desc) > 500 {
		err = errors.New("desc length must be less than 500")
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckDescLength)
		return err
	}

	return nil
}

// ReqForUpdateBranchType 更新分支类型请求
type ReqForUpdateBranchType struct {
	Name     string   `json:"name" binding:"required,max=100" example:"功能分支"`
	Prefixes []string `json:"prefixes" binding:"required,min=1,max=5,unique"`
	Desc     string   `json:"desc" binding:"omitempty,max=500" example:"用于开发新功能的分支"`
}

func (r *ReqForUpdateBranchType) Check() error {
	// 名称选填,如果提供则校验长度
	if len(r.Name) > 100 {
		return errors.New("name length must be less than 100")
	}

	// 前缀选填,如果提供则校验,1-5条,每条最多50个字符,必须以分隔符(/-_.)结尾
	if len(r.Prefixes) > 0 {
		if len(r.Prefixes) > 5 {
			return errors.New("prefixes length must be less than 5")
		}
		for _, prefix := range r.Prefixes {
			if len(prefix) > 50 {
				return errors.New("prefix length must be less than 50")
			}
			if !validPrefixRegex.MatchString(prefix) {
				return errors.New("prefix can only contain letters and numbers, and must end with a separator (/-_.)")
			}
		}
	}

	// 描述选填,如果提供则校验长度
	if len(r.Desc) > 500 {
		return errors.New("desc length must be less than 500")
	}

	return nil
}

// ReqFilterBranchType 分支类型查询过滤器
type ReqFilterBranchType struct {
	Name     string `json:"name" form:"name"`
	Key      string `json:"key" form:"key"`
	Prefix   string `json:"prefix" form:"prefix"`
	Desc     string `json:"desc" form:"desc"`
	Page     int    `json:"page" form:"page" default:"1"`
	PageSize int    `json:"pageSize" form:"pageSize" default:"20"`
}
