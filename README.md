# OpenConstructionERP

**Open-Source ERP Platform for Infrastructure Construction Projects**

Metro · Tunnels · Railways · Hydraulic Structures · EPC/EPCM

---

## Overview

OpenConstructionERP is a next-generation digital platform for managing the full lifecycle of large infrastructure construction projects — from lead to operation. Built on modern architecture with AI, BIM, GIS, and real-time analytics.

## Key Features

- **Project Management** — WBS, milestones, portfolio
- **Cost Management** — Budget, EVM, CBS, cash flow, multi-currency
- **Schedule Management** — CPM, P6-compatible, resource leveling
- **Document Control** — RFI, NCR, submittals, transmittals, correspondence
- **Contract Management** — EPC/EPCM, subcontracts, variations, claims
- **TBM Module** — Real-time telemetry, ring building, segment tracking
- **BIM Module** — IFC viewer, clash detection, 4D/5D/6D/7D
- **AI Module** — Copilot, document analysis, delay prediction, risk analysis
- **GIS Module** — Maps, geolocation, survey data, drone orthophoto
- **Finance** — Invoicing, payments, retention, guarantees, bank loans
- **HSE** — Incidents, permits, safety analytics
- **Quality** — ITP, NCR, inspection, test records
- **Equipment** — TBM, cranes, fleet, maintenance, fuel
- **Procurement** — RFQ, PO, supplier management, warehouse
- **Reporting** — Dashboards, KPI, Power BI export, custom reports

## Architecture

```
Frontend:    React 18 + TypeScript + Next.js 14
Backend:     NestJS (Node.js) + FastAPI (Python) + Go + Rust
Database:    PostgreSQL 16 + TimescaleDB + Neo4j + Elasticsearch
Storage:     MinIO (S3-compatible)
Message:     RabbitMQ + Kafka + NATS
AI/ML:       LangChain + LlamaIndex + vLLM + Qdrant
BIM:         IFC.js + IfcOpenShell + Three.js
GIS:         MapLibre GL + CesiumJS + GeoServer
Auth:        Keycloak (OAuth 2.0 / OIDC / SAML)
Deployment:  Docker + Kubernetes + Helm
CI/CD:       GitHub Actions + ArgoCD
Monitoring:  Prometheus + Grafana + Loki + OpenTelemetry
```

## Modules

100+ modules organized into domains:
- Core (20) — Project, Cost, Schedule, Document, Contract, etc.
- Tunnel (15) — TBM, NATM, Microtunnelling, Rings, Segments, etc.
- BIM (10) — IFC, Clash, 4D-7D, Digital Twin
- AI (15) — Copilot, Planner, Risk, Claims, Vision
- Finance (12) — Budget, EVM, Cash Flow, Funding, Invoicing
- Document Control (10) — RFI, NCR, ITP, Submittals, Reports
- Integration (8) — Primavera, SAP, Autodesk, Bentley, Telegram

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Node.js 22+
- Python 3.12+
- Go 1.22+

### Quick Start
```bash
git clone https://github.com/ua5191086-boop/OpenConstructionERP.git
cd OpenConstructionERP
docker compose -f infrastructure/docker/docker-compose.dev.yml up -d
```

## License

GNU Affero General Public License v3.0 (AGPL-3.0)

## Roadmap

| Phase | Timeline | Focus |
|-------|----------|-------|
| MVP | 6-9 months | Core modules + BOQ + Schedule + Cost |
| Beta | 9-12 months | TBM + BIM + AI + GIS |
| Release | 12-18 months | All modules + Enterprise features |
| Enterprise | 18-24 months | Digital Twin + Global scale |

## Contact

Project Owner: Ruslan Sarybaev
Architecture: OpenConstructionERP Team
