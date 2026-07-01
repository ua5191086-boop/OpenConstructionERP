-- ============================================================================
-- V020__Change_Management_Module.sql
-- Модуль Change Management (CM) — Change Requests, Impact Analysis, Approval
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Change Requests (запросы на изменения, расширение project_changes)
-- ============================================================================
CREATE TABLE change_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    cr_number       INTEGER NOT NULL,
    cr_code         VARCHAR(30) NOT NULL,                    -- 'CR-0001'
    cr_name         VARCHAR(500) NOT NULL,
    cr_type         VARCHAR(30) NOT NULL DEFAULT 'scope'
        CHECK (cr_type IN ('scope','design','specification','schedule','cost','contract','regulatory','quality','hse','other')),
    source          VARCHAR(30) NOT NULL DEFAULT 'owner'
        CHECK (source IN ('owner','contractor','designer','supplier','regulatory','internal','other')),
    priority        VARCHAR(15) NOT NULL DEFAULT 'medium'
        CHECK (priority IN ('low','medium','high','emergency')),
    description     TEXT NOT NULL,
    reason          TEXT NOT NULL,
    proposed_by     VARCHAR(200),
    proposed_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    required_by_date DATE,
    category        VARCHAR(50)
        CHECK (category IN ('scope_change','design_change','omission','error','value_engineering','regulatory','site_condition','owner_request','other')),
    related_contract_id UUID,
    related_wbs_code VARCHAR(50),
    documents       JSONB DEFAULT '[]'::jsonb,              -- [{name, url, type}]
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','under_review','approved','rejected','deferred','implemented','closed','cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, cr_number),
    UNIQUE (project_id, cr_code)
);

CREATE INDEX idx_cm_cr_project ON change_requests(project_id);
CREATE INDEX idx_cm_cr_type ON change_requests(project_id, cr_type);
CREATE INDEX idx_cm_cr_status ON change_requests(project_id, status);
CREATE INDEX idx_cm_cr_priority ON change_requests(project_id, priority);

COMMENT ON TABLE change_requests IS 'Change Requests — запросы на изменения (расширение project_changes)';

-- ============================================================================
-- 2. Change Orders (приказы об изменениях)
-- ============================================================================
CREATE TABLE change_orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    change_request_id UUID REFERENCES change_requests(id) ON DELETE SET NULL,
    co_number       INTEGER NOT NULL,
    co_code         VARCHAR(30) NOT NULL,                    -- 'CO-0001'
    co_name         VARCHAR(500) NOT NULL,
    co_type         VARCHAR(30) NOT NULL DEFAULT 'variation'
        CHECK (co_type IN ('variation','change_directive','claim_settlement','compensation_event','other')),
    scope_change    TEXT,
    cost_change     NUMERIC(14,2) NOT NULL DEFAULT 0,
    cost_currency   VARCHAR(3) DEFAULT 'USD',
    schedule_change_days INTEGER DEFAULT 0,
    new_end_date    DATE,
    old_end_date    DATE,
    justification   TEXT,
    approved_by     VARCHAR(200),
    approved_at     TIMESTAMPTZ,
    contractor_name VARCHAR(300),
    contract_ref    VARCHAR(100),
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','under_review','approved','rejected','executed','closed','cancelled')),
    executed_at     TIMESTAMPTZ,
    closed_at       TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, co_number),
    UNIQUE (project_id, co_code)
);

CREATE INDEX idx_cm_co_project ON change_orders(project_id);
CREATE INDEX idx_cm_co_cr ON change_orders(change_request_id);
CREATE INDEX idx_cm_co_type ON change_orders(project_id, co_type);
CREATE INDEX idx_cm_co_status ON change_orders(project_id, status);

COMMENT ON TABLE change_orders IS 'Change Orders — приказы и распоряжения об изменениях (Variation Orders)';

-- ============================================================================
-- 3. Change Impact Analysis (анализ воздействия изменений)
-- ============================================================================
CREATE TABLE change_impact_analysis (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    change_request_id UUID NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    impact_type     VARCHAR(30) NOT NULL
        CHECK (impact_type IN ('cost','schedule','scope','quality','safety','environment','stakeholder','resource','contract','risk','other')),
    description     TEXT NOT NULL,
    impact_level    VARCHAR(15) NOT NULL DEFAULT 'medium'
        CHECK (impact_level IN ('very_low','low','medium','high','very_high')),
    likelihood_pct  NUMERIC(5,2),
    cost_impact     NUMERIC(14,2),
    schedule_impact_days INTEGER,
    resource_impact TEXT,
    mitigation      TEXT,
    analyzed_by     VARCHAR(200),
    analysis_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','reviewed','approved','superseded')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (change_request_id, impact_type)
);

CREATE INDEX idx_cm_cia_project ON change_impact_analysis(project_id);
CREATE INDEX idx_cm_cia_cr ON change_impact_analysis(change_request_id);

COMMENT ON TABLE change_impact_analysis IS 'Change Impact Analysis — анализ воздействия изменений на стоимость, сроки, объем';

-- ============================================================================
-- 4. Change Approval Workflow (цепочка утверждения изменений)
-- ============================================================================
CREATE TABLE change_approval_workflow (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    change_request_id UUID NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    step_order      INTEGER NOT NULL,
    step_name       VARCHAR(200) NOT NULL,
    approver_role   VARCHAR(200) NOT NULL,
    approver_name   VARCHAR(200),
    min_approval_level VARCHAR(20) DEFAULT 'approve'
        CHECK (min_approval_level IN ('review','approve','sign_off')),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','approved','rejected','skipped','waived')),
    decision        VARCHAR(30),
    comments        TEXT,
    decided_at      TIMESTAMPTZ,
    decided_by      VARCHAR(200),
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_at TIMESTAMPTZ,
    escalation_days INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (change_request_id, step_order)
);

CREATE INDEX idx_cm_aw_project ON change_approval_workflow(project_id);
CREATE INDEX idx_cm_aw_cr ON change_approval_workflow(change_request_id);
CREATE INDEX idx_cm_aw_status ON change_approval_workflow(project_id, status);

COMMENT ON TABLE change_approval_workflow IS 'Change Approval Workflow — поэтапное утверждение изменений';

-- ============================================================================
-- 5. Change Log (журнал всех изменений)
-- ============================================================================
CREATE TABLE change_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    change_request_id UUID REFERENCES change_requests(id) ON DELETE SET NULL,
    log_type        VARCHAR(30) NOT NULL DEFAULT 'status_change'
        CHECK (log_type IN ('status_change','comment','document_added','approval','rejection','deferral','implementation','close','other')),
    previous_status VARCHAR(30),
    new_status      VARCHAR(30),
    description     TEXT NOT NULL,
    changed_by      VARCHAR(200),
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata        JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_cm_cl_project ON change_log(project_id);
CREATE INDEX idx_cm_cl_cr ON change_log(change_request_id);
CREATE INDEX idx_cm_cl_type ON change_log(project_id, log_type);
CREATE INDEX idx_cm_cl_time ON change_log(project_id, changed_at DESC);

COMMENT ON TABLE change_log IS 'Change Log — журнал всех событий по изменениям';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('change_request',       'Change Request',       'edit',            'CM'),
('change_order',         'Change Order',         'file-text',       'CM'),
('change_impact',        'Impact Analysis',      'activity',        'CM'),
('change_approval',      'Approval Workflow',    'check-square',    'CM'),
('change_log',           'Change Log',           'clipboard',       'CM')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW change_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM change_requests WHERE project_id = p.id AND status NOT IN ('closed','cancelled')) AS open_crs,
    (SELECT COUNT(*) FROM change_requests WHERE project_id = p.id) AS total_crs,
    (SELECT COUNT(*) FROM change_requests WHERE project_id = p.id AND status = 'approved') AS approved_crs,
    (SELECT COUNT(*) FROM change_requests WHERE project_id = p.id AND status = 'implemented') AS implemented_crs,
    (SELECT COUNT(*) FROM change_requests WHERE project_id = p.id AND status = 'rejected') AS rejected_crs,
    (SELECT COUNT(*) FROM change_requests WHERE project_id = p.id AND priority IN ('high','emergency') AND status NOT IN ('closed','cancelled')) AS high_priority_open,
    (SELECT COALESCE(SUM(cost_change),0) FROM change_orders WHERE project_id = p.id AND status NOT IN ('cancelled')) AS total_cost_change,
    (SELECT COALESCE(SUM(schedule_change_days),0) FROM change_orders WHERE project_id = p.id AND status NOT IN ('cancelled')) AS total_schedule_impact,
    (SELECT COUNT(*) FROM change_approval_workflow WHERE project_id = p.id AND status = 'pending') AS pending_approvals,
    (SELECT COUNT(*) FROM change_impact_analysis WHERE project_id = p.id AND impact_level IN ('high','very_high')) AS high_impact_analyses
FROM projects p;