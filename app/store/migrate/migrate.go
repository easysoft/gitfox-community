// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/migrate"
)

//go:embed sqlite/*.sql
var sqlite embed.FS

//go:embed mysql/*.sql
var mysql embed.FS

//go:embed dm/*.sql
var dm embed.FS

const (
	tableName = "ops_migrations"

	sqliteDriverName = "sqlite3"
	sqliteSourceDir  = "sqlite"

	mysqlDriverName = "mysql"
	mysqlSourceDir  = "mysql"

	dmDriverName = "dm"
	dmSourceDir  = "dm"
)

// Migrate performs the database migration.
func Migrate(ctx context.Context, db *sql.DB, driverName string, log *logger.SugaredMateLogger) error {
	log.Infow("[Migrate] Starting database migration", "driver", driverName)


	// 获取当前版本
	currentVersion, err := Current(ctx, db, driverName)
	if err != nil {
		log.Errorw("[Migrate] Failed to get current version", "error", err)
		return fmt.Errorf("failed to get current version: %w", err)
	}
	log.Infow("[Migrate] Current database version", "version", currentVersion)

	opts, err := getMigrator(db, driverName, log)
	if err != nil {
		log.Errorw("[Migrate] Failed to get migrator", "error", err)
		return fmt.Errorf("failed to get migrator: %w", err)
	}

	log.Info("[Migrate] Executing migration...")
	if err := migrate.New(opts).MigrateUp(ctx); err != nil {
		log.Errorw("[Migrate] Migration failed", "error", err)
		return err
	}

	// 获取迁移后的版本
	newVersion, err := Current(ctx, db, driverName)
	if err != nil {
		log.Errorw("[Migrate] Failed to get new version", "error", err)
	} else {
		log.Infow("[Migrate] Migration completed successfully", "version", newVersion)
	}

	return nil
}

// To performs the database migration to the specific version.
func To(ctx context.Context, db *sql.DB, driverName string, version string, log *logger.SugaredMateLogger) error {
	opts, err := getMigrator(db, driverName, log)
	if err != nil {
		return fmt.Errorf("failed to get migrator: %w", err)
	}
	return migrate.New(opts).MigrateTo(ctx, version)
}

// Current returns the current version ID (the latest migration applied) of the database.
func Current(ctx context.Context, db *sql.DB, driverName string) (string, error) {
	var (
		query               string
		migrationTableCount int
		queryTableName      = tableName
	)

	switch driverName {
	case sqliteDriverName:
		query = `
			SELECT count(*)
			FROM sqlite_master
			WHERE name = ? and type = 'table'`
	case mysqlDriverName:
		query = `
			SELECT count(*)
			FROM information_schema.TABLES
			WHERE TABLE_NAME= ? AND TABLE_SCHEMA = database()`
	case dmDriverName:
		// 达梦数据库：由于建表时使用了双引号 "ops_migrations"，
		// 在 ALL_TABLES 中 TABLE_NAME 存储的是小写，但是需要区分 OWNER
		query = `
			SELECT count(*)
			FROM USER_TABLES
			WHERE TABLE_NAME= ?`
	default:
		return "", fmt.Errorf("unsupported driver '%s'", driverName)
	}

	if err := db.QueryRowContext(ctx, query, queryTableName).Scan(&migrationTableCount); err != nil {
		return "", fmt.Errorf("failed to check migration table existence (table=%s, query=%s): %w", queryTableName, query, err)
	}

	// 表不存在，返回空版本（这是正常情况，表示首次初始化）
	if migrationTableCount == 0 {
		return "", nil
	}

	var version string

	// 达梦数据库需要使用双引号括起来保持大小写
	if driverName == dmDriverName {
		query = `select "version" from "` + tableName + `"`
	} else {
		query = "select version from " + tableName + " limit 1"
	}

	if err := db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		// 如果表存在但没有记录，返回空版本
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("migration table exists but has no rows, database may be corrupted")
		}
		return "", fmt.Errorf("failed to read current DB version from migration table (query=%s): %w", query, err)
	}

	return version, nil
}

func getMigrator(db *sql.DB, driverName string, log *logger.SugaredMateLogger) (migrate.Options, error) {
	before := func(ctx context.Context, _ *sql.Tx, version string) error {
		log.Infow("[Migrate][Before] Starting migration for version", "version", version)
		return nil
	}

	after := func(ctx context.Context, dbtx *sql.Tx, version string) error {
		log.Infow("[Migrate][After] Completed migration for version", "version", version)

		// TODO: Add version-specific migration hooks
		switch version {
		// case "0039_alter_table_webhooks_uid":
		// 	return migrateAfter_0039_alter_table_webhooks_uid(ctx, dbtx)
		// case "0042_alter_table_rules":
		// 	return migrateAfter_0042_alter_table_rules(ctx, dbtx)
		default:
			return nil
		}
	}

	opts := migrate.Options{
		After:      after,
		Before:     before,
		DriverName: driverName,
		DB:         db,
		FS:         sqlite,
		Table:      tableName,
	}

	switch driverName {
	case sqliteDriverName:
		folder, _ := fs.Sub(sqlite, sqliteSourceDir)
		opts.FS = folder
	case mysqlDriverName:
		folder, _ := fs.Sub(mysql, mysqlSourceDir)
		opts.FS = folder
	case dmDriverName:
		folder, _ := fs.Sub(dm, dmSourceDir)
		opts.FS = folder
	default:
		return migrate.Options{}, fmt.Errorf("unsupported driver '%s'", driverName)
	}

	return opts, nil
}
