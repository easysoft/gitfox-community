// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package config

import "testing"

func TestComposeDatasource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  *Config
		want string
	}{
		{
			name: "mysql keeps existing tcp dsn",
			cfg: &Config{
				Database: DatabaseConfig{
					Driver:   "mysql",
					Host:     "cy.oop.cc",
					Port:     "13306",
					User:     "root",
					Password: "123456",
					Name:     "zentaoipd",
					Timezone: "Local",
				},
			},
			want: "root:123456@tcp(cy.oop.cc:13306)/zentaoipd?parseTime=true&loc=Local",
		},
		{
			name: "dm with schema name",
			cfg: &Config{
				Database: DatabaseConfig{
					Driver:   "dm",
					Host:     "10.0.4.16",
					Port:     "5236",
					User:     "SYSDBA",
					Password: "Zxcvbnm123",
					Name:     "GITFOXCOMMUNITY",
				},
			},
			want: "dm://SYSDBA:Zxcvbnm123@10.0.4.16:5236?schema=GITFOXCOMMUNITY",
		},
		{
			name: "dm without schema name uses default user schema",
			cfg: &Config{
				Database: DatabaseConfig{
					Driver:   "dm",
					Host:     "10.0.4.16",
					Port:     "5236",
					User:     "SYSDBA",
					Password: "Zxcvbnm123",
				},
			},
			want: "dm://SYSDBA:Zxcvbnm123@10.0.4.16:5236",
		},
		{
			name: "gaussdb uses key=value dsn",
			cfg: &Config{
				Database: DatabaseConfig{
					Driver:   "gaussdb",
					Host:     "10.0.4.16",
					Port:     "5432",
					User:     "gaussdb",
					Password: "Zxcvbnm123",
					Name:     "gitfox",
				},
			},
			want: "host=10.0.4.16 port=5432 user=gaussdb password=Zxcvbnm123 dbname=gitfox sslmode=disable",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := composeDatasource(tt.cfg)
			if got != tt.want {
				t.Fatalf("composeDatasource() = %q, want %q", got, tt.want)
			}
		})
	}
}
