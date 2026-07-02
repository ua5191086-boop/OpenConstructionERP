-- ============================================================================
-- V047__Integration_Framework.sql
-- Фреймворк внешних интеграций
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Реестр внешних систем
-- ============================================================================
CREATE TABLE integration_systems (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(300) NOT NULL,
    system_type         VARCHAR(100) NOT NULL,                  -- sap, autodesk, bentley, sharepoint, telegram, bim360, powerbi, custom
    vendor              VARCHAR(200),
    version             VARCHAR(50),
    base_url            VARCHAR(500),
    auth_type           VARCHAR(50) NOT NULL DEFAULT 'api_key', -- api_key, oauth2, basic, certificate, none
    auth_config         JSONB,                                  -- зашифрованные credentials
    webhook_url         VARCHAR(500),
    health_check_url    VARCHAR(500),
    capabilities        JSONB DEFAULT '[]'::JSONB,             -- ["read_projects","write_documents","sync_schedule"]
    is_active           BOOLEAN DEFAULT TRUE,
    sync_frequency_min  INTEGER DEFAULT 60,                     -- периодичность синхронизации
    last_sync_at        TIMESTAMPTZ,
    last_sync_status    VARCHAR(50),                            -- success, failed, timeout
    error_count         INTEGER DEFAULT 0,
    retry_policy        JSONB DEFAULT '{"max_retries":3,"backoff_min":5,"backoff_max":60}'::JSONB,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_integration_type ON integration_systems(system_type);
CREATE INDEX idx_integration_active ON integration_systems(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE integration_systems IS 'Реестр внешних систем для интеграции';

-- ============================================================================
-- 2. Журнал синхронизации
-- ============================================================================
CREATE TABLE integration_sync_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id           UUID NOT NULL REFERENCES integration_systems(id) ON DELETE CASCADE,
    sync_type           VARCHAR(100) NOT NULL,                  -- full, incremental, one_way, bi_directional
    entity_type         VARCHAR(100) NOT NULL,                  -- project, schedule, document, boq, contract, invoice
    direction           VARCHAR(10) NOT NULL,                   -- inbound, outbound, both
    started_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ,
    status              VARCHAR(50) NOT NULL DEFAULT 'running', -- running, completed, failed, partial
    records_processed   INTEGER DEFAULT 0,
    records_succeeded   INTEGER DEFAULT 0,
    records_failed      INTEGER DEFAULT 0,
    error_details       JSONB,
    duration_sec        INTEGER,
    triggered_by        VARCHAR(50) DEFAULT 'schedule',        -- schedule, manual, webhook, api
    notes               TEXT
);

CREATE INDEX idx_sync_log_system ON integration_sync_log(system_id);
CREATE INDEX idx_sync_log_status ON integration_sync_log(status);
CREATE INDEX idx_sync_log_started ON integration_sync_log(started_at);

COMMENT ON TABLE integration_sync_log IS 'Журнал синхронизации с внешними системами';

-- ============================================================================
-- 3. Маппинг сущностей (внешний ID ↔ локальный ID)
-- ============================================================================
CREATE TABLE integration_entity_mappings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id           UUID NOT NULL REFERENCES integration_systems(id) ON DELETE CASCADE,
    entity_type         VARCHAR(100) NOT NULL,
    local_id            UUID NOT NULL,
    external_id         VARCHAR(500) NOT NULL,
    external_url        VARCHAR(500),
    external_version    VARCHAR(100),
    last_sync_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sync_status         VARCHAR(50) DEFAULT 'synced',           -- synced, pending, conflict, error
    conflict_details    TEXT,
    custom_data         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(system_id, entity_type, local_id),
    UNIQUE(system_id, entity_type, external_id)
);

CREATE INDEX idx_mappings_system ON integration_entity_mappings(system_id);
CREATE INDEX idx_mappings_type ON integration_entity_mappings(entity_type);
CREATE INDEX idx_mappings_local ON integration_entity_mappings(local_id);
CREATE INDEX idx_mappings_external ON integration_entity_mappings(external_id);

COMMENT ON TABLE integration_entity_mappings IS 'Маппинг между внешними ID и локальными ID сущностей';

-- ============================================================================
-- 4. Очередь событий для webhook-уведомлений
-- ============================================================================
CREATE TABLE integration_webhook_queue (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id           UUID REFERENCES integration_systems(id) ON DELETE CASCADE,
    event_type          VARCHAR(100) NOT NULL,                  -- project.updated, contract.approved, document.uploaded
    payload             JSONB NOT NULL,
    priority            INTEGER DEFAULT 0,
    status              VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, delivered, failed, expired
    retry_count         INTEGER DEFAULT 0,
    max_retries         INTEGER DEFAULT 3,
    last_attempt_at     TIMESTAMPTZ,
    next_attempt_at     TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    response_status     INTEGER,
    response_body       TEXT,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_queue_status ON integration_webhook_queue(status);
CREATE INDEX idx_webhook_queue_next ON integration_webhook_queue(next_attempt_at) WHERE status = 'pending';
CREATE INDEX idx_webhook_queue_system ON integration_webhook_queue(system_id);

COMMENT ON TABLE integration_webhook_queue IS 'Очередь событий для отправки во внешние системы через webhook';