package db

import (
	"database/sql"
	"fmt"
)

type migration struct {
	Version int
	SQL     string
}

var migrations = []migration{
	{
		Version: 1,
		SQL: `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INT NOT NULL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS urls (
	id BIGINT UNSIGNED NOT NULL PRIMARY KEY,
	short_code VARCHAR(16) NOT NULL,
	original_url TEXT NOT NULL,
	original_url_sha256 BINARY(32) NOT NULL,
	expire_at DATETIME(6) NULL,
	deleted_at DATETIME(6) NULL,
	created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),

	UNIQUE KEY uq_urls_short_code (short_code),
	KEY idx_urls_sha256_active (original_url_sha256, deleted_at, expire_at),
	KEY idx_urls_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`,
	},
}

func Migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version INT NOT NULL PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[int]bool{}
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scan schema_migrations: %w", err)
		}
		applied[v] = true
	}

	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.Version, err)
		}
		if _, err := tx.Exec(m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec migration %d: %w", m.Version, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, m.Version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}
	}

	return nil
}
