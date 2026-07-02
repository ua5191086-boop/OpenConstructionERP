-- ============================================================================
-- V030__Permits_Module.sql
-- Permits: Applications, Inspections, Regulatory Compliance
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Regulatory Bodies
-- ============================================================================
CREATE TABLE regulatory_bodies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    body_name       VARCHAR(500) NOT NULL,
    body_code       VARCHAR(100),
    jurisdiction    VARCHAR(300),
    contact_info    JSONB,
    website         VARCHAR(500),
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. Permit Applications
-- ============================================================================
CREATE TABLE permit_applications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    regulatory_body_id UUID REFERENCES regulatory_bodies(id),
    permit_number   VARCHAR(200),
    permit_type     VARCHAR(200) NOT NULL,                    -- construction, environmental, occupancy, zoning
    description     TEXT,
    application_date DATE,
    decision_date   DATE,
    status          VARCHAR(50) DEFAULT 'draft',              -- draft, submitted, under_review, approved, rejected, expired
    approved_by     VARCHAR(300),
    expiry_date     DATE,
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_permit_applications_project ON permit_applications(project_id);
CREATE INDEX idx_permit_applications_status ON permit_applications(status);

-- ============================================================================
-- 3. Permit Documents
-- ============================================================================
CREATE TABLE permit_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permit_application_id UUID NOT NULL REFERENCES permit_applications(id),
    document_type   VARCHAR(200) NOT NULL,
    document_name   VARCHAR(500),
    document_url    TEXT,
    version         VARCHAR(50),
    submitted_date  DATE,
    status          VARCHAR(50) DEFAULT 'pending',            -- pending, submitted, accepted, rejected
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_permit_documents_app ON permit_documents(permit_application_id);

-- ============================================================================
-- 4. Permit Inspections
-- ============================================================================
CREATE TABLE permit_inspections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permit_application_id UUID NOT NULL REFERENCES permit_applications(id),
    inspection_type VARCHAR(200) NOT NULL,
    inspection_date DATE,
    inspector_name  VARCHAR(300),
    inspector_agency VARCHAR(300),
    result          VARCHAR(50),                              -- pass, fail, conditional_pass
    findings        TEXT,
    corrective_actions TEXT,
    scheduled_date  DATE,
    completed_date  DATE,
    status          VARCHAR(50) DEFAULT 'scheduled',          -- scheduled, in_progress, completed, cancelled
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_permit_inspections_app ON permit_inspections(permit_application_id);

-- ============================================================================
-- 5. Permit Renewals
-- ============================================================================
CREATE TABLE permit_renewals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permit_application_id UUID NOT NULL REFERENCES permit_applications(id),
    renewal_number  VARCHAR(100),
    renewal_date    DATE,
    expiry_date     DATE,
    fee_amount      NUMERIC(14,2),
    fee_currency    VARCHAR(3) DEFAULT 'USD',
    status          VARCHAR(50) DEFAULT 'pending',            -- pending, approved, rejected
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_permit_renewals_app ON permit_renewals(permit_application_id);

-- ============================================================================
-- 6. Permit Conditions
-- ============================================================================
CREATE TABLE permit_conditions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permit_application_id UUID NOT NULL REFERENCES permit_applications(id),
    condition_number VARCHAR(100),
    description     TEXT NOT NULL,
    condition_type  VARCHAR(100),                             -- prerequisite, ongoing, reporting
    due_date        DATE,
    status          VARCHAR(50) DEFAULT 'pending',            -- pending, satisfied, waived, breached
    satisfied_date  DATE,
    verified_by     VARCHAR(300),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_permit_conditions_app ON permit_conditions(permit_application_id);