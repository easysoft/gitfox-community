package stage

import (
	"database/sql"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type Stage struct {
	ID           uint            `gorm:"column:id;primaryKey;autoIncrement"`
	RepoID       uint            `gorm:"column:repoID"`
	ExecutionID  uint            `gorm:"column:executionID"`
	Parent       uint            `gorm:"column:parent"`
	Name         string          `gorm:"column:name"`
	Code         string          `gorm:"column:code"`
	SN           int             `gorm:"column:sn"`
	Type         string          `gorm:"column:type"`
	Host         string          `gorm:"column:host"`
	OS           string          `gorm:"column:os"`
	Arch         string          `gorm:"column:arch"`
	Labels       string          `gorm:"column:labels"`
	Depends      string          `gorm:"column:depends"`
	RunPolicy    string          `gorm:"column:runPolicy"`
	CustomPolicy string          `gorm:"column:customPolicy"`
	FailPolicy   string          `gorm:"column:failPolicy"`
	Started      *datetime.Local `gorm:"column:started"`
	Finished     *datetime.Local `gorm:"column:finished"`
	Duration     uint            `gorm:"column:duration"`
	Status       string          `gorm:"column:status"`
	ErrorType    string          `gorm:"column:errorType"`
	ErrorCode    int             `gorm:"column:errorCode"`
	ErrorLog     sql.NullString  `gorm:"column:errorLog"`
}

func (s *Stage) TableName() string {
	return "ops_stage_executions"
}
