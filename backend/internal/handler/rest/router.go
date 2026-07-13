package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/gianluca-v/filas-backend/internal/handler/rest/middleware"
)

// RouterDeps holds the dependencies NewRouter needs to wire the Gin engine.
// It grows as later PRs add resource handlers (admins, orders, ...).
type RouterDeps struct {
	CORSAllowedOrigins []string
	HealthDB           dbPinger

	ProductService      ProductService
	NewsService         NewsService
	GalleryService      GalleryService
	FamilyService       FamilyService
	OrganizationService OrganizationService
}

// NewRouter builds the Gin engine with global middleware (recovery, request
// logging, CORS, centralized error handling) and registers routes: /health
// plus the 5 public (unauthenticated) GET resource endpoints added in PR2.
// Auth-gated writes and the admins/orders resources are wired in later PRs.
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

	return r
}
