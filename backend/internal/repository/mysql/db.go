// Package mysql implements domain repository interfaces on top of sqlx +
// the MySQL driver, using parameterized queries only.
package mysql

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // registers the "mysql" sql.DB driver
	"github.com/jmoiron/sqlx"
)

// Modest local-dev connection pool limits (fix #10 from the PR1 gate
// review). This is a single-instance local portfolio deployment, not a
// production fleet — these values just prevent an unbounded pool and stale
// idle connections outliving a MySQL-side wait_timeout.
const (
	maxOpenConns    = 10
	maxIdleConns    = 5
	connMaxLifetime = 5 * time.Minute
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

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	return db
}
