// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package storage

type DriverType string

const (
	DriverFilesystem DriverType = "filesystem"
)

type DriverConfig struct {
	Driver     DriverType
	Parameters map[string]interface{}
}

type Config struct {
	Local DriverConfig
	S3    DriverConfig
}
