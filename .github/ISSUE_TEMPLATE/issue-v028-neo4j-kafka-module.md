---
title: "[MODULE] Neo4j + Kafka — Knowledge Graph & Event Bus V028"
labels: enhancement, module, infrastructure
assignees: ""
---

## Описание модуля Neo4j + Kafka (V028)

Инфраструктурный модуль для OpenConstructionERP: граф знаний (Neo4j синхронизация) и событийная шина (Kafka).

### Миграция БД
- ✅ `database/migrations/V028__Neo4j_Kafka_Module.sql` — 6 таблиц:
  - `knowledge_graph_nodes` — узлы графа (project, contract, document, etc.)
  - `knowledge_graph_edges` — рёбра связей
  - `graph_sync_queue` — очередь синхронизации с Neo4j
  - `kafka_topics` — топики Kafka
  - `kafka_events` — события
  - `kafka_consumers` — потребители

### Go API хендлеры
- ✅ `services/core/internal/handlers/neo4j_kafka.go` — два роутера:
  - `/knowledge/graph`, `/knowledge/nodes`, `/knowledge/edges`, `/knowledge/sync`
  - `/events/topics`, `/events/publish`, `/events/events`, `/events/consumers`
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 6 типов (KnowledgeGraphNode, KnowledgeGraphEdge, GraphSyncQueue, KafkaTopic, KafkaEvent, KafkaConsumer)

### Генератор тестовых данных
- ✅ `scripts/generate_knowledge.py` — 30 узлов, 45 рёбер, 12 записей синхронизации, 6 топиков, 20 событий, 4 потребителя

### HTML-дашборды
- ✅ `apps/web/knowledge-graph-dashboard.html` — тёмная тема, Chart.js:
  - Nodes by Type (doughnut), Edges by Type (doughnut)
  - Events (bar), Sync Queue (pie)
  - Nodes table, Events table
- ✅ `apps/web/events-dashboard.html` — тёмная тема:
  - Events by Type (bar), Consumer Status (pie)
  - Topics table, Consumers table

### React-страница
- ✅ `apps/frontend/src/pages/KnowledgeGraphPage.tsx` — новая страница с KPI, таблицами узлов, топиков, событий

### API-клиенты (frontend)
- ✅ `apps/frontend/src/api.ts` — `knowledgeApi` + `eventsApi`

### Типы TypeScript
- ✅ `apps/frontend/src/types.ts` — 5 интерфейсов (KnowledgeGraphNode, KnowledgeGraphEdge, KafkaTopic, KafkaEvent, KafkaConsumer)

### Docker
- Для продакшена требуется добавить Neo4j и Kafka в `docker-compose.yml`