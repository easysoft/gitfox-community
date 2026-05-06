package execution

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

type Execution struct {
	ID          uint            `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID  uint            `gorm:"column:pipelineID;not null"`
	Parent      uint            `gorm:"column:parent;not null;default:0"`
	Trigger     string          `gorm:"column:trigger;not null"`
	Commit       string          `gorm:"column:commit;not null;default:''"`
	CommitBefore string          `gorm:"column:commitBefore;not null;default:''"`
	Ref         string          `gorm:"column:ref;not null;default:''"`
	Params      string          `gorm:"column:params;type:text;not null;default:''"`
	Started     *datetime.Local `gorm:"column:started"`
	Finished    *datetime.Local `gorm:"column:finished"`
	Duration    uint            `gorm:"column:duration;not null;default:0"`
	Status      string          `gorm:"column:status;not null;default:''"`
	Error       string          `gorm:"column:error;not null;default:''"`
	CreatedBy   string          `gorm:"column:createdBy;not null"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	EditedDate  *datetime.Local `gorm:"column:editedDate"`
}

func (e *Execution) TableName() string {
	return "ops_pipeline_executions"
}

// BeforeCreate GORM 钩子：创建前自动设置时间
func (e *Execution) BeforeCreate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	e.CreatedDate = now
	e.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (e *Execution) BeforeUpdate(tx *gorm.DB) error {
	now := datetime.NowLocalTime()
	e.EditedDate = now
	return nil
}
