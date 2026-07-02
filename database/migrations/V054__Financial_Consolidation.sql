-- ============================================================================
-- V054__Financial_Consolidation.sql
-- Консолидация по группе компаний, межпроектные транзакции
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Юридические лица группы
-- ============================================================================
CREATE TABLE group_legal_entities (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    legal_name          VARCHAR(500) NOT NULL,
    registration_number VARCHAR(200),
    tax_id              VARCHAR(200),
    country             VARCHAR(100),
    address             TEXT,
    legal_form          VARCHAR(100),                           -- LLC, JSC, subsidiary, branch
    parent_entity_id    UUID REFERENCES group_legal_entities(id),
    consolidation_method VARCHAR(50) DEFAULT 'full',           -- full, proportional, equity, cost
    ownership_pct       NUMERIC(5,2),
    currency            VARCHAR(3) DEFAULT 'USD',
    fiscal_year_end     VARCHAR(10) DEFAULT '2026-12-31',
    is_active           BOOLEAN DEFAULT TRUE,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gle_parent ON group_legal_entities(parent_entity_id);

COMMENT ON TABLE group_legal_entities IS 'Юридические лица группы компаний';

-- ============================================================================
-- 2. Межпроектные транзакции (intercompany)
-- ============================================================================
CREATE TABLE intercompany_transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_project_id     UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    to_project_id       UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    from_entity_id      UUID REFERENCES group_legal_entities(id),
    to_entity_id        UUID REFERENCES group_legal_entities(id),
    transaction_type    VARCHAR(100) NOT NULL,                  -- service_charge, material_transfer, loan, allocation, dividend
    reference_doc       VARCHAR(200),
    description         TEXT NOT NULL,
    amount              NUMERIC(18,2) NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate       NUMERIC(18,6) DEFAULT 1,
    amount_base_ccy     NUMERIC(18,2),
    transaction_date    DATE NOT NULL,
    elimination_entry   BOOLEAN DEFAULT FALSE,                  -- TRUE = элиминационная проводка
    elimination_date    DATE,
    status              VARCHAR(50) DEFAULT 'posted',           -- draft, posted, eliminated, reversed
    created_by          VARCHAR(200),
    approved_by         VARCHAR(200),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ict_from ON intercompany_transactions(from_project_id);
CREATE INDEX idx_ict_to ON intercompany_transactions(to_project_id);
CREATE INDEX idx_ict_date ON intercompany_transactions(transaction_date);
CREATE INDEX idx_ict_status ON intercompany_transactions(status);

COMMENT ON TABLE intercompany_transactions IS 'Межпроектные и внутригрупповые транзакции';

-- ============================================================================
-- 3. Консолидированная отчётность
-- ============================================================================
CREATE TABLE consolidation_reports (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_name         VARCHAR(500) NOT NULL,
    report_type         VARCHAR(100) NOT NULL,                 -- balance_sheet, income_statement, cash_flow, trial_balance
    period_from         DATE NOT NULL,
    period_to           DATE NOT NULL,
    currency            VARCHAR(3) DEFAULT 'USD',
    entities_included   JSONB DEFAULT '[]'::JSONB,             -- список ID юрлиц
    exclusion_rules     JSONB,                                  -- правила элиминации
    total_assets        NUMERIC(18,2),
    total_liabilities   NUMERIC(18,2),
    total_equity        NUMERIC(18,2),
    total_revenue       NUMERIC(18,2),
    net_income          NUMERIC(18,2),
    status              VARCHAR(50) DEFAULT 'draft',            -- draft, generated, reviewed, approved, published
    generated_by        VARCHAR(200),
    generated_at        TIMESTAMPTZ,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cr_period ON consolidation_reports(period_from, period_to);

COMMENT ON TABLE consolidation_reports IS 'Консолидированная отчётность по группе';

-- ============================================================================
-- 4. Элиминационные проводки
-- ============================================================================
CREATE TABLE consolidation_eliminations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id           UUID NOT NULL REFERENCES consolidation_reports(id) ON DELETE CASCADE,
    ic_transaction_id   UUID REFERENCES intercompany_transactions(id) ON DELETE SET NULL,
    account_code        VARCHAR(100) NOT NULL,
    account_name        VARCHAR(300) NOT NULL,
    amount              NUMERIC(18,2) NOT NULL,
    direction           VARCHAR(10) NOT NULL,                   -- debit, credit
    entity_id           UUID REFERENCES group_legal_entities(id),
    description         TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ce_report ON consolidation_eliminations(report_id);

COMMENT ON TABLE consolidation_eliminations IS 'Элиминационные проводки для консолидации';