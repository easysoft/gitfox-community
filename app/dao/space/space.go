package space

import "time"

// Space 空间模型
type Space struct {
	ID          int64     `gorm:"column:id;type:bigint;primaryKey;autoIncrement;comment:ID"` // 主键ID，自增
	Version     int       `gorm:"column:version;type:int;not null;default:0;comment:版本"`    // 版本号，默认0
	Name        string    `gorm:"column:name;type:varchar(255);not null;comment:名称"`       // 名称，非空
	Description string    `gorm:"column:description;type:text;default:null;comment:描述"`    // 描述，可空
	CreatedBy   string    `gorm:"column:createdBy;type:varchar(255);not null;comment:创建者"` // 创建者，非空
	Created     time.Time `gorm:"column:created;type:datetime;not null;default:current_timestamp;comment:创建时间"` // 创建时间，默认当前时间
	Updated     time.Time `gorm:"column:updated;type:datetime;not null;default:current_timestamp on update current_timestamp;autoUpdateTime;comment:更新时间"` // 更新时间，自动更新
	Deleted     int       `gorm:"column:deleted;type:tinyint;not null;default:0;comment:删除标识"` // 删除标识（0-未删除，1-已删除）
}

// TableName 显式指定结构体对应的数据表名为 ops_space
func (Space) TableName() string {
	return "ops_space"
}