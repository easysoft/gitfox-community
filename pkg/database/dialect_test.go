package database

import "testing"

func TestGaussDBSQLRewriter_RewriteSQL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "纯驼峰列名",
			in:   `WHERE isCustom = 0 AND deleted = 0`,
			want: `WHERE "isCustom" = 0 AND deleted = 0`,
		},
		{
			name: "别名前缀驼峰列名",
			in:   `WHERE i.repoID = ? AND i.deleted = ?`,
			want: `WHERE i."repoID" = ? AND i.deleted = ?`,
		},
		{
			name: "已加引号的驼峰列名不重复加引号",
			in:   `SELECT * FROM "ops_scan_rules" WHERE "isCustom" = 0`,
			want: `SELECT * FROM "ops_scan_rules" WHERE "isCustom" = 0`,
		},
		{
			name: "单引号字符串值不受影响",
			in:   `WHERE name = 'isCustomValue' AND deleted = 0`,
			want: `WHERE name = 'isCustomValue' AND deleted = 0`,
		},
		{
			name: "全小写列名不动",
			in:   `WHERE name = ? AND deleted = 0`,
			want: `WHERE name = ? AND deleted = 0`,
		},
		{
			name: "SQL关键字不受影响",
			in:   `SELECT COUNT(*) FROM ops_scan_rules`,
			want: `SELECT COUNT(*) FROM ops_scan_rules`,
		},
		{
			name: "递归CTE中的驼峰列名",
			in:   `SELECT g.id, g.parentID FROM ops_artifact_groups g WHERE g.id = ?`,
			want: `SELECT g.id, g."parentID" FROM ops_artifact_groups g WHERE g.id = ?`,
		},
		{
			name: "JOIN中的驼峰列名",
			in:   `SELECT i.*, t.planID FROM ops_scan_issues i LEFT JOIN ops_scan_tasks t ON i.updatedByTaskID = t.id`,
			want: `SELECT i.*, t."planID" FROM ops_scan_issues i LEFT JOIN ops_scan_tasks t ON i."updatedByTaskID" = t.id`,
		},
		{
			name: "子查询中的驼峰列名",
			in:   `SELECT COUNT(*) FROM ops_scan_solution_rulesets WHERE solutionID = s.id AND deleted = ?`,
			want: `SELECT COUNT(*) FROM ops_scan_solution_rulesets WHERE "solutionID" = s.id AND deleted = ?`,
		},
		{
			name: "表名中的驼峰不使用双引号前缀",
			in:   `DELETE FROM ops_scan_rules WHERE isCustom = 0`,
			want: `DELETE FROM ops_scan_rules WHERE "isCustom" = 0`,
		},
		{
			name: "ORDER BY 中已引号的表名列",
			in:   `ORDER BY "ops_scan_rules"."createdDate" DESC`,
			want: `ORDER BY "ops_scan_rules"."createdDate" DESC`,
		},
	}

	rw := GaussDBSQLRewriter{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rw.RewriteSQL(tt.in)
			if got != tt.want {
				t.Errorf("RewriteSQL(%q)\n  got:  %q\n  want: %q", tt.in, got, tt.want)
			}
		})
	}
}
