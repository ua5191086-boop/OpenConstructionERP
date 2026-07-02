-- ============================================================================
-- V027__Funding_Module.sql
-- Finance: Funding Sources, Multi-Currency, Guarantees
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Funding Sources (banks, ECA, investors, grants)
-- ============================================================================
CREATE TABLE funding_sources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_type     VARCHAR(50) NOT NULL,                     -- bank, eca, investor, grant, internal
    source_name     VARCHAR(500) NOT NULL,
    source_code     VARCHAR(100),
    description     TEXT,
    contact_info    JSONB,
    commitment_amount NUMERIC(18,2) DEFAULT 0,
    currency        VARCHAR(3) DEFAULT 'USD',
    status          VARCHAR(50) DEFAULT 'active',             -- active, pending, closed
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_funding_src_project ON funding_sources(project_id);
CREATE INDEX idx_funding_src_type ON funding_sources(source_type);

-- ============================================================================
-- 2. Funding Tranches (транши с датами)
-- ============================================================================
CREATE TABLE funding_tranches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    funding_source_id UUID NOT NULL REFERENCES funding_sources(id),
    tranche_name    VARCHAR(300) NOT NULL,
    amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(3) DEFAULT 'USD',
    expected_date   DATE,
    actual_date     DATE,
    status          VARCHAR(50) DEFAULT 'planned',            -- planned, disbursed, cancelled
    terms           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_funding_tranches_source ON funding_tranches(funding_source_id);
CREATE INDEX idx_funding_tranches_status ON funding_tranches(status);

-- ============================================================================
-- 3. Funding Drawdowns (выборки)
-- ============================================================================
CREATE TABLE funding_drawdowns (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    funding_source_id UUID NOT NULL REFERENCES funding_sources(id),
    tranche_id      UUID REFERENCES funding_tranches(id),
    drawdown_date   DATE NOT NULL,
    amount          NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    exchange_rate   NUMERIC(12,6) DEFAULT 1,
    reference       VARCHAR(200),
    status          VARCHAR(50) DEFAULT 'completed',          -- completed, pending, rejected
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_funding_drawdowns_source ON funding_drawdowns(funding_source_id);

-- ============================================================================
-- 4. Funding Covenants (ковенанты)
-- ============================================================================
CREATE TABLE funding_covenants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    funding_source_id UUID NOT NULL REFERENCES funding_sources(id),
    covenant_type   VARCHAR(100) NOT NULL,                    -- financial, operational, reporting
    covenant_name   VARCHAR(500) NOT NULL,
    description     TEXT,
    metric          VARCHAR(200),
    threshold       VARCHAR(200),
    status          VARCHAR(50) DEFAULT 'active',             -- active, breached, waived, fulfilled
    breach_date     DATE,
    breach_notes    TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_funding_covenants_source ON funding_covenants(funding_source_id);

-- ============================================================================
-- 5. Multi-Currency Rates (расширение exchange_rates)
-- ============================================================================
CREATE TABLE multi_currency_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency   VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate            NUMERIC(14,6) NOT NULL,
    rate_date       DATE NOT NULL,
    source          VARCHAR(100),                             -- central_bank, bloomberg, manual
    is_historical   BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(base_currency, target_currency, rate_date)
);
CREATE INDEX idx_currency_rates_date ON multi_currency_rates(rate_date);

-- ============================================================================
-- 6. Currency Hedges (хеджи)
-- ============================================================================
CREATE TABLE currency_hedges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    hedge_type      VARCHAR(50) NOT NULL,                     -- forward, future, option, swap
    base_currency   VARCHAR(3) NOT NULL,
    hedge_currency  VARCHAR(3) NOT NULL,
    notional_amount NUMERIC(18,2) NOT NULL,
    strike_rate     NUMERIC(14,6),
    maturity_date   DATE,
    counterparty    VARCHAR(300),
    status          VARCHAR(50) DEFAULT 'active',             -- active, matured, closed
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_currency_hedges_project ON currency_hedges(project_id);

-- ============================================================================
-- 7. Guarantees (performance bonds, bank guarantees, retention)
-- ============================================================================
CREATE TABLE guarantees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    contract_id     UUID REFERENCES contracts(id),
    guarantee_type  VARCHAR(100) NOT NULL,                    -- performance_bond, bid_bond, bank_guarantee, retention, advance_payment
    guarantee_number VARCHAR(200),
    issuing_bank    VARCHAR(500),
    beneficiary     VARCHAR(500),
    applicant       VARCHAR(500),
    amount          NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) DEFAULT 'USD',
    issue_date      DATE,
    expiry_date     DATE,
    claim_expiry_date DATE,
    status          VARCHAR(50) DEFAULT 'active',             -- active, claimed, expired, released
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_guarantees_project ON guarantees(project_id);
CREATE INDEX idx_guarantees_type ON guarantees(guarantee_type);

-- ============================================================================
-- 8. Guarantee Claims (требования по гарантиям)
-- ============================================================================
CREATE TABLE guarantee_claims (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    guarantee_id    UUID NOT NULL REFERENCES guarantees(id),
    claim_date      DATE NOT NULL,
    claim_amount    NUMERIC(18,2) NOT NULL,
    claim_reason    TEXT,
    claim_status    VARCHAR(50) DEFAULT 'submitted',          -- submitted, under_review, approved, rejected, paid
    response_date   DATE,
    settlement_amount NUMERIC(18,2),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_guarantee_claims_guarantee ON guarantee_claims(guarantee_id);

-- ============================================================================
-- 9. Guarantee Amendments (изменения гарантий)
-- ============================================================================
CREATE TABLE guarantee_amendments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    guarantee_id    UUID NOT NULL REFERENCES guarantees(id),
    amendment_number VARCHAR(100),
    amendment_date  DATE,
    description     TEXT,
    new_amount      NUMERIC(18,2),
    new_expiry_date DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_guarantee_amendments ON guarantee_amendments(guarantee_id);