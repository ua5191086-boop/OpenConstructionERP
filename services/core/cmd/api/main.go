package main

import (
	"context"
	"encoding/json"
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
	"github.com/openconstructionerp/oce/services/core/internal/auth"
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

	// Keycloak JWT Auth
	keycloakCfg := auth.NewKeycloakConfig()

	// Public routes (no auth required)
	r.Route("/api/v1", func(r chi.Router) {
		// Auth endpoint — returns current user info from JWT (public with optional auth)
		r.With(auth.PublicAuthMiddleware(keycloakCfg)).Get("/auth/me", auth.AuthHandler())
	})

	// Protected API v1 routes (JWT required)
	r.Route("/api/v1/protected", func(r chi.Router) {
		r.Use(auth.JWTAuthMiddleware(keycloakCfg))

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

		// AI Assistant Framework Module (V048)
		aiAssistantHandler := handlers.NewAIHandler(database.DB)
		aiAssistantHandler.RegisterRoutes(r)

		// Project Management Module
		pmHandler := handlers.NewPMHandler(database.DB)
		pmHandler.RegisterRoutes(r)

		// Document Control Module
		docControlHandler := handlers.NewDocControlHandler(database.DB)
		docControlHandler.RegisterRoutes(r)

		// Schedule Management Module
		scheduleHandler := handlers.NewScheduleHandler(database.DB)
		scheduleHandler.RegisterRoutes(r)

		// Equipment Management Module
		equipmentHandler := handlers.NewEquipmentHandler(database.DB)
		equipmentHandler.RegisterRoutes(r)

		// HSE Module
		hseHandler := handlers.NewHSEHandler(database.DB)
		hseHandler.RegisterRoutes(r)

		// Quality Module
		qualityHandler := handlers.NewQualityHandler(database.DB)
		qualityHandler.RegisterRoutes(r)

		// GIS & Survey Module
		gisHandler := handlers.NewGISHandler(database.DB)
		gisHandler.RegisterRoutes(r)

		// Risk Management Module
		riskHandler := handlers.NewRiskHandler(database.DB)
		riskHandler.RegisterRoutes(r)

		// TBM Management Module
		tbmHandler := handlers.NewTBMHandler(database.DB)
		tbmHandler.RegisterRoutes(r)

		// Ring Builder & Segment Tracking Module
		ringBuilderHandler := handlers.NewRingBuilderHandler(database.DB)
		ringBuilderHandler.RegisterRoutes(r)

		// NATM & Microtunnelling Module
		natmHandler := handlers.NewNATMHandler(database.DB)
		natmHandler.RegisterRoutes(r)

		// Change Management Module
		changeHandler := handlers.NewChangeHandler(database.DB)
		changeHandler.RegisterRoutes(r)

		// EVM Module (V025)
		evmHandler := handlers.NewEVMHandler(database.DB)
		evmHandler.RegisterRoutes(r)

		// P6 Connector Module (V026)
		p6Handler := handlers.NewP6Handler(database.DB)
		p6Handler.RegisterRoutes(r)

		// Funding Module (V027)
		fundingHandler := handlers.NewFundingHandler(database.DB)
		fundingHandler.RegisterRoutes(r)

		// Neo4j + Kafka Module (V028)
		neo4jKafkaHandler := handlers.NewNeo4jKafkaHandler(database.DB)
		neo4jKafkaHandler.RegisterRoutes(r)

		// Laboratory Module (V029)
		labHandler := handlers.NewLaboratoryHandler(database.DB)
		labHandler.RegisterRoutes(r)

		// Permits Module (V030)
		permitsHandler := handlers.NewPermitsHandler(database.DB)
		permitsHandler.RegisterRoutes(r)

		// Insurance Module (V031)
		insuranceHandler := handlers.NewInsuranceHandler(database.DB)
		insuranceHandler.RegisterRoutes(r)

		// Fleet Module (V032)
		fleetHandler := handlers.NewFleetHandler(database.DB)
		fleetHandler.RegisterRoutes(r)

		// Time & Attendance Module (V034)
		attendanceHandler := handlers.NewAttendanceHandler(database.DB)
		attendanceHandler.RegisterRoutes(r)

		// Training & Certifications Module (V035)
		trainingHandler := handlers.NewTrainingHandler(database.DB)
		trainingHandler.RegisterRoutes(r)

		// Segment Factory Module (V036)
		segFactoryHandler := handlers.NewSegmentFactoryHandler(database.DB)
		segFactoryHandler.RegisterRoutes(r)

		// Shaft Management Module (V037)
		shaftHandler := handlers.NewShaftHandler(database.DB)
		shaftHandler.RegisterRoutes(r)

		// Cross Passage + Geology Module (V038)
		cpGeoHandler := handlers.NewCPGeoHandler(database.DB)
		cpGeoHandler.RegisterRoutes(r)

		// Settlement, Grouting, Ventilation Module (V039)
		sgvHandler := handlers.NewSGVHandler(database.DB)
		sgvHandler.RegisterRoutes(r)

		// Instrumentation, Dewatering, TBM Maintenance Module (V040)
		tunnelSvcHandler := handlers.NewTBMServiceHandler(database.DB)
		tunnelSvcHandler.RegisterRoutes(r)

		// Retention, Guarantees, Multi-Currency Module (V041)
		retentionHandler := handlers.NewRetentionHandler(database.DB)
		retentionHandler.RegisterRoutes(r)

		// Audit Trail, Tax, Stakeholders, ESG, Carbon Module (V042-V043)
		auditHandler := handlers.NewAuditHandler(database.DB)
		auditHandler.RegisterRoutes(r)

		// Reporting Builder Module (V044)
		reportHandler := handlers.NewReportHandler(database.DB)
		reportHandler.RegisterRoutes(r)

		// Asset Management Module (V045)
		assetHandler := handlers.NewAssetHandler(database.DB)
		assetHandler.RegisterRoutes(r)

		// Performance Benchmarking Module (V046)
		benchmarkHandler := handlers.NewBenchmarkHandler(database.DB)
		benchmarkHandler.RegisterRoutes(r)

		// Integration Framework Module (V047)
		integrationHandler := handlers.NewIntegrationHandler(database.DB)
		integrationHandler.RegisterRoutes(r)

		// Audit & Compliance Module (V049)
		auditComplianceHandler := handlers.NewAuditComplianceHandler(database.DB)
		auditComplianceHandler.RegisterRoutes(r)

		// Mobile API, Notifications, Activity, Comments (V050)
		mobileHandler := handlers.NewMobileHandler(database.DB)
		mobileHandler.RegisterRoutes(r)

		// Tunnel Logistics Module (V051)
		tunnelLogisticsHandler := handlers.NewTunnelLogisticsHandler(database.DB)
		tunnelLogisticsHandler.RegisterRoutes(r)
	})

	// Legacy /api/v1 routes (backward compatible, also protected)
	r.Route("/api/v1/legacy", func(r chi.Router) {
		// Use the same pattern — for full migration, replace with protected routes
		r.Use(auth.JWTAuthMiddleware(keycloakCfg))

		// Admin-only routes example
		r.With(auth.RequiredRoleMiddleware("admin")).Get("/admin/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "admin-ok"})
		})
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
