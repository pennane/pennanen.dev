package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"math"
	"sort"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

const singleWriter = 1

func Open(ctx context.Context, path string) (*sql.DB, error) {
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(singleWriter)
	if err := migrate(ctx, d); err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

const everyMigration = math.MaxInt

func migrate(ctx context.Context, d *sql.DB) error {
	return migrateUpTo(ctx, d, everyMigration)
}

func migrateUpTo(ctx context.Context, d *sql.DB, lastVersion int) error {
	var appliedVersion int
	if err := d.QueryRowContext(ctx, "PRAGMA user_version").Scan(&appliedVersion); err != nil {
		return err
	}
	names, err := fs.Glob(migrations, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		version, err := migrationVersion(name)
		if err != nil {
			return err
		}
		if version <= appliedVersion || version > lastVersion {
			continue
		}
		if err := applyMigration(ctx, d, name, version); err != nil {
			return err
		}
	}
	return nil
}

func migrationVersion(name string) (int, error) {
	base := strings.TrimPrefix(name, "migrations/")
	version, err := strconv.Atoi(strings.SplitN(base, "_", 2)[0])
	if err != nil {
		return 0, fmt.Errorf("migration %s: name must start with a number", name)
	}
	return version, nil
}

func applyMigration(ctx context.Context, d *sql.DB, name string, version int) error {
	body, err := migrations.ReadFile(name)
	if err != nil {
		return err
	}
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("migration %s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", version)); err != nil {
		return err
	}
	return tx.Commit()
}
