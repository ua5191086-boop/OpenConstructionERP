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
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Logger()

	cfg := loadConfig()
	db := connectDB(cfg, &logger)
	defer db.Close()

	redis := connectRedis(cfg, &logger)
	defer redis.Close()

	kafka := connectKafka(cfg, &logger)
	defer kafka.Close()

	minio := connectMinIO(cfg, &logger)

	search := connectOpenSearch(cfg, &logger)

	// Initialize domain services
	ontologySvc := ontology.NewService(db, redis, kafka, &logger)
	iamSvc := iam.NewService(db, redis, cfg.JWTSecret, &logger)
	cdeSvc := cde.NewService(db, minio, kafka, &logger)
	workflowSvc := workflow.NewService(db, kafka, &logger)
	notifySvc := notify.NewService(db, kafka, &logger)
	reportSvc := report.NewService(db, search, &logger)

	// API router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","version":"0.1.0","time":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(iamSvc.AuthMiddleware)

		// Ontology
		r.Route("/ontology", func(r chi.Router) {
			r.Get("/object-types", ontologySvc.ListObjectTypes)
			r.Post("/object-types", ontologySvc.CreateObjectType)
			r.Get("/object-types/{id}", ontologySvc.GetObjectType)
			r.Put("/object-types/{id}", ontologySvc.UpdateObjectType)
			r.Get("/objects", ontologySvc.ListObjects)
			r.Post("/objects", ontologySvc.CreateObject)
			r.Get("/objects/{id}", ontologySvc.GetObject)
			r.Put("/objects/{id}", ontologySvc.UpdateObject)
			r.Get("/links", ontologySvc.ListLinks)
			r.Post("/links", ontologySvc.CreateLink)
		})

		// IAM
		r.Route("/iam", func(r chi.Router) {
			r.Post("/login", iamSvc.Login)
			r.Post("/refresh", iamSvc.RefreshToken)
			r.Post("/users", iamSvc.CreateUser)
			r.Get("/users", iamSvc.ListUsers)
			r.Get("/users/me", iamSvc.GetCurrentUser)
			r.Put("/users/me", iamSvc.UpdateCurrentUser)
			r.Get("/roles", iamSvc.ListRoles)
			r.Post("/roles", iamSvc.CreateRole)
		})

		// CDE (Document Control)
		r.Route("/projects/{projectId}/documents", func(r chi.Router) {
			r.Get("/", cdeSvc.ListDocuments)
			r.Post("/", cdeSvc.CreateDocument)
			r.Get("/{id}", cdeSvc.GetDocument)
			r.Put("/{id}", cdeSvc.UpdateDocument)
			r.Post("/{id}/upload", cdeSvc.UploadFile)
			r.Get("/{id}/download", cdeSvc.DownloadFile)
			r.Post("/{id}/transmit", cdeSvc.TransmitDocument)
			r.Post("/{id}/approve", cdeSvc.ApproveDocument)
			r.Post("/{id}/reject", cdeSvc.RejectDocument)
		})

		// Workflow
		r.Route("/workflows", func(r chi.Router) {
			r.Get("/", workflowSvc.ListWorkflows)
			r.Post("/", workflowSvc.CreateWorkflow)
			r.Get("/{id}", workflowSvc.GetWorkflow)
			r.Post("/{id}/execute", workflowSvc.ExecuteWorkflow)
			r.Get("/instances", workflowSvc.ListInstances)
			r.Get("/instances/{id}", workflowSvc.GetInstance)
			r.Post("/instances/{id}/transition", workflowSvc.Transition)
		})

		// Notifications
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notifySvc.ListNotifications)
			r.Post("/", notifySvc.CreateNotification)
			r.Put("/{id}/read", notifySvc.MarkAsRead)
			r.Get("/unread-count", notifySvc.UnreadCount)
		})

		// Reports
		r.Route("/reports", func(r chi.Router) {
			r.Get("/", reportSvc.ListReports)
			r.Post("/", reportSvc.CreateReport)
			r.Get("/{id}", reportSvc.GetReport)
			r.Post("/{id}/generate", reportSvc.GenerateReport)
			r.Get("/{id}/download", reportSvc.DownloadReport)
		})
	})

	// GraphQL endpoint
	r.Post("/graphql", graphQLHandler(ontologySvc, cdeSvc, iamSvc, &logger))

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info().Str("port", cfg.Port).Msg("OpenConstructionERP Core API starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	KafkaURL    string
	MinIOEndpoint string
	MinIOAccessKey string
	MinIOSecretKey string
	OpenSearchURL string
	JWTSecret   string
}

func loadConfig() *Config {
	return &Config{
		Port:           getEnv("PORT", "8081"),
		DatabaseURL:    fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("DB_USER", "oce"), getEnv("DB_PASSWORD", "oce_secret"),
			getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "oce")),
		RedisURL:       fmt.Sprintf("redis://:%s@%s:%s/0",
			getEnv("REDIS_PASSWORD", "oce_secret"),
			getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379")),
		KafkaURL:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "oce"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "oce_secret"),
		OpenSearchURL:  fmt.Sprintf("http://%s:%s",
			getEnv("OPENSEARCH_HOST", "localhost"), getEnv("OPENSEARCH_PORT", "9200")),
		JWTSecret:      getEnv("JWT_SECRET", "oce-jwt-secret"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func connectDB(cfg *Config, log *zerolog.Logger) *pgxpool.Pool {
	// In real implementation: pgxpool.New(ctx, cfg.DatabaseURL)
	log.Info().Msg("connecting to PostgreSQL...")
	return nil
}

func connectRedis(cfg *Config, log *zerolog.Logger) *redis.Client {
	log.Info().Msg("connecting to Redis...")
	return nil
}

func connectKafka(cfg *Config, log *zerolog.Logger) *kafka.Conn {
	log.Info().Msg("connecting to Kafka...")
	return nil
}

func connectMinIO(cfg *Config, log *zerolog.Logger) *minio.Client {
	log.Info().Msg("connecting to MinIO...")
	return nil
}

func connectOpenSearch(cfg *Config, log *zerolog.Logger) *opensearch.Client {
	log.Info().Msg("connecting to OpenSearch...")
	return nil
}

func graphQLHandler(ontologySvc *ontology.Service, cdeSvc *cde.Service, iamSvc *iam.Service, log *zerolog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"data":{"__schema":{"types":[]}}}`)
	}
}
