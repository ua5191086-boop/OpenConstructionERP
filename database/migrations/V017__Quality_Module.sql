-- ============================================================================
-- V017__Quality_Module.sql
-- Модуль Quality Management (QM) — Inspection & Test Plans, NCR, Calibration
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. QM Inspection & Test Plans (ITP)
-- ============================================================================
CREATE TABLE qm_itp (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    itp_number      INTEGER NOT NULL,
    itp_code        VARCHAR(30) NOT NULL,                    -- 'ITP-0001'
    itp_name        VARCHAR(500) NOT NULL,
    itp_type        VARCHAR(50) NOT NULL DEFAULT 'inspection'
        CHECK (itp_type IN ('inspection','test','combined','witness','hold_point','review','surveillance')),
    wbs_code        VARCHAR(50),
    boq_item_id     UUID,
    description     TEXT NOT NULL,
    scope           TEXT,
    applicable_standards TEXT,
    acceptance_criteria TEXT,
    responsible_party VARCHAR(200),
    inspection_frequency VARCHAR(100),
    revision        VARCHAR(10) NOT NULL DEFAULT '1.0',
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','active','completed','superseded','cancelled')),
    approved_by     VARCHAR(200),
    approved_at     TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, itp_number),
    UNIQUE (project_id, itp_code)
);

CREATE INDEX idx_qm_itp_project ON qm_itp(project_id);
CREATE INDEX idx_qm_itp_status ON qm_itp(project_id, status);
CREATE INDEX idx_qm_itp_type ON qm_itp(project_id, itp_type);

COMMENT ON TABLE qm_itp IS 'QM ITP — Inspection & Test Plans';

-- ============================================================================
-- 2. QM Inspection Records (чек-листы инспекций)
-- ============================================================================
CREATE TABLE qm_inspection_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    itp_id          UUID REFERENCES qm_itp(id) ON DELETE SET NULL,
    record_number   INTEGER NOT NULL,
    record_code     VARCHAR(30) NOT NULL,                    -- 'INS-0001'
    inspection_type VARCHAR(50) NOT NULL DEFAULT 'visual'
        CHECK (inspection_type IN ('visual','dimensional','weld','coating','concrete','soil','electrical','mechanical','piping','civil','structural','nondestructive','pressure_test','functional','other')),
    title           VARCHAR(500) NOT NULL,
    location        VARCHAR(300),
    area            VARCHAR(200),
    inspector       VARCHAR(200),
    inspection_date DATE NOT NULL,
    check_items     JSONB DEFAULT '[]'::jsonb,               -- checklist items with pass/fail
    result          VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (result IN ('pass','fail','conditional_pass','pending','not_applicable')),
    defects_found   INTEGER DEFAULT 0,
    remarks         TEXT,
    attachment_url  VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','closed','void')),
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, record_number),
    UNIQUE (project_id, record_code)
);

CREATE INDEX idx_qm_ir_project ON qm_inspection_records(project_id);
CREATE INDEX idx_qm_ir_itp ON qm_inspection_records(itp_id);
CREATE INDEX idx_qm_ir_type ON qm_inspection_records(project_id, inspection_type);
CREATE INDEX idx_qm_ir_result ON qm_inspection_records(project_id, result);
CREATE INDEX idx_qm_ir_date ON qm_inspection_records(project_id, inspection_date DESC);

COMMENT ON TABLE qm_inspection_records IS 'QM Inspection Records — чек-листы и результаты инспекций';

-- ============================================================================
-- 3. QM Test Results (результаты испытаний)
-- ============================================================================
CREATE TABLE qm_test_results (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    itp_id          UUID REFERENCES qm_itp(id) ON DELETE SET NULL,
    test_number     INTEGER NOT NULL,
    test_code       VARCHAR(30) NOT NULL,                    -- 'TST-0001'
    test_name       VARCHAR(500) NOT NULL,
    test_type       VARCHAR(50) NOT NULL DEFAULT 'material'
        CHECK (test_type IN ('material','concrete','steel','soil','weld','coating','electrical','mechanical','pressure','leak','functional','environmental','other')),
    sample_id       VARCHAR(100),
    lab_name        VARCHAR(300),
    technician      VARCHAR(200),
    test_date       DATE NOT NULL,
    test_method     VARCHAR(200),
    specification   TEXT,
    measured_value  NUMERIC(14,4),
    unit            VARCHAR(30),
    min_acceptable  NUMERIC(14,4),
    max_acceptable  NUMERIC(14,4),
    result          VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (result IN ('pass','fail','conditional','pending','invalid','retest')),
    deviation       NUMERIC(10,4),
    remarks         TEXT,
    report_url      VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','closed','void')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, test_number),
    UNIQUE (project_id, test_code)
);

CREATE INDEX idx_qm_tr_project ON qm_test_results(project_id);
CREATE INDEX idx_qm_tr_itp ON qm_test_results(itp_id);
CREATE INDEX idx_qm_tr_type ON qm_test_results(project_id, test_type);
CREATE INDEX idx_qm_tr_result ON qm_test_results(project_id, result);

COMMENT ON TABLE qm_test_results IS 'QM Test Results — результаты лабораторных и полевых испытаний';

-- ============================================================================
-- 4. QM Non-Conformance Reports (NCR)
-- ============================================================================
CREATE TABLE qm_ncr (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ncr_number      INTEGER NOT NULL,
    ncr_code        VARCHAR(30) NOT NULL,                    -- 'NCR-0001'
    title           VARCHAR(500) NOT NULL,
    ncr_category    VARCHAR(50) NOT NULL DEFAULT 'material'
        CHECK (ncr_category IN ('material','workmanship','design','dimensional','documentation','safety','environmental','other')),
    severity        VARCHAR(15) NOT NULL DEFAULT 'minor'
        CHECK (severity IN ('minor','major','critical')),
    source          VARCHAR(100) NOT NULL DEFAULT 'inspection'
        CHECK (source IN ('inspection','test','audit','customer','supplier','design_review','other')),
    description     TEXT NOT NULL,
    location        VARCHAR(300),
    discovered_date DATE NOT NULL,
    discovered_by   VARCHAR(200),
    related_item_type VARCHAR(50),
    related_item_id UUID,
    root_cause      TEXT,
    root_cause_category VARCHAR(50)
        CHECK (root_cause_category IN ('human_error','procedure','material','equipment','design','training','external','other')),
    proposed_disposition TEXT,
    approved_disposition TEXT,
    disposition_type VARCHAR(30)
        CHECK (disposition_type IN ('rework','repair','replace','use_as_is','scrap','concession','other')),
    rework_cost     NUMERIC(12,2),
    schedule_impact INTEGER,                                -- days
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','investigating','disposition','implementing','verification','closed','void')),
    closed_at       TIMESTAMPTZ,
    closed_by       VARCHAR(200),
    verification_method TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ncr_number),
    UNIQUE (project_id, ncr_code)
);

CREATE INDEX idx_qm_ncr_project ON qm_ncr(project_id);
CREATE INDEX idx_qm_ncr_category ON qm_ncr(project_id, ncr_category);
CREATE INDEX idx_qm_ncr_severity ON qm_ncr(project_id, severity);
CREATE INDEX idx_qm_ncr_status ON qm_ncr(project_id, status);

COMMENT ON TABLE qm_ncr IS 'QM NCR — Non-Conformance Reports (отчёты о несоответствиях)';

-- ============================================================================
-- 5. QM Corrective Actions (корректирующие действия)
-- ============================================================================
CREATE TABLE qm_corrective_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ncr_id          UUID REFERENCES qm_ncr(id) ON DELETE CASCADE,
    ca_number       INTEGER NOT NULL,
    ca_code         VARCHAR(30) NOT NULL,                    -- 'CA-0001'
    title           VARCHAR(500) NOT NULL,
    action_type     VARCHAR(30) NOT NULL DEFAULT 'corrective'
        CHECK (action_type IN ('corrective','preventive','improvement')),
    description     TEXT NOT NULL,
    assigned_to     VARCHAR(200),
    priority        VARCHAR(15) NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('low','medium','high','critical')),
    due_date        DATE,
    completed_at    TIMESTAMPTZ,
    verification_method TEXT,
    verified_by     VARCHAR(200),
    verified_at     TIMESTAMPTZ,
    effectiveness   VARCHAR(20)
        CHECK (effectiveness IN ('effective','partially_effective','not_effective','not_verified')),
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','in_progress','implemented','verified','closed','cancelled')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ca_number),
    UNIQUE (project_id, ca_code)
);

CREATE INDEX idx_qm_ca_project ON qm_corrective_actions(project_id);
CREATE INDEX idx_qm_ca_ncr ON qm_corrective_actions(ncr_id);
CREATE INDEX idx_qm_ca_status ON qm_corrective_actions(project_id, status);

COMMENT ON TABLE qm_corrective_actions IS 'QM Corrective Actions — корректирующие и предупреждающие действия';

-- ============================================================================
-- 6. QM Calibration (калибровка оборудования)
-- ============================================================================
CREATE TABLE qm_calibration (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_name  VARCHAR(300) NOT NULL,
    equipment_model VARCHAR(200),
    serial_number   VARCHAR(100) NOT NULL,
    calibration_type VARCHAR(30) NOT NULL DEFAULT 'internal'
        CHECK (calibration_type IN ('internal','external','factory')),
    calibration_frequency_days INTEGER DEFAULT 365,
    last_calibration_date DATE,
    next_calibration_date DATE NOT NULL,
    calibration_standard VARCHAR(300),
    calibration_result VARCHAR(20) NOT NULL DEFAULT 'pass'
        CHECK (calibration_result IN ('pass','fail','conditional','expired','not_calibrated')),
    certificate_number VARCHAR(100),
    certificate_file VARCHAR(500),
    calibrated_by   VARCHAR(200),
    calibration_lab VARCHAR(300),
    remarks         TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','expired','retired','lost')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, serial_number)
);

CREATE INDEX idx_qm_cal_project ON qm_calibration(project_id);
CREATE INDEX idx_qm_cal_status ON qm_calibration(project_id, status);
CREATE INDEX idx_qm_cal_next_date ON qm_calibration(project_id, next_calibration_date) WHERE status = 'active';

COMMENT ON TABLE qm_calibration IS 'QM Calibration — калибровка измерительного оборудования';

-- ============================================================================
-- 7. QM Quality Metrics (метрики качества)
-- ============================================================================
CREATE TABLE qm_quality_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    report_month    DATE NOT NULL,                           -- first day of month
    total_inspections INTEGER DEFAULT 0,
    inspections_passed INTEGER DEFAULT 0,
    inspections_failed INTEGER DEFAULT 0,
    total_tests     INTEGER DEFAULT 0,
    tests_passed    INTEGER DEFAULT 0,
    tests_failed    INTEGER DEFAULT 0,
    ncr_opened      INTEGER DEFAULT 0,
    ncr_closed      INTEGER DEFAULT 0,
    ncr_critical    INTEGER DEFAULT 0,
    ncr_major       INTEGER DEFAULT 0,
    ncr_minor       INTEGER DEFAULT 0,
    ca_opened       INTEGER DEFAULT 0,
    ca_closed       INTEGER DEFAULT 0,
    rework_cost     NUMERIC(14,2) DEFAULT 0,
    first_pass_yield NUMERIC(6,2),                          -- FPY percentage
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, report_month)
);

CREATE INDEX idx_qm_qm_project ON qm_quality_metrics(project_id);
CREATE INDEX idx_qm_qm_month ON qm_quality_metrics(project_id, report_month DESC);

COMMENT ON TABLE qm_quality_metrics IS 'QM Quality Metrics — ежемесячные KPI качества';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('qm_itp',                 'QM ITP',              'clipboard-check','QM'),
('qm_inspection_record',   'QM Inspection',       'search',        'QM'),
('qm_test_result',         'QM Test Result',      'flask',         'QM'),
('qm_ncr',                 'QM NCR',              'alert-triangle','QM'),
('qm_corrective_action',   'QM Corrective Action','check-circle',  'QM'),
('qm_calibration',         'QM Calibration',      'sliders',       'QM'),
('qm_quality_metric',      'QM Quality Metric',   'bar-chart',     'QM')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW qm_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM qm_itp WHERE project_id = p.id AND status = 'active') AS active_itps,
    (SELECT COUNT(*) FROM qm_itp WHERE project_id = p.id) AS total_itps,
    (SELECT COUNT(*) FROM qm_inspection_records WHERE project_id = p.id AND result = 'fail') AS failed_inspections,
    (SELECT COUNT(*) FROM qm_inspection_records WHERE project_id = p.id AND inspection_date >= CURRENT_DATE - 30) AS inspections_30d,
    (SELECT COUNT(*) FROM qm_inspection_records WHERE project_id = p.id) AS total_inspections,
    (SELECT COUNT(*) FROM qm_test_results WHERE project_id = p.id AND result = 'fail') AS failed_tests,
    (SELECT COUNT(*) FROM qm_ncr WHERE project_id = p.id AND status NOT IN ('closed','void')) AS open_ncrs,
    (SELECT COUNT(*) FROM qm_ncr WHERE project_id = p.id AND severity = 'critical' AND status NOT IN ('closed','void')) AS critical_ncrs,
    (SELECT COUNT(*) FROM qm_ncr WHERE project_id = p.id) AS total_ncrs,
    (SELECT COUNT(*) FROM qm_corrective_actions WHERE project_id = p.id AND status NOT IN ('closed','cancelled')) AS open_cas,
    (SELECT COUNT(*) FROM qm_calibration WHERE project_id = p.id AND status = 'active' AND next_calibration_date < CURRENT_DATE) AS overdue_calibrations,
    (SELECT COALESCE(SUM(rework_cost),0) FROM qm_ncr WHERE project_id = p.id) AS total_rework_cost
FROM projects p;