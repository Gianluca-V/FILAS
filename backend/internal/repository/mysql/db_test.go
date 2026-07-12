package mysql_test

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return sqlx.NewDb(sqlDB, "sqlmock"), mock
}

func TestPing_Success(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectPing()

	if err := mysql.Ping(context.Background(), db); err != nil {
		t.Errorf("Ping() error = %v, want nil", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPing_PropagatesFailure(t *testing.T) {
	db, mock := newMockDB(t)
	mock.ExpectPing().WillReturnError(context.DeadlineExceeded)

	err := mysql.Ping(context.Background(), db)
	if err == nil {
		t.Fatal("Ping() error = nil, want error when the driver ping fails")
	}
}

func TestMustOpen_PanicsOnUnopenableDSN(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustOpen() did not panic on an unreachable DSN")
		}
	}()
	// Loopback port unlikely to have a MySQL server; Ping must fail fast.
	mysql.MustOpen("root:root@tcp(127.0.0.1:1)/nonexistent?timeout=1s")
}
