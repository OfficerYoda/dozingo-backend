package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/officeryoda/dozingo/internal/config"
	"github.com/officeryoda/dozingo/internal/handler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	pool, err := connectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	router := createRouter(cfg)
	registerRoutes(router, pool)
	startServer(cfg.Port, router)
}

// connectDB creates a connection pool to PostgreSQL and verifies the connection.
func connectDB(databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	log.Println("Connected to database")
	return pool, nil
}

// createRouter creates a Chi router with standard middleware and a root health page.
func createRouter(cfg *config.Config) *chi.Mux {
	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := fmt.Fprintf(w, "Dozingo API is running\nDocs: http://localhost:%s/docs", cfg.Port); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	return router
}

// registerRoutes sets up the Huma API and registers all handler groups.
func registerRoutes(router *chi.Mux, pool *pgxpool.Pool) {
	api := humachi.New(router, huma.DefaultConfig("Dozingo API", "0.1.0"))
	apiGroup := huma.NewGroup(api, "/api")

	handler.RegisterHealth(apiGroup)
	handler.RegisterLecturers(apiGroup, pool)
	handler.RegisterBoards(apiGroup, pool)
}

// startServer begins listening on the given port and blocks until the server exits.
func startServer(port string, handler http.Handler) {
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server starting on %s", addr)
	log.Printf("API docs available at http://localhost:%s/docs", port)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
