// Package mysql implements domain repository interfaces on top of sqlx +
// the MySQL driver, using parameterized queries only.
package mysql

import (
	"context"
	"fmt"

	_ "github.com/go-sql-driver/mysql" // registers the "mysql" sql.DB driver
	"github.com/jmoiron/sqlx"
)

// MustOpen opens a MySQL connection pool via sqlx and pings it immediately,
// panicking on failure. Intended for the composition root only (a
// non-functional DB is a fatal startup condition for the API process).
func MustOpen(dsn string) *sqlx.DB {
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		panic(fmt.Sprintf("mysql: open failed: %v", err))
	}
	if err := db.Ping(); err != nil {
		panic(fmt.Sprintf("mysql: ping failed: %v", err))
	}
	return db
}

// Ping checks DB connectivity with the given context. Used by the health
// handler to report liveness without crashing the process.
func Ping(ctx context.Context, db *sqlx.DB) error {
	return db.PingContext(ctx)
}
