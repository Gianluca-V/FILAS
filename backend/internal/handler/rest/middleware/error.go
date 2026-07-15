package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// ErrorHandler centralizes domain error -> HTTP status translation. Handlers
// push errors via c.Error(err) and return early without writing a response;
// this middleware runs after the handler chain and maps the last pushed
// error to a status code + JSON body, keeping status mapping in one
// auditable place instead of scattered across handlers. The underlying
// error is always logged server-side (method + path + error) before the
// response is written, even though the response body itself stays generic
// (fix #6 from the PR1 gate review — do not let real errors vanish silently).
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Written() {
			return
		}
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status, message := mapDomainError(err)
		log.Printf("%s %s -> %d: %v", c.Request.Method, c.Request.URL.Path, status, err)
		c.JSON(status, gin.H{"message": message})
	}
}

func mapDomainError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "resource not found"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrInsufficientStock):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
