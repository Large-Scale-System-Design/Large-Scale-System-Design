package migrate

import (
	"database/sql"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS urls (
  id                 BIGINT UNSIGNED NOT NULL,
  short_code         VARCHAR(32) NOT NULL,
  original_url       TEXT NOT NULL,
  original_url_sha256 BINARY(32) NOT NULL,
  expire_at          DATETIME NULL,
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at         DATETIME NULL,
  PRIMARY KEY (id),
  UNIQUE KEY ux_short_code (short_code),
  KEY ix_sha_active (original_url_sha256),
  KEY ix_deleted_at (deleted_at),
  KEY ix_expire_at (expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}
