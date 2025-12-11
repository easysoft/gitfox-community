package repo

import (
	"database/sql/driver"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/gorm"
)

// 先复用之前定义的代码库状态枚举（与表中status字段对应）
type RepoState int

const (
	RepoStateActive            RepoState = iota // 0-正常活跃
	RepoStateGitImport                          // 1-Git导入中
	RepoStateMigrateGitPush                     // 2-迁移-Git推送中
	RepoStateMigrateDataImport                  // 3-迁移-数据导入中
)

// Repository 代码库表（ops_repository）Store层结构体
type Repository struct {
	ID            uint       `gorm:"column:id;comment:'代码库ID'"`
	SpaceID       uint       `gorm:"column:spaceID;comment:'空间ID'"`
	Identifier    string     `gorm:"column:identifier;comment:'代码库唯一标识'"`
	Desc          string     `gorm:"column:desc;comment:'代码库描述'"` // 关键字兼容：映射数据库 desc 字段
	GitUID        string     `gorm:"column:gitUID;comment:'Git仓库唯一标识'"`
	DefaultBranch string     `gorm:"column:defaultBranch;comment:'默认分支名'"`
	ForkID        *uint      `gorm:"column:forkID;comment:'派生来源代码库ID'"` // 指针类型：兼容数据库 NULL
	Empty         bool       `gorm:"column:empty;comment:'是否为空库（0=非空，1=空库）'"`
	ACL           string     `gorm:"column:acl;comment:'权限类型（private=私有，open=公开）'"`
	Mirror        bool       `gorm:"column:mirror;comment:'是否为镜像库（0=普通库，1=镜像库）'"`
	Status        int        `gorm:"column:status;comment:'代码库状态（0=正常，1=Git导入中，2=迁移Git推送中，3=迁移数据导入中）'"`
	Mutex         int        `gorm:"column:mutex;comment:'乐观锁版本号'"`
	CreatedBy     string     `gorm:"column:createdBy;comment:'创建人'"`
	CreatedDate   *datetime.Local `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy      string     `gorm:"column:editedBy;comment:'更新人'"`
	EditedDate    *datetime.Local `gorm:"column:editedDate;comment:'更新时间'"`
	Deleted       uint8      `gorm:"column:deleted;comment:'删除标识（0=未删除，1=已删除）'"`
}

// TableName 指定数据库表名（GORM自动映射时使用）
func (r *Repository) TableName() string {
	return "ops_repository"
}

func (r *Repository) BeforeCreate(tx *gorm.DB) error {
	now:= datetime.NowLocalTime()
	r.CreatedDate = now
	r.EditedDate = now
	return nil
}

// BeforeUpdate GORM 钩子：更新前自动设置 EditedDate
func (r *Repository) BeforeUpdate(tx *gorm.DB) error {
	now:= datetime.NowLocalTime()
	r.EditedDate = now
	return nil
}

// Scan 实现 Scan 方法（数据库→Go：tinyint → RepoState）
func (s *RepoState) Scan(value interface{}) error {
	val, ok := value.(int64)
	if !ok {
		return fmt.Errorf("invalid RepoState value type: %T", value)
	}
	*s = RepoState(val)
	return nil
}

func (s RepoState) Value() (driver.Value, error) {
	return int64(s), nil
}
