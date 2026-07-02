-- ============================================================================
-- V049__Audit_Compliance.sql
-- Внутренний аудит и соответствие требованиям
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Внутренние аудиты
-- ============================================================================
CREATE TABLE internal_audits (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    audit_type          VARCHAR(100) NOT NULL,                  -- financial, quality, safety, environmental, compliance, operational
    audit_scope         TEXT NOT NULL,
    audit_period_from   DATE,
    audit_period_to     DATE,
    lead_auditor        VARCHAR(300) NOT NULL,
    audit_team          JSONB DEFAULT '[]'::JSONB,
    standards           JSONB DEFAULT '[]'::JSONB,              -- ISO 9001, ISO 14001, ISO 45001, etc.
    scheduled_date      DATE,
    actual_date         DATE,
    duration_days       INTEGER,
    status              VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, in_progress, completed, cancelled
    overall_rating      VARCHAR(50),                            -- excellent, good, satisfactory, needs_improvement, non_compliant
    findings_count      INTEGER DEFAULT 0,
    critical_findings   INTEGER DEFAULT 0,
    major_findings      INTEGER DEFAULT 0,
    minor_findings      INTEGER DEFAULT 0,
    opportunities       INTEGER DEFAULT 0,
    summary             TEXT,
    conclusion          TEXT,
    report_document     VARCHAR(500),
    is_confidential     BOOLEAN DEFAULT FALSE,
    created_by          VARCHAR(200),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_internal_audits_project ON internal_audits(project_id);
CREATE INDEX idx_internal_audits_type ON internal_audits(audit_type);
CREATE INDEX idx_internal_audits_status ON internal_audits(status);

COMMENT ON TABLE internal_audits IS 'Внутренние аудиты проектов';

-- ============================================================================
-- 2. Находки аудита
-- ============================================================================
CREATE TABLE audit_findings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_id            UUID NOT NULL REFERENCES internal_audits(id) ON DELETE CASCADE,
    finding_type        VARCHAR(50) NOT NULL,                   -- critical, major, minor, opportunity, observation
    finding_code        VARCHAR(100) NOT NULL,
    title               VARCHAR(500) NOT NULL,
    description         TEXT NOT NULL,
    reference_standard  VARCHAR(200),
    reference_clause    VARCHAR(100),
    root_cause          TEXT,
    impact              TEXT,
    corrective_action   TEXT,
    preventive_action   TEXT,
    deadline_date       DATE,
    assigned_to         VARCHAR(300),
    status              VARCHAR(50) NOT NULL DEFAULT 'open',    -- open, in_progress, resolved, verified, closed
    resolution_notes    TEXT,
    resolved_at         TIMESTAMPTZ,
    verified_by         VARCHAR(200),
    verified_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_findings_audit ON audit_findings(audit_id);
CREATE INDEX idx_audit_findings_type ON audit_findings(finding_type);
CREATE INDEX idx_audit_findings_status ON audit_findings(status);
CREATE INDEX idx_audit_findings_deadline ON audit_findings(deadline_date) WHERE status != 'closed';

COMMENT ON TABLE audit_findings IS 'Находки внутренних аудитов';

-- ============================================================================
-- 3. Реестр нормативных требований
-- ============================================================================
CREATE TABLE compliance_requirements (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    requirement_code    VARCHAR(100) NOT NULL,
    requirement_type    VARCHAR(100) NOT NULL,                  -- legal, regulatory, contractual, internal, industry_standard
    title               VARCHAR(500) NOT NULL,
    description         TEXT,
    authority           VARCHAR(300),                           -- issuing body
    effective_date      DATE,
    expiry_date         DATE,
    applies_to          VARCHAR(200),                           -- all_projects, tunnel_only, building, specific_client
    risk_if_noncompliant TEXT,
    control_measure     TEXT,
    review_frequency    VARCHAR(50),                            -- monthly, quarterly, annually, event_driven
    last_review_date    DATE,
    next_review_date    DATE,
    status              VARCHAR(50) DEFAULT 'active',           -- active, under_review, superseded, expired
    responsible_party   VARCHAR(300),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_compliance_project ON compliance_requirements(project_id);
CREATE INDEX idx_compliance_type ON compliance_requirements(requirement_type);
CREATE INDEX idx_compliance_status ON compliance_requirements(status);

COMMENT ON TABLE compliance_requirements IS 'Реестр нормативных требований и обязательств';

-- ============================================================================
-- 4. Проверки соответствия
-- ============================================================================
CREATE TABLE compliance_checks (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requirement_id      UUID NOT NULL REFERENCES compliance_requirements(id) ON DELETE CASCADE,
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    check_type          VARCHAR(100) NOT NULL,                  -- self_assessment, audit, inspection, automated
    check_date          DATE NOT NULL,
    checked_by          VARCHAR(300),
    result              VARCHAR(50) NOT NULL,                   -- compliant, non_compliant, partially_compliant, not_applicable
    evidence            TEXT,
    non_compliance_detail TEXT,
    corrective_action   TEXT,
    deadline_date       DATE,
    closure_date        DATE,
    status              VARCHAR(50) DEFAULT 'completed',        -- completed, overdue, waived
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_compliance_checks_req ON compliance_checks(requirement_id);
CREATE INDEX idx_compliance_checks_result ON compliance_checks(result);
CREATE INDEX idx_compliance_checks_deadline ON compliance_checks(deadline_date) WHERE status='overdue';

COMMENT ON TABLE compliance_checks IS 'Результаты проверок соответствия требованиям';