package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/uptrace/bun"
	bunmigrate "github.com/uptrace/bun/migrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var migrations = mustLoadMigrations()

func mustLoadMigrations() *bunmigrate.Migrations {
	fsys, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		panic(err)
	}

	loaded := bunmigrate.NewMigrations()
	if err := loaded.Discover(fsys); err != nil {
		panic(err)
	}

	return loaded
}

func NewMigrator(db *bun.DB) *bunmigrate.Migrator {
	return bunmigrate.NewMigrator(
		db,
		migrations,
		bunmigrate.WithMarkAppliedOnSuccess(true),
		bunmigrate.WithUpsert(true),
	)
}

func Migrate(ctx context.Context, db *bun.DB) (*bunmigrate.MigrationGroup, error) {
	migrator := NewMigrator(db)
	if err := migrator.Init(ctx); err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	if err := migrator.Lock(ctx); err != nil {
		return nil, fmt.Errorf("lock migrator: %w", err)
	}
	defer func() {
		_ = migrator.Unlock(ctx)
	}()

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return group, nil
}

func Rollback(ctx context.Context, db *bun.DB) (*bunmigrate.MigrationGroup, error) {
	migrator := NewMigrator(db)
	if err := migrator.Init(ctx); err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	if err := migrator.Lock(ctx); err != nil {
		return nil, fmt.Errorf("lock migrator: %w", err)
	}
	defer func() {
		_ = migrator.Unlock(ctx)
	}()

	group, err := migrator.Rollback(ctx)
	if err != nil {
		return nil, fmt.Errorf("rollback migrations: %w", err)
	}

	return group, nil
}

func MigrationStatus(ctx context.Context, db *bun.DB) (bunmigrate.MigrationSlice, error) {
	migrator := NewMigrator(db)
	if err := migrator.Init(ctx); err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	statuses, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("load migration status: %w", err)
	}

	return statuses, nil
}

func FormatMigrationGroup(group *bunmigrate.MigrationGroup) string {
	if group == nil || group.IsZero() {
		return "no migrations applied"
	}

	return group.String()
}

func FormatMigrationStatus(statuses bunmigrate.MigrationSlice) string {
	if len(statuses) == 0 {
		return "no migrations registered"
	}

	lines := make([]string, 0, len(statuses))
	for _, migration := range statuses {
		state := "pending"
		if migration.IsApplied() {
			state = "applied"
		}

		lines = append(lines, fmt.Sprintf("%s_%s\t%s", migration.Name, migration.Comment, state))
	}

	return strings.Join(lines, "\n")
}
