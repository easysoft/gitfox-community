// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package queue

import (
	"database/sql"
	"time"
)

const rfc3339Milli = "2006-01-02T15:04:05.000Z07:00"

type SQLFlavor int

const (
	SQLFlavorSQLite SQLFlavor = iota
	SQLFlavorPostgreSQL
	SQLFlavorMYSQL
)

type NewOpts struct {
	DB *sql.DB
	// 消息的最大接收次数。
	// 如果你希望重试2次，应该设置 MaxReceive = 3
	// 1次正常处理 + 2次重试 = 3次总接收
	MaxReceive int
	Name       string
	SQLFlavor  SQLFlavor
	// 在消息可被其他消费者重新接收之前，每个消费者的处理时间限制。
	// 如果你希望每个消费者最多处理10秒，应该设置 Timeout = 10 * time.Second
	Timeout time.Duration
}

type Queue struct {
	db         *sql.DB
	flavor     SQLFlavor
	maxReceive int
	name       string
	timeout    time.Duration
}

type ID string

type Message struct {
	ID       ID
	Body     []byte
	Delay    time.Duration
	Priority int // 首先接收更高优先级的消息
}
