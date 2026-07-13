// Package rest contains Gin handlers, request/response DTOs, and the
// route/middleware registration for the API. Named "rest" (not "http") so
// callers never need to alias net/http when importing this package.
package rest

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// dbPinger is satisfied by *sqlx.DB; kept narrow so the handler is testable
// without a real database connection.
type dbPinger interface {
	PingContext(ctx context.Context) error
}

// NewHealthHandler returns a handler for GET /health that reports 200 when
// the database is reachable and 503 otherwise, per the skeleton boot gate
// ("container boots, DB pings, /health 200"). The underlying DB error is
// logged server-side only — unauthenticated callers get a generic body
// (fix #7 from the PR1 gate review; do not leak internal error text).
func NewHealthHandler(db dbPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			log.Printf("health check failed: db ping error: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
