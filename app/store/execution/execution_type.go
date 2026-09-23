// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package execution

import (
	"database/sql"

	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

type Execution struct {
	ID           uint            `gorm:"column:id;primaryKey;autoIncrement"`
	PipelineID   uint            `gorm:"column:pipelineID"`
	Parent       uint            `gorm:"column:parent"`
	Trigger      string          `gorm:"column:trigger"`
	Commit       string          `gorm:"column:commit"`
	CommitBefore string          `gorm:"column:commitBefore"`
	Ref          string          `gorm:"column:ref"`
	Params       sql.NullString  `gorm:"column:params"`
	StartedDate  *datetime.Local `gorm:"column:startedDate"`
	FinishedDate *datetime.Local `gorm:"column:finishedDate"`
	Duration     uint            `gorm:"column:duration"`
	Status       string          `gorm:"column:status"`
	Error        string          `gorm:"column:error"`
	Number       uint            `gorm:"column:number"`
	Logs         string          `gorm:"column:logs"`
	CreatedBy    string          `gorm:"column:createdBy"`
	CreatedDate  *datetime.Local `gorm:"column:createdDate"`
	EditedDate   *datetime.Local `gorm:"column:editedDate"`
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
