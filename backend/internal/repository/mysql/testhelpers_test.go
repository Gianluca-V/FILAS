package mysql_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

// newTestDB is the single shared sqlmock fixture for every repository test
// in this package. It used to be copy-pasted byte-for-byte five times (one
// per resource); centralizing it here removes that duplication (gate #34
// review, fix #5).
func newTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlx.NewDb(db, "sqlmock"), mock
}
