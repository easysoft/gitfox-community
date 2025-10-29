package space

import "time"

// 定义 Acl 字段的可选值常量（规范业务使用）
const (
	AclPrivate = "private" // 私有权限
	AclOpen    = "open"    // 公开权限
)

// Space 空间模型
type Space struct {
	ID          uint16     `gorm:"column:id;type:smallint unsigned;primaryKey;autoIncrement;comment:'ID'" json:"id"`
	Name        string     `gorm:"column:name;type:varchar(200);not null;uniqueIndex:uk_name;comment:'名称'" json:"name"`
	Owner       string     `gorm:"column:owner;type:varchar(30);not null;default:'';comment:'负责人'" json:"owner"`
	Permission  string     `gorm:"column:acl;type:varchar(30);not null;default:'open';comment:'权限:private-私有,open-公开'" json:"acl"`
	Desc        string     `gorm:"column:desc;type:text;null;default:null;comment:'描述'" json:"desc,omitempty"` // 允许为null
	CreatedBy   string     `gorm:"column:createdBy;type:varchar(30);not null;default:'';comment:'创建人'" json:"createdBy"`
	CreatedDate *time.Time `gorm:"column:createdDate;type:datetime;null;comment:'创建时间'" json:"createdDate,omitempty"`
	EditedBy    string     `gorm:"column:editedBy;type:varchar(30);not null;default:'';comment:'修改人'" json:"editedBy"`
	EditedDate  *time.Time `gorm:"column:editedDate;type:datetime;null;comment:'更新时间'" json:"editedDate,omitempty"`
	Deleted     int8       `gorm:"column:deleted;type:tinyint;not null;default:0;comment:'删除标识，0-未删除，1-已删除'" json:"deleted"`
}

// TableName 显式指定结构体对应的数据表名为 ops_space
func (Space) TableName() string {
	return "ops_space"
}

