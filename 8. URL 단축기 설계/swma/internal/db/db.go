package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"example.com/urlshortener/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

func DSN(cfg config.Config) string {
	// parseTime=true is required for time.Time scanning
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&multiStatements=true",
		cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
}

func OpenWithRetry(cfg config.Config, timeout time.Duration) (*sql.DB, error) {
	dsn := DSN(cfg)
	deadline := time.Now().Add(timeout)

	var lastErr error
	for time.Now().Before(deadline) {
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		db.SetMaxOpenConns(cfg.DBMaxOpen)
		db.SetMaxIdleConns(cfg.DBMaxIdle)
		db.SetConnMaxLifetime(cfg.DBMaxLife)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return db, nil
		}
		_ = db.Close()
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	return nil, lastErr
}
