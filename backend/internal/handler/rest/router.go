package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

// RouterDeps holds the dependencies NewRouter needs to wire the Gin engine.
// It grows as later PRs add resource handlers (orders, ...).
type RouterDeps struct {
	CORSAllowedOrigins []string
	HealthDB           dbPinger

	ProductService      ProductService
	NewsService         NewsService
	GalleryService      GalleryService
	FamilyService       FamilyService
	OrganizationService OrganizationService

	AdminService AdminService
	AuthService  AuthService
	JWTService   tokenParser
}

// NewRouter builds the Gin engine with global middleware (recovery, request
// logging, CORS, centralized error handling) and registers routes: /health,
// the 5 public (unauthenticated) GET resource endpoints added in PR2, and
// the admins resource added in PR3 (login is public; list/get/create
// (non-login)/update/delete require a valid JWT — this is the first real
// wiring of middleware.RequireAuth). Auth-gated writes for the other 4
// resources and the orders resource are wired in later PRs.
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(deps.CORSAllowedOrigins))
	r.Use(middleware.ErrorHandler())

	r.GET("/health", NewHealthHandler(deps.HealthDB))

	products := NewProductHandler(deps.ProductService)
	r.GET("/api/products", products.List)
	r.GET("/api/products/:id", products.Get)

	news := NewNewsHandler(deps.NewsService)
	r.GET("/api/news", news.List)
	r.GET("/api/news/:id", news.Get)

	gallery := NewGalleryHandler(deps.GalleryService)
	r.GET("/api/gallery", gallery.List)
	r.GET("/api/gallery/:id", gallery.Get)

	family := NewFamilyHandler(deps.FamilyService)
	r.GET("/api/family", family.List)
	r.GET("/api/family/:id", family.Get)

	organizations := NewOrganizationHandler(deps.OrganizationService)
	r.GET("/api/organizations", organizations.List)
	r.GET("/api/organizations/:id", organizations.Get)

	admins := NewAdminHandler(deps.AdminService, deps.AuthService, deps.JWTService)
	r.POST("/api/admins", admins.Create) // login (public) vs create (auth-gated inline)
	r.GET("/api/admins", middleware.RequireAuth(deps.JWTService), admins.List)
	r.GET("/api/admins/:id", middleware.RequireAuth(deps.JWTService), admins.Get)
	r.PUT("/api/admins/:id", middleware.RequireAuth(deps.JWTService), admins.Update)
	r.DELETE("/api/admins/:id", middleware.RequireAuth(deps.JWTService), admins.Delete)

	return r
}
