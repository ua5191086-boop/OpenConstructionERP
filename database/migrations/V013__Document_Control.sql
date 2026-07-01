-- ============================================================================
-- V013__Document_Control.sql
-- Модуль Document Control (D-10) — Управление документацией
-- RFI, NCR, Submittals, Method Statements, Shop Drawings,
-- Correspondence, Minutes of Meeting, Daily Reports,
-- Document Transmittals, Document Revisions
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. RFI (Request for Information)
-- ============================================================================
CREATE TABLE rfi_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    rfi_number      INTEGER NOT NULL,
    rfi_code        VARCHAR(30) NOT NULL,                  -- 'RFI-0001'
    subject         VARCHAR(500) NOT NULL,
    question        TEXT NOT NULL,
    answer          TEXT,
    discipline      VARCHAR(50),                           -- civil/structural/MEP/geotech/arch
    priority        VARCHAR(15) NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low','normal','high','urgent')),
    raised_by       VARCHAR(200),
    assigned_to     VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','answered','closed','void','overdue')),
    due_date        DATE,
    raised_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at     TIMESTAMPTZ,
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, rfi_number),
    UNIQUE (project_id, rfi_code)
);

CREATE INDEX idx_rfi_project_status ON rfi_documents(project_id, status);
CREATE INDEX idx_rfi_due ON rfi_documents(project_id, due_date) WHERE status = 'open';
CREATE INDEX idx_rfi_discipline ON rfi_documents(project_id, discipline);

COMMENT ON TABLE rfi_documents IS 'Requests for Information — запросы на разъяснение документации';

-- ============================================================================
-- 2. NCR (Non-Conformance Report)
-- ============================================================================
CREATE TABLE ncr_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ncr_number      INTEGER NOT NULL,
    ncr_code        VARCHAR(30) NOT NULL,                  -- 'NCR-0001'
    title           VARCHAR(500) NOT NULL,
    description     TEXT NOT NULL,
    location        VARCHAR(300),
    ncr_type        VARCHAR(50) NOT NULL DEFAULT 'material'
        CHECK (ncr_type IN ('material','workmanship','design','dimensional','documentation','safety','other')),
    severity        VARCHAR(15) NOT NULL DEFAULT 'minor'
        CHECK (severity IN ('minor','major','critical')),
    source          VARCHAR(100),                          -- inspection, test, audit, complaint
    reported_by     VARCHAR(200),
    assigned_to     VARCHAR(200),
    root_cause      TEXT,
    corrective_action TEXT,
    preventive_action TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','investigating','action_planned','action_taken','verified','closed','void')),
    due_date        DATE,
    reported_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ncr_number),
    UNIQUE (project_id, ncr_code)
);

CREATE INDEX idx_ncr_project_status ON ncr_documents(project_id, status);
CREATE INDEX idx_ncr_severity ON ncr_documents(project_id, severity);
CREATE INDEX idx_ncr_due ON ncr_documents(project_id, due_date) WHERE status IN ('open','investigating','action_planned');

COMMENT ON TABLE ncr_documents IS 'Non-Conformance Reports — отчёты о несоответствиях';

-- ============================================================================
-- 3. Submittals
-- ============================================================================
CREATE TABLE submittals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    submittal_number INTEGER NOT NULL,
    submittal_code  VARCHAR(30) NOT NULL,                  -- 'SUB-0001'
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    submittal_type  VARCHAR(50) NOT NULL DEFAULT 'material'
        CHECK (submittal_type IN ('material','equipment','drawing','sample','document','method','product_data','other')),
    specification_ref VARCHAR(200),
    submitted_by    VARCHAR(200),
    submitted_to    VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','under_review','reviewed','approved','approved_with_comments','rejected','resubmit','closed')),
    review_notes    TEXT,
    resubmit_count  INTEGER DEFAULT 0,
    submitted_at    TIMESTAMPTZ,
    reviewed_at     TIMESTAMPTZ,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, submittal_number),
    UNIQUE (project_id, submittal_code)
);

CREATE INDEX idx_submittals_project_status ON submittals(project_id, status);
CREATE INDEX idx_submittals_type ON submittals(project_id, submittal_type);

COMMENT ON TABLE submittals IS 'Submittals — подача материалов на утверждение';

-- ============================================================================
-- 4. Method Statements
-- ============================================================================
CREATE TABLE method_statements (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ms_number       INTEGER NOT NULL,
    ms_code         VARCHAR(30) NOT NULL,                  -- 'MS-0001'
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    work_area       VARCHAR(300),
    activity        VARCHAR(200),
    method          TEXT,                                  -- описание метода работ
    resources       TEXT,                                  -- оборудование, материалы
    hse_aspects     TEXT,                                  -- HSE аспекты
    quality_checks  TEXT,                                  -- контроль качества
    submitted_by    VARCHAR(200),
    reviewed_by     VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','under_review','approved','rejected','revised','closed')),
    submitted_at    TIMESTAMPTZ,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ms_number),
    UNIQUE (project_id, ms_code)
);

CREATE INDEX idx_ms_project_status ON method_statements(project_id, status);

COMMENT ON TABLE method_statements IS 'Method Statements — описания методов производства работ';

-- ============================================================================
-- 5. Shop Drawings
-- ============================================================================
CREATE TABLE shop_drawings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    drawing_number  INTEGER NOT NULL,
    drawing_code    VARCHAR(30) NOT NULL,                  -- 'SD-0001'
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    discipline      VARCHAR(50),                           -- civil/structural/MEP/arch/landscape
    drawing_format  VARCHAR(20) DEFAULT 'pdf',             -- pdf, dwg, ifc, rvt
    revision        VARCHAR(10) NOT NULL DEFAULT 'A',
    file_path       VARCHAR(500),
    submitted_by    VARCHAR(200),
    checked_by      VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','under_review','approved','approved_with_comments','rejected','resubmit','closed')),
    review_notes    TEXT,
    resubmit_count  INTEGER DEFAULT 0,
    submitted_at    TIMESTAMPTZ,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, drawing_number),
    UNIQUE (project_id, drawing_code)
);

CREATE INDEX idx_sd_project_status ON shop_drawings(project_id, status);
CREATE INDEX idx_sd_discipline ON shop_drawings(project_id, discipline);

COMMENT ON TABLE shop_drawings IS 'Shop Drawings — деталировочные чертежи';

-- ============================================================================
-- 6. Correspondence
-- ============================================================================
CREATE TABLE correspondence (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    corr_number     INTEGER NOT NULL,
    corr_code       VARCHAR(30) NOT NULL,                  -- 'LTR-0001'
    subject         VARCHAR(500) NOT NULL,
    body            TEXT NOT NULL,
    corr_type       VARCHAR(30) NOT NULL DEFAULT 'letter'
        CHECK (corr_type IN ('letter','email','memo','fax','transmittal','notice','instruction')),
    direction       VARCHAR(10) NOT NULL DEFAULT 'outgoing'
        CHECK (direction IN ('incoming','outgoing','internal')),
    from_entity     VARCHAR(200) NOT NULL,
    to_entity       VARCHAR(200) NOT NULL,
    cc_entity       VARCHAR(500),
    priority        VARCHAR(15) DEFAULT 'normal'
        CHECK (priority IN ('low','normal','high','urgent')),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','sent','received','acknowledged','replied','archived')),
    replied_to_id   UUID REFERENCES correspondence(id),
    sent_at         TIMESTAMPTZ,
    received_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, corr_number),
    UNIQUE (project_id, corr_code)
);

CREATE INDEX idx_corr_project_type ON correspondence(project_id, corr_type);
CREATE INDEX idx_corr_direction ON correspondence(project_id, direction);
CREATE INDEX idx_corr_status ON correspondence(project_id, status);

COMMENT ON TABLE correspondence IS 'Correspondence — переписка по проекту';

-- ============================================================================
-- 7. Minutes of Meeting
-- ============================================================================
CREATE TABLE minutes_of_meeting (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    mom_number      INTEGER NOT NULL,
    mom_code        VARCHAR(30) NOT NULL,                  -- 'MOM-0001'
    meeting_title   VARCHAR(500) NOT NULL,
    meeting_type    VARCHAR(50) NOT NULL DEFAULT 'progress'
        CHECK (meeting_type IN ('progress','technical','coordination','hse','design_review','kickoff','closeout','other')),
    meeting_date    DATE NOT NULL,
    location        VARCHAR(300),
    chairperson     VARCHAR(200),
    attendees       TEXT,                                  -- список участников
    minutes         TEXT NOT NULL,                         -- протокол
    action_items    TEXT,                                  -- пункты с ответственным и сроком
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','distributed','approved','closed')),
    distributed_at  TIMESTAMPTZ,
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, mom_number),
    UNIQUE (project_id, mom_code)
);

CREATE INDEX idx_mom_project_date ON minutes_of_meeting(project_id, meeting_date DESC);
CREATE INDEX idx_mom_type ON minutes_of_meeting(project_id, meeting_type);

COMMENT ON TABLE minutes_of_meeting IS 'Minutes of Meeting — протоколы совещаний';

-- ============================================================================
-- 8. Daily Reports (расширенная версия, замещает V005)
-- ============================================================================
CREATE TABLE doc_daily_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    report_date     DATE NOT NULL,
    shift           VARCHAR(10) NOT NULL DEFAULT 'day'
        CHECK (shift IN ('day','night','A','B','C')),
    weather         VARCHAR(100),
    temp_c          NUMERIC(4,1),
    manpower_total  INTEGER,
    equipment_total INTEGER,
    narrative       TEXT,
    hse_notes       TEXT,
    delays          TEXT,
    work_completed  TEXT,                                  -- выполненные работы
    planned_tomorrow TEXT,                                 -- планы на завтра
    author          VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','approved','rejected')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, report_date, shift)
);

CREATE INDEX idx_ddr_project_date ON doc_daily_reports(project_id, report_date DESC);
CREATE INDEX idx_ddr_status ON doc_daily_reports(project_id, status);

COMMENT ON TABLE doc_daily_reports IS 'Daily Reports — ежедневные отчёты о производстве работ';

-- ============================================================================
-- 9. Document Transmittals
-- ============================================================================
CREATE TABLE document_transmittals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    transmittal_number INTEGER NOT NULL,
    transmittal_code VARCHAR(30) NOT NULL,                 -- 'DT-0001'
    title           VARCHAR(500) NOT NULL,
    purpose         VARCHAR(100) NOT NULL DEFAULT 'for_review'
        CHECK (purpose IN ('for_review','for_approval','for_construction','for_record','for_comment','for_information')),
    from_entity     VARCHAR(200) NOT NULL,
    to_entity       VARCHAR(200) NOT NULL,
    document_list   TEXT NOT NULL,                         -- список передаваемых документов
    notes           TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'prepared'
        CHECK (status IN ('prepared','sent','received','acknowledged','rejected','closed')),
    sent_at         TIMESTAMPTZ,
    received_at     TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, transmittal_number),
    UNIQUE (project_id, transmittal_code)
);

CREATE INDEX idx_dt_project_status ON document_transmittals(project_id, status);

COMMENT ON TABLE document_transmittals IS 'Document Transmittals — сопроводительные письма передачи документов';

-- ============================================================================
-- 10. Document Revisions (история версий документов)
-- ============================================================================
CREATE TABLE document_revisions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    document_type   VARCHAR(50) NOT NULL,                  -- rfi, ncr, submittal, ms, sd, correspondence, mom, transmittal
    document_id     UUID NOT NULL,
    revision        VARCHAR(10) NOT NULL,
    change_summary  TEXT,
    file_path       VARCHAR(500),
    file_size       BIGINT,
    file_hash       VARCHAR(64),                           -- SHA-256
    created_by      VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'current'
        CHECK (status IN ('current','superseded','archived')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_drev_project ON document_revisions(project_id);
CREATE INDEX idx_drev_document ON document_revisions(document_type, document_id);
CREATE INDEX idx_drev_status ON document_revisions(project_id, status);

COMMENT ON TABLE document_revisions IS 'Document Revisions — история версий документов';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('rfi',              'RFI',                 'question',     'D-10'),
('ncr',              'NCR',                 'exclamation',  'D-10'),
('submittal',        'Submittal',           'upload',       'D-10'),
('method_statement', 'Method Statement',    'clipboard',    'D-10'),
('shop_drawing',     'Shop Drawing',        'pencil',       'D-10'),
('correspondence',   'Correspondence',      'envelope',     'D-10'),
('minutes_of_meeting','Minutes of Meeting', 'calendar',     'D-10'),
('daily_report',     'Daily Report',        'chart',        'D-10'),
('transmittal',      'Document Transmittal','send',         'D-10'),
('revision',         'Document Revision',   'history',      'D-10')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW doc_control_summary AS
SELECT
    project_id,
    (SELECT COUNT(*) FROM rfi_documents WHERE project_id = p.id) AS total_rfi,
    (SELECT COUNT(*) FROM rfi_documents WHERE project_id = p.id AND status = 'open') AS open_rfi,
    (SELECT COUNT(*) FROM ncr_documents WHERE project_id = p.id) AS total_ncr,
    (SELECT COUNT(*) FROM ncr_documents WHERE project_id = p.id AND status IN ('open','investigating','action_planned')) AS open_ncr,
    (SELECT COUNT(*) FROM ncr_documents WHERE project_id = p.id AND severity = 'critical') AS critical_ncr,
    (SELECT COUNT(*) FROM submittals WHERE project_id = p.id) AS total_submittals,
    (SELECT COUNT(*) FROM submittals WHERE project_id = p.id AND status IN ('submitted','under_review')) AS pending_submittals,
    (SELECT COUNT(*) FROM submittals WHERE project_id = p.id AND status = 'rejected') AS rejected_submittals,
    (SELECT COUNT(*) FROM method_statements WHERE project_id = p.id) AS total_ms,
    (SELECT COUNT(*) FROM method_statements WHERE project_id = p.id AND status IN ('draft','submitted','under_review')) AS pending_ms,
    (SELECT COUNT(*) FROM shop_drawings WHERE project_id = p.id) AS total_sd,
    (SELECT COUNT(*) FROM shop_drawings WHERE project_id = p.id AND status IN ('submitted','under_review')) AS pending_sd,
    (SELECT COUNT(*) FROM correspondence WHERE project_id = p.id) AS total_corr,
    (SELECT COUNT(*) FROM correspondence WHERE project_id = p.id AND status NOT IN ('archived')) AS active_corr,
    (SELECT COUNT(*) FROM minutes_of_meeting WHERE project_id = p.id) AS total_mom,
    (SELECT COUNT(*) FROM doc_daily_reports WHERE project_id = p.id AND report_date >= CURRENT_DATE - 7) AS recent_reports,
    (SELECT COUNT(*) FROM document_transmittals WHERE project_id = p.id) AS total_transmittals,
    (SELECT COUNT(*) FROM document_transmittals WHERE project_id = p.id AND status = 'prepared') AS pending_transmittals
FROM projects p;