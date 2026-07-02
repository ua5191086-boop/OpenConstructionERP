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

	// Public API v1 routes (no auth required for GET, POST/PUT/DELETE require auth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/auth/me", auth.AuthHandler())

		// sqlDB is the standard *sql.DB for legacy handlers
		sqlDB := database.StdDB()

		// Projects (direct, schema-matched)
		projectHandler := handlers.NewProjectHandler(sqlDB)
		projectHandler.RegisterRoutes(r)

		// BOQ Module
		boqHandler := handlers.NewBOQHandler(sqlDB)
		boqHandler.RegisterRoutes(r)

		// Tenders Module
		tendersHandler := handlers.NewTendersHandler(sqlDB)
		tendersHandler.RegisterRoutes(r)

		// Contracts Module
		contractsHandler := handlers.NewContractsHandler(sqlDB)
		contractsHandler.RegisterRoutes(r)

		// HR Module
		hrHandler := handlers.NewHRHandler(sqlDB)
		hrHandler.RegisterRoutes(r)

		// Finance Module
		financeHandler := handlers.NewFinanceHandler(sqlDB)
		financeHandler.RegisterRoutes(r)

		// Financial Consolidation & Loans (V054-V055)
		financialExtHandler := handlers.NewFinancialHandler(database.DB)
		financialExtHandler.RegisterRoutes(r)

		// Procurement Module
		procurementHandler := handlers.NewProcurementHandler(sqlDB)
		procurementHandler.RegisterRoutes(r)

		// BIM Module
		bimHandler := handlers.NewBIMHandler(sqlDB)
		bimHandler.RegisterRoutes(r)

		// Document Control Module
		docControlHandler := handlers.NewDocControlHandler(sqlDB)
		docControlHandler.RegisterRoutes(r)

		// Schedule Management Module
		scheduleHandler := handlers.NewScheduleHandler(sqlDB)
		scheduleHandler.RegisterRoutes(r)

		// Equipment Management Module
		equipmentHandler := handlers.NewEquipmentHandler(sqlDB)
		equipmentHandler.RegisterRoutes(r)

		// HSE Module
		hseHandler := handlers.NewHSEHandler(sqlDB)
		hseHandler.RegisterRoutes(r)

		// Quality Module
		qualityHandler := handlers.NewQualityHandler(sqlDB)
		qualityHandler.RegisterRoutes(r)

		// GIS & Survey Module
		gisHandler := handlers.NewGISHandler(sqlDB)
		gisHandler.RegisterRoutes(r)

		// Risk Management Module
		riskHandler := handlers.NewRiskHandler(sqlDB)
		riskHandler.RegisterRoutes(r)

		// TBM Management Module
		tbmHandler := handlers.NewTBMHandler(sqlDB)
		tbmHandler.RegisterRoutes(r)

		// Ring Builder & Segment Tracking Module
		ringBuilderHandler := handlers.NewRingBuilderHandler(sqlDB)
		ringBuilderHandler.RegisterRoutes(r)

		// NATM & Microtunnelling Module
		natmHandler := handlers.NewNATMHandler(sqlDB)
		natmHandler.RegisterRoutes(r)

		// Change Management Module
		changeHandler := handlers.NewChangeHandler(sqlDB)
		changeHandler.RegisterRoutes(r)

		// EVM Module (V025)
		evmHandler := handlers.NewEVMHandler(sqlDB)
		evmHandler.RegisterRoutes(r)

		// P6 Connector Module (V026)
		p6Handler := handlers.NewP6Handler(sqlDB)
		p6Handler.RegisterRoutes(r)

		// Funding Module (V027)
		fundingHandler := handlers.NewFundingHandler(sqlDB)
		fundingHandler.RegisterRoutes(r)

		// Neo4j + Kafka Module (V028)
		neo4jKafkaHandler := handlers.NewNeo4jKafkaHandler(sqlDB)
		neo4jKafkaHandler.RegisterRoutes(r)

		// Laboratory Module (V029)
		labHandler := handlers.NewLaboratoryHandler(sqlDB)
		labHandler.RegisterRoutes(r)

		// Permits Module (V030)
		permitsHandler := handlers.NewPermitsHandler(sqlDB)
		permitsHandler.RegisterRoutes(r)

		// Insurance Module (V031)
		insuranceHandler := handlers.NewInsuranceHandler(sqlDB)
		insuranceHandler.RegisterRoutes(r)

		// Fleet Module (V032)
		fleetHandler := handlers.NewFleetHandler(sqlDB)
		fleetHandler.RegisterRoutes(r)

		// Time & Attendance Module (V034)
		attendanceHandler := handlers.NewAttendanceHandler(sqlDB)
		attendanceHandler.RegisterRoutes(r)

		// Training & Certifications Module (V035)
		trainingHandler := handlers.NewTrainingHandler(sqlDB)
		trainingHandler.RegisterRoutes(r)

		// Segment Factory Module (V036)
		segFactoryHandler := handlers.NewSegmentFactoryHandler(sqlDB)
		segFactoryHandler.RegisterRoutes(r)

		// Shaft Management Module (V037)
		shaftHandler := handlers.NewShaftHandler(sqlDB)
		shaftHandler.RegisterRoutes(r)

		// Cross Passage + Geology Module (V038)
		cpGeoHandler := handlers.NewCPGeoHandler(sqlDB)
		cpGeoHandler.RegisterRoutes(r)

		// Settlement, Grouting, Ventilation Module (V039)
		sgvHandler := handlers.NewSGVHandler(sqlDB)
		sgvHandler.RegisterRoutes(r)

		// Instrumentation, Dewatering, TBM Maintenance Module (V040)
		tunnelSvcHandler := handlers.NewTBMServiceHandler(sqlDB)
		tunnelSvcHandler.RegisterRoutes(r)

		// Tunnel Logistics Module (V051-V053)
		tunnelLogisticsHandler := handlers.NewTunnelLogisticsHandler(database.DB)
		tunnelLogisticsHandler.RegisterRoutes(r)

		// AI Assistant Framework Module (V048)
		aiAssistantHandler := handlers.NewAIHandler(database.DB)
		aiAssistantHandler.RegisterRoutes(r)

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
	})

	// Legacy /api/v1 routes (backward compatible, also protected)
	r.Route("/api/v1/legacy", func(r chi.Router) {
		r.Use(auth.JWTAuthMiddleware(keycloakCfg))
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
