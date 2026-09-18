package pipeline

import (
	"database/sql"

	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

type Pipeline struct {
	ID               int64           `gorm:"column:id;primaryKey;autoIncrement;comment:'流水线ID，自增主键'"`
	Name             string          `gorm:"column:name;not null;comment:'流水线名称'"`
	Engine           string          `gorm:"column:engine;not null;default:'gitfox';comment:'引擎: gitlab, jenkins, gitfox'"`
	ProviderID       int64           `gorm:"column:providerID;not null;default:0;comment:'关联的外部服务ID'"`
	Scope            string          `gorm:"column:scope;not null;comment:'流水线类型：系统/空间/仓库'"`
	SpaceID          int64           `gorm:"column:spaceID;not null;comment:'关联的空间ID'"`
	RepoID           int64           `gorm:"column:repoID;not null;comment:'关联的代码仓库ID'"`
	Desc             string          `gorm:"column:desc;not null;default:'';comment:'流水线描述信息'"`
	Status           string          `gorm:"column:status;not null;comment:'流水线状态：草稿/激活'"`
	LatestVersion    int             `gorm:"column:latestVersion;not null;default:0;comment:'流水线最新版本'"`
	DefaultBranch    string          `gorm:"column:defaultBranch;not null;comment:'流水线默认分支'"`
	YamlPath         string          `gorm:"column:yamlPath;not null;default:'';comment:'流水线配置文件路径'"`
	CustomParam      string          `gorm:"column:customParam;type:text;default:null;comment:'流水线执行传参'"`
	ExternalPipeline string          `gorm:"column:externalPipeline;not null;default:'';comment:'Jenkins/GitLab pipeline'"`
	LastExec         *datetime.Local `gorm:"column:lastExec;comment:'最后一次执行时间'"`
	LastResult       string          `gorm:"column:lastResult;not null;default:'';comment:'最后一次执行状态'"`
	CreatedBy        string          `gorm:"column:createdBy;not null;default:'';comment:'创建人'"`
	CreatedDate      *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy         string          `gorm:"column:editedBy;not null;default:'';comment:'更新人'"`
	EditedDate       *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
	Deleted          uint8           `gorm:"column:deleted;not null;default:0;comment:'删除标识，0-未删除，1-已删除'"`
}

func (p *Pipeline) TableName() string {
	return "ops_pipeline"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (p *Pipeline) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (p *Pipeline) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.EditedDate = now
	return nil
}

type PipelineContent struct {
	ID          int64           `gorm:"column:id;primaryKey;autoIncrement;comment:'自增主键'"`
	PipelineID  int64           `gorm:"column:pipelineID;not null;comment:'关联的流水线ID'"`
	Version     int             `gorm:"column:version;not null;default:0;comment:'流水线版本'"`
	Data        []byte          `gorm:"column:data"`
	Variables   sql.NullString  `gorm:"column:variables"`
	CreatedBy   string          `gorm:"column:createdBy;not null;default:'';comment:'创建人'"`
	CreatedDate *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy    string          `gorm:"column:editedBy;not null;default:'';comment:'更新人'"`
	EditedDate  *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
}

func (p *PipelineContent) TableName() string {
	return "ops_pipeline_content"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (p *PipelineContent) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.CreatedDate = now
	p.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (p *PipelineContent) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	p.EditedDate = now
	return nil
}
