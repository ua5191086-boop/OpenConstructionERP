-- ============================================================================
-- V045__Asset_Management.sql
-- Управление активами и оборудованием
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Реестр активов
-- ============================================================================
CREATE TABLE asset_registry (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    asset_type          VARCHAR(100) NOT NULL,                  -- equipment, vehicle, tool, building, infrastructure, software, intangible
    asset_code          VARCHAR(100) NOT NULL,
    asset_name          VARCHAR(500) NOT NULL,
    serial_number       VARCHAR(200),
    manufacturer        VARCHAR(300),
    model               VARCHAR(200),
    year_manufactured   INTEGER,
    purchase_date       DATE,
    purchase_cost       NUMERIC(18,2),
    currency            VARCHAR(3) DEFAULT 'USD',
    current_value       NUMERIC(18,2),                          -- остаточная стоимость
    depreciation_method VARCHAR(50),                            -- straight_line, declining, units_of_production
    useful_life_years   INTEGER,
    salvage_value       NUMERIC(18,2),
    depreciation_rate   NUMERIC(5,2),
    location            VARCHAR(500),
    gps_coordinates     JSONB,                                  -- {"lat":...,"lng":...}
    assigned_to         VARCHAR(300),
    department          VARCHAR(200),
    status              VARCHAR(50) NOT NULL DEFAULT 'operational', -- operational, maintenance, idle, disposed, sold, lost
    condition           VARCHAR(50) DEFAULT 'good',             -- excellent, good, fair, poor, critical
    warranty_expiry     DATE,
    insurance_policy    VARCHAR(200),
    insurance_value     NUMERIC(18,2),
    qr_code             VARCHAR(500),
    documents           JSONB DEFAULT '[]'::JSONB,              -- ссылки на документы
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_asset_project ON asset_registry(project_id);
CREATE INDEX idx_asset_type ON asset_registry(asset_type);
CREATE INDEX idx_asset_status ON asset_registry(status);
CREATE INDEX idx_asset_code ON asset_registry(asset_code);

COMMENT ON TABLE asset_registry IS 'Реестр всех активов проекта';

-- ============================================================================
-- 2. Движение активов
-- ============================================================================
CREATE TABLE asset_movements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id            UUID NOT NULL REFERENCES asset_registry(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    movement_type       VARCHAR(50) NOT NULL,                   -- transfer, assign, return, disposal, sale, write_off
    from_location       VARCHAR(500),
    to_location         VARCHAR(500),
    from_assignee       VARCHAR(300),
    to_assignee         VARCHAR(300),
    movement_date       DATE NOT NULL,
    reference_doc       VARCHAR(500),                            -- номер акта/накладной
    authorized_by       VARCHAR(200),
    reason              TEXT,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_asset_movements_asset ON asset_movements(asset_id);
CREATE INDEX idx_asset_movements_project ON asset_movements(project_id);
CREATE INDEX idx_asset_movements_date ON asset_movements(movement_date);

COMMENT ON TABLE asset_movements IS 'История перемещений и изменений статуса активов';

-- ============================================================================
-- 3. Проверки и инспекции активов
-- ============================================================================
CREATE TABLE asset_inspections (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id            UUID NOT NULL REFERENCES asset_registry(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    inspection_type     VARCHAR(100) NOT NULL,                  -- safety, maintenance, calibration, regulatory, insurance
    inspection_date     DATE NOT NULL,
    next_inspection_date DATE,
    inspector           VARCHAR(300),
    inspector_company   VARCHAR(300),
    result              VARCHAR(50) NOT NULL DEFAULT 'pass',    -- pass, fail, conditional, not_applicable
    findings            TEXT,
    recommendations     TEXT,
    action_taken        TEXT,
    cost                NUMERIC(18,2),
    document_ref        VARCHAR(500),
    status              VARCHAR(50) DEFAULT 'completed',        -- scheduled, completed, overdue, cancelled
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_asset_inspections_asset ON asset_inspections(asset_id);
CREATE INDEX idx_asset_inspections_date ON asset_inspections(next_inspection_date);
CREATE INDEX idx_asset_inspections_type ON asset_inspections(inspection_type);

COMMENT ON TABLE asset_inspections IS 'Инспекции и проверки состояния активов';

-- ============================================================================
-- 4. Амортизация активов
-- ============================================================================
CREATE TABLE asset_depreciation (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id            UUID NOT NULL REFERENCES asset_registry(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    period_start        DATE NOT NULL,
    period_end          DATE NOT NULL,
    depreciation_amount NUMERIC(18,2) NOT NULL,
    accumulated_depr    NUMERIC(18,2) NOT NULL,
    book_value          NUMERIC(18,2) NOT NULL,
    method              VARCHAR(50) NOT NULL,
    posted              BOOLEAN DEFAULT FALSE,
    posted_at           TIMESTAMPTZ,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(asset_id, period_start, period_end)
);

CREATE INDEX idx_asset_depr_asset ON asset_depreciation(asset_id);
CREATE INDEX idx_asset_depr_period ON asset_depreciation(period_start, period_end);
CREATE INDEX idx_asset_depr_posted ON asset_depreciation(posted) WHERE posted = FALSE;

COMMENT ON TABLE asset_depreciation IS 'Амортизация активов по периодам';