package notify

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// Notify 对应数据库表 zt_notify
type Notify struct {
	ID          uint            `gorm:"column:id;primaryKey;autoIncrement"`
	ObjectType  string          `gorm:"column:objectType;not null;default:''"`
	ObjectID    uint            `gorm:"column:objectID;not null;default:0"`
	Action      uint            `gorm:"column:action;not null;default:0"`
	ToList      *string         `gorm:"column:toList"`
	CcList      *string         `gorm:"column:ccList"`
	Subject     *string         `gorm:"column:subject"`
	Data        *string         `gorm:"column:data"`
	CreatedBy   string          `gorm:"column:createdBy;not null;default:''"`
	CreatedDate *datetime.Local `gorm:"column:createdDate"`
	SendTime    *datetime.Local `gorm:"column:sendTime"`
	Status      string          `gorm:"column:status;not null;default:'wait'"`
	FailReason  *string         `gorm:"column:failReason"`
}

func (n *Notify) TableName() string {
	return "zt_notify"
}

func (n *Notify) BeforeCreate(tx *gorm.DB) error {
	if n.CreatedDate == nil {
		n.CreatedDate = datetime.NowLocalTime()
	}
	if n.Status == "" {
		n.Status = "wait"
	}
	return nil
}
