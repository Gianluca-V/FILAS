// Command api is the composition root for the FILAS backend: it loads
// config, opens the MySQL connection pool, wires repositories -> usecases
// -> handlers (grows in later PRs), and starts the Gin HTTP server.
package main

import (
	"log"

	"github.com/gianluca-v/filas-backend/internal/config"
	handlerhttp "github.com/gianluca-v/filas-backend/internal/handler/http"
	"github.com/gianluca-v/filas-backend/internal/repository/mysql"
)

func main() {
	cfg := config.Load()

	db := mysql.MustOpen(cfg.DSN())
	defer db.Close()

	router := handlerhttp.NewRouter(handlerhttp.RouterDeps{
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		HealthDB:           db,
	})

	log.Printf("filas-backend listening on :%s", cfg.APIPort)
	if err := router.Run(":" + cfg.APIPort); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
