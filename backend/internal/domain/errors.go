// Package domain holds framework-free entities, repository interfaces, and
// domain-level sentinel errors shared across the usecase, repository, and
// handler layers.
package domain

import "errors"

// Sentinel domain errors. Usecases wrap these with fmt.Errorf("...: %w", ErrX)
// so the HTTP layer can map them to status codes via errors.Is.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")
	// ErrUnauthorized indicates missing or invalid authentication.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrValidation indicates the request payload failed a domain invariant.
	ErrValidation = errors.New("validation failed")
	// ErrConflict indicates a state conflict (e.g. invalid transition).
	ErrConflict = errors.New("conflict")
	// ErrInsufficientStock indicates an order would drive product stock below zero.
	ErrInsufficientStock = errors.New("insufficient stock")
)
