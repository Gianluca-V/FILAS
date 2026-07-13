// Package middleware holds Gin middleware: JWT auth, CORS, and domain
// error -> HTTP status translation.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/auth"
)

// jwtParser is satisfied by *auth.JWTService; kept narrow for testability.
type jwtParser interface {
	Parse(tokenString string) (int64, error)
}

// RequireAuth returns middleware that validates the Authorization header
// as a JWT, aborting with 401 on a missing or invalid token. On success it
// stores the admin ID in the Gin context under "adminID". The optional
// "Bearer " prefix is tolerated for parity with the legacy client, which
// sends the raw token.
func RequireAuth(svc jwtParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")
		adminID, err := svc.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
			return
		}

		c.Set("adminID", adminID)
		c.Next()
	}
}

var _ jwtParser = (*auth.JWTService)(nil)
