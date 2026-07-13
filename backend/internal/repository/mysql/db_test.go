package mysql_test

import (
	"testing"

	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
)

func TestMustOpen_PanicsOnMalformedDSN(t *testing.T) {
	// Hermetic: a malformed DSN fails synchronously in the mysql driver's
	// DSN parser (sqlx.Open), so this never attempts a real network dial
	// (fix #13 from the PR1 gate review).
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustOpen() did not panic on a malformed DSN")
		}
	}()
	mysql.MustOpen("this-is-not-a-valid-dsn-at-all")
}
