// Package http contains Gin handlers, request/response DTOs, and the
// route/middleware registration for the API.
package http

import (
	"context"
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
// ("container boots, DB pings, /health 200").
func NewHealthHandler(db dbPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "down",
				"error":  err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
