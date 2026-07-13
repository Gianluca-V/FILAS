package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

// RouterDeps holds the dependencies NewRouter needs to wire the Gin engine.
// It grows as later PRs add resource handlers (products, news, orders, ...).
type RouterDeps struct {
	CORSAllowedOrigins []string
	HealthDB           dbPinger
}

// NewRouter builds the Gin engine with global middleware (recovery, request
// logging, CORS, centralized error handling) and registers routes. Only
// /health is wired in the skeleton; resource routes are added by later PRs.
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(deps.CORSAllowedOrigins))
	r.Use(middleware.ErrorHandler())

	r.GET("/health", NewHealthHandler(deps.HealthDB))

	return r
}
