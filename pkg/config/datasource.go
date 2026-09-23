// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package config

import (
	"fmt"
	"net/url"
)

// DatasourceComposer 把拆分字段按数据库驱动拼成 DSN。
type DatasourceComposer interface {
	// Compose 返回该驱动对应的连接串。
	Compose(db DatabaseConfig) string
}

// datasourceComposers 保存 driver -> DatasourceComposer 的映射，仅注册于 init 期。
var datasourceComposers = make(map[string]DatasourceComposer)

// registerDatasourceComposer 注册某数据库驱动的 DSN 拼接实现。
// 同一驱动重复注册或注册空实现会直接 panic，用于尽早暴露拼写/实现错误。
func registerDatasourceComposer(driver string, composer DatasourceComposer) {
	if composer == nil {
		panic("datasource composer must not be nil")
	}
	if _, registered := datasourceComposers[driver]; registered {
		panic(fmt.Sprintf("datasource composer %q already registered", driver))
	}
	datasourceComposers[driver] = composer
}

// composerFor 返回 driver 对应的拼接实现；未注册的驱动回退到 mysql 实现，保持默认行为。
func composerFor(driver string) DatasourceComposer {
	if composer, ok := datasourceComposers[driver]; ok {
		return composer
	}
	return mysqlDatasourceComposer{}
}

type mysqlDatasourceComposer struct{}

func (mysqlDatasourceComposer) Compose(db DatabaseConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=%s",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
		url.QueryEscape(db.Timezone))
}

type dmDatasourceComposer struct{}

// Compose 拼接达梦数据库 DSN。
// 达梦驱动对 user:pass@host:port 原样切分且不做 URL 解码，因此用户名/密码需避免包含 @ : ? 字符。
// Name 在达梦中即 schema；为空时不携带 ?schema=，使用账号默认 schema。
func (dmDatasourceComposer) Compose(db DatabaseConfig) string {
	dsn := fmt.Sprintf("dm://%s:%s@%s:%s",
		db.User,
		db.Password,
		db.Host,
		db.Port)
	if db.Name != "" {
		dsn += "?schema=" + db.Name
	}
	return dsn
}

type gaussdbDatasourceComposer struct{}

// Compose 拼接 GaussDB（openGauss 兼容 PostgreSQL 协议）DSN，使用 key=value 形式。
func (gaussdbDatasourceComposer) Compose(db DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name)
}

func init() {
	registerDatasourceComposer("mysql", mysqlDatasourceComposer{})
	registerDatasourceComposer("dm", dmDatasourceComposer{})
	registerDatasourceComposer("gaussdb", gaussdbDatasourceComposer{})
}
