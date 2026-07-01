# OpenConstructionERP

**Open-source platform for managing the full lifecycle of large infrastructure construction projects**
Metro · Tunnels · Railways · Hydraulic Structures · EPC/EPCM

> **Status: pre-alpha.** Architecture is defined and baselined; the working code today is the
> BOQ (Bill of Quantities) module: schema, data generator and dashboard prototype.
> Everything else is a specification, not a promise of existing code. We build in the open.

---

## What this is

Not "an accounting system with BIM bolted on", but an **ontology-first project operating system**:
every entity — a lining ring, an RFI, a payment certificate, an excavator, a borehole — is an object
in one semantic graph with full bitemporal history. Modules are applications over that shared graph.
Domain state changes are Kafka events, which gives banks/auditors a complete audit trail and gives
delay analysis an as-of-any-date view out of the box.

The reference domain is what generic tools handle worst: **tunneling** — TBM telemetry, ring register,
segment traceability, NATM, microtunnelling, instrumentation & settlement — designed by practitioners
running metro and tunnel EPC projects, offline-first for crews underground.

## Authoritative documentation

| Document | Purpose |
|----------|---------|
| [SAD Tom 1 — Architecture](docs/sad/SAD-Tom1-Architecture-v1.0.md) | System architecture, 15 general decisions, 112-module registry, role model, data conventions, roadmap |
| [ADR-001…016](docs/adr/README.md) | Architecture Decision Records — every core decision with rejected alternatives |
| [Archived draft v0](docs/archive/architecture-draft-v0-SUPERSEDED.md) | Early draft, superseded by SAD Tom 1 |

## Stack (baselined — see ADRs)

```
Core:        Go modular monolith (ontology, IAM, workflow, CDE) — ADR-002
Services:    Python (AI / BIM / analytics), Go (telemetry, CPM)  — ADR-003
Frontend:    React 18 + TypeScript, Module Federation            — ADR-007
Mobile:      Flutter, offline-first (SQLite + event-log sync)    — ADR-008
Database:    PostgreSQL 16 + PostGIS + TimescaleDB + pgvector + Apache AGE — ADR-004
Analytics:   ClickHouse (CDC via Debezium)                       — ADR-005
Events:      Kafka (event store) + NATS (realtime)               — ADR-006
Storage:     MinIO (S3, WORM for contractual docs)               — ADR-010
Search:      OpenSearch (BM25 + kNN hybrid RAG)                  — ADR-011
BIM:         IfcOpenShell (IFC 4.3) + xeokit + BCF 3.0 server    — ADR-012
AI:          LiteLLM gateway + LangGraph agents, LLM-agnostic    — ADR-013
Auth:        Keycloak + OPA (RBAC x ABAC)                        — ADR-009
Deploy:      Kubernetes/Helm/ArgoCD + single-node Compose profile — ADR-015
```

## What works today (Prototype 0.1)

- **`services/core-py/`** — Python reference implementation of the platform core (FastAPI):
  - Minimal **ontology API** (ADR-001): object types, objects, links, graph neighbourhood — every project and BOQ item lives in the semantic graph
  - **Projects API** + **BOQ vertical**: Excel import with RU/EN header auto-mapping, summary by CBS chapter with **regional coefficients**, styled Excel export
  - **CDE core** (D-01 lite): document register with per-type mask numbering, revision history, ISO 19650 state machine (WIP→Shared→Published→Archived) with suitability codes, transmittals with revision snapshots
  - **Executive report** (P-01 lite): one-call `.xlsx` for leadership — KPI status, cost by chapter, tunnel drives, open/overdue RFIs, recent daily reports
  - **Cost control vertical** (F-01/F-02 lite): cost transactions (Actual/Commitment/Forecast) against BOQ items, plan-vs-actual summary by CBS chapter with variance and spent%, budget version snapshots
  - **Tunnel vertical** (L-01/L-03): drives, bulk shift ring registration, chainage derivation, progress analytics (rings/day, S-curve, ETA to breakthrough) + tunnel dashboard at `/tunnel.html`
  - **Live dashboard** at `http://localhost:8000` (KPI cards, chapter chart, searchable items, import/export from the browser); OpenAPI docs at `/docs`
- `database/migrations/` — V000 foundation, V001 BOQ module, V002 ontology core + regional coefficients (full chain installed by CI on every PR)
- `scripts/generate_boq.py` — test-data generator

> Per ADR-003 the production core is Go; `core-py` is the executable specification
> and the tool the team uses today. Business logic stays within the BOQ vertical.

## Quick start

```bash
git clone https://github.com/ua5191086-boop/OpenConstructionERP.git
cd OpenConstructionERP
docker compose -f infrastructure/docker/docker-compose.dev.yml up -d --build
# BOQ dashboard:  http://localhost:8000      API docs: http://localhost:8000/docs
# PostgreSQL :5432 (oce/oce_dev_only), MinIO console :9001, Adminer :8080
# Migrations from database/migrations/ are applied automatically on first start.
python3 scripts/seed_reference_project.py   # optional: seed ALM-L3-REF reference project
# -> instantly a living system: $101.5M BOQ, twin TBM drives with 60 days of rings,
#    daily reports, RFIs, CDE documents, cost transactions
```

## Roadmap (summary — full version with exit criteria in SAD Tom 1, §11)

| Phase | Focus | Exit criterion |
|-------|-------|----------------|
| MVP | Ontology core, CDE, schedule import (XER), daily reports + offline foreman app, budget/IPC, tunnel minimum, exec dashboard | A real project runs its daily cycle only in the system for 60 days |
| Beta | Full CPM + baselines, EVM/cash flow, VO+claims, procurement/warehouse, TBM telemetry realtime, BIM viewer + BCF + 4D, AI copilot | Parallel run vs Primavera/Aconex, <1% divergence, legacy switched off |
| 1.0 | All 112 modules baseline, public GraphQL API, delay analysis, 5D, i18n (EN/RU/DE/TR/UZ) | Market-ready |
| Enterprise | Multi-cluster/DR, holding consolidation, digital twin 6D/7D, ML forecasting, SOC2 | Commercial open-core offering |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Architecture changes go through ADRs, never through drive-by PRs.
Licensing note: the open-core split (ADR-016) is pending sign-off; external contributions may require a CLA.

## License

AGPL-3.0 (open-core split under evaluation — see [ADR-016](docs/adr/ADR-016-licensing.md)).

## Contact

Project Owner: Ruslan Sarybaev — Deputy Chairman of the Board, CAI Interbudmontazh GmbH
