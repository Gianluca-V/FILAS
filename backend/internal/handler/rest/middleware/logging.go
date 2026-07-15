package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger returns middleware that logs one line per request: method,
// path, status code, and latency. This is the minimum observability needed
// to correlate the generic error responses ErrorHandler writes with what
// actually happened server-side (fix #6 from the PR1 gate review).
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		log.Printf("%s %s -> %d (%s)", method, path, c.Writer.Status(), time.Since(start))
	}
}
