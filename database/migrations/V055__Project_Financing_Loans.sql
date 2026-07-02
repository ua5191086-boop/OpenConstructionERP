-- ============================================================================
-- V055__Project_Financing_Loans.sql
-- Кредитные линии, графики выборки, проценты, ковенанты
-- ============================================================================

CREATE TABLE financing_facilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    facility_type VARCHAR(100) NOT NULL, -- term_loan, revolving_credit, overdraft, guarantee_facility, bridge_loan
    facility_number VARCHAR(200) NOT NULL,
    lender VARCHAR(300) NOT NULL,
    total_amount NUMERIC(18,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    drawn_amount NUMERIC(18,2) DEFAULT 0,
    available_amount NUMERIC(18,2),
    interest_rate_type VARCHAR(50), -- fixed, floating, mixed
    base_rate VARCHAR(50), -- SOFR, EURIBOR, TONA, fixed
    spread NUMERIC(8,4), -- margin over base rate
    all_in_rate NUMERIC(8,4),
    commitment_fee NUMERIC(8,4), -- %
    arrangement_fee NUMERIC(18,2),
    start_date DATE NOT NULL,
    maturity_date DATE NOT NULL,
    grace_period_months INTEGER DEFAULT 0,
    repayment_schedule JSONB, -- [{"date":"...","amount":...}]
    secured_by TEXT,
    covenants JSONB, -- [{"type":"dscr","threshold":1.2,"current":1.15}]
    status VARCHAR(50) DEFAULT 'active', -- active, fully_drawn, repaid, closed, defaulted
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE financing_drawdowns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id UUID NOT NULL REFERENCES financing_facilities(id),
    drawdown_number VARCHAR(100) NOT NULL,
    amount NUMERIC(18,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    drawdown_date DATE NOT NULL,
    interest_period_start DATE,
    interest_period_end DATE,
    interest_amount NUMERIC(18,2),
    all_in_rate NUMERIC(8,4),
    purpose TEXT,
    approved_by VARCHAR(200),
    approval_date DATE,
    status VARCHAR(50) DEFAULT 'requested', -- requested, approved, disbursed, cancelled
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE financing_repayments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id UUID NOT NULL REFERENCES financing_facilities(id),
    drawdown_id UUID REFERENCES financing_drawdowns(id),
    repayment_type VARCHAR(50) NOT NULL, -- principal, interest, fee, prepayment
    amount NUMERIC(18,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    repayment_date DATE NOT NULL,
    reference VARCHAR(200),
    status VARCHAR(50) DEFAULT 'scheduled', -- scheduled, paid, overdue, waived
    paid_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE financing_covenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    facility_id UUID NOT NULL REFERENCES financing_facilities(id),
    covenant_type VARCHAR(100) NOT NULL, -- DSCR, LLCR, PLCR, LTV, ICR, debt_equity, min_cash
    description TEXT NOT NULL,
    threshold NUMERIC(18,4) NOT NULL,
    current_value NUMERIC(18,4),
    measurement_frequency VARCHAR(50), -- monthly, quarterly, annually
    last_measured_at TIMESTAMPTZ,
    next_measurement_at DATE,
    status VARCHAR(50) DEFAULT 'compliant', -- compliant, warning, breached, waived
    waiver_ref VARCHAR(200),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ff_project ON financing_facilities(project_id);
CREATE INDEX idx_fd_facility ON financing_drawdowns(facility_id);
CREATE INDEX idx_fr_facility ON financing_repayments(facility_id);
CREATE INDEX idx_fc_facility ON financing_covenants(facility_id);
CREATE INDEX idx_fc_status ON financing_covenants(status) WHERE status != 'compliant';