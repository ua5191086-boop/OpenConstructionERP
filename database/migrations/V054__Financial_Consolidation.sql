-- ============================================================================
-- V054__Financial_Consolidation.sql
-- Консолидация по группе компаний, межпроектные транзакции
-- ============================================================================

CREATE TABLE consolidation_entities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES consolidation_entities(id),
    entity_code VARCHAR(50) NOT NULL UNIQUE,
    entity_name VARCHAR(500) NOT NULL,
    entity_type VARCHAR(50) NOT NULL, -- parent, subsidiary, joint_venture, spv
    registration_country VARCHAR(100),
    tax_id VARCHAR(200),
    ownership_pct NUMERIC(5,2),
    consolidation_method VARCHAR(50), -- full, equity, proportional, cost
    fiscal_year_end VARCHAR(10) DEFAULT '12-31',
    base_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE consolidation_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id UUID NOT NULL REFERENCES consolidation_entities(id),
    fiscal_year INTEGER NOT NULL,
    period_type VARCHAR(50) NOT NULL, -- monthly, quarterly, annual
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    period_status VARCHAR(50) DEFAULT 'open', -- open, closing, consolidated, locked, archived
    local_currency VARCHAR(3),
    fx_rate_used NUMERIC(18,6),
    intercompany_reconciled BOOLEAN DEFAULT FALSE,
    consolidated_at TIMESTAMPTZ,
    locked_at TIMESTAMPTZ,
    notes TEXT,
    UNIQUE(entity_id, fiscal_year, period_type, period_start)
);

CREATE TABLE intercompany_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    from_entity UUID NOT NULL REFERENCES consolidation_entities(id),
    to_entity UUID NOT NULL REFERENCES consolidation_entities(id),
    transaction_type VARCHAR(100) NOT NULL, -- loan, management_fee, dividend, service, royalty, recharges
    reference_number VARCHAR(200),
    amount NUMERIC(18,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    fx_rate NUMERIC(18,6),
    amount_consolidated NUMERIC(18,2),
    transaction_date DATE NOT NULL,
    description TEXT,
    elimination_entry BOOLEAN DEFAULT FALSE,
    elimination_period UUID REFERENCES consolidation_periods(id),
    reconciled BOOLEAN DEFAULT FALSE,
    reconciled_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'posted',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE consolidation_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period_id UUID NOT NULL REFERENCES consolidation_periods(id),
    adjustment_type VARCHAR(100) NOT NULL, -- elimination, reclassification, revaluation, correction
    description TEXT NOT NULL,
    debit_account VARCHAR(100),
    credit_account VARCHAR(100),
    amount NUMERIC(18,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    posted BOOLEAN DEFAULT FALSE,
    posted_at TIMESTAMPTZ,
    posted_by VARCHAR(200),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ce_parent ON consolidation_entities(parent_id);
CREATE INDEX idx_cp_entity ON consolidation_periods(entity_id);
CREATE INDEX idx_ict_from ON intercompany_transactions(from_entity);
CREATE INDEX idx_ict_to ON intercompany_transactions(to_entity);
CREATE INDEX idx_ict_period ON intercompany_transactions(elimination_period);
CREATE INDEX idx_ca_period ON consolidation_adjustments(period_id);