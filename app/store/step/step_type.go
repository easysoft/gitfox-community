package step

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

type Step struct {
	ID         uint            `gorm:"column:id;primaryKey;autoIncrement"`
	StageID    uint            `gorm:"column:stageID"`
	Name       string          `gorm:"column:name"`
	Code       string          `gorm:"column:code"`
	SN         int             `gorm:"column:sn"`
	Image      string          `gorm:"column:image"`
	Depends    string          `gorm:"column:depends"`
	GroupID    uint            `gorm:"column:groupID"`
	FailPolicy string          `gorm:"column:failPolicy"`
	Started    *datetime.Local `gorm:"column:started"`
	Finished   *datetime.Local `gorm:"column:finished"`
	Duration   uint            `gorm:"column:duration"`
	Status     string          `gorm:"column:status"`
	ErrorCode  int             `gorm:"column:errorCode"`
	ErrorLog   string          `gorm:"column:errorLog"`
}

func (s *Step) TableName() string {
	return "ops_step_executions"
}
