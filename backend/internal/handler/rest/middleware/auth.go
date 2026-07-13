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

// TokenParser is the exported counterpart of jwtParser, satisfied by
// *auth.JWTService, used as the parameter type of AuthenticateRequest so
// callers outside this package (e.g. AdminHandler.requireAuthInline) can
// invoke it without needing their own copy of the header-parsing logic.
// The narrow-interface-per-consumer pattern used elsewhere (jwtParser
// here, tokenParser in handler/rest, jwtIssuer in usecase) is deliberately
// kept as-is: only the LOGIC below is deduplicated, not the interface
// types (see AuthenticateRequest doc comment).
type TokenParser interface {
	Parse(tokenString string) (int64, error)
}

// AuthenticateRequest is the single implementation of bearer-token
// authentication: extract the Authorization header, tolerate an optional
// "Bearer " prefix (parity with the legacy client, which sends the raw
// token), parse/validate the JWT, and on failure abort the request with a
// 401 JSON body. Both RequireAuth below (route-group middleware, used by
// every gated GET/PUT/DELETE route) and AdminHandler.requireAuthInline
// (used by POST /api/admins, which cannot use route-group middleware
// because the same route/method also serves the public login call) call
// this, so the header-parsing logic and the two 401 message strings
// ("missing authorization token" / "invalid or expired token") cannot
// drift between the two call sites — previously hand-duplicated (gate #36
// readability finding).
//
// UNIFIED-401 SIMPLIFICATION (documented, not a legacy-parity requirement
// any decision covers): legacy PHP (e.g. admins.php) distinguishes a
// MISSING Authorization header (400, empty body) from an INVALID token
// (401, {"message":"Token is invalid"}) on every gated route. This
// implementation unifies BOTH cases to 401 with the two messages above,
// using the ONE auth contract every gated route shares rather than forking
// a second status-code scheme per resource. See
// backend/docs/legacy-quirks.md §5 for the full characterization.
func AuthenticateRequest(c *gin.Context, svc TokenParser) (adminID int64, ok bool) {
	header := c.GetHeader("Authorization")
	if header == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing authorization token"})
		return 0, false
	}

	token := strings.TrimPrefix(header, "Bearer ")
	id, err := svc.Parse(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
		return 0, false
	}
	return id, true
}

// RequireAuth returns middleware that validates the Authorization header
// as a JWT, aborting with 401 on a missing or invalid token. On success it
// stores the admin ID in the Gin context under "adminID".
func RequireAuth(svc jwtParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, ok := AuthenticateRequest(c, svc)
		if !ok {
			return
		}
		c.Set("adminID", adminID)
		c.Next()
	}
}

var _ jwtParser = (*auth.JWTService)(nil)
var _ TokenParser = (*auth.JWTService)(nil)
