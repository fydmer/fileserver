package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
)

//go:embed migrations
var f embed.FS

const dir = "migrations"

func applyEntry(ctx context.Context, db *sql.DB, entry fs.DirEntry) error {
	content, err := f.ReadFile(filepath.Join(dir, entry.Name()))
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, string(content)); err != nil {
		return errors.Join(err, tx.Rollback())
	}

	return tx.Commit()
}

func Apply(ctx context.Context, db *sql.DB) error {
	entries, err := f.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".sql") {
			continue
		}

		slog.Info("applying migration", slog.String("name", entry.Name()))
		if err = applyEntry(ctx, db, entry); err != nil {
			return err
		}

		slog.Info("migration applied", slog.String("name", entry.Name()))
	}

	return nil
}
