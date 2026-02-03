package stage

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type Stage struct {
	ID           uint            `gorm:"column:id;primaryKey;autoIncrement"`
	RepoID       uint            `gorm:"column:repoID;not null;default:0"`
	ExecutionID  uint            `gorm:"column:executionID;not null"`
	Parent       uint            `gorm:"column:parent;not null;default:0"`
	Name         string          `gorm:"column:name;not null;default:''"`
	Code         string          `gorm:"column:code;not null"`
	SN           int             `gorm:"column:sn;not null"`
	Type         string          `gorm:"column:type;not null;default:''"`
	Host         string          `gorm:"column:host;not null;default:''"`
	OS           string          `gorm:"column:os;not null;default:''"`
	Arch         string          `gorm:"column:arch;not null;default:''"`
	Labels       string          `gorm:"column:labels;not null;default:''"`
	Depends      string          `gorm:"column:depends;not null;default:''"`
	RunPolicy    string          `gorm:"column:runPolicy;not null;default:'onSuccess'"`
	CustomPolicy string          `gorm:"column:customPolicy;not null;default:''"`
	FailPolicy   string          `gorm:"column:failPolicy;not null;default:'fail'"`
	Started      *datetime.Local `gorm:"column:started"`
	Finished     *datetime.Local `gorm:"column:finished"`
	Duration     uint            `gorm:"column:duration"`
	Status       string          `gorm:"column:status;not null"`
	ErrorType    string          `gorm:"column:errorType;not null"`
	ErrorCode    int             `gorm:"column:errorCode;not null"`
	ErrorLog     string          `gorm:"column:errorLog;type:text;not null"`
}

func (s *Stage) TableName() string {
	return "ops_stage_executions"
}
