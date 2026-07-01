# Architecture Decision Records

ADR-001..015 codify the General Decisions (GD-01..GD-15) from [SAD Tom 1](../sad/SAD-Tom1-Architecture-v1.0.md). Changes to any accepted ADR go through a new superseding ADR, not by editing history.

| ADR | Decision | Status |
|-----|----------|--------|
| [001](ADR-001.md) | Ontology-first core (Foundry-style semantic layer) | Accepted |
| [002](ADR-002.md) | Modular monolith core + microservices at the periphery | Accepted |
| [003](ADR-003.md) | Backend languages: Go (core/services) + Python (AI, analytics, BIM processing) + TypeScript (BFF/gateway) | Accepted |
| [004](ADR-004.md) | Primary database: PostgreSQL 16 with extensions — PostGIS, TimescaleDB, pgvector, Apache AGE, ltree | Accepted |
| [005](ADR-005.md) | Analytical database: ClickHouse (CDC via Debezium -> Kafka) | Accepted |
| [006](ADR-006.md) | Event backbone: Apache Kafka (domain events, permanent retention) + NATS (realtime/mobile sync) | Accepted |
| [007](ADR-007.md) | Frontend: React 18 + TypeScript, Module Federation, own design system | Accepted |
| [008](ADR-008.md) | Mobile: Flutter, offline-first (SQLite + event-log sync, CRDT conflict resolution) | Accepted |
| [009](ADR-009.md) | AuthN/AuthZ: Keycloak (OIDC/SAML/LDAP) + OPA/Rego for ABAC policies | Accepted |
| [010](ADR-010.md) | File storage: MinIO (S3-compatible), versioning, WORM Object Lock for contractual documents | Accepted |
| [011](ADR-011.md) | Search: OpenSearch (BM25 + kNN hybrid for RAG) | Accepted |
| [012](ADR-012.md) | BIM: full OpenBIM — IfcOpenShell (IFC 4.3 incl. infrastructure entities) + xeokit viewer + own BCF 3.0 server | Accepted |
| [013](ADR-013.md) | AI layer: LLM-agnostic gateway (LiteLLM: Claude/GPT/local Llama via vLLM), LangGraph agents, hybrid RAG (pgvector + OpenSearch) | Accepted |
| [014](ADR-014.md) | Multi-tenancy: row-level security within an installation + separate installations per holding | Accepted |
| [015](ADR-015.md) | Deployment: Kubernetes-first (Helm, ArgoCD) with a mandatory single-node Docker Compose profile | Accepted |
| [016](ADR-016-licensing.md) | Open-core licensing | **Proposed** |
