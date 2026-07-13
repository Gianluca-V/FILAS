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
// the 5 public (unauthenticated) GET resource endpoints added in PR2, the
// admins resource added in PR3 (login is public; list/get/create
// (non-login)/update/delete require a valid JWT — the first real wiring of
// middleware.RequireAuth), and the auth-gated writes for
// products/gallery/family/organizations added in PR4 (task 3.1): GETs stay
// public, POST/PUT/DELETE require a valid JWT via middleware.RequireAuth,
// same pattern as the admins GET/PUT/DELETE routes. News writes are wired
// with the FIXED auth behavior in PR5 (task 4.1), not here. The orders
// resource is wired in a later PR.
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
	r.POST("/api/products", middleware.RequireAuth(deps.JWTService), products.Create)
	r.PUT("/api/products/:id", middleware.RequireAuth(deps.JWTService), products.Update)
	r.DELETE("/api/products/:id", middleware.RequireAuth(deps.JWTService), products.Delete)

	news := NewNewsHandler(deps.NewsService)
	r.GET("/api/news", news.List)
	r.GET("/api/news/:id", news.Get)

	gallery := NewGalleryHandler(deps.GalleryService)
	r.GET("/api/gallery", gallery.List)
	r.GET("/api/gallery/:id", gallery.Get)
	r.POST("/api/gallery", middleware.RequireAuth(deps.JWTService), gallery.Create)
	r.PUT("/api/gallery/:id", middleware.RequireAuth(deps.JWTService), gallery.Update)
	r.DELETE("/api/gallery/:id", middleware.RequireAuth(deps.JWTService), gallery.Delete)

	family := NewFamilyHandler(deps.FamilyService)
	r.GET("/api/family", family.List)
	r.GET("/api/family/:id", family.Get)
	r.POST("/api/family", middleware.RequireAuth(deps.JWTService), family.Create)
	r.PUT("/api/family/:id", middleware.RequireAuth(deps.JWTService), family.Update)
	r.DELETE("/api/family/:id", middleware.RequireAuth(deps.JWTService), family.Delete)

	organizations := NewOrganizationHandler(deps.OrganizationService)
	r.GET("/api/organizations", organizations.List)
	r.GET("/api/organizations/:id", organizations.Get)
	r.POST("/api/organizations", middleware.RequireAuth(deps.JWTService), organizations.Create)
	r.PUT("/api/organizations/:id", middleware.RequireAuth(deps.JWTService), organizations.Update)
	r.DELETE("/api/organizations/:id", middleware.RequireAuth(deps.JWTService), organizations.Delete)

	// Legacy admins.php has a routing bug where GET /api/admins (no
	// trailing slash, no ID) misroutes into getUser(0) -> 404 instead of
	// the list, an artifact of its manual explode('/') URL parsing. Gin's
	// :id path-param routing below has no equivalent failure mode, so
	// there is nothing to deliberately NOT reproduce — the bug class does
	// not exist here. See backend/docs/legacy-quirks.md §6.
	admins := NewAdminHandler(deps.AdminService, deps.AuthService, deps.JWTService)
	r.POST("/api/admins", admins.Create) // login (public) vs create (auth-gated inline)
	r.GET("/api/admins", middleware.RequireAuth(deps.JWTService), admins.List)
	r.GET("/api/admins/:id", middleware.RequireAuth(deps.JWTService), admins.Get)
	r.PUT("/api/admins/:id", middleware.RequireAuth(deps.JWTService), admins.Update)
	r.DELETE("/api/admins/:id", middleware.RequireAuth(deps.JWTService), admins.Delete)

	return r
}
