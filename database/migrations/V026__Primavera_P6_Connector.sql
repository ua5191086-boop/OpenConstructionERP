-- ============================================================================
-- V026__Primavera_P6_Connector.sql
-- Primavera P6 Connector — маппинг P6 ↔ OpenConstructionERP
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Маппинг проектов P6
-- ============================================================================
CREATE TABLE p6_projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    p6_project_id   VARCHAR(100) NOT NULL,                   -- P6 internal ID
    p6_uid          VARCHAR(100),                             -- P6 GUID
    p6_project_code VARCHAR(200),                             -- код проекта в P6
    p6_project_name VARCHAR(500),
    last_sync_at    TIMESTAMPTZ,
    sync_status     VARCHAR(50) DEFAULT 'pending',           -- pending, synced, error
    sync_error      TEXT,
    config          JSONB,                                   -- доп. настройки синхронизации
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(p6_project_id)
);

CREATE INDEX idx_p6_projects_uid ON p6_projects(p6_project_id);
CREATE INDEX idx_p6_projects_status ON p6_projects(sync_status);

-- ============================================================================
-- 2. Маппинг WBS
-- ============================================================================
CREATE TABLE p6_wbs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    p6_project_id   VARCHAR(100) NOT NULL,                   -- P6 project reference
    p6_wbs_id       VARCHAR(100) NOT NULL,                   -- P6 WBS ID
    p6_wbs_code     VARCHAR(200),
    p6_wbs_name     VARCHAR(500),
    p6_parent_wbs_id VARCHAR(100),                           -- иерархия
    level           INTEGER DEFAULT 0,
    wbs_path        TEXT,                                    -- полный путь
    mapped_element_type VARCHAR(50),                         -- control_account, wbs_element
    mapped_element_id   UUID,                                -- ссылка на OCE entity
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(p6_project_id, p6_wbs_id)
);

CREATE INDEX idx_p6_wbs_project ON p6_wbs(p6_project_id);
CREATE INDEX idx_p6_wbs_parent ON p6_wbs(p6_parent_wbs_id);

-- ============================================================================
-- 3. Маппинг активностей (Activities)
-- ============================================================================
CREATE TABLE p6_activities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    p6_project_id   VARCHAR(100) NOT NULL,
    p6_wbs_id       VARCHAR(100),
    p6_activity_id  VARCHAR(100) NOT NULL,                   -- P6 Activity ID
    p6_activity_code VARCHAR(200),
    p6_activity_name VARCHAR(500),
    activity_type   VARCHAR(100),                            -- Task, Milestone, WBS Summary, Resource Dependent
    status          VARCHAR(50),                              -- Not Started, In Progress, Completed
    planned_start   TIMESTAMPTZ,
    planned_finish  TIMESTAMPTZ,
    actual_start    TIMESTAMPTZ,
    actual_finish   TIMESTAMPTZ,
    remaining_duration INTEGER,                               -- days
    at_completion_duration INTEGER,
    percent_complete NUMERIC(5,2),
    physical_complete NUMERIC(5,2),
    duration_type   VARCHAR(50),                              -- Fixed Duration, Fixed Units, Fixed Work
    mapped_to_type  VARCHAR(50),                              -- activity, milestone, wbs_element
    mapped_element_id UUID,                                   -- ссылка на OCE
    is_active       BOOLEAN DEFAULT TRUE,
    last_sync_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(p6_project_id, p6_activity_id)
);

CREATE INDEX idx_p6_activities_project ON p6_activities(p6_project_id);
CREATE INDEX idx_p6_activities_wbs ON p6_activities(p6_wbs_id);
CREATE INDEX idx_p6_activities_code ON p6_activities(p6_activity_code);

-- ============================================================================
-- 4. Связи (Predecessors / Successors)
-- ============================================================================
CREATE TABLE p6_relationships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    p6_project_id   VARCHAR(100) NOT NULL,
    predecessor_id  VARCHAR(100) NOT NULL,                   -- p6_activity_id of predecessor
    successor_id    VARCHAR(100) NOT NULL,                   -- p6_activity_id of successor
    relationship_type VARCHAR(20) NOT NULL DEFAULT 'FS',     -- FS, FF, SS, SF
    lag_days        INTEGER DEFAULT 0,
    lag_type        VARCHAR(20) DEFAULT 'finish_to_start',   -- finish_to_start, start_to_start, etc
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(p6_project_id, predecessor_id, successor_id)
);

CREATE INDEX idx_p6_rel_pred ON p6_relationships(predecessor_id);
CREATE INDEX idx_p6_rel_succ ON p6_relationships(successor_id);

-- ============================================================================
-- 5. Маппинг ресурсов
-- ============================================================================
CREATE TABLE p6_resources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    p6_project_id   VARCHAR(100) NOT NULL,
    p6_resource_id  VARCHAR(100) NOT NULL,
    p6_resource_name VARCHAR(500),
    resource_type   VARCHAR(50),                              -- Labor, Material, Equipment, Expense
    unit_of_measure VARCHAR(50),
    unit_price      NUMERIC(18,4),
    currency        VARCHAR(3) DEFAULT 'USD',
    mapped_to_type  VARCHAR(50),                              -- employee, equipment, material
    mapped_element_id UUID,                                   -- ссылка на OCE
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(p6_project_id, p6_resource_id)
);

CREATE INDEX idx_p6_resources_project ON p6_resources(p6_project_id);

-- ============================================================================
-- 6. Лог синхронизации
-- ============================================================================
CREATE TABLE p6_sync_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id),
    p6_project_id   VARCHAR(100),
    sync_type       VARCHAR(50) NOT NULL,                    -- full, incremental, activities, resources, wbs
    status          VARCHAR(50) NOT NULL DEFAULT 'running',  -- running, completed, failed
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ,
    duration_sec    INTEGER,
    records_processed INTEGER DEFAULT 0,
    records_created INTEGER DEFAULT 0,
    records_updated INTEGER DEFAULT 0,
    records_deleted INTEGER DEFAULT 0,
    sync_file       TEXT,                                     -- имя XER-файла
    error_message   TEXT,
    details         JSONB
);

CREATE INDEX idx_p6_sync_project ON p6_sync_log(project_id);
CREATE INDEX idx_p6_sync_status ON p6_sync_log(status);
CREATE INDEX idx_p6_sync_started ON p6_sync_log(started_at DESC);

-- ============================================================================
-- Функция быстрого поиска по маппингу P6 → OCE
-- ============================================================================
CREATE OR REPLACE FUNCTION p6_find_local_entity(
    p_p6_project_id VARCHAR(100),
    p_p6_entity_id VARCHAR(100),
    p_entity_type VARCHAR(20)   -- 'activity', 'wbs', 'resource'
) RETURNS UUID AS $$
DECLARE
    v_mapped_id UUID;
BEGIN
    IF p_entity_type = 'activity' THEN
        SELECT mapped_element_id INTO v_mapped_id
        FROM p6_activities
        WHERE p6_project_id = p_p6_project_id AND p6_activity_id = p_p6_entity_id AND is_active = TRUE;
    ELSIF p_entity_type = 'wbs' THEN
        SELECT mapped_element_id INTO v_mapped_id
        FROM p6_wbs
        WHERE p6_project_id = p_p6_project_id AND p6_wbs_id = p_p6_entity_id AND is_active = TRUE;
    ELSIF p_entity_type = 'resource' THEN
        SELECT mapped_element_id INTO v_mapped_id
        FROM p6_resources
        WHERE p6_project_id = p_p6_project_id AND p6_resource_id = p_p6_entity_id AND is_active = TRUE;
    END IF;
    RETURN v_mapped_id;
END;
$$ LANGUAGE plpgsql;