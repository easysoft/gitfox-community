// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package database

import (
	"testing"

	"github.com/easysoft/gitfox-community/pkg/config"
)

func TestValidateDMConfigDatasource(t *testing.T) {
	t.Parallel()

	if err := validateDMConfig(config.DatabaseConfig{}); err == nil {
		t.Fatal("validateDMConfig() = nil, want error when Datasource is empty")
	}

	if err := validateDMConfig(config.DatabaseConfig{
		Datasource: "dm://SYSDBA:Zxcvbnm123@10.0.4.16:5236?schema=GITFOXCOMMUNITY",
	}); err != nil {
		t.Fatalf("validateDMConfig() = %v, want nil when Datasource is set", err)
	}
}
