# OpenConstructionERP

**Open-source platform for managing the full lifecycle of large infrastructure construction projects**
Metro · Tunnels · Railways · Hydraulic Structures · EPC/EPCM

> **Status: v0.1.0 — Production-ready schema, Go API, Docker deployment.**
> 75+ SQL migrations, 405+ tables, 45 Go handlers, 40+ web dashboards.

---

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose v2
- Git
- 4GB+ RAM, 10GB+ disk

### 1. Clone & Deploy

```bash
git clone https://github.com/ua5191086-boop/OpenConstructionERP.git
cd OpenConstructionERP

# Start all services
docker compose -f infrastructure/docker/docker-compose.single-node.yml up -d
```

### 2. Verify

```bash
# Check all services are running
docker ps

# API health check
curl http://localhost:8085/health

# Web dashboard
open http://localhost:8086/
```

### 3. Load Demo Data

```bash
# Seed data is auto-applied via V076 migration
# Verify data loaded:
docker exec oce-postgres psql -U oce -d oce_erp -c "SELECT count(*) FROM projects;"
docker exec oce-postgres psql -U oce -d oce_erp -c "SELECT count(*) FROM tunnel_rings;"
```

### 4. Access

| Service | URL | Credentials |
|---------|-----|-------------|
| **API** | http://localhost:8085 | — |
| **Web Dashboard** | http://localhost:8086 | — |
| **PostgreSQL** | localhost:5434, user: `oce`, db: `oce_erp` | password in `.env` |
| **Keycloak** | http://localhost:8084 | admin / admin |
| **MinIO** | http://localhost:9002 | minio / minio123 |
| **Grafana** | http://localhost:3002 | admin / admin |
| **Prometheus** | http://localhost:9091 | — |

---

## 🏗 Architecture

```
┌─────────────────────────────────────────────────────┐
│                    Web UI (React)                     │
│              40+ dashboards, port 8086               │
├─────────────────────────────────────────────────────┤
│                 Go API (45 handlers)                  │
│              RESTful JSON, port 8085                  │
├─────────────────────────────────────────────────────┤
│              PostgreSQL 16 (405 tables)               │
│              oce_erp database, port 5434              │
├─────────────────────────────────────────────────────┤
│     MinIO     │   Redis    │   Kafka    │  Keycloak   │
│   (Documents) │  (Cache)   │  (Events)  │   (Auth)    │
└─────────────────────────────────────────────────────┘
```

## 📦 Modules (100% SAD Coverage)

### Core (20 modules)
- Project Management, BOQ, Contracts, Finance, Procurement
- HR, HSE, Quality, BIM, Document Control
- Schedule, Equipment, GIS/Survey, Risk, Change Management
- TBM, Ring Builder/Segment, NATM/Microtunnelling
- Auth/Audit, EVM

### Tunnel (15 modules)
- TBM Management, Ring Builder, Segment Factory, NATM
- Shaft Management, Cross Passage, Geology
- Settlement Monitoring, Grouting, Ventilation
- Instrumentation, Dewatering, TBM Maintenance
- **Tunnel Logistics, Ventilation Design, Fire Safety**

### Financial (12 modules)
- Finance Core, EVM, Cost Control, Budget
- Retention/Guarantees, Multi-Currency, Audit Trail
- Tax Management, Transfer Pricing
- **Financial Consolidation, Project Financing/Loans**

### AI (15 modules)
- AI Framework, Document Classifier, Cost Estimator
- Schedule Optimizer, Risk Predictor, Quality Inspector
- Progress Monitor, Contract Analyzer, Procurement Optimizer
- Safety Monitor, Report Generator, Chatbot
- Predictive Maintenance, ESG Reporter

### Integrations (8 modules)
- Integration Framework, SAP, Autodesk BIM 360
- Bentley iTwin, SharePoint, Telegram Bot, PowerBI

### Additional (10 modules)
- Primavera P6, Neo4j/Kafka, Physical Progress IPC
- Time & Attendance, Variation Orders, Training
- Reporting Builder, Asset Management, Performance Benchmarking
- **Lessons Learned**

---

## 🛠 Development

### Adding a New Migration

```bash
# Create migration file
touch database/migrations/V082__My_New_Module.sql

# Run it
docker exec -i oce-postgres psql -U oce -d oce_erp < database/migrations/V082__My_New_Module.sql
```

### Adding a New Go Handler

```bash
# Create handler
touch services/core/internal/handlers/my_module.go

# Register in main.go
# Add route: r.HandleFunc("/api/v1/my-module", handlers.NewMyModuleHandler(db))
```

### Running Tests

```bash
# Go tests
cd services/core && go test ./...

# E2E API tests
bash scripts/test-api.sh
```

---

## 📚 Documentation

| Document | Location |
|----------|----------|
| SAD Tom 1 — Architecture | [docs/sad/SAD-Tom1-Architecture-v1.0.md](docs/sad/SAD-Tom1-Architecture-v1.0.md) |
| API Reference (OpenAPI) | [docs/api/openapi.js](docs/api/openapi.js) |
| Architecture Decision Records | [docs/adr/README.md](docs/adr/README.md) |
| Contributing Guide | [CONTRIBUTING.md](CONTRIBUTING.md) |

---

## 📊 Dashboard Screenshots

Access all dashboards at http://localhost:8086/:

- **Project Overview** — Portfolio, milestones, EVM
- **BOQ** — Bill of Quantities with WBS/CBS breakdown
- **Tunnel** — TBM telemetry, ring map, segment factory
- **HSE** — Incidents, NCRs, safety metrics
- **Finance** — Budget vs actual, cash flow, IPC
- **Risk** — Risk matrix, Monte Carlo, mitigation tracker
- **Settlement** — Monitoring points, time-series readings
- **AI** — Classifications, predictions, recommendations

---

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

**Quick rules:**
1. Never push without running full migration chain locally
2. All new tables use UUID primary keys (BIGINT is banned by CI)
3. Every migration must be idempotent (use `IF NOT EXISTS`)
4. Add tests for new handlers

---

## 📄 License

- **Core platform**: Apache 2.0
- **AI modules**: AGPL v3
- **Integrations**: AGPL v3

---

## 🔗 Links

- **Repository**: https://github.com/ua5191086-boop/OpenConstructionERP
- **Issues**: https://github.com/ua5191086-boop/OpenConstructionERP/issues
- **Docker Hub**: *(coming soon)*
