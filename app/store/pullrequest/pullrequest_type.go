package pullrequest

import (
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

// PullRequest 合并请求表（ops_ppmrequests）Store层结构体
type PullRequest struct {
	ID                uint            `gorm:"column:id;primaryKey;autoIncrement;comment:'自增主键'"`
	Title             string          `gorm:"column:title;not null;comment:'PR 标题'"`
	Desc              string          `gorm:"column:desc;type:text;not null;comment:'拉取请求详细描述'"`
	RepoID            uint            `gorm:"column:repoID;not null;comment:'关联zt_repo表的id'"`
	SourceRepoID      uint            `gorm:"column:sourceRepoID;not null;comment:'源仓库ID'"`
	SourceBranch      string          `gorm:"column:sourceBranch;not null;comment:'源分支名称'"`
	SourceSHA         string          `gorm:"column:sourceSHA;type:char(40);not null;comment:'源分支最新提交的SHA-1值'"`
	TargetRepoID      uint            `gorm:"column:targetRepoID;not null;comment:'目标仓库ID'"`
	TargetBranch      string          `gorm:"column:targetBranch;not null;comment:'目标分支名称'"`
	MergeDate         *datetime.Local `gorm:"column:mergeDate;comment:'合并时间'"`
	MergedBy          *string         `gorm:"column:mergedBy;comment:'执行合并操作的用户ID'"`
	MergeMethod       string          `gorm:"column:mergeMethod;comment:'合并方法'"`
	MergeTargetSHA    string          `gorm:"column:mergeTargetSHA;type:char(40);comment:'合并时目标分支的SHA'"`
	MergeBaseSHA      string          `gorm:"column:mergeBaseSHA;type:char(40);not null;default:'';comment:'合并基准SHA'"`
	MergeSHA          string          `gorm:"column:mergeSHA;type:char(40);comment:'合并后生成的commit SHA'"`
	MergeCheckStatus  string          `gorm:"column:mergeCheckStatus;not null;comment:'合并前检查状态'"`
	MergeConflicts    StringSlice     `gorm:"column:mergeConflicts;type:text;comment:'冲突文件列表'"`
	RebaseCheckStatus string          `gorm:"column:rebaseCheckStatus;not null;default:'unchecked';comment:'变基检查状态'"`
	RebaseConflicts   StringSlice     `gorm:"column:rebaseConflicts;type:text;comment:'变基冲突详情'"`
	CommitCount       *uint           `gorm:"column:commitCount;comment:'PR包含的commit数量'"`
	FileCount         *uint           `gorm:"column:fileCount;comment:'PR修改的文件数量'"`
	Flow              int8            `gorm:"column:flow;not null;default:0;comment:'流程标识'"`
	ReviewFlowID      uint            `gorm:"column:reviewFlowID;not null;default:0;comment:'关联的评审流程ID'"`
	Additions         *uint           `gorm:"column:additions;comment:'新增的代码行数'"`
	Deletions         *uint           `gorm:"column:deletions;comment:'删除的代码行数'"`
	Status            string          `gorm:"column:status;not null;comment:'拉取请求状态'"`
	CreatedBy         string          `gorm:"column:createdBy;not null;default:'';comment:'创建人'"`
	CreatedDate       *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy          string          `gorm:"column:editedBy;not null;default:'';comment:'更新人'"`
	EditedDate        *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
}

// TableName 指定数据库表名
func (p *PullRequest) TableName() string {
	return "ops_ppmrequests"
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
