// Command api is the composition root for the FILAS backend: it loads
// config, opens the MySQL connection pool, wires repositories -> usecases
// -> handlers (grows in later PRs), and starts the Gin HTTP server with
// graceful shutdown on SIGINT/SIGTERM.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gianluca-v/filas-backend/internal/auth"
	"github.com/gianluca-v/filas-backend/internal/config"
	"github.com/gianluca-v/filas-backend/internal/handler/rest"
	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
	"github.com/gianluca-v/filas-backend/internal/usecase"
)

// shutdownTimeout bounds how long in-flight requests get to finish once a
// shutdown signal arrives before the server is forced closed.
const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run holds all cleanup in one place (via defer) so both the signal path
// and the error path close the DB pool — no bare log.Fatalf after
// resources are open, which used to skip deferred cleanup entirely (fix #9
// from the PR1 gate review).
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db := mysql.MustOpen(cfg.DSN())
	defer db.Close()

	productSvc := usecase.NewProductService(mysql.NewProductRepository(db))
	newsSvc := usecase.NewNewsService(mysql.NewNewsRepository(db))
	gallerySvc := usecase.NewGalleryService(mysql.NewGalleryRepository(db))
	familySvc := usecase.NewFamilyService(mysql.NewFamilyRepository(db))
	organizationSvc := usecase.NewOrganizationService(mysql.NewOrganizationRepository(db))

	adminRepo := mysql.NewAdminRepository(db)
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, time.Duration(cfg.JWTExpiryHours)*time.Hour)
	adminSvc := usecase.NewAdminService(adminRepo)
	authSvc := usecase.NewAuthService(adminRepo, jwtSvc)

	router := rest.NewRouter(rest.RouterDeps{
		CORSAllowedOrigins:  cfg.CORSAllowedOrigins,
		HealthDB:            db,
		ProductService:      productSvc,
		NewsService:         newsSvc,
		GalleryService:      gallerySvc,
		FamilyService:       familySvc,
		OrganizationService: organizationSvc,
		AdminService:        adminSvc,
		AuthService:         authSvc,
		JWTService:          jwtSvc,
	})

	srv := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: router,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Printf("filas-backend listening on :%s", cfg.APIPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serveErr:
		return err
	case sig := <-stop:
		log.Printf("received %s, shutting down gracefully (timeout %s)", sig, shutdownTimeout)
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	log.Println("server shut down cleanly")
	return nil
}
