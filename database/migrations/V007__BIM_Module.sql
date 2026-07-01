-- ============================================================================
-- V007__BIM_Module.sql
-- Модуль информационного моделирования (BIM Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. BIM-модели / проекты
-- ============================================================================
CREATE TABLE bim_models (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    model_name      VARCHAR(500) NOT NULL,
    model_version   VARCHAR(50) NOT NULL DEFAULT '1.0',
    description     TEXT,
    discipline      VARCHAR(100) NOT NULL,                 -- architectural, structural, MEP, civil, geotechnical
    author          VARCHAR(200),
    software        VARCHAR(200),                          -- Revit, Tekla, Allplan, Civil3D
    file_format     VARCHAR(50),                            -- IFC, RVT, DGN, DWG, SKP
    file_path       VARCHAR(1000),
    file_size       BIGINT,
    ifc_schema      VARCHAR(50),                            -- IFC2X3, IFC4, IFC4x3
    lod             VARCHAR(20),                            -- LOD 100-500
    status          VARCHAR(50) NOT NULL DEFAULT 'uploaded', -- uploaded, processing, published, archived
    checksum        VARCHAR(64),
    is_latest       BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    uploaded_by     BIGINT REFERENCES employees(id),
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, model_name, model_version)
);

CREATE INDEX idx_bim_project ON bim_models(project_id);
CREATE INDEX idx_bim_discipline ON bim_models(discipline);
CREATE INDEX idx_bim_latest ON bim_models(project_id, is_latest) WHERE is_latest = TRUE;

-- ============================================================================
-- 2. Элементы модели (IFC-объекты)
-- ============================================================================
CREATE TABLE bim_elements (
    id              BIGSERIAL PRIMARY KEY,
    model_id        BIGINT NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    ifc_global_id   VARCHAR(100),                          -- IfcGloballyUniqueId
    ifc_type        VARCHAR(200) NOT NULL,                  -- IfcWall, IfcSlab, IfcBeam, IfcColumn, IfcPipeSegment
    ifc_class       VARCHAR(100),                           -- IfcBuildingElement, IfcFlowSegment
    name            VARCHAR(500),
    description     TEXT,
    level           VARCHAR(200),                          -- этаж / уровень
    material        VARCHAR(200),
    volume          NUMERIC(18,4),                          -- объём, м³
    area            NUMERIC(18,4),                          -- площадь, м²
    length          NUMERIC(18,4),                          -- длина, м
    weight          NUMERIC(18,4),                          -- вес, кг
    elevation       NUMERIC(10,2),                          -- отметка
    x_position      NUMERIC(12,4),
    y_position      NUMERIC(12,4),
    z_position      NUMERIC(12,4),
    properties      JSONB,                                  -- пользовательские свойства
    status          VARCHAR(50) NOT NULL DEFAULT 'active',  -- active, modified, demolished, temporary
    boq_item_id     BIGINT REFERENCES boq_items(id),       -- привязка к смете
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bim_elements_model ON bim_elements(model_id);
CREATE INDEX idx_bim_elements_type ON bim_elements(ifc_type);
CREATE INDEX idx_bim_elements_level ON bim_elements(level);

-- ============================================================================
-- 3. Связи между элементами
-- ============================================================================
CREATE TABLE bim_relationships (
    id              BIGSERIAL PRIMARY KEY,
    model_id        BIGINT NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    source_element_id BIGINT NOT NULL REFERENCES bim_elements(id) ON DELETE CASCADE,
    target_element_id BIGINT NOT NULL REFERENCES bim_elements(id) ON DELETE CASCADE,
    relationship_type VARCHAR(100) NOT NULL,                -- IfcRelContainedInSpatialStructure, IfcRelVoidsElement, IfcRelFillsElement
    description     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bim_rel_source ON bim_relationships(source_element_id);
CREATE INDEX idx_bim_rel_target ON bim_relationships(target_element_id);

-- ============================================================================
-- 4. Clash Detection / коллизии
-- ============================================================================
CREATE TABLE bim_clashes (
    id              BIGSERIAL PRIMARY KEY,
    model_id        BIGINT NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    clash_group     VARCHAR(200),
    clash_type      VARCHAR(100) NOT NULL,                 -- hard, soft, clearance
    severity        VARCHAR(50) NOT NULL DEFAULT 'medium', -- critical, high, medium, low
    status          VARCHAR(50) NOT NULL DEFAULT 'open',   -- open, in_progress, resolved, approved
    element_a_id    BIGINT REFERENCES bim_elements(id),
    element_b_id    BIGINT REFERENCES bim_elements(id),
    element_a_name  VARCHAR(500),
    element_b_name  VARCHAR(500),
    distance        NUMERIC(10,4),                          -- расстояние между элементами
    tolerance       NUMERIC(10,4),                          -- допуск
    location_x      NUMERIC(12,4),
    location_y      NUMERIC(12,4),
    location_z      NUMERIC(12,4),
    screenshot_path VARCHAR(1000),
    assigned_to     BIGINT REFERENCES employees(id),
    resolution      TEXT,
    resolved_by     BIGINT REFERENCES employees(id),
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bim_clashes_model ON bim_clashes(model_id);
CREATE INDEX idx_bim_clashes_status ON bim_clashes(status);
CREATE INDEX idx_bim_clashes_severity ON bim_clashes(severity);

-- ============================================================================
-- 5. Версии модели / ревизии
-- ============================================================================
CREATE TABLE bim_revisions (
    id              BIGSERIAL PRIMARY KEY,
    model_id        BIGINT NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    revision_number INTEGER NOT NULL,
    revision_date   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    author          VARCHAR(200),
    description     TEXT,
    changes         TEXT,                                   -- описание изменений
    file_path       VARCHAR(1000),
    file_size       BIGINT,
    status          VARCHAR(50) NOT NULL DEFAULT 'submitted', -- submitted, reviewed, approved, rejected
    reviewed_by     BIGINT REFERENCES employees(id),
    reviewed_at     TIMESTAMPTZ,
    approved_by     BIGINT REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    UNIQUE(model_id, revision_number)
);

-- ============================================================================
-- 6. Комментарии к модели / разметка
-- ============================================================================
CREATE TABLE bim_markups (
    id              BIGSERIAL PRIMARY KEY,
    model_id        BIGINT NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    element_id      BIGINT REFERENCES bim_elements(id),
    author_id       BIGINT REFERENCES employees(id),
    markup_type     VARCHAR(50) NOT NULL,                   -- comment, issue, question, approval, redline
    title           VARCHAR(500),
    description     TEXT,
    position_x      NUMERIC(12,4),
    position_y      NUMERIC(12,4),
    position_z      NUMERIC(12,4),
    status          VARCHAR(50) NOT NULL DEFAULT 'open',    -- open, resolved, closed
    assigned_to     BIGINT REFERENCES employees(id),
    parent_id       BIGINT REFERENCES bim_markups(id),      -- для тредов
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bim_markups_model ON bim_markups(model_id);
CREATE INDEX idx_bim_markups_element ON bim_markups(element_id);

-- ============================================================================
-- 7. 4D / 5D связи (BIM + Время + Стоимость)
-- ============================================================================
CREATE TABLE bim_4d_links (
    id              BIGSERIAL PRIMARY KEY,
    element_id      BIGINT NOT NULL REFERENCES bim_elements(id) ON DELETE CASCADE,
    milestone_id    BIGINT REFERENCES contract_milestones(id),
    activity_code   VARCHAR(100),
    planned_start   DATE,
    planned_end     DATE,
    actual_start    DATE,
    actual_end      DATE,
    duration_days   INTEGER,
    status          VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, in_progress, completed, delayed
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bim_4d_element ON bim_4d_links(element_id);

-- ============================================================================
-- 8. BIM-метрики / статистика
-- ============================================================================
CREATE TABLE bim_metrics (
    id              BIGSERIAL PRIMARY KEY,
    model_id        BIGINT NOT NULL REFERENCES bim_models(id) ON DELETE CASCADE,
    report_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    total_elements  INTEGER,
    unique_types    INTEGER,
    total_clashes   INTEGER,
    open_clashes    INTEGER,
    resolved_clashes INTEGER,
    model_size_mb   NUMERIC(10,2),
    lod_achieved    VARCHAR(20),
    completeness_pct NUMERIC(5,2),                          -- % завершённости модели
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(model_id, report_date)
);

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE bim_models IS 'BIM-модели / проекты';
COMMENT ON TABLE bim_elements IS 'Элементы модели (IFC-объекты)';
COMMENT ON TABLE bim_relationships IS 'Связи между элементами';
COMMENT ON TABLE bim_clashes IS 'Коллизии / Clash Detection';
COMMENT ON TABLE bim_revisions IS 'Версии / ревизии модели';
COMMENT ON TABLE bim_markups IS 'Комментарии и разметка';
COMMENT ON TABLE bim_4d_links IS '4D/5D связи (BIM + время + стоимость)';
COMMENT ON TABLE bim_metrics IS 'Метрики и статистика модели';
