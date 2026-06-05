package branch

import (
	"fmt"
	"regexp"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
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
	// branchNameForbiddenChars 分支名禁用字符集合(不含空白,空白单独检测)
	branchNameForbiddenChars = "~^:?*[\\{}@!$&()=+|<>;\"'"
)

type ReqForCreate struct {
	Name   string `json:"name" binding:"required,max=100" example:"dev"`
	Source string `json:"source"`
}

func (r *ReqForCreate) Validate() error {
	return validateBranchName(r.Name)
}

// validateBranchName 校验分支名是否符合 Git 引用命名规范及项目约束
func validateBranchName(name string) error {
	// 长度限制(Git 工具对超过 255 字符的引用名显示异常)
	if len(name) > 255 {
		return errors.WithI18nKey(fmt.Errorf("branch name too long"), i18n.BranchNameInvalidLength)
	}
	// 开头限制
	if strings.HasPrefix(name, ".") {
		return errors.WithI18nKey(fmt.Errorf("branch name cannot start with dot"), i18n.BranchNameInvalidLeadingDot)
	}
	if strings.HasPrefix(name, "-") {
		return errors.WithI18nKey(fmt.Errorf("branch name cannot start with hyphen"), i18n.BranchNameInvalidLeadingHyphen)
	}
	if strings.HasPrefix(name, "/") {
		return errors.WithI18nKey(fmt.Errorf("branch name cannot start with slash"), i18n.BranchNameInvalidLeadingSlash)
	}
	// 结尾限制
	if strings.HasSuffix(name, "/") {
		return errors.WithI18nKey(fmt.Errorf("branch name cannot end with slash"), i18n.BranchNameInvalidTrailingSlash)
	}
	if strings.HasSuffix(name, ".") {
		return errors.WithI18nKey(fmt.Errorf("branch name cannot end with dot"), i18n.BranchNameInvalidTrailingDot)
	}
	if strings.HasSuffix(name, ".lock") {
		return errors.WithI18nKey(fmt.Errorf("branch name cannot end with .lock"), i18n.BranchNameInvalidTrailingLock)
	}
	// 连续序列限制
	if strings.Contains(name, "..") {
		return errors.WithI18nKey(fmt.Errorf("branch name contains ..."), i18n.BranchNameInvalidDotDot)
	}
	if strings.Contains(name, "//") {
		return errors.WithI18nKey(fmt.Errorf("branch name contains //"), i18n.BranchNameInvalidDoubleSlash)
	}
	// 字符集合禁用(空白 + 特殊符号)
	for _, ch := range name {
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			return errors.WithI18nKey(fmt.Errorf("branch name contains whitespace"), i18n.BranchNameInvalidChars)
		}
		if strings.ContainsRune(branchNameForbiddenChars, ch) {
			return errors.WithI18nKey(fmt.Errorf("branch name contains forbidden char %q", ch), i18n.BranchNameInvalidChars)
		}
	}

	return nil
}

type ReqForDelete struct {
	Name string `json:"name" binding:"required,max=100" example:"dev"`
}

// ReqForArchive 批量归档/激活分支请求
type ReqForArchive struct {
	BranchNames []string `json:"branchNames" binding:"required,min=1" example:"dev,feature/login"`
}

// ArchiveResult 单个分支归档结果
type ArchiveResult struct {
	BranchName string `json:"branchName"`
	Status     bool   `json:"status"`
	Message    string `json:"message"`
}

// ActivateResult 单个分支激活结果
type ActivateResult struct {
	BranchName string `json:"branchName"`
	Status     bool   `json:"status"`
	Message    string `json:"message"`
}

// ReqFilter stores branch query parameters.
type ReqFilter struct {
	Query           string `json:"query" form:"query"`
	Prefix          string `json:"prefix" form:"prefix"`
	BranchTypeIDs   []uint `json:"branchTypeIDs" form:"branchTypeIDs"`
	IncludeArchived bool   `json:"includeArchived" form:"includeArchived"`
	ArchivedOnly    bool   `json:"archivedOnly" form:"archivedOnly"`
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
	Repo        uint            `json:"repo"`
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
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckNameLength)
		return err
	}

	// Key必填,以字母开头,仅能输入英文字母、数字与符号（/-_.），最多50个字符
	if len(r.Key) > 50 {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckKeyLength)
		return err
	}
	if !alphanumStartRegex.MatchString(r.Key) {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckKeyFormat)
		return err
	}

	// 前缀必填,1-5条,每条最多50个字符,仅能输入英文字母和数字,必须以分隔符(/-_.)结尾
	if len(r.Prefixes) < 1 || len(r.Prefixes) > 5 {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixNum)
		return err
	}
	for _, prefix := range r.Prefixes {
		if len(prefix) > 50 {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixLength)
			return err
		}
		if !validPrefixRegex.MatchString(prefix) {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixFormat)
			return err
		}
		if strings.HasPrefix(prefix, constant.BranchArchivedPrefix) || strings.TrimRight(prefix, "/-_.") == "archived" {
			err = errors.WithI18nKey(err, i18n.BranchTypeArchivedPrefixForbidden)
			return err
		}
	}

	// 描述选填,最多500个字符
	if len(r.Desc) > 500 {
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
	var err error
	// 名称选填,如果提供则校验长度
	if len(r.Name) > 100 {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckNameLength)
		return err
	}

	// 前缀选填,如果提供则校验,1-5条,每条最多50个字符,必须以分隔符(/-_.)结尾
	if len(r.Prefixes) > 0 {
		if len(r.Prefixes) > 5 {
			err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixNum)
			return err
		}
		for _, prefix := range r.Prefixes {
			if len(prefix) > 50 {
				err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixLength)
				return err
			}
			if !validPrefixRegex.MatchString(prefix) {
				err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckPrefixFormat)
				return err
			}
			if strings.HasPrefix(prefix, constant.BranchArchivedPrefix) || strings.TrimRight(prefix, "/-_.") == "archived" {
				err = errors.WithI18nKey(err, i18n.BranchTypeArchivedPrefixForbidden)
				return err
			}
		}
	}

	// 描述选填,如果提供则校验长度
	if len(r.Desc) > 500 {
		err = errors.WithI18nKey(err, i18n.BranchTypeBodyCheckDescLength)
		return err
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
