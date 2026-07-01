-- ============================================================================
-- V016__HSE_Module.sql
-- Модуль HSE (H-13) — Health, Safety & Environment
-- Safety permits, incident investigation, audits, PPE tracking, emergency response
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. HSE Incidents (расширенная таблица инцидентов)
-- ============================================================================
CREATE TABLE hse_incidents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    incident_number INTEGER NOT NULL,
    incident_code   VARCHAR(30) NOT NULL,                    -- 'INC-0001'
    title           VARCHAR(500) NOT NULL,
    description     TEXT NOT NULL,
    incident_type   VARCHAR(50) NOT NULL DEFAULT 'near_miss'
        CHECK (incident_type IN ('near_miss','first_aid','medical_treatment','lost_time_injury','restricted_work','fatality','property_damage','environmental','fire','explosion','chemical_spill','vehicle','other')),
    severity        VARCHAR(15) NOT NULL DEFAULT 'minor'
        CHECK (severity IN ('minor','moderate','major','critical','catastrophic')),
    incident_date   DATE NOT NULL,
    incident_time   TIME,
    location        VARCHAR(300),
    area            VARCHAR(200),
    activity_at_time VARCHAR(300),
    reported_by     VARCHAR(200),
    reported_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    affected_person VARCHAR(300),
    affected_part   VARCHAR(100),                            -- body part injured
    lost_days       INTEGER DEFAULT 0,
    restricted_days INTEGER DEFAULT 0,
    medical_cost    NUMERIC(12,2),
    property_cost   NUMERIC(12,2),
    total_cost      NUMERIC(12,2),
    root_cause      TEXT,
    direct_cause    TEXT,
    contributing_factors TEXT,
    corrective_action TEXT,
    preventive_action TEXT,
    investigation_status VARCHAR(30) NOT NULL DEFAULT 'open'
        CHECK (investigation_status IN ('open','investigating','report_draft','report_approved','closed','void')),
    investigation_lead VARCHAR(200),
    investigation_team TEXT,
    investigation_findings TEXT,
    lessons_learned TEXT,
    is_reportable   BOOLEAN DEFAULT FALSE,
    authority_notified BOOLEAN DEFAULT FALSE,
    authority_ref   VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','closed','void')),         -- overall incident status
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, incident_number),
    UNIQUE (project_id, incident_code)
);

CREATE INDEX idx_hsi_project ON hse_incidents(project_id);
CREATE INDEX idx_hsi_type ON hse_incidents(project_id, incident_type);
CREATE INDEX idx_hsi_severity ON hse_incidents(project_id, severity);
CREATE INDEX idx_hsi_investigation ON hse_incidents(project_id, investigation_status);
CREATE INDEX idx_hsi_date ON hse_incidents(project_id, incident_date DESC);

COMMENT ON TABLE hse_incidents IS 'HSE Incidents — регистрация происшествий (расширенная)';

-- ============================================================================
-- 2. HSE Permits (наряды-допуски)
-- ============================================================================
CREATE TABLE hse_permits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    permit_number   INTEGER NOT NULL,
    permit_code     VARCHAR(30) NOT NULL,                    -- 'PTW-0001'
    permit_type     VARCHAR(50) NOT NULL
        CHECK (permit_type IN ('hot_work','cold_work','confined_space','work_at_height','excavation','electrical','lifting','demolition','chemical','radiation','pressure_test','blasting','general')),
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    location        VARCHAR(300),
    work_description TEXT NOT NULL,
    issuing_authority VARCHAR(200),
    permit_holder   VARCHAR(200),
    responsible_person VARCHAR(200),
    hazard_assessment TEXT,
    control_measures TEXT,
    ppe_required    TEXT,
    valid_from      TIMESTAMPTZ NOT NULL,
    valid_to        TIMESTAMPTZ NOT NULL,
    extension_count INTEGER DEFAULT 0,
    extended_to     TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','issued','active','suspended','cancelled','expired','closed')),
    issued_at       TIMESTAMPTZ,
    issued_by       VARCHAR(200),
    closed_at       TIMESTAMPTZ,
    closed_by       VARCHAR(200),
    remarks         TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, permit_number),
    UNIQUE (project_id, permit_code)
);

CREATE INDEX idx_hsp_project ON hse_permits(project_id);
CREATE INDEX idx_hsp_type ON hse_permits(project_id, permit_type);
CREATE INDEX idx_hsp_status ON hse_permits(project_id, status);
CREATE INDEX idx_hsp_valid ON hse_permits(valid_from, valid_to) WHERE status = 'active';

COMMENT ON TABLE hse_permits IS 'HSE Permits — наряды-допуски на опасные работы';

-- ============================================================================
-- 3. HSE Audits (аудиты)
-- ============================================================================
CREATE TABLE hse_audits (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    audit_number    INTEGER NOT NULL,
    audit_code      VARCHAR(30) NOT NULL,                    -- 'AUD-0001'
    audit_type      VARCHAR(50) NOT NULL DEFAULT 'internal'
        CHECK (audit_type IN ('internal','external','regulatory','certification','supplier','corporate')),
    title           VARCHAR(500) NOT NULL,
    scope           TEXT NOT NULL,
    criteria        TEXT,                                    -- audit criteria / standards
    lead_auditor    VARCHAR(200),
    audit_team      TEXT,
    audit_date      DATE NOT NULL,
    location        VARCHAR(300),
    findings_count  INTEGER DEFAULT 0,
    non_conformities INTEGER DEFAULT 0,
    observations    INTEGER DEFAULT 0,
    opportunities   INTEGER DEFAULT 0,
    score_pct       NUMERIC(5,2),
    status          VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','reviewed','closed')),
    findings_summary TEXT,
    conclusion      TEXT,
    report_file     VARCHAR(500),
    follow_up_date  DATE,
    completed_at    DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, audit_number),
    UNIQUE (project_id, audit_code)
);

CREATE INDEX idx_hsa_project ON hse_audits(project_id);
CREATE INDEX idx_hsa_type ON hse_audits(project_id, audit_type);
CREATE INDEX idx_hsa_status ON hse_audits(project_id, status);

COMMENT ON TABLE hse_audits IS 'HSE Audits — аудиты безопасности';

-- ============================================================================
-- 4. HSE Inspections (инспекции)
-- ============================================================================
CREATE TABLE hse_inspections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    inspection_number INTEGER NOT NULL,
    inspection_code VARCHAR(30) NOT NULL,                    -- 'INSP-0001'
    inspection_type VARCHAR(50) NOT NULL DEFAULT 'routine'
        CHECK (inspection_type IN ('routine','focused','pre_work','toolbox_talk','safety_walk','regulatory','third_party','daily')),
    title           VARCHAR(500) NOT NULL,
    location        VARCHAR(300),
    area            VARCHAR(200),
    inspector       VARCHAR(200),
    inspection_date DATE NOT NULL,
    findings        TEXT,
    positive_observations TEXT,
    violations_found INTEGER DEFAULT 0,
    violations_resolved INTEGER DEFAULT 0,
    severity        VARCHAR(15) DEFAULT 'low'
        CHECK (severity IN ('low','medium','high','critical')),
    status          VARCHAR(20) NOT NULL DEFAULT 'completed'
        CHECK (status IN ('planned','in_progress','completed','reviewed','closed')),
    action_items    TEXT,
    follow_up_date  DATE,
    closed_at       DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, inspection_number),
    UNIQUE (project_id, inspection_code)
);

CREATE INDEX idx_hsi2_project ON hse_inspections(project_id);
CREATE INDEX idx_hsi2_type ON hse_inspections(project_id, inspection_type);
CREATE INDEX idx_hsi2_status ON hse_inspections(project_id, status);

COMMENT ON TABLE hse_inspections IS 'HSE Inspections — инспекции по безопасности';

-- ============================================================================
-- 5. HSE Training (обучение)
-- ============================================================================
CREATE TABLE hse_training (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    training_number INTEGER NOT NULL,
    training_code   VARCHAR(30) NOT NULL,                    -- 'TRN-0001'
    training_name   VARCHAR(500) NOT NULL,
    training_type   VARCHAR(50) NOT NULL DEFAULT 'safety_induction'
        CHECK (training_type IN ('safety_induction','refresher','first_aid','fire_safety','work_at_height','confined_space','chemical_handling','lifting_ops','emergency_response','defensive_driving','hazcom','scaffold','electrical_safety','excavation','other')),
    description     TEXT,
    trainer         VARCHAR(200),
    training_date   DATE NOT NULL,
    duration_hours  NUMERIC(5,1),
    location        VARCHAR(300),
    attendees       INTEGER DEFAULT 0,
    max_attendees   INTEGER,
    status          VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','scheduled','in_progress','completed','cancelled')),
    certificate_type VARCHAR(100),
    certificate_validity_days INTEGER,
    cost_per_person NUMERIC(10,2),
    total_cost      NUMERIC(12,2),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, training_number),
    UNIQUE (project_id, training_code)
);

CREATE INDEX idx_hst_project ON hse_training(project_id);
CREATE INDEX idx_hst_type ON hse_training(project_id, training_type);
CREATE INDEX idx_hst_date ON hse_training(project_id, training_date DESC);

COMMENT ON TABLE hse_training IS 'HSE Training — обучение по безопасности';

-- ============================================================================
-- 6. HSE PPE (СИЗ — средства индивидуальной защиты)
-- ============================================================================
CREATE TABLE hse_ppe (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ppe_code        VARCHAR(30) NOT NULL,
    ppe_name        VARCHAR(300) NOT NULL,
    ppe_category    VARCHAR(50) NOT NULL DEFAULT 'head'
        CHECK (ppe_category IN ('head','eye','face','hearing','respiratory','hand','foot','body','fall_protection','high_visibility','drowning')),
    manufacturer    VARCHAR(200),
    model           VARCHAR(200),
    size            VARCHAR(30),
    certification   VARCHAR(200),
    quantity_issued INTEGER DEFAULT 0,
    quantity_stock  INTEGER DEFAULT 0,
    reorder_level   INTEGER DEFAULT 5,
    unit_cost       NUMERIC(10,2),
    shelf_life_days INTEGER,
    expiry_date     DATE,
    storage_location VARCHAR(200),
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ppe_code)
);

CREATE INDEX idx_hsppe_project ON hse_ppe(project_id);
CREATE INDEX idx_hsppe_category ON hse_ppe(project_id, ppe_category);

COMMENT ON TABLE hse_ppe IS 'HSE PPE — средства индивидуальной защиты';

-- ============================================================================
-- 7. HSE Drill (учёт тренировок/учений)
-- ============================================================================
CREATE TABLE hse_drill (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    drill_number    INTEGER NOT NULL,
    drill_code      VARCHAR(30) NOT NULL,                    -- 'DRL-0001'
    drill_name      VARCHAR(500) NOT NULL,
    drill_type      VARCHAR(50) NOT NULL DEFAULT 'fire'
        CHECK (drill_type IN ('fire','first_aid','evacuation','confined_space_rescue','height_rescue','chemical_spill','earthquake','lockdown','medical_emergency','other')),
    description     TEXT,
    location        VARCHAR(300),
    drill_date      DATE NOT NULL,
    participants    INTEGER DEFAULT 0,
    duration_minutes INTEGER,
    evaluator       VARCHAR(200),
    score_pct       NUMERIC(5,2),
    observations    TEXT,
    improvements    TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','debriefed','cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, drill_number),
    UNIQUE (project_id, drill_code)
);

CREATE INDEX idx_hsd_project ON hse_drill(project_id);
CREATE INDEX idx_hsd_type ON hse_drill(project_id, drill_type);

COMMENT ON TABLE hse_drill IS 'HSE Drill — учебные тревоги и тренировки';

-- ============================================================================
-- 8. HSE Statistics (статистика)
-- ============================================================================
CREATE TABLE hse_statistics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    report_month    DATE NOT NULL,                           -- first day of month
    manhours        NUMERIC(12,0) DEFAULT 0,
    lost_time_injuries INTEGER DEFAULT 0,
    recordable_injuries INTEGER DEFAULT 0,
    fatalities      INTEGER DEFAULT 0,
    near_misses     INTEGER DEFAULT 0,
    first_aid_cases INTEGER DEFAULT 0,
    property_damage INTEGER DEFAULT 0,
    environmental_incidents INTEGER DEFAULT 0,
    vehicle_incidents INTEGER DEFAULT 0,
    fire_incidents  INTEGER DEFAULT 0,
    lti_frequency   NUMERIC(8,2),                            -- Lost Time Injury Frequency (per 1M hrs)
    lti_severity    NUMERIC(8,2),                            -- Lost Time Injury Severity (days lost per 1M hrs)
    total_recordable_rate NUMERIC(8,2),                     -- TRIR
    days_since_last_lti   INTEGER,
    days_since_last_fatality INTEGER,
    safety_training_hours  NUMERIC(10,0) DEFAULT 0,
    inspections_conducted  INTEGER DEFAULT 0,
    audits_conducted       INTEGER DEFAULT 0,
    permits_issued         INTEGER DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, report_month)
);

CREATE INDEX idx_hsstats_project ON hse_statistics(project_id);
CREATE INDEX idx_hsstats_month ON hse_statistics(project_id, report_month DESC);

COMMENT ON TABLE hse_statistics IS 'HSE Statistics — сводная статистика безопасности';

-- ============================================================================
-- 9. HSE Emergency Plans (планы действий в ЧС)
-- ============================================================================
CREATE TABLE hse_emergency_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    plan_number     INTEGER NOT NULL,
    plan_code       VARCHAR(30) NOT NULL,                    -- 'EP-0001'
    plan_name       VARCHAR(500) NOT NULL,
    plan_type       VARCHAR(50) NOT NULL DEFAULT 'general'
        CHECK (plan_type IN ('general_emergency','fire','spill','collapse','flood','earthquake','medical','terrorism','power_outage','confined_space','height_rescue','other')),
    description     TEXT NOT NULL,
    hazard_type     VARCHAR(200),
    trigger_conditions TEXT,
    response_procedure TEXT NOT NULL,
    evacuation_routes  TEXT,
    assembly_points    TEXT,
    emergency_contacts TEXT,
    resources_required TEXT,
    responsible_person VARCHAR(200),
    deputy_person   VARCHAR(200),
    drill_frequency VARCHAR(100),
    last_reviewed   DATE,
    next_review     DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','approved','active','superseded','archived')),
    version         VARCHAR(10) NOT NULL DEFAULT '1.0',
    approval_date   DATE,
    approved_by     VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, plan_number),
    UNIQUE (project_id, plan_code)
);

CREATE INDEX idx_hsep_project ON hse_emergency_plans(project_id);
CREATE INDEX idx_hsep_type ON hse_emergency_plans(project_id, plan_type);
CREATE INDEX idx_hsep_active ON hse_emergency_plans(project_id, status) WHERE status = 'active';

COMMENT ON TABLE hse_emergency_plans IS 'HSE Emergency Plans — планы действий в чрезвычайных ситуациях';

-- ============================================================================
-- 10. HSE Chemicals (учёт химических веществ)
-- ============================================================================
CREATE TABLE hse_chemicals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    chemical_code   VARCHAR(30) NOT NULL,
    chemical_name   VARCHAR(500) NOT NULL,
    cas_number      VARCHAR(50),                             -- CAS registry number
    hazard_class    VARCHAR(100),
    ghs_symbols     TEXT,                                    -- JSON array
    risk_phrases    TEXT,
    safety_phrases  TEXT,
    manufacturer    VARCHAR(300),
    supplier        VARCHAR(300),
    storage_location VARCHAR(200),
    max_quantity    NUMERIC(10,2),
    unit            VARCHAR(20) DEFAULT 'L',
    is_hazardous    BOOLEAN DEFAULT TRUE,
    is_flammable    BOOLEAN DEFAULT FALSE,
    is_toxic        BOOLEAN DEFAULT FALSE,
    is_corrosive    BOOLEAN DEFAULT FALSE,
    is_environmentally_hazardous BOOLEAN DEFAULT FALSE,
    sds_file        VARCHAR(500),                            -- Safety Data Sheet file path
    sds_revision_date DATE,
    expiry_date     DATE,
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, chemical_code)
);

CREATE INDEX idx_hsch_project ON hse_chemicals(project_id);
CREATE INDEX idx_hsch_hazard ON hse_chemicals(project_id, is_hazardous) WHERE is_hazardous = TRUE;

COMMENT ON TABLE hse_chemicals IS 'HSE Chemicals — учёт опасных химических веществ';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('hse_incident',        'HSE Incident',         'alert-triangle','H-13'),
('hse_permit',          'HSE Permit',           'shield',       'H-13'),
('hse_audit',           'HSE Audit',            'clipboard',    'H-13'),
('hse_inspection',      'HSE Inspection',       'search',       'H-13'),
('hse_training',        'HSE Training',         'book',         'H-13'),
('hse_ppe',             'HSE PPE',              'eye',          'H-13'),
('hse_drill',           'HSE Drill',            'bell',         'H-13'),
('hse_statistics',      'HSE Statistics',       'bar-chart',    'H-13'),
('hse_emergency_plan',  'Emergency Plan',       'map',          'H-13'),
('hse_chemical',        'HSE Chemical',         'flask',        'H-13')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW hse_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM hse_incidents WHERE project_id = p.id AND status = 'open') AS open_incidents,
    (SELECT COUNT(*) FROM hse_incidents WHERE project_id = p.id AND severity IN ('major','critical','catastrophic')) AS high_severity_incidents,
    (SELECT COUNT(*) FROM hse_incidents WHERE project_id = p.id) AS total_incidents,
    (SELECT COUNT(*) FROM hse_incidents WHERE project_id = p.id AND incident_type = 'near_miss') AS near_misses,
    (SELECT COUNT(*) FROM hse_incidents WHERE project_id = p.id AND incident_type = 'lost_time_injury') AS lti,
    (SELECT COUNT(*) FROM hse_incidents WHERE project_id = p.id AND incident_type = 'fatality') AS fatalities,
    (SELECT COUNT(*) FROM hse_permits WHERE project_id = p.id AND status = 'active') AS active_permits,
    (SELECT COUNT(*) FROM hse_permits WHERE project_id = p.id AND status IN ('issued','active')) AS issued_permits,
    (SELECT COUNT(*) FROM hse_audits WHERE project_id = p.id AND status IN ('planned','in_progress')) AS pending_audits,
    (SELECT COUNT(*) FROM hse_audits WHERE project_id = p.id) AS total_audits,
    (SELECT COUNT(*) FROM hse_inspections WHERE project_id = p.id AND inspection_date >= CURRENT_DATE - 7) AS inspections_7d,
    (SELECT COUNT(*) FROM hse_training WHERE project_id = p.id AND training_date >= CURRENT_DATE - 90) AS trainings_90d,
    (SELECT COUNT(*) FROM hse_ppe WHERE project_id = p.id) AS total_ppe_items,
    (SELECT COUNT(*) FROM hse_ppe WHERE project_id = p.id AND quantity_stock <= reorder_level) AS low_stock_ppe,
    (SELECT COUNT(*) FROM hse_drill WHERE project_id = p.id AND status IN ('planned','in_progress')) AS planned_drills,
    (SELECT COUNT(*) FROM hse_emergency_plans WHERE project_id = p.id AND status = 'active') AS active_plans,
    (SELECT COUNT(*) FROM hse_chemicals WHERE project_id = p.id AND is_hazardous = TRUE) AS hazardous_chemicals
FROM projects p;