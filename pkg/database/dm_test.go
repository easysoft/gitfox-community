package database

import (
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
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
