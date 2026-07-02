-- ============================================================================
-- V033__Neo4j_Kafka_Module.sql
-- Infrastructure: Neo4j Knowledge Graph + Kafka Event Bus
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Knowledge Graph Nodes (типы узлов, метки)
-- ============================================================================
CREATE TABLE knowledge_graph_nodes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_type       VARCHAR(100) NOT NULL,                    -- project, contract, document, equipment, etc.
    node_label      VARCHAR(500),
    node_properties JSONB DEFAULT '{}',
    neo4j_id        BIGINT,                                   -- ID in Neo4j after sync
    is_synced       BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_kg_nodes_type ON knowledge_graph_nodes(node_type);
CREATE INDEX idx_kg_nodes_synced ON knowledge_graph_nodes(is_synced);

-- ============================================================================
-- 2. Knowledge Graph Edges (типы связей)
-- ============================================================================
CREATE TABLE knowledge_graph_edges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    edge_type       VARCHAR(100) NOT NULL,                    -- belongs_to, references, depends_on, etc.
    source_node_id  UUID NOT NULL REFERENCES knowledge_graph_nodes(id),
    target_node_id  UUID NOT NULL REFERENCES knowledge_graph_nodes(id),
    edge_properties JSONB DEFAULT '{}',
    neo4j_id        BIGINT,
    is_synced       BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_kg_edges_type ON knowledge_graph_edges(edge_type);
CREATE INDEX idx_kg_edges_source ON knowledge_graph_edges(source_node_id);
CREATE INDEX idx_kg_edges_target ON knowledge_graph_edges(target_node_id);

-- ============================================================================
-- 3. Graph Sync Queue (очередь для синхронизации с Neo4j)
-- ============================================================================
CREATE TABLE graph_sync_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation       VARCHAR(20) NOT NULL,                     -- create_node, update_node, delete_node, create_edge, delete_edge
    entity_type     VARCHAR(50) NOT NULL,                     -- node, edge
    entity_id       UUID NOT NULL,
    payload         JSONB,
    status          VARCHAR(50) DEFAULT 'pending',             -- pending, processing, done, error
    error_message   TEXT,
    retry_count     INTEGER DEFAULT 0,
    max_retries     INTEGER DEFAULT 3,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at    TIMESTAMPTZ
);
CREATE INDEX idx_graph_sync_status ON graph_sync_queue(status);
CREATE INDEX idx_graph_sync_created ON graph_sync_queue(created_at);

-- ============================================================================
-- 4. Kafka Topics (реестр топиков)
-- ============================================================================
CREATE TABLE kafka_topics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_name      VARCHAR(200) NOT NULL UNIQUE,
    description     TEXT,
    partitions      INTEGER DEFAULT 1,
    replication_factor INTEGER DEFAULT 1,
    config          JSONB DEFAULT '{}',
    is_internal     BOOLEAN DEFAULT FALSE,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 5. Kafka Events (логи событий)
-- ============================================================================
CREATE TABLE kafka_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id        UUID REFERENCES kafka_topics(id),
    topic_name      VARCHAR(200) NOT NULL,
    event_type      VARCHAR(200),
    event_key       VARCHAR(500),
    event_value     JSONB,
    headers         JSONB DEFAULT '{}',
    partition_id    INTEGER,
    offset_id       BIGINT,
    producer        VARCHAR(200),
    status          VARCHAR(50) DEFAULT 'published',           -- published, consumed, failed
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_kafka_events_topic ON kafka_events(topic_name);
CREATE INDEX idx_kafka_events_created ON kafka_events(created_at);

-- ============================================================================
-- 6. Kafka Consumers (реестр consumers)
-- ============================================================================
CREATE TABLE kafka_consumers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    consumer_name   VARCHAR(300) NOT NULL,
    group_id        VARCHAR(200) NOT NULL,
    topic_pattern   VARCHAR(500),
    description     TEXT,
    status          VARCHAR(50) DEFAULT 'active',              -- active, paused, stopped
    config          JSONB DEFAULT '{}',
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_kafka_consumers_group ON kafka_consumers(group_id);

-- Seed default Kafka topics
INSERT INTO kafka_topics (topic_name, description, partitions, is_internal) VALUES
    ('oce.project.created', 'Project created events', 3, TRUE),
    ('oce.project.updated', 'Project updated events', 3, TRUE),
    ('oce.document.uploaded', 'Document uploaded events', 3, TRUE),
    ('oce.contract.signed', 'Contract signed events', 3, TRUE),
    ('oce.finance.invoice.paid', 'Invoice paid events', 3, TRUE),
    ('oce.quality.ncr.raised', 'NCR raised events', 2, TRUE),
    ('oce.hse.incident.reported', 'HSE incident reported events', 2, TRUE),
    ('oce.sync.graph', 'Graph sync trigger events', 2, TRUE);