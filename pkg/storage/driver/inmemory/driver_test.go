// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package inmemory

import (
	"testing"

	storagedriver "github.com/easysoft/gitfox-community/pkg/storage/driver"
	"github.com/easysoft/gitfox-community/pkg/storage/driver/testsuites"
)

func newDriverConstructor() (storagedriver.StorageDriver, error) {
	return New(), nil
}

func TestInMemoryDriverSuite(t *testing.T) {
	testsuites.Driver(t, newDriverConstructor)
}

func BenchmarkInMemoryDriverSuite(b *testing.B) {
	testsuites.BenchDriver(b, newDriverConstructor)
}
