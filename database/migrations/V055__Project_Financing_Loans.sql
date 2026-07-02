-- ============================================================================
-- V055__Project_Financing_Loans.sql
-- Кредитные линии, графики выборки, проценты, ковенанты
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Кредитные линии (facilities)
-- ============================================================================
CREATE TABLE loan_facilities (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    facility_type       VARCHAR(100) NOT NULL,                  -- term_loan, revolving, overdraft, guarantee_line, bridge_financing
    facility_number     VARCHAR(200) NOT NULL,
    lender              VARCHAR(300) NOT NULL,
    lender_country      VARCHAR(100),
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    total_limit         NUMERIC(18,2) NOT NULL,
    drawn_amount        NUMERIC(18,2) DEFAULT 0,
    available_amount    NUMERIC(18,2),
    interest_rate_type  VARCHAR(50) DEFAULT 'fixed',            -- fixed, floating, blended, fixed_floating
    base_rate           VARCHAR(50),                            -- SOFR, EURIBOR, TONAR, CentralBank, fixed
    margin_bps          INTEGER,                                -- спред в базисных пунктах
    all_in_rate         NUMERIC(8,4),                           -- полная ставка %
    repayment_schedule  VARCHAR(100),                           -- bullet, amortizing, grace_period, custom
    grace_period_months INTEGER,
    tenor_months        INTEGER NOT NULL,
    maturity_date       DATE NOT NULL,
    arrangement_fee_pct NUMERIC(5,2),
    commitment_fee_pct  NUMERIC(5,2),
    prepayment_penalty  TEXT,
    covenants           JSONB DEFAULT '[]'::JSONB,              -- [{type:"DSCR",requirement:">1.2x",current:1.35,status:"compliant"}]
    security            TEXT,                                    -- обеспечение
    status              VARCHAR(50) DEFAULT 'pending',          -- pending, approved, active, fully_drawn, repaid, defaulted
    approved_by         VARCHAR(200),
    approval_date       DATE,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lf_project ON loan_facilities(project_id);
CREATE INDEX idx_lf_status ON loan_facilities(status);
CREATE INDEX idx_lf_lender ON loan_facilities(lender);

COMMENT ON TABLE loan_facilities IS 'Кредитные линии и заёмные средства проекта';

-- ============================================================================
-- 2. Выборки по кредитам (drawdowns)
-- ============================================================================
CREATE TABLE loan_drawdowns (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id         UUID NOT NULL REFERENCES loan_facilities(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    drawdown_number     VARCHAR(100) NOT NULL,
    drawdown_date       DATE NOT NULL,
    amount              NUMERIC(18,2) NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate       NUMERIC(18,6) DEFAULT 1,
    amount_base_ccy     NUMERIC(18,2),
    interest_start_date DATE,
    interest_rate       NUMERIC(8,4),
    repayment_date      DATE,
    repaid_amount       NUMERIC(18,2) DEFAULT 0,
    outstanding         NUMERIC(18,2),
    purpose             TEXT,
    reference_doc       VARCHAR(500),
    status              VARCHAR(50) DEFAULT 'disbursed',        -- disbursed, outstanding, repaid, defaulted
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ld_facility ON loan_drawdowns(facility_id);
CREATE INDEX idx_ld_status ON loan_drawdowns(status);

COMMENT ON TABLE loan_drawdowns IS 'Выборки по кредитным линиям';

-- ============================================================================
-- 3. График погашения (amortization schedule)
-- ============================================================================
CREATE TABLE loan_repayment_schedule (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drawdown_id         UUID NOT NULL REFERENCES loan_drawdowns(id) ON DELETE CASCADE,
    facility_id         UUID NOT NULL REFERENCES loan_facilities(id) ON DELETE CASCADE,
    installment_number  INTEGER NOT NULL,
    due_date            DATE NOT NULL,
    principal_amount    NUMERIC(18,2) NOT NULL,
    interest_amount     NUMERIC(18,2) NOT NULL,
    total_amount        NUMERIC(18,2) NOT NULL,
    paid_principal      NUMERIC(18,2) DEFAULT 0,
    paid_interest       NUMERIC(18,2) DEFAULT 0,
    paid_date           DATE,
    days_accrued        INTEGER,
    status              VARCHAR(50) DEFAULT 'pending',          -- pending, paid, overdue, partial, defaulted
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lrs_drawdown ON loan_repayment_schedule(drawdown_id);
CREATE INDEX idx_lrs_due ON loan_repayment_schedule(due_date) WHERE status = 'pending';
CREATE INDEX idx_lrs_status ON loan_repayment_schedule(status);

COMMENT ON TABLE loan_repayment_schedule IS 'График погашения кредитов';

-- ============================================================================
-- 4. Ковенанты и их мониторинг
-- ============================================================================
CREATE TABLE loan_covenant_monitoring (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id         UUID NOT NULL REFERENCES loan_facilities(id) ON DELETE CASCADE,
    covenant_type       VARCHAR(100) NOT NULL,                  -- DSCR, LLCR, PLCR, LTV, DebtEquity, ICR, min_equity
    description         TEXT NOT NULL,
    requirement         VARCHAR(200) NOT NULL,                  -- ">= 1.2x", "<= 60%", "> $5M"
    current_value       NUMERIC(12,4),
    threshold_min       NUMERIC(12,4),
    threshold_max       NUMERIC(12,4),
    measurement_date    DATE NOT NULL,
    measurement_period  VARCHAR(50),                            -- quarterly, semi_annual, annual
    status              VARCHAR(50) DEFAULT 'compliant',        -- compliant, early_warning, breach, waived
    breach_notice_date  DATE,
    waiver_date         DATE,
    waiver_expiry       DATE,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lcm_facility ON loan_covenant_monitoring(facility_id);
CREATE INDEX idx_lcm_status ON loan_covenant_monitoring(status);

COMMENT ON TABLE loan_covenant_monitoring IS 'Мониторинг ковенант по кредитам';

-- ============================================================================
-- 5. View: кредитный портфель проекта
-- ============================================================================
CREATE OR REPLACE VIEW project_financing_summary AS
SELECT
    lf.project_id,
    p.code as project_code,
    p.name as project_name,
    COUNT(DISTINCT lf.id) as facility_count,
    SUM(lf.total_limit) as total_limit,
    SUM(lf.drawn_amount) as total_drawn,
    SUM(lf.available_amount) as total_available,
    COUNT(DISTINCT CASE WHEN lf.status='active' THEN lf.id END) as active_facilities,
    COUNT(DISTINCT CASE WHEN lcm.status='breach' THEN lcm.id END) as covenant_breaches,
    MIN(lf.maturity_date) as earliest_maturity,
    MAX(lf.maturity_date) as latest_maturity
FROM loan_facilities lf
JOIN projects p ON p.id = lf.project_id
LEFT JOIN loan_covenant_monitoring lcm ON lcm.facility_id = lf.id
GROUP BY lf.project_id, p.code, p.name;

COMMENT ON VIEW project_financing_summary IS 'Сводка по финансированию проекта — кредитный портфель';