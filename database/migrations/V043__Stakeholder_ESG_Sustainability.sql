-- ============================================================================
-- V043__Stakeholder_ESG_Sustainability.sql
-- Управление стейкхолдерами, ESG и устойчивое развитие
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Стейкхолдеры проекта
-- ============================================================================
CREATE TABLE stakeholders (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    stakeholder_type    VARCHAR(100) NOT NULL,                  -- client, investor, regulator, community, ngo, contractor, supplier, designer
    name                VARCHAR(500) NOT NULL,
    organization        VARCHAR(500),
    contact_person      VARCHAR(300),
    email               VARCHAR(300),
    phone               VARCHAR(100),
    interest_level      VARCHAR(50) DEFAULT 'medium',           -- high, medium, low
    influence_level     VARCHAR(50) DEFAULT 'medium',           -- high, medium, low
    engagement_strategy TEXT,
    expectations        TEXT,
    concerns            TEXT,
    communication_freq  VARCHAR(100),                           -- daily, weekly, monthly, quarterly
    last_contact_date   DATE,
    next_contact_date   DATE,
    status              VARCHAR(50) DEFAULT 'active',           -- active, dormant, archived
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stakeholders_project ON stakeholders(project_id);
CREATE INDEX idx_stakeholders_type ON stakeholders(stakeholder_type);
CREATE INDEX idx_stakeholders_status ON stakeholders(status);

COMMENT ON TABLE stakeholders IS 'Реестр стейкхолдеров проекта';

-- ============================================================================
-- 2. Коммуникации со стейкхолдерами
-- ============================================================================
CREATE TABLE stakeholder_communications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stakeholder_id      UUID NOT NULL REFERENCES stakeholders(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    communication_type  VARCHAR(100) NOT NULL,                  -- meeting, email, call, report, presentation, site_visit
    subject             VARCHAR(500) NOT NULL,
    summary             TEXT,
    outcome             TEXT,
    action_items        TEXT,                                   -- JSON-массив: [{"what":"...","owner":"...","due":"..."}]
    communication_date  DATE NOT NULL,
    duration_minutes    INTEGER,
    conducted_by        VARCHAR(200),
    participants        TEXT,                                   -- JSON-массив участников
    document_ref        VARCHAR(500),
    follow_up_date      DATE,
    follow_up_status    VARCHAR(50) DEFAULT 'pending',          -- pending, completed, overdue
    satisfaction_score  INTEGER CHECK (satisfaction_score BETWEEN 1 AND 5),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stakeholder_comm_stakeholder ON stakeholder_communications(stakeholder_id);
CREATE INDEX idx_stakeholder_comm_project ON stakeholder_communications(project_id);
CREATE INDEX idx_stakeholder_comm_date ON stakeholder_communications(communication_date);
CREATE INDEX idx_stakeholder_comm_followup ON stakeholder_communications(follow_up_status) WHERE follow_up_status = 'pending';

COMMENT ON TABLE stakeholder_communications IS 'Журнал коммуникаций со стейкхолдерами';

-- ============================================================================
-- 3. ESG-показатели
-- ============================================================================
CREATE TABLE esg_metrics (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    category            VARCHAR(50) NOT NULL,                   -- environmental, social, governance
    metric_code         VARCHAR(100) NOT NULL,
    metric_name         VARCHAR(500) NOT NULL,
    metric_description  TEXT,
    unit                VARCHAR(50),
    target_value        NUMERIC(18,4),
    current_value       NUMERIC(18,4),
    measurement_date    DATE NOT NULL,
    reporting_period    VARCHAR(50),                            -- Q1-2026, Y2026, etc.
    data_source         VARCHAR(200),
    verified_by         VARCHAR(200),
    verified_at         TIMESTAMPTZ,
    status              VARCHAR(50) DEFAULT 'on_track',         -- on_track, at_risk, off_track, not_measured
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_esg_project ON esg_metrics(project_id);
CREATE INDEX idx_esg_category ON esg_metrics(category);
CREATE INDEX idx_esg_period ON esg_metrics(reporting_period);
CREATE INDEX idx_esg_status ON esg_metrics(status);

COMMENT ON TABLE esg_metrics IS 'ESG-показатели проекта — Environmental, Social, Governance';

-- Environmental
COMMENT ON COLUMN esg_metrics.metric_code IS 'E—CO2: carbon_emissions, E—ENERGY: energy_consumption, E—WATER: water_usage, E—WASTE: waste_generation, E—NOISE: noise_level, E—BIODIVERSITY: biodiversity_impact';
COMMENT ON COLUMN esg_metrics.metric_code IS 'S—SAFETY: safety_incidents, S—TRAINING: training_hours, S—LOCAL: local_employment, S—COMMUNITY: community_complaints, S—DIVERSITY: workforce_diversity';
COMMENT ON COLUMN esg_metrics.metric_code IS 'G—COMPLIANCE: compliance_breaches, G—ETHICS: ethics_reports, G—TRANSPARENCY: disclosure_score, G—RISK: esg_risk_score';

-- ============================================================================
-- 4. Устойчивость цепочки поставок (Sustainability)
-- ============================================================================
CREATE TABLE supply_chain_sustainability (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    supplier_id         UUID REFERENCES procurement_suppliers(id) ON DELETE SET NULL,
    supplier_name       VARCHAR(500) NOT NULL,
    assessment_date     DATE NOT NULL,
    assessment_type     VARCHAR(100) NOT NULL,                  -- environmental, social, ethical, compliance
    score               NUMERIC(5,2) NOT NULL,                  -- 0-100
    max_score           NUMERIC(5,2) DEFAULT 100,
    certification       VARCHAR(500),                           -- ISO 14001, SA 8000, etc.
    certification_expiry DATE,
    risk_level          VARCHAR(50) DEFAULT 'low',             -- low, medium, high, critical
    findings            TEXT,                                   -- ключевые находки
    improvement_plan    TEXT,                                   -- план улучшений
    reassessment_date   DATE,
    status              VARCHAR(50) DEFAULT 'compliant',       -- compliant, non_compliant, under_review, improvement
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scs_project ON supply_chain_sustainability(project_id);
CREATE INDEX idx_scs_supplier ON supply_chain_sustainability(supplier_id);
CREATE INDEX idx_scs_risk ON supply_chain_sustainability(risk_level);

COMMENT ON TABLE supply_chain_sustainability IS 'Оценка устойчивости цепочки поставок';

-- ============================================================================
-- 5. Углеродный след проекта
-- ============================================================================
CREATE TABLE carbon_footprint (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scope               INTEGER NOT NULL,                      -- 1: прямые выбросы, 2: энергия, 3: цепочка поставок
    category            VARCHAR(200) NOT NULL,                  -- concrete, steel, transport, energy, waste, travel
    source_description  TEXT,
    co2_amount          NUMERIC(18,4) NOT NULL,                -- тонн CO2-eq
    co2_unit            VARCHAR(50) DEFAULT 'tCO2e',
    activity_data       NUMERIC(18,4),                          -- объём деятельности (например, тонн бетона)
    emission_factor     NUMERIC(18,6),                          -- коэффициент выбросов
    emission_factor_source VARCHAR(200),
    reporting_period    VARCHAR(50),
    verified_by         VARCHAR(200),
    verified_at         TIMESTAMPTZ,
    status              VARCHAR(50) DEFAULT 'estimated',        -- estimated, calculated, verified
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_carbon_project ON carbon_footprint(project_id);
CREATE INDEX idx_carbon_scope ON carbon_footprint(scope);
CREATE INDEX idx_carbon_period ON carbon_footprint(reporting_period);

COMMENT ON TABLE carbon_footprint IS 'Учёт углеродного следа проекта по Scope 1, 2, 3';