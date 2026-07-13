package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	corsAllowMethods = "GET, POST, PUT, DELETE, PATCH, OPTIONS"
	corsAllowHeaders = "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With"
)

// CORS returns middleware replacing the legacy manual header()+OPTIONS-204
// handling. When allowedOrigins is empty it falls back to "*" (permissive
// local dev); otherwise only an exact match is echoed back with Vary:
// Origin, and the header is omitted for any other origin.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		switch {
		case len(allowed) == 0:
			c.Header("Access-Control-Allow-Origin", "*")
		case origin != "":
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}

		c.Header("Access-Control-Allow-Methods", corsAllowMethods)
		c.Header("Access-Control-Allow-Headers", corsAllowHeaders)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
