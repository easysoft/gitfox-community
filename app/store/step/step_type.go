package step

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type Step struct {
	ID         uint            `gorm:"column:id;primaryKey;autoIncrement"`
	StageID    uint            `gorm:"column:stageID;not null"`
	Name       string          `gorm:"column:name;not null;default:''"`
	Code       string          `gorm:"column:code;not null"`
	SN         int             `gorm:"column:sn;not null"`
	Image      string          `gorm:"column:image;not null;default:''"`
	Depends    string          `gorm:"column:depends;not null;default:''"`
	GroupID    uint            `gorm:"column:groupID;not null;default:0"`
	FailPolicy string          `gorm:"column:failPolicy;not null;default:'fail'"`
	Started    *datetime.Local `gorm:"column:started;not null"`
	Finished   *datetime.Local `gorm:"column:finished;not null"`
	Duration   uint            `gorm:"column:duration"`
	Status     string          `gorm:"column:status;not null"`
	ErrorCode  int             `gorm:"column:errorCode;not null"`
	ErrorLog   string          `gorm:"column:errorLog;not null"`
}

func (s *Step) TableName() string {
	return "ops_step_executions"
}
