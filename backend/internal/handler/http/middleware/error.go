package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/domain"
)

// ErrorHandler centralizes domain error -> HTTP status translation. Handlers
// push errors via c.Error(err) and return early without writing a response;
// this middleware runs after the handler chain and maps the last pushed
// error to a status code + JSON body, keeping status mapping in one
// auditable place instead of scattered across handlers.
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
