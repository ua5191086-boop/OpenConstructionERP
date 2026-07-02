-- ============================================================================
-- V031__Insurance_Module.sql
-- Insurance: Policies, Claims, Coverage, Brokers
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Insurance Brokers
-- ============================================================================
CREATE TABLE insurance_brokers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    broker_name     VARCHAR(500) NOT NULL,
    contact_person  VARCHAR(300),
    email           VARCHAR(300),
    phone           VARCHAR(100),
    address         TEXT,
    license_number  VARCHAR(200),
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. Insurance Policies
-- ============================================================================
CREATE TABLE insurance_policies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    policy_number   VARCHAR(200) NOT NULL,
    policy_type     VARCHAR(200) NOT NULL,                    -- constructor_all_risk, third_party_liability, professional_indemnity, workers_comp, marine_cargo, motor, property, environmental
    insurer         VARCHAR(500) NOT NULL,
    broker_id       UUID REFERENCES insurance_brokers(id),
    insured_party   VARCHAR(500),
    sum_insured     NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    premium_amount  NUMERIC(18,2),
    deductible      NUMERIC(18,2),
    excess          NUMERIC(18,2),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    renewal_date    DATE,
    territory       VARCHAR(500),
    status          VARCHAR(50) DEFAULT 'active',             -- active, expired, cancelled, renewed
    description     TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ins_policies_project ON insurance_policies(project_id);
CREATE INDEX idx_ins_policies_type ON insurance_policies(policy_type);
CREATE INDEX idx_ins_policies_status ON insurance_policies(status);

-- ============================================================================
-- 3. Insurance Coverage
-- ============================================================================
CREATE TABLE insurance_coverage (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES insurance_policies(id),
    coverage_type   VARCHAR(300) NOT NULL,
    coverage_limit  NUMERIC(18,2),
    currency        VARCHAR(3) DEFAULT 'USD',
    deductible      NUMERIC(18,2),
    sublimit        NUMERIC(18,2),
    description     TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ins_coverage_policy ON insurance_coverage(policy_id);

-- ============================================================================
-- 4. Insurance Premiums
-- ============================================================================
CREATE TABLE insurance_premiums (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES insurance_policies(id),
    premium_number  VARCHAR(200),
    amount          NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    due_date        DATE,
    paid_date       DATE,
    payment_method  VARCHAR(100),
    status          VARCHAR(50) DEFAULT 'pending',            -- pending, paid, overdue, refunded
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ins_premiums_policy ON insurance_premiums(policy_id);

-- ============================================================================
-- 5. Insurance Claims
-- ============================================================================
CREATE TABLE insurance_claims (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    policy_id       UUID NOT NULL REFERENCES insurance_policies(id),
    claim_number    VARCHAR(200) NOT NULL,
    claim_date      DATE NOT NULL,
    incident_date   DATE,
    incident_type   VARCHAR(300),
    cause           TEXT,
    description     TEXT,
    claimed_amount  NUMERIC(18,2),
    currency        VARCHAR(3) DEFAULT 'USD',
    settled_amount  NUMERIC(18,2),
    status          VARCHAR(50) DEFAULT 'submitted',          -- submitted, under_review, approved, rejected, settled, closed
    adjuster_name   VARCHAR(300),
    decision_date   DATE,
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ins_claims_policy ON insurance_claims(policy_id);
CREATE INDEX idx_ins_claims_status ON insurance_claims(status);

-- ============================================================================
-- 6. Certificates of Insurance
-- ============================================================================
CREATE TABLE certificates_of_insurance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_id       UUID NOT NULL REFERENCES insurance_policies(id),
    certificate_number VARCHAR(200) NOT NULL,
    certificate_holder VARCHAR(500),
    issue_date      DATE,
    expiry_date     DATE,
    description     TEXT,
    document_url    TEXT,
    status          VARCHAR(50) DEFAULT 'valid',              -- valid, expired, cancelled
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ins_certificates_policy ON certificates_of_insurance(policy_id);