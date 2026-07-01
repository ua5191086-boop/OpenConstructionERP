package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/openconstructionerp/oce/services/core/internal/db"
	"github.com/openconstructionerp/oce/services/core/internal/handlers"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("[Core API] OpenConstructionERP Core API starting...")

	// Connect to database
	database, err := db.New()
	if err != nil {
		log.Fatalf("[Core API] Failed to connect to database: %v", err)
	}
	defer database.Close()

	// API router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key", "X-API-Key"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","service":"oce-core-api","version":"0.1.0","time":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// BOQ Module
		boqHandler := handlers.NewBOQHandler(database.DB)
		boqHandler.RegisterRoutes(r)

		// Tenders Module
		tendersHandler := handlers.NewTendersHandler(database.DB)
		tendersHandler.RegisterRoutes(r)

		// Contracts Module
		contractsHandler := handlers.NewContractsHandler(database.DB)
		contractsHandler.RegisterRoutes(r)

		// HR Module
		hrHandler := handlers.NewHRHandler(database.DB)
		hrHandler.RegisterRoutes(r)

		// Finance Module
		financeHandler := handlers.NewFinanceHandler(database.DB)
		financeHandler.RegisterRoutes(r)

		// Procurement Module
		procurementHandler := handlers.NewProcurementHandler(database.DB)
		procurementHandler.RegisterRoutes(r)

		// BIM Module
		bimHandler := handlers.NewBIMHandler(database.DB)
		bimHandler.RegisterRoutes(r)

		// AI Module
		aiHandler := handlers.NewAIHandler(database.DB)
		aiHandler.RegisterRoutes(r)

		// Project Management Module
		pmHandler := handlers.NewPMHandler(database.DB)
		pmHandler.RegisterRoutes(r)

		// Document Control Module
		docControlHandler := handlers.NewDocControlHandler(database.DB)
		docControlHandler.RegisterRoutes(r)
	})

	// Start server
	port := getEnv("PORT", "8081")
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("[Core API] Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Core API] Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Core API] Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
