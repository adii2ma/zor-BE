package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/uptrace/bun"

	"be-zor/internal/models"
)

func ApplySchema(ctx context.Context, db *bun.DB) error {
	if _, err := db.NewCreateTable().
		Model((*models.UserRecord)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	if _, err := db.NewCreateTable().
		Model((*models.SessionRecord)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		return fmt.Errorf("create sessions table: %w", err)
	}

	indexes := []struct {
		name   string
		table  string
		column string
		unique bool
	}{
		{name: "users_google_subject_uidx", table: "users", column: "google_subject", unique: true},
		{name: "users_email_uidx", table: "users", column: "email", unique: true},
		{name: "sessions_token_uidx", table: "sessions", column: "token", unique: true},
		{name: "sessions_user_id_idx", table: "sessions", column: "user_id"},
	}

	for _, index := range indexes {
		query := db.NewCreateIndex().
			Table(index.table).
			Index(index.name).
			Column(index.column).
			IfNotExists()

		if index.unique {
			query = query.Unique()
		}

		if _, err := query.Exec(ctx); err != nil && !isDuplicateIndexError(err) {
			return fmt.Errorf("create index %s: %w", index.name, err)
		}
	}

	return nil
}

func isDuplicateIndexError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "already exists")
}
