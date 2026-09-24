// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package notify

import (
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
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
