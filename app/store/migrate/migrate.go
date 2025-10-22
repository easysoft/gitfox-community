package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/migrate"
	// "github.com/rs/zerolog/log"
)

//go:embed sqlite/*.sql
var sqlite embed.FS

//go:embed mysql/*.sql
var mysql embed.FS

const (
	tableName = "migrations"

	sqliteDriverName = "sqlite3"
	sqliteSourceDir  = "sqlite"

	mysqlDriverName = "mysql"
	mysqlSourceDir  = "mysql"
)

// Migrate performs the database migration.
func Migrate(ctx context.Context, db *sql.DB, driverName string) error {
	opts, err := getMigrator(db, driverName)
	if err != nil {
		return fmt.Errorf("failed to get migrator: %w", err)
	}
	return migrate.New(opts).MigrateUp(ctx)
}

// To performs the database migration to the specific version.
func To(ctx context.Context, db *sql.DB, driverName string, version string) error {
	opts, err := getMigrator(db, driverName)
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
			WHERE TABLE_NAME= ?`
	default:
		return "", fmt.Errorf("unsupported driver '%s'", driverName)
	}

	if err := db.QueryRowContext(ctx, query, tableName).Scan(&migrationTableCount); err != nil {
		return "", fmt.Errorf("failed to check migration table existence: %w", err)
	}

	if migrationTableCount == 0 {
		return "", nil
	}

	var version string

	query = "select version from " + tableName + " limit 1"
	if err := db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		return "", fmt.Errorf("failed to read current DB version from migration table: %w", err)
	}

	return version, nil
}

func getMigrator(db *sql.DB, driverName string) (migrate.Options, error) {
	before := func(ctx context.Context, _ *sql.Tx, version string) error {
		// TODO: Add logging back later
		// ctx = log.Ctx(ctx).With().
		// 	Str("migrate.version", version).
		// 	Str("migrate.phase", "before").
		// 	Logger().WithContext(ctx)
		// log := log.Ctx(ctx)
		//
		// log.Info().Msg("[START]")
		// defer log.Info().Msg("[DONE]")
		_ = ctx
		_ = version

		return nil
	}

	after := func(ctx context.Context, dbtx *sql.Tx, version string) error {
		// TODO: Add logging back later
		// ctx = log.Ctx(ctx).With().
		// 	Str("migrate.version", version).
		// 	Str("migrate.phase", "after").
		// 	Logger().WithContext(ctx)
		// log := log.Ctx(ctx)
		//
		// log.Info().Msg("[START]")
		// defer log.Info().Msg("[DONE]")
		_ = ctx
		_ = dbtx

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

	default:
		return migrate.Options{}, fmt.Errorf("unsupported driver '%s'", driverName)
	}

	return opts, nil
}
