# OpenConstructionERP — Software Architecture Document

## Executive Summary

**Видение:** Цифровая платформа мирового стандарта для управления жизненным циклом инфраструктурных проектов — от лида до эксплуатации.

**Архитектурный принцип:** Palantir Foundry для строительства — единая операционная система проекта (Project Operating System), где данные, BIM, AI и финансы существуют в одном semantic layer.

**Ключевое отличие от существующих ERP:** Не «учётная система с прикрученным BIM», а **data-first платформа**, где каждая сущность (бетонный блок, RFI, TBM-сегмент, бюджетная строка) — first-class citizen в графе знаний.

---

## 1. Архитектура системы (High-Level)

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CLIENT LAYER                                  │
│  Web (React 18 + TypeScript) │ Mobile (React Native) │ Desktop (Tauri)│
└──────────────────────────────┬──────────────────────────────────────┘
                               │ GraphQL / REST / WebSocket
┌──────────────────────────────┴──────────────────────────────────────┐
│                        API GATEWAY (Kong / Envoy)                   │
│  Auth │ Rate Limit │ Routing │ Caching │ WAF                        │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
┌──────────────────────────────┴──────────────────────────────────────┐
│                    SERVICE MESH (Istio / Linkerd)                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Core     │ │ BIM      │ │ AI       │ │ GIS      │ │ Document │  │
│  │ Services │ │ Services │ │ Services │ │ Services │ │ Services │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Finance  │ │ Equipment│ │ HSE      │ │ Planning │ │ Analytics│  │
│  │ Services │ │ Services │ │ Services │ │ Services │ │ Services │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │ Contract │ │ Procure  │ │ Quality  │ │ HR       │ │ Notific  │  │
│  │ Services │ │ Services │ │ Services │ │ Services │ │ Services │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│                                                                     │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
┌──────────────────────────────┴──────────────────────────────────────┐
│                    DATA & INFRASTRUCTURE LAYER                       │
│                                                                     │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │PostgreSQL│ │ Timescale│ │ Neo4j    │ │ Elastic  │ │ MinIO    │  │
│  │ (OLTP)   │ │ (Metrics)│ │ (Graph)  │ │ (Search) │ │ (S3)     │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐  │
│  │Redis     │ │RabbitMQ  │ │Kafka     │ │Trino     │ │Spark     │  │
│  │(Cache)   │ │(Queue)   │ │(Stream)  │ │(Query)   │ │(Batch)   │  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.1 Frontend
- **Framework:** React 18 + TypeScript + Next.js 14 (SSR для публичных страниц)
- **State:** Zustand + React Query (TanStack Query)
- **UI:** Radix UI + Tailwind CSS + Tremor (дашборды)
- **BIM Viewer:** Three.js + IFC.js (web-ifc-viewer)
- **GIS Map:** MapLibre GL JS + CesiumJS
- **Graphs:** D3.js + Cytoscape.js (граф знаний)
- **Forms:** React Hook Form + Zod
- **i18n:** react-i18next (многоязычность)
- **PWA:** Workbox
- **Testing:** Vitest + Playwright + Storybook

### 1.2 Backend (Microservices)
- **Runtime:** Node.js 22 (NestJS) для API-сервисов + Python 3.12 (FastAPI) для AI/ML
- **Go** для high-throughput сервисов (BIM-конвертация, GIS-рендеринг)
- **Rust** для критичных по производительности модулей (segment tracking real-time)
- **Communication:** gRPC (inter-service) + GraphQL (client-facing) + Kafka (events)
- **API Style:** GraphQL Federation (Apollo Federation) + REST для внешних интеграций

### 1.3 Database
- **Primary OLTP:** PostgreSQL 16 + pg_partman (шардирование по проектам)
- **Time-series:** TimescaleDB (TBM data, sensors, IoT)
- **Graph:** Neo4j 5 (knowledge graph, project relationships)
- **Search:** Elasticsearch 8 (полнотекстовый + геопоиск)
- **Object Storage:** MinIO (S3-compatible) — документы, BIM-файлы, чертежи
- **Vector DB:** Qdrant (AI embeddings, semantic search)
- **Cache:** Redis 7 (session, rate limit, hot data)
- **Data Lake:** Apache Iceberg on MinIO (аналитика, ML)

### 1.4 Message Broker
- **RabbitMQ:** Task queues (email, notifications, PDF generation)
- **Apache Kafka:** Event sourcing, CDC (Debezium), stream processing
- **NATS:** Real-time TBM telemetry, IoT sensor data

### 1.5 Authentication & Authorization
- **Auth:** Keycloak (OAuth 2.0 / OIDC / SAML)
- **MFA:** TOTP + WebAuthn
- **RBAC/ABAC:** Custom policy engine (Casbin + OPA)
- **SSO:** Azure AD, Google Workspace, SAML 2.0

### 1.6 AI Layer
- **LLM:** Local (Llama 3 / Mistral Large) via vLLM + OpenAI API-compatible
- **RAG:** LangChain + LlamaIndex + Qdrant
- **ML Pipeline:** MLflow + Kubeflow
- **Computer Vision:** YOLOv8 (site safety, progress tracking)
- **NLP:** spaCy + transformers (contract analysis, claims)
- **Time-series ML:** Prophet + Kats (forecasting, delay prediction)

### 1.7 GIS Layer
- **Engine:** CesiumJS + MapLibre GL
- **Server:** GeoServer + pgRouting
- **Formats:** GeoJSON, MBTiles, 3D Tiles, CityGML
- **Data:** OpenStreetMap + proprietary survey data + drone orthophoto

### 1.8 BIM Layer
- **Core:** IFC (Industry Foundation Classes) — IFC2x3 / IFC4x3
- **Viewer:** web-ifc-viewer (Three.js-based)
- **BCF:** BCF API (BIM Collaboration Format)
- **Conversion:** IfcOpenShell + xBIM Toolkit
- **Digital Twin:** Azure Digital Twins / Eclipse Ditto
- **4D/5D/6D/7D:** Custom engine on top of IFC + schedule + cost + sustainability

### 1.9 Deployment
- **Container:** Docker + Docker Compose (dev) / Kubernetes (prod)
- **K8s:** Rancher / OpenShift
- **Helm Charts:** Все микросервисы
- **Service Mesh:** Istio
- **API Gateway:** Kong / Envoy
- **CDN:** Cloudflare / Fastly
- **DNS:** Cloudflare

### 1.10 CI/CD
- **Pipeline:** GitHub Actions + ArgoCD (GitOps)
- **Registry:** GitHub Container Registry + Harbor
- **Testing:** Unit (Vitest/Jest) + Integration (Testcontainers) + E2E (Playwright)
- **Security:** Trivy + Snyk + SonarQube + Dependabot
- **Secrets:** HashiCorp Vault + External Secrets Operator

### 1.11 Monitoring & Observability
- **Metrics:** Prometheus + Grafana
- **Logging:** Loki + OpenTelemetry + Fluentd
- **Tracing:** Jaeger / Tempo
- **Alerting:** Alertmanager + PagerDuty
- **APM:** Grafana Faro (frontend) + OpenTelemetry (backend)
- **Uptime:** Checkly / Grafana Synthetic Monitoring

### 1.12 Security
- **WAF:** ModSecurity + Coraza
- **DDoS:** Cloudflare
- **Secrets:** Vault
- **Audit:** AuditLog (immutable, append-only)
- **Encryption:** AES-256 at rest, TLS 1.3 in transit
- **DLP:** Data Loss Prevention layer
- **Compliance:** ISO 27001, SOC 2, GDPR

---

## 2. Полный перечень модулей (100+)

### Core Modules (20)
| # | Модуль | Назначение |
|---|--------|-----------|
| 1 | **Project Management** | Управление проектами, WBS, milestones |
| 2 | **Lead Management** | CRM, тендеры, pre-qualification |
| 3 | **Contract Management** | EPC/EPCM, subcontracts, amendments |
| 4 | **Document Control** | RFI, NCR, submittals, transmittals |
| 5 | **Cost Management** | Budget, forecast, EVM, CBS |
| 6 | **Schedule Management** | P6-compatible, CPM, Gantt, resource |
| 7 | **Procurement** | RFQ, PO, supplier management |
| 8 | **Warehouse** | Inventory, material tracking, barcode |
| 9 | **Equipment Management** | TBM, cranes, fleet, maintenance |
| 10 | **HR & Payroll** | Personnel, timesheets, payroll |
| 11 | **HSE** | Safety, incidents, permits, audits |
| 12 | **Quality** | ITP, NCR, inspection, test records |
| 13 | **BIM** | IFC viewer, model management, clash |
| 14 | **GIS** | Maps, geolocation, survey data |
| 15 | **Finance** | Invoicing, payments, cash flow |
| 16 | **Risk Management** | Risk register, mitigation, Monte Carlo |
| 17 | **Change Management** | VO, CO, claims, variations |
| 18 | **Reporting** | Dashboards, KPI, Power BI export |
| 19 | **AI Assistant** | Copilot, chat, document analysis |
| 20 | **Integration Hub** | API gateway, connectors, ETL |

### Tunnel Construction Modules (15)
| # | Модуль | Назначение |
|---|--------|-----------|
| 21 | **TBM Management** | EPB/Slurry TBM telemetry, parameters |
| 22 | **Ring Builder** | Ring assembly, segment tracking |
| 23 | **Segment Tracking** | Production, curing, transport, install |
| 24 | **Segment Factory** | Factory management, QC, stock |
| 25 | **NATM** | Sequential excavation, shotcrete |
| 26 | **Microtunnelling** | Pipe jacking, thrust, lubrication |
| 27 | **Shaft Management** | Launch/reception shafts |
| 28 | **Cross Passage** | Design, construction, waterproofing |
| 29 | **Geology** | GPR, boreholes, face mapping |
| 30 | **Instrumentation** | Sensors, monitoring, alerts |
| 31 | **Settlement** | Monitoring, triggers, mitigation |
| 32 | **Grouting** | Backfill, consolidation, records |
| 33 | **Ventilation** | Tunnel ventilation design/monitoring |
| 34 | **Dewatering** | Groundwater control |
| 35 | **TBM Maintenance** | Cutterhead, seals, gearbox |

### BIM Modules (10)
| # | Модуль | Назначение |
|---|--------|-----------|
| 36 | **Model Viewer** | IFC/BIM viewer in browser |
| 37 | **Clash Detection** | Automated clash analysis |
| 38 | **4D Simulation** | Schedule-linked model |
| 39 | **5D Cost** | Model-linked cost |
| 40 | **6D Sustainability** | Carbon, energy analysis |
| 41 | **7D Facility Management** | O&M data |
| 42 | **BCF Collaboration** | Issue tracking on model |
| 43 | **Model Comparison** | Version diff |
| 44 | **Quantity Takeoff** | Automated QTO from IFC |
| 45 | **Digital Twin** | Real-time mirror of physical asset |

### AI Modules (15)
| # | Модуль | Назначение |
|---|--------|-----------|
| 46 | **AI Planner** | Auto-generate schedules from specs |
| 47 | **AI Scheduler** | Optimize resource allocation |
| 48 | **AI Cost Engineer** | Estimate from historical data |
| 49 | **AI Contract Manager** | Clause extraction, compliance |
| 50 | **AI Risk Manager** | Predict risks from project data |
| 51 | **AI Procurement** | Supplier matching, price prediction |
| 52 | **AI Document Analyzer** | OCR, classification, extraction |
| 53 | **AI Claims** | Auto-detect claim events |
| 54 | **AI Delay Analysis** | Root cause, float consumption |
| 55 | **AI Forecast** | Cost/schedule prediction |
| 56 | **AI Chat** | Natural language query on project data |
| 57 | **AI Knowledge Graph** | Entity extraction, relationships |
| 58 | **AI Copilot** | Context-aware assistant |
| 59 | **AI Vision** | Site camera analysis, safety |
| 60 | **AI Quality** | Defect detection from photos |

### Financial Modules (12)
| # | Модуль | Назначение |
|---|--------|-----------|
| 61 | **Budget Management** | Project budget, revisions |
| 62 | **Cash Flow** | Forecast, actual, variance |
| 63 | **EVM** | Earned Value Management |
| 64 | **Cost Breakdown** | CBS, cost codes |
| 65 | **Funding** | Bank loans, ECA, investor tracking |
| 66 | **Invoicing** | Progress billing, milestone billing |
| 67 | **Payments** | Supplier/subcontractor payments |
| 68 | **Retention** | Retention tracking, release |
| 69 | **Guarantees** | Performance bonds, bank guarantees |
| 70 | **Tax** | VAT, withholding tax, cross-border |
| 71 | **Audit Trail** | Immutable financial log |
| 72 | **Multi-Currency** | FX, hedging |

### Document Control Modules (10)
| # | Модуль | Назначение |
|---|--------|-----------|
| 73 | **RFI** | Request for Information workflow |
| 74 | **RFQ** | Request for Quotation |
| 75 | **NCR** | Non-Conformance Report |
| 76 | **ITP** | Inspection & Test Plan |
| 77 | **Method Statement** | Approval workflow |
| 78 | **Shop Drawings** | Review, approval cycle |
| 79 | **Submittals** | Material/equipment approval |
| 80 | **Correspondence** | Letters, emails, transmittals |
| 81 | **Minutes of Meeting** | Action items, distribution |
| 82 | **Daily Reports** | Site diary, progress photos |

### Integration Modules (8)
| # | Модуль | Назначение |
|---|--------|-----------|
| 83 | **Primavera Connector** | Bi-directional P6 sync |
| 84 | **SAP Connector** | Financial/MM integration |
| 85 | **Autodesk Connector** | ACC/BIM 360 sync |
| 86 | **Bentley Connector** | iTwin/ProjectWise sync |
| 87 | **SharePoint Connector** | Document sync |
| 88 | **Nextcloud Connector** | Self-hosted file sync |
| 89 | **Telegram/WhatsApp** | Notification, chat |
| 90 | **Power BI Connector** | Live data export |

### Additional Modules (10)
| # | Модуль | Назначение |
|---|--------|-----------|
| 91 | **Survey** | Topographic, geodetic |
| 92 | **Laboratory** | Material testing, concrete |
| 93 | **Permits** | Regulatory approvals |
| 94 | **Insurance** | Policy management, claims |
| 95 | **Stakeholder** | Community relations |
| 96 | **Sustainability** | ESG, carbon tracking |
| 97 | **Training** | Competency, certifications |
| 98 | **Fleet** | Vehicle tracking, fuel |
| 99 | **Time & Attendance** | Biometric, gate access |
| 100 | **Offshore Module** | Marine works, dredging |

---

## 3. Структура базы данных (ключевые таблицы)

### 3.1 Core Schema

```sql
-- Projects
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    project_type VARCHAR(50) NOT NULL, -- 'EPC','EPCM','Design-Build','BOT'
    status VARCHAR(20) NOT NULL DEFAULT 'lead', -- lead, tender, awarded, execution, closeout, completed
    client_id UUID REFERENCES organizations(id),
    consultant_id UUID REFERENCES organizations(id),
    contract_value DECIMAL(20,2),
    currency VARCHAR(3) DEFAULT 'USD',
    country VARCHAR(2),
    city VARCHAR(100),
    lat DECIMAL(10,7),
    lng DECIMAL(10,7),
    start_date DATE,
    end_date DATE,
    duration_days INTEGER,
    wbs_json JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Organizations (Clients, Contractors, Suppliers, Consultants)
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50), -- client, contractor, supplier, consultant, bank
    tax_id VARCHAR(50),
    registration_country VARCHAR(2),
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(255),
    duns_number VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(50),
    avatar_url TEXT,
    organization_id UUID REFERENCES organizations(id),
    is_active BOOLEAN DEFAULT true,
    mfa_enabled BOOLEAN DEFAULT false,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT false,
    permissions JSONB NOT NULL DEFAULT '{}'
);

-- Project Members
CREATE TABLE project_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id),
    start_date DATE,
    end_date DATE,
    allocation_percent DECIMAL(5,2),
    UNIQUE(project_id, user_id)
);

-- WBS (Work Breakdown Structure)
CREATE TABLE wbs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES wbs(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    level INTEGER NOT NULL,
    sort_order INTEGER,
    path ltree, -- materialized path
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- CBS (Cost Breakdown Structure)
CREATE TABLE cbs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES cbs(id),
    level INTEGER,
    path ltree,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Budget Lines
CREATE TABLE budget_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    wbs_id UUID REFERENCES wbs(id),
    cbs_id UUID REFERENCES cbs(id),
    original_amount DECIMAL(20,2),
    revised_amount DECIMAL(20,2),
    committed_amount DECIMAL(20,2),
    spent_amount DECIMAL(20,2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Schedule Activities
CREATE TABLE schedule_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    wbs_id UUID REFERENCES wbs(id),
    activity_id VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    original_duration INTEGER,
    remaining_duration INTEGER,
    actual_duration INTEGER,
    early_start DATE,
    early_finish DATE,
    late_start DATE,
    late_finish DATE,
    actual_start DATE,
    actual_finish DATE,
    total_float INTEGER,
    free_float INTEGER,
    status VARCHAR(20), -- not_started, in_progress, completed, delayed
    percent_complete DECIMAL(5,2),
    constraint_type VARCHAR(20), -- FS, SS, FF, SF
    lag_days INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- TBM Data
CREATE TABLE tbm_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    tbm_id VARCHAR(50) NOT NULL,
    ring_number INTEGER NOT NULL,
    date DATE NOT NULL,
    advance_rate DECIMAL(10,2), -- mm/min
    thrust_force DECIMAL(10,2), -- kN
    torque DECIMAL(10,2), -- kNm
    penetration_rate DECIMAL(10,2), -- mm/rev
    screw_speed DECIMAL(10,2), -- rpm
    screw_pressure DECIMAL(10,2), -- bar
    face_pressure DECIMAL(10,2), -- bar
    slurry_density DECIMAL(10,2),
    slurry_flow DECIMAL(10,2),
    grout_volume DECIMAL(10,2),
    grout_pressure DECIMAL(10,2),
    chainage DECIMAL(10,2), -- meters from start
    pitch DECIMAL(10,2),
    yaw DECIMAL(10,2),
    roll DECIMAL(10,2),
    geology_code VARCHAR(50),
    water_inflow DECIMAL(10,2),
    temperature DECIMAL(5,2),
    vibration DECIMAL(10,2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Ring Tracking
CREATE TABLE rings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    ring_number INTEGER NOT NULL,
    segment_count INTEGER DEFAULT 7,
    outer_diameter DECIMAL(10,2),
    inner_diameter DECIMAL(10,2),
    width DECIMAL(10,2),
    concrete_grade VARCHAR(20),
    reinforcement_kg DECIMAL(10,2),
    production_date DATE,
    curing_end_date DATE,
    transport_date DATE,
    install_date DATE,
    chainage DECIMAL(10,2),
    status VARCHAR(20), -- produced, cured, transported, installed, inspected
    qc_status VARCHAR(20), -- pending, passed, failed
    defects JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Documents
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL, -- RFI, NCR, submittal, drawing, contract
    document_number VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'draft',
    version INTEGER DEFAULT 1,
    file_path TEXT,
    file_size BIGINT,
    mime_type VARCHAR(100),
    created_by UUID REFERENCES users(id),
    assigned_to UUID REFERENCES users(id),
    due_date DATE,
    approved_date TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- RFI
CREATE TABLE rfis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    rfi_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    question TEXT NOT NULL,
    response TEXT,
    priority VARCHAR(20), -- low, medium, high, urgent
    status VARCHAR(20) DEFAULT 'open', -- open, answered, closed
    submitted_by UUID REFERENCES users(id),
    assigned_to UUID REFERENCES users(id),
    due_date DATE,
    answered_date TIMESTAMPTZ,
    category VARCHAR(100),
    related_document_id UUID REFERENCES documents(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- NCR
CREATE TABLE ncrs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    ncr_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(20), -- minor, major, critical
    status VARCHAR(20) DEFAULT 'open', -- open, in_review, closed
    raised_by UUID REFERENCES users(id),
    assigned_to UUID REFERENCES users(id),
    root_cause TEXT,
    corrective_action TEXT,
    preventive_action TEXT,
    due_date DATE,
    closed_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Contracts
CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    contract_number VARCHAR(100) NOT NULL,
    contract_type VARCHAR(50), -- lump_sum, unit_price, cost_plus, target_price
    party_id UUID REFERENCES organizations(id),
    value DECIMAL(20,2),
    currency VARCHAR(3) DEFAULT 'USD',
    start_date DATE,
    end_date DATE,
    status VARCHAR(20), -- draft, signed, active, completed, terminated
    payment_terms TEXT,
    retention_percent DECIMAL(5,2),
    warranty_period INTEGER, -- months
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Invoices
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    contract_id UUID REFERENCES contracts(id),
    invoice_number VARCHAR(100) NOT NULL,
    invoice_type VARCHAR(20), -- progress, milestone, advance, final
    amount DECIMAL(20,2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20), -- draft, submitted, approved, paid, disputed
    submitted_date DATE,
    approved_date DATE,
    paid_date DATE,
    retention_amount DECIMAL(20,2),
    retention_release_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Change Orders
CREATE TABLE change_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    contract_id UUID REFERENCES contracts(id),
    co_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(20), -- variation, change, claim, extension
    status VARCHAR(20), -- proposed, approved, rejected, implemented
    amount DECIMAL(20,2),
    time_extension INTEGER, -- days
    submitted_by UUID REFERENCES users(id),
    approved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Risk Register
CREATE TABLE risks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    risk_id VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100), -- technical, financial, schedule, HSE, legal
    probability INTEGER, -- 1-5
    impact INTEGER, -- 1-5
    risk_score INTEGER GENERATED ALWAYS AS (probability * impact) STORED,
    mitigation_strategy TEXT,
    contingency_amount DECIMAL(20,2),
    owner UUID REFERENCES users(id),
    status VARCHAR(20), -- identified, mitigated, realized, closed
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- HSE Incidents
CREATE TABLE hse_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    incident_number VARCHAR(50) NOT NULL,
    incident_type VARCHAR(50), -- LTI, MTI, first_aid, near_miss, property_damage
    severity VARCHAR(20), -- low, medium, high, fatal
    description TEXT NOT NULL,
    location POINT,
    reported_by UUID REFERENCES users(id),
    investigation TEXT,
    root_cause TEXT,
    corrective_actions TEXT,
    lost_days INTEGER,
    status VARCHAR(20), -- reported, investigating, closed
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Equipment
CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    equipment_code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50), -- TBM, crane, excavator, locomotive, batch_plant
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    serial_number VARCHAR(100),
    year INTEGER,
    status VARCHAR(20), -- active, maintenance, idle, retired
    location POINT,
    operator_id UUID REFERENCES users(id),
    hourly_rate DECIMAL(10,2),
    fuel_consumption DECIMAL(10,2),
    last_maintenance_date DATE,
    next_maintenance_date DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Daily Reports
CREATE TABLE daily_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    report_date DATE NOT NULL,
    weather VARCHAR(100),
    temperature DECIMAL(5,2),
    shift VARCHAR(20), -- day, night, both
    total_workers INTEGER,
    total_hours DECIMAL(10,2),
    summary TEXT,
    created_by UUID REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Knowledge Graph (Neo4j)
-- Stored in Neo4j, synced via CDC
-- Nodes: Project, Activity, Resource, Document, Person, Organization, Equipment, Material, Location, Risk, Contract
-- Relationships: DEPENDS_ON, ASSIGNED_TO, BELONGS_TO, LOCATED_AT, SUPPLIES, MANAGES, APPROVES, CAUSES, MITIGATES

-- Vector Store (Qdrant)
-- Collections: document_embeddings, contract_clauses, spec_sections, chat_history
-- Dimension: 1536 (text-embedding-3-small) or 4096 (local model)
```

### 3.2 Indexes (ключевые)

```sql
-- Performance indexes
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_client ON projects(client_id);
CREATE INDEX idx_wbs_path ON wbs USING gist(path);
CREATE INDEX idx_cbs_path ON cbs USING gist(path);
CREATE INDEX idx_schedule_project ON schedule_activities(project_id, status);
CREATE INDEX idx_tbm_session_ring ON tbm_sessions(project_id, ring_number);
CREATE INDEX idx_tbm_session_date ON tbm_sessions(project_id, date);
CREATE INDEX idx_rings_status ON rings(project_id, status);
CREATE INDEX idx_documents_type ON documents(project_id, document_type, status);
CREATE INDEX idx_rfis_status ON rfis(project_id, status);
CREATE INDEX idx_ncrs_status ON ncrs(project_id, status);
CREATE INDEX idx_contracts_status ON contracts(project_id, status);
CREATE INDEX idx_invoices_status ON invoices(project_id, status);
CREATE INDEX idx_risks_score ON risks(project_id, risk_score DESC);
CREATE INDEX idx_hse_severity ON hse_incidents(project_id, severity);
CREATE INDEX idx_daily_reports_date ON daily_reports(project_id, report_date DESC);
CREATE INDEX idx_equipment_type ON equipment(project_id, type);
CREATE INDEX idx_budget_lines_wbs ON budget_lines(project_id, wbs_id);
CREATE INDEX idx_budget_lines_cbs ON budget_lines(project_id, cbs_id);

-- Full-text search
CREATE INDEX idx_documents_fts ON documents USING gin(to_tsvector('english', title || ' ' || COALESCE(description, '')));
CREATE INDEX idx_projects_fts ON projects USING gin(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- Time-series (TimescaleDB)
SELECT create_hypertable('tbm_sessions', 'date');
SELECT create_hypertable('equipment_telemetry', 'timestamp');
```

---

## 4. Пользовательские роли и права доступа

### 4.1 Роли

| Роль | Уровень | Описание |
|------|---------|----------|
| **System Admin** | System | Полный доступ ко всем проектам, настройки системы |
| **CEO** | Organization | Все проекты организации, финансы, стратегия |
| **Project Director** | Project | Полный доступ к проекту, все модули |
| **Construction Manager** | Project | Оперативное управление строительством |
| **Chief Engineer** | Project | Технические решения, BIM, design |
| **Planning Engineer** | Project | Сchedules, progress, EVM |
| **Scheduler** | Project | P6, detailed scheduling |
| **Cost Engineer** | Project | Budget, cost control, forecasting |
| **Contract Manager** | Project | Contracts, claims, variations |
| **Procurement Manager** | Project | RFQ, PO, supplier management |
| **Warehouse Manager** | Project | Inventory, material receipt |
| **HSE Manager** | Project | Safety, incidents, permits |
| **QA/QC Manager** | Project | Quality, ITP, NCR, testing |
| **Surveyor** | Project | Geodetic, topographic |
| **TBM Manager** | Project | TBM operations, ring building |
| **Foreman** | Project | Daily operations, crew management |
| **Subcontractor** | Project | Limited: timesheets, RFI, submittals |
| **Client** | Project | Read-only + approvals, RFI response |
| **Consultant** | Project | Design review, inspection |
| **Bank** | Project | Financial reports, guarantees |
| **Auditor** | Project | Read-only, audit trail |
| **Engineer** | Project | Technical documents, shop drawings |
| **Inspector** | Project | Inspection, NCR, test records |
| **Safety Officer** | Project | HSE inspections, permits |
| **Environmental Officer** | Project | Environmental monitoring |
| **Community Liaison** | Project | Stakeholder management |
| **IT Support** | Organization | System maintenance, user management |

### 4.2 Permission Matrix (пример для ключевых модулей)

```
Module: Schedule
  - Project Director: CRUD
  - Planning Engineer: CRUD
  - Scheduler: CRUD
  - Construction Manager: Read
  - Chief Engineer: Read
  - Client: Read
  - Subcontractor: Read (own activities)

Module: Cost
  - Project Director: CRUD
  - Cost Engineer: CRUD
  - CEO: Read
  - Client: Read (summary)
  - Bank: Read (financial)
  - Construction Manager: Read

Module: Documents (RFI/NCR)
  - Project Director: CRUD + Approve
  - Engineer: Create + Read
  - Construction Manager: Create + Read
  - Client: Read + Respond (RFI)
  - Consultant: Read + Approve
  - Subcontractor: Create (RFI) + Read (own)

Module: TBM
  - TBM Manager: CRUD
  - Chief Engineer: Read
  - Construction Manager: Read
  - Foreman: Read + Create (daily data)

Module: Finance
  - CEO: Read
  - Project Director: Read
  - Cost Engineer: CRUD
  - Bank: Read (specific reports)
  - Auditor: Read
```

---

## 5. REST API Structure

### 5.1 Core Endpoints

```
# Projects
GET    /api/v1/projects
POST   /api/v1/projects
GET    /api/v1/projects/:id
PUT    /api/v1/projects/:id
DELETE /api/v1/projects/:id
GET    /api/v1/projects/:id/dashboard
GET    /api/v1/projects/:id/kpi

# WBS
GET    /api/v1/projects/:id/wbs
POST   /api/v1/projects/:id/wbs
PUT    /api/v1/projects/:id/wbs/:wbsId
DELETE /api/v1/projects/:id/wbs/:wbsId

# Schedule
GET    /api/v1/projects/:id/schedule
POST   /api/v1/projects/:id/schedule/activities
PUT    /api/v1/projects/:id/schedule/activities/:activityId
POST   /api/v1/projects/:id/schedule/relationships
GET    /api/v1/projects/:id/schedule/gantt
POST   /api/v1/projects/:id/schedule/update-progress
POST   /api/v1/projects/:id/schedule/export-msproject
POST   /api/v1/projects/:id/schedule/import-primavera

# Cost
GET    /api/v1/projects/:id/cost/budget
POST   /api/v1/projects/:id/cost/budget
GET    /api/v1/projects/:id/cost/evm
GET    /api/v1/projects/:id/cost/cashflow
POST   /api/v1/projects/:id/cost/forecast

# TBM
GET    /api/v1/projects/:id/tbm/sessions
POST   /api/v1/projects/:id/tbm/sessions
GET    /api/v1/projects/:id/tbm/rings
POST   /api/v1/projects/:id/tbm/rings
GET    /api/v1/projects/:id/tbm/dashboard
POST   /api/v1/projects/:id/tbm/telemetry

# Documents
GET    /api/v1/projects/:id/documents
POST   /api/v1/projects/:id/documents
GET    /api/v1/projects/:id/documents/:docId
PUT    /api/v1/projects/:id/documents/:docId
POST   /api/v1/projects/:id/documents/:docId/upload
GET    /api/v1/projects/:id/documents/:docId/download
POST   /api/v1/projects/:id/documents/:docId/approve
POST   /api/v1/projects/:id/documents/:docId/reject

# RFI
GET    /api/v1/projects/:id/rfis
POST   /api/v1/projects/:id/rfis
GET    /api/v1/projects/:id/rfis/:rfiId
PUT    /api/v1/projects/:id/rfis/:rfiId
POST   /api/v1/projects/:id/rfis/:rfiId/respond
POST   /api/v1/projects/:id/rfis/:rfiId/close

# NCR
GET    /api/v1/projects/:id/ncrs
POST   /api/v1/projects/:id/ncrs
GET    /api/v1/projects/:id/ncrs/:ncrId
PUT    /api/v1/projects/:id/ncrs/:ncrId
POST   /api/v1/projects/:id/ncrs/:ncrId/close

# Contracts
GET    /api/v1/projects/:id/contracts
POST   /api/v1/projects/:id/contracts
GET    /api/v1/projects/:id/contracts/:contractId
PUT    /api/v1/projects/:id/contracts/:contractId
POST   /api/v1/projects/:id/contracts/:contractId/amend

# Invoices
GET    /api/v1/projects/:id/invoices
POST   /api/v1/projects/:id/invoices
GET    /api/v1/projects/:id/invoices/:invoiceId
POST   /api/v1/projects/:id/invoices/:invoiceId/approve
POST   /api/v1/projects/:id/invoices/:invoiceId/pay

# Change Orders
GET    /api/v1/projects/:id/change-orders
POST   /api/v1/projects/:id/change-orders
GET    /api/v1/projects/:id/change-orders/:coId
POST   /api/v1/projects/:id/change-orders/:coId/approve

# Risks
GET    /api/v1/projects/:id/risks
POST   /api/v1/projects/:id/risks
PUT    /api/v1/projects/:id/risks/:riskId
POST   /api/v1/projects/:id/risks/monte-carlo

# HSE
GET    /api/v1/projects/:id/hse/incidents
POST   /api/v1/projects/:id/hse/incidents
GET    /api/v1/projects/:id/hse/dashboard
GET    /api/v1/projects/:id/hse/statistics

# Equipment
GET    /api/v1/projects/:id/equipment
POST   /api/v1/projects/:id/equipment
GET    /api/v1/projects/:id/equipment/:equipId
POST   /api/v1/projects/:id/equipment/:equipId/maintenance

# Daily Reports
GET    /api/v1/projects/:id/daily-reports
POST   /api/v1/projects/:id/daily-reports
GET    /api/v1/projects/:id/daily-reports/:reportId

# BIM
GET    /api/v1/projects/:id/bim/models
POST   /api/v1/projects/:id/bim/models/upload
GET    /api/v1/projects/:id/bim/models/:modelId
GET    /api/v1/projects/:id/bim/clash-detection
POST   /api/v1/projects/:id/bim/clash-detection/run
GET    /api/v1/projects/:id/bim/quantity-takeoff
POST   /api/v1/projects/:id/bim/4d-simulation

# AI
POST   /api/v1/ai/chat
POST   /api/v1/ai/analyze-document
POST   /api/v1/ai/predict-cost
POST   /api/v1/ai/predict-schedule
POST   /api/v1/ai/detect-risks
POST   /api/v1/ai/analyze-contract
POST   /api/v1/ai/generate-report
POST   /api/v1/ai/copilot

# GIS
GET    /api/v1/projects/:id/gis/layers
POST   /api/v1/projects/:id/gis/layers
GET    /api/v1/projects/:id/gis/features
POST   /api/v1/projects/:id/gis/features

# Reports
GET    /api/v1/projects/:id/reports
POST   /api/v1/projects/:id/reports/generate
GET    /api/v1/projects/:id/reports/:reportId/export

# Users & Auth
POST   /api/v1/auth/login
POST   /api/v1/auth/register
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/mfa/setup
POST   /api/v1/auth/mfa/verify
GET    /api/v1/users/me
PUT    /api/v1/users/me
GET    /api/v1/users
POST   /api/v1/users/invite

# Organizations
GET    /api/v1/organizations
POST   /api/v1/organizations
GET    /api/v1/organizations/:id
PUT    /api/v1/organizations/:id

# Admin
GET    /api/v1/admin/audit-log
GET    /api/v1/admin/system-health
GET    /api/v1/admin/metrics
POST   /api/v1/admin/backup
POST   /api/v1/admin/restore
```

---

## 6. GraphQL API

```graphql
type Query {
  # Projects
  project(id: ID!): Project
  projects(filter: ProjectFilter, pagination: Pagination): ProjectConnection
  projectDashboard(id: ID!): ProjectDashboard
  projectKPI(id: ID!): [KPI!]

  # Schedule
  schedule(projectId: ID!): Schedule!
  activities(projectId: ID!, filter: ActivityFilter): [Activity!]
  ganttData(projectId: ID!): GanttData!

  # Cost
  budget(projectId: ID!): Budget!
  evm(projectId: ID!): EVMData!
  cashFlow(projectId: ID!): CashFlow!

  # TBM
  tbmSessions(projectId: ID!, filter: TBMSessionFilter): [TBMSession!]
  rings(projectId: ID!, filter: RingFilter): [Ring!]
  tbmDashboard(projectId: ID!): TBMDashboard!

  # Documents
  documents(projectId: ID!, filter: DocumentFilter): [Document!]
  rfi(projectId: ID!, id: ID!): RFI
  rfis(projectId: ID!, filter: RFIFilter): [RFI!]
  ncr(projectId: ID!, id: ID!): NCR
  ncrs(projectId: ID!, filter: NCRFilter): [NCR!]

  # Contracts
  contracts(projectId: ID!): [Contract!]
  contract(projectId: ID!, id: ID!): Contract

  # Invoices
  invoices(projectId: ID!, filter: InvoiceFilter): [Invoice!]

  # Risks
  risks(projectId: ID!): [Risk!]
  riskMatrix(projectId: ID!): RiskMatrix!

  # HSE
  hseDashboard(projectId: ID!): HSEDashboard!
  hseIncidents(projectId: ID!): [HSEIncident!]

  # Equipment
  equipment(projectId: ID!): [Equipment!]
  equipmentTelemetry(projectId: ID!, equipmentId: ID!, from: DateTime, to: DateTime): [TelemetryPoint!]

  # BIM
  bimModels(projectId: ID!): [BIMModel!]
  clashResults(projectId: ID!, modelId: ID): [ClashResult!]
  quantityTakeoff(projectId: ID!, modelId: ID!): [QuantityItem!]

  # AI
  aiChat(projectId: ID!, sessionId: ID): AIChatSession
  aiAnalysis(documentId: ID!): AIAnalysis

  # Search
  search(projectId: ID!, query: String!, type: SearchType): SearchResults!

  # Reports
  reports(projectId: ID!): [Report!]
  report(projectId: ID!, id: ID!): Report
}

type Mutation {
  # Projects
  createProject(input: CreateProjectInput!): Project!
  updateProject(id: ID!, input: UpdateProjectInput!): Project!
  deleteProject(id: ID!): Boolean!

  # Schedule
  createActivity(projectId: ID!, input: CreateActivityInput!): Activity!
  updateActivity(projectId: ID!, id: ID!, input: UpdateActivityInput!): Activity!
  updateProgress(projectId: ID!, input: ProgressUpdateInput!): Activity!
  createRelationship(projectId: ID!, input: CreateRelationshipInput!): Relationship!

  # Cost
  updateBudget(projectId: ID!, input: BudgetInput!): Budget!
  createForecast(projectId: ID!, input: ForecastInput!): Forecast!

  # TBM
  createTBMSession(projectId: ID!, input: CreateTBMSessionInput!): TBMSession!
  createRing(projectId: ID!, input: CreateRingInput!): Ring!
  updateRingStatus(projectId: ID!, id: ID!, status: RingStatus!): Ring!

  # Documents
  createDocument(projectId: ID!, input: CreateDocumentInput!): Document!
  uploadDocument(projectId: ID!, id: ID!, file: Upload!): Document!
  approveDocument(projectId: ID!, id: ID!, comment: String): Document!
  rejectDocument(projectId: ID!, id: ID!, reason: String!): Document!

  # RFI
  createRFI(projectId: ID!, input: CreateRFIInput!): RFI!
  respondToRFI(projectId: ID!, id: ID!, response: String!): RFI!
  closeRFI(projectId: ID!, id: ID!): RFI!

  # NCR
  createNCR(projectId: ID!, input: CreateNCRInput!): NCR!
  closeNCR(projectId: ID!, id: ID!, action: String!): NCR!

  # Contracts
  createContract(projectId: ID!, input: CreateContractInput!): Contract!
  amendContract(projectId: ID!, id: ID!, input: AmendContractInput!): Contract!

  # Invoices
  createInvoice(projectId: ID!, input: CreateInvoiceInput!): Invoice!
  approveInvoice(projectId: ID!, id: ID!): Invoice!
  payInvoice(projectId: ID!, id: ID!): Invoice!

  # Change Orders
  createChangeOrder(projectId: ID!, input: CreateChangeOrderInput!): ChangeOrder!
  approveChangeOrder(projectId: ID!, id: ID!): ChangeOrder!

  # Risks
  createRisk(projectId: ID!, input: CreateRiskInput!): Risk!
  updateRisk(projectId: ID!, id: ID!, input: UpdateRiskInput!): Risk!
  runMonteCarlo(projectId: ID!): RiskAnalysisResult!

  # HSE
  reportIncident(projectId: ID!, input: CreateIncidentInput!): HSEIncident!
  closeIncident(projectId: ID!, id: ID!, input: CloseIncidentInput!): HSEIncident!

  # Equipment
  createEquipment(projectId: ID!, input: CreateEquipmentInput!): Equipment!
  scheduleMaintenance(projectId: ID!, equipmentId: ID!, input: MaintenanceInput!): Equipment!

  # BIM
  uploadModel(projectId: ID!, file: Upload!): BIMModel!
  runClashDetection(projectId: ID!, modelIds: [ID!]!): ClashDetectionJob!
  generate4DSimulation(projectId: ID!, modelId: ID!): SimulationResult!

  # AI
  sendChatMessage(projectId: ID!, sessionId: ID, message: String!): AIChatResponse!
  analyzeDocument(projectId: ID!, documentId: ID!): AIAnalysis!
  generateReport(projectId: ID!, input: ReportGenerationInput!): Report!

  # Daily Reports
  createDailyReport(projectId: ID!, input: CreateDailyReportInput!): DailyReport!
  submitDailyReport(projectId: ID!, id: ID!): DailyReport!
}

type Subscription {
  tbmTelemetry(projectId: ID!, tbmId: String!): TBSTelemetryPoint!
  equipmentTelemetry(projectId: ID!, equipmentId: ID!): TelemetryPoint!
  notification(projectId: ID!): Notification!
  scheduleUpdate(projectId: ID!): ScheduleUpdate!
  documentUpdate(projectId: ID!): DocumentUpdate!
}
```

---

## 7. Файловая структура проекта

```
openconstructionerp/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml
│   │   ├── cd.yml
│   │   ├── release.yml
│   │   ├── security-scan.yml
│   │   └── docs.yml
│   ├── CODEOWNERS
│   └── ISSUE_TEMPLATE/
├── apps/
│   ├── web/                          # Next.js frontend
│   │   ├── src/
│   │   │   ├── app/                  # App router pages
│   │   │   ├── components/           # Shared components
│   │   │   ├── lib/                  # Utilities, API client
│   │   │   ├── hooks/               # Custom hooks
│   │   │   ├── stores/              # Zustand stores
│   │   │   ├── graphql/             # GraphQL queries/mutations
│   │   │   └── styles/              # Tailwind config
│   │   ├── public/
│   │   ├── tests/
│   │   ├── Dockerfile
│   │   └── package.json
│   ├── mobile/                       # React Native
│   │   ├── src/
│   │   ├── ios/
│   │   ├── android/
│   │   └── package.json
│   └── desktop/                      # Tauri
│       ├── src-tauri/
│       └── src/
├── services/
│   ├── api-gateway/                  # Kong/Envoy config
│   ├── auth-service/                 # Keycloak + custom
│   ├── project-service/              # NestJS
│   ├── schedule-service/             # NestJS
│   ├── cost-service/                 # NestJS
│   ├── document-service/             # NestJS
│   ├── contract-service/             # NestJS
│   ├── procurement-service/          # NestJS
│   ├── warehouse-service/            # NestJS
│   ├── equipment-service/            # Go (high perf)
│   ├── tbm-service/                  # Go (real-time)
│   ├── hse-service/                  # NestJS
│   ├── quality-service/              # NestJS
│   ├── bim-service/                  # Python (IfcOpenShell)
│   ├── gis-service/                  # Go + GeoServer
│   ├── finance-service/              # NestJS
│   ├── risk-service/                 # Python (Monte Carlo)
│   ├── hr-service/                   # NestJS
│   ├── notification-service/         # Node.js
│   ├── reporting-service/            # Python
│   ├── ai-service/                   # Python FastAPI
│   │   ├── api/
│   │   ├── models/
│   │   ├── pipelines/
│   │   ├── agents/
│   │   └── embeddings/
│   ├── integration-service/          # Node.js
│   ├── search-service/               # Node.js + Elastic
│   ├── audit-service/                # Go
│   └── analytics-service/            # Python + Spark
├── libs/
│   ├── shared/                       # Shared TypeScript types
│   ├── graphql-schema/               # Federation schema
│   ├── protos/                       # gRPC protobuf definitions
│   └── python-common/                # Shared Python libs
├── infrastructure/
│   ├── docker/
│   │   ├── docker-compose.yml
│   │   ├── docker-compose.dev.yml
│   │   └── docker-compose.prod.yml
│   ├── k8s/
│   │   ├── namespaces/
│   │   ├── deployments/
│   │   ├── services/
│   │   ├── configmaps/
│   │   ├── secrets/
│   │   ├── ingress/
│   │   └── helm/
│   ├── terraform/
│   │   ├── modules/
│   │   └── environments/
│   └── monitoring/
│       ├── prometheus/
│       ├── grafana/
│       ├── loki/
│       └── alertmanager/
├── database/
│   ├── migrations/                   # Flyway/Sqitch
│   ├── seeds/
│   ├── functions/
│   └── triggers/
├── docs/
│   ├── architecture/
│   ├── api/
│   ├── user-guides/
│   ├── developer-guides/
│   └── deployment/
├── scripts/
│   ├── setup.sh
│   ├── backup.sh
│   ├── restore.sh
│   └── migrate.sh
├── tests/
│   ├── e2e/
│   ├── integration/
│   └── performance/
├── .env.example
├── .gitignore
├── .prettierrc
├── .eslintrc.js
├── tsconfig.json
├── package.json
├── lerna.json
├── nx.json
├── README.md
├── CONTRIBUTING.md
├── LICENSE
└── SECURITY.md
```

---

## 8. Docker & Kubernetes

### 8.1 Docker Compose (Development)

```yaml
version: '3.9'
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ocerp
      POSTGRES_USER: ocerp
      POSTGRES_PASSWORD: ocerp_dev
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  timescaledb:
    image: timescale/timescaledb:2-pg16
    environment:
      POSTGRES_DB: ocerp_ts
      POSTGRES_USER: ocerp
      POSTGRES_PASSWORD: ocerp_dev
    volumes:
      - tsdata:/var/lib/postgresql/data
    ports:
      - "5433:5432"

  neo4j:
    image: neo4j:5-community
    environment:
      NEO4J_AUTH: neo4j/ocerp_dev
    volumes:
      - neo4jdata:/data
    ports:
      - "7687:7687"
      - "7474:7474"

  elasticsearch:
    image: elasticsearch:8.11
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - esdata:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3-management
    ports:
      - "5672:5672"
      - "15672:15672"

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"

  zookeeper:
    image: confluentinc/cp-zookeeper:latest

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data

  qdrant:
    image: qdrant/qdrant
    ports:
      - "6333:6333"
    volumes:
      - qdrantdata:/qdrant/storage

  keycloak:
    image: quay.io/keycloak/keycloak:22.0
    environment:
      KC_DB: postgres
      KC_DB_URL: jdbc:postgresql://postgres:5432/ocerp
      KC_DB_USERNAME: ocerp
      KC_DB_PASSWORD: ocerp_dev
      KEYCLOAK_ADMIN: admin
      KEYCLOAK_ADMIN_PASSWORD: admin
    ports:
      - "8080:8080"

  api-gateway:
    image: kong:3.5
    depends_on:
      - postgres
    ports:
      - "8000:8000"
      - "8001:8001"

volumes:
  pgdata:
  tsdata:
  neo4jdata:
  esdata:
  miniodata:
  qdrantdata:
```

### 8.2 Kubernetes (Production)

```yaml
# helm/values.yaml (ключевые параметры)
global:
  environment: production
  domain: ocerp.example.com
  imageRegistry: ghcr.io/openconstructionerp

postgresql:
  enabled: true
  architecture: replication
  replicaCount: 3
  primary:
    resources:
      requests: { cpu: "4", memory: "16Gi" }
      limits: { cpu: "8", memory: "32Gi" }
  readReplicas:
    resources:
      requests: { cpu: "2", memory: "8Gi" }

timescaledb:
  replicaCount: 2
  resources:
    requests: { cpu: "4", memory: "16Gi" }

neo4j:
  core: 3
  readReplicas: 2
  resources:
    requests: { cpu: "2", memory: "8Gi" }

elasticsearch:
  nodeCount: 3
  resources:
    requests: { cpu: "2", memory: "8Gi" }

kafka:
  replicaCount: 3
  resources:
    requests: { cpu: "2", memory: "4Gi" }

minio:
  mode: distributed
  statefulsetCount: 4
  drivesPerNode: 2
  resources:
    requests: { cpu: "1", memory: "2Gi" }

qdrant:
  replicaCount: 2
  resources:
    requests: { cpu: "1", memory: "4Gi" }

services:
  project-service:
    replicaCount: 3
    resources:
      requests: { cpu: "500m", memory: "1Gi" }
    autoscaling:
      enabled: true
      minReplicas: 3
      maxReplicas: 20
      targetCPUUtilizationPercentage: 70

  tbm-service:
    replicaCount: 2
    resources:
      requests: { cpu: "1", memory: "2Gi" }
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10

  ai-service:
    replicaCount: 2
    resources:
      requests: { cpu: "4", memory: "16Gi" }  # GPU node
    nodeSelector:
      cloud.google.com/gke-accelerator: nvidia-tesla-t4
```

---

## 9. CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '22' }
      - run: npm ci
      - run: npm run lint
      - run: npm run type-check

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: ocerp_test
          POSTGRES_USER: ocerp
          POSTGRES_PASSWORD: ocerp_test
        ports: ['5432:5432']
    steps:
      - uses: actions/checkout@v4
      - run: npm ci
      - run: npm run test:unit
      - run: npm run test:integration
      - run: npm run test:e2e

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          format: 'sarif'
          output: 'trivy-results.sarif'
      - uses: github/codeql-action/upload-sarif@v3

  build:
    runs-on: ubuntu-latest
    needs: [lint, test, security]
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t ocerp-web -f apps/web/Dockerfile .
      - run: docker build -t ocerp-api -f services/api-gateway/Dockerfile .
      # ... build all services
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - run: docker push ghcr.io/openconstructionerp/ocerp-web:latest

  deploy-staging:
    needs: [build]
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      - uses: azure/setup-kubectl@v3
      - run: kubectl set image deployment/web ocerp-web=ghcr.io/openconstructionerp/ocerp-web:${{ github.sha }}
```

---

## 10. Дорожная карта разработки

### Phase 1: MVP (6-9 месяцев, 15 разработчиков)
**Цель:** Работающая система для одного проекта с базовыми модулями

- [ ] Project Management (WBS, milestones)
- [ ] Document Control (RFI, NCR, submittals)
- [ ] Schedule Management (Gantt, CPM, P6 import)
- [ ] Cost Management (budget, EVM, cash flow)
- [ ] Basic BIM Viewer (IFC)
- [ ] User Management + RBAC
- [ ] REST API + GraphQL
- [ ] Docker Compose deployment
- [ ] PostgreSQL + MinIO + Redis

### Phase 2: Beta (9-12 месяцев, 25 разработчиков)
**Цель:** Полноценная ERP для тоннельного строительства

- [ ] TBM Module (telemetry, rings, segments)
- [ ] Contract Management
- [ ] Procurement + Warehouse
- [ ] Equipment Management
- [ ] HSE Module
- [ ] Quality Module
- [ ] Daily Reports
- [ ] AI Chat + Document Analysis
- [ ] GIS Integration
- [ ] BIM 4D/5D
- [ ] Mobile App (React Native)
- [ ] Kubernetes deployment
- [ ] Integration: Primavera, SAP, Excel

### Phase 3: Release (12-18 месяцев, 35 разработчиков)
**Цель:** Enterprise-grade платформа

- [ ] All 100+ modules
- [ ] AI Copilot + Knowledge Graph
- [ ] Digital Twin
- [ ] Advanced BIM (clash, QTO, 6D/7D)
- [ ] Real-time TBM telemetry (NATS)
- [ ] Multi-project portfolio management
- [ ] Advanced analytics + Power BI
- [ ] Multi-currency + multi-language
- [ ] Audit + Compliance (ISO 27001)
- [ ] Performance optimization
- [ ] Load testing (100+ concurrent projects)

### Phase 4: Enterprise (18-24 месяцев, 50+ разработчиков)
**Цель:** Мировой стандарт

- [ ] AI Delay Analysis + Claims
- [ ] AI Vision (site cameras)
- [ ] Full Digital Twin
- [ ] Edge computing for TBM
- [ ] Marketplace for plugins
- [ ] Open-source community
- [ ] Certification programs
- [ ] Global CDN deployment
- [ ] 99.99% SLA
- [ ] SOC 2 + ISO 27001 certified

---

## 11. KPI по модулям

### Project Management
- Schedule Performance Index (SPI)
- Cost Performance Index (CPI)
- Planned vs Actual Progress (%)
- Milestone Achievement Rate (%)

### TBM
- Advance Rate (mm/min)
- Ring Build Time (min)
- Segment Installation Rate (rings/day)
- TBM Utilization (%)
- MTBF (Mean Time Between Failures)

### Cost
- Cost Variance (CV)
- Schedule Variance (SV)
- Estimate at Completion (EAC)
- Budget Utilization (%)
- Cash Flow Variance

### Quality
- NCR Closure Rate (%)
- First Pass Yield (%)
- Rework Cost (% of total)
- Inspection Pass Rate (%)

### HSE
- Lost Time Injury Frequency (LTIF)
- Total Recordable Incident Rate (TRIR)
- Near Miss Reporting Rate
- Safety Training Compliance (%)

### Procurement
- PO Cycle Time (days)
- Supplier On-Time Delivery (%)
- Material Availability (%)
- Cost Savings (%)

### Equipment
- Equipment Utilization (%)
- Maintenance Cost per Hour
- Fuel Efficiency
- Availability (%)

### Document Control
- RFI Response Time (days)
- Submittal Approval Cycle (days)
- Document Transmittal Accuracy (%)
- NCR Closure Time (days)

---

## 12. Заключение

**OpenConstructionERP** — это не ERP в традиционном понимании. Это **Project Operating System**, построенная на принципах:

1. **Data-first** — каждая сущность проекта first-class citizen
2. **Graph-native** — всё связано через knowledge graph
3. **AI-native** — AI не надстройка, а встроенный слой
4. **BIM-native** — модель — источник истины для всех данных
5. **Real-time** — от TBM telemetry до финансовых потоков
6. **Open** — открытый код, открытые API, открытые стандарты

Система спроектирована так, чтобы через 10 лет стать мировым стандартом управления строительством инфраструктурных проектов — так же, как Palantir Foundry стал стандартом для разведки и обороны.

**Следующий шаг:** Создание репозитория, настройка CI/CD, реализация MVP.
