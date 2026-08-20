package pullrequest

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// StringSlice 自定义字符串切片类型，用于处理数据库中的换行符分隔字符串
type StringSlice []string

// Scan implements the Scanner interface.
func (ss *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*ss = nil
		return nil
	}
	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return errors.New("failed to scan StringSlice: invalid value type")
	}
	if str == "" {
		*ss = []string{}
		return nil
	}
	// Normalize line endings
	str = strings.ReplaceAll(str, "\r\n", "\n")
	*ss = strings.Split(str, "\n")
	return nil
}

// Value implements the driver Valuer interface.
func (ss StringSlice) Value() (driver.Value, error) {
	if len(ss) == 0 {
		return "", nil
	}
	return strings.Join(ss, "\n"), nil
}

// PullRequest 合并请求表（ops_ppm）Store层结构体
type PullRequest struct {
	ID                uint            `gorm:"column:id;primaryKey;autoIncrement;comment:'自增主键'"`
	Title             string          `gorm:"column:title"`
	Desc              sql.NullString  `gorm:"column:desc"`
	RepoID            uint            `gorm:"column:repoID"`
	SourceRepoID      uint            `gorm:"column:sourceRepoID"`
	SourceBranch      string          `gorm:"column:sourceBranch"`
	SourceSHA         string          `gorm:"column:sourceSHA"`
	TargetRepoID      uint            `gorm:"column:targetRepoID"`
	TargetBranch      string          `gorm:"column:targetBranch"`
	MergedDate        *datetime.Local `gorm:"column:mergedDate"`
	MergedBy          string          `gorm:"column:mergedBy"`
	MergeMethod       string          `gorm:"column:mergeMethod"`
	MergeTargetSHA    string          `gorm:"column:mergeTargetSHA"`
	MergeBaseSHA      string          `gorm:"column:mergeBaseSHA"`
	MergeSHA          string          `gorm:"column:mergeSHA"`
	MergeCheckStatus  string          `gorm:"column:mergeCheckStatus"`
	MergeConflicts    StringSlice     `gorm:"column:mergeConflicts"`
	RebaseCheckStatus string          `gorm:"column:rebaseCheckStatus"`
	RebaseConflicts   StringSlice     `gorm:"column:rebaseConflicts"`
	CommitCount       uint            `gorm:"column:commitCount"`
	FileCount         uint            `gorm:"column:fileCount"`
	Flow              int8            `gorm:"column:flow"`
	ReviewFlowID      uint            `gorm:"column:reviewFlowID"`
	Additions         uint            `gorm:"column:additions"`
	Deletions         uint            `gorm:"column:deletions"`
	Status            string          `gorm:"column:status"`
	CreatedBy         string          `gorm:"column:createdBy"`
	CreatedDate       *datetime.Local `gorm:"column:createdDate"`
	EditedBy          string          `gorm:"column:editedBy"`
	EditedDate        *datetime.Local `gorm:"column:editedDate"`
}

// TableName 指定数据库表名
func (p *PullRequest) TableName() string {
	return "ops_ppm"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (p *PullRequest) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (p *PullRequest) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.EditedDate = now
	return nil
}

// RequestReviewer 评审人员表（ops_request_reviewers）Store层结构体
type RequestReviewer struct {
	ID          uint            `gorm:"column:id;primaryKey;autoIncrement;comment:'ID'"`
	RequestID   uint            `gorm:"column:requestID;comment:'关联合并请求ID'"`
	RepoID      uint            `gorm:"column:repoID;comment:'关联仓库ID'"`
	Decision    string          `gorm:"column:decision;comment:'最新审核决策'"`
	SHA         string          `gorm:"column:sha;comment:'审核对应的代码提交SHA校验值'"`
	Account     string          `gorm:"column:account;comment:'评审人'"`
	Opinion     sql.NullString  `gorm:"column:opinion;comment:'评审意见'"`
	CreatedBy   string          `gorm:"column:createdBy;comment:'创建人'"`
	CreatedDate *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedDate  *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
}

// TableName 指定数据库表名
func (r *RequestReviewer) TableName() string {
	return "ops_request_reviewers"
}
