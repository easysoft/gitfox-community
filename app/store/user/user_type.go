package user

import (
	"time"

	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
)

type User struct {
	ID           int64      `db:"id" gorm:"column:id;primaryKey"`
	Company      uint       `db:"company" gorm:"column:company"`
	Type         string     `db:"type" gorm:"column:type"`
	Dept         uint       `db:"dept" gorm:"column:dept"`
	Account      string     `db:"account" gorm:"column:account"`
	Password     string     `db:"password" gorm:"column:password"`
	Role         string     `db:"role" gorm:"column:role"`
	Realname     string     `db:"realname" gorm:"column:realname"`
	Superior     *string    `db:"superior" gorm:"column:superior"`
	Pinyin       string     `db:"pinyin" gorm:"column:pinyin"`
	Nickname     string     `db:"nickname" gorm:"column:nickname"`
	Commiter     string     `db:"commiter" gorm:"column:commiter"`
	Avatar       *string    `db:"avatar" gorm:"column:avatar"`
	Birthday     *time.Time `db:"birthday" gorm:"column:birthday"`
	Gender       string     `db:"gender" gorm:"column:gender"`
	Email        string     `db:"email" gorm:"column:email"`
	Skype        string     `db:"skype" gorm:"column:skype"`
	QQ           string     `db:"qq" gorm:"column:qq"`
	Mobile       string     `db:"mobile" gorm:"column:mobile"`
	Phone        string     `db:"phone" gorm:"column:phone"`
	Weixin       string     `db:"weixin" gorm:"column:weixin"`
	Dingding     string     `db:"dingding" gorm:"column:dingding"`
	Slack        string     `db:"slack" gorm:"column:slack"`
	Whatsapp     string     `db:"whatsapp" gorm:"column:whatsapp"`
	Address      string     `db:"address" gorm:"column:address"`
	Zipcode      string     `db:"zipcode" gorm:"column:zipcode"`
	Nature       *string    `db:"nature" gorm:"column:nature"`
	Analysis     *string    `db:"analysis" gorm:"column:analysis"`
	Strategy     *string    `db:"strategy" gorm:"column:strategy"`
	Join         *time.Time `db:"join" gorm:"column:join"`
	Visits       uint       `db:"visits" gorm:"column:visits"`
	Visions      string     `db:"visions" gorm:"column:visions"`
	IP           string     `db:"ip" gorm:"column:ip"`
	Last         *time.Time `db:"last" gorm:"column:last"`
	Fails        int8       `db:"fails" gorm:"column:fails"`
	Locked       *time.Time `db:"locked" gorm:"column:locked"`
	Feedback     string     `db:"feedback" gorm:"column:feedback"`
	Ranzhi       string     `db:"ranzhi" gorm:"column:ranzhi"`
	LDAP         string     `db:"ldap" gorm:"column:ldap"`
	Score        int        `db:"score" gorm:"column:score"`
	ScoreLevel   int        `db:"scoreLevel" gorm:"column:scoreLevel"`
	ResetToken   string     `db:"resetToken" gorm:"column:resetToken"`
	Deleted      string     `db:"deleted" gorm:"column:deleted"`
	ClientStatus string     `db:"clientStatus" gorm:"column:clientStatus"`
	ClientLang   string     `db:"clientLang" gorm:"column:clientLang"`
}

func (p *User) TableName() string {
	return "zt_user"
}

func (p *User) toUser() *principal.User {
	return &principal.User{
		ID:       p.ID,
		Account:  p.Account,
		Email:    p.Email,
		Password: p.Password,
		IsAdmin:  p.Role == "admin",
	}
}
