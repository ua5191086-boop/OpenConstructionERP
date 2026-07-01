-- ============================================================================
-- V005__Finance_Module.sql
-- Модуль финансов и бюджетирования (Finance Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Бюджеты проектов
-- ============================================================================
CREATE TABLE project_budgets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version         VARCHAR(50) NOT NULL,                  -- v1.0, v2.0, approved, actual
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    budget_type     VARCHAR(50) NOT NULL DEFAULT 'original', -- original, revised, actual, forecast
    total_amount    NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    contingency_pct NUMERIC(5,2) DEFAULT 5.0,              -- резерв, %
    contingency_amount NUMERIC(18,2),
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',   -- draft, approved, locked, archived
    approved_by     VARCHAR(200),
    approved_at     TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, version)
);

CREATE INDEX IF NOT EXISTS idx_budgets_project ON project_budgets(project_id);
CREATE INDEX IF NOT EXISTS idx_budgets_active ON project_budgets(project_id, is_active) WHERE is_active = TRUE;

-- ============================================================================
-- 2. Статьи бюджета
-- ============================================================================
CREATE TABLE budget_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id       UUID NOT NULL REFERENCES project_budgets(id) ON DELETE CASCADE,
    parent_id       UUID REFERENCES budget_items(id),
    item_code       VARCHAR(100) NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    item_type       VARCHAR(50) NOT NULL DEFAULT 'cost',    -- cost, revenue, profit
    cbs_code        VARCHAR(50),                             -- привязка к CBS
    planned_amount  NUMERIC(18,2) NOT NULL,
    actual_amount   NUMERIC(18,2) DEFAULT 0,
    committed_amount NUMERIC(18,2) DEFAULT 0,               -- законтрактовано
    remaining_amount NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    sort_order      INTEGER DEFAULT 0,
    is_leaf         BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_budget_items_budget ON budget_items(budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_items_parent ON budget_items(parent_id);

-- ============================================================================
-- 3. Cash Flow (план/факт)
-- ============================================================================
CREATE TABLE cash_flow (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id),
    contract_id     UUID REFERENCES contracts(id),
    entry_date      DATE NOT NULL,
    entry_type      VARCHAR(50) NOT NULL,                   -- inflow, outflow
    category        VARCHAR(100) NOT NULL,                   -- advance, progress_payment, material, salary, equipment, overhead, tax, other
    amount          NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_planned      BOOLEAN DEFAULT TRUE,                   -- TRUE = план, FALSE = факт
    description     TEXT,
    reference_type  VARCHAR(50),                             -- contract, acceptance, invoice, payroll
    reference_id    UUID,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, confirmed, reconciled
    reconciled_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cashflow_project ON cash_flow(project_id);
CREATE INDEX IF NOT EXISTS idx_cashflow_date ON cash_flow(entry_date);
CREATE INDEX IF NOT EXISTS idx_cashflow_type ON cash_flow(entry_type);
CREATE INDEX IF NOT EXISTS idx_cashflow_category ON cash_flow(category);

-- ============================================================================
-- 4. Счета-фактуры
-- ============================================================================
CREATE TABLE invoices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number  VARCHAR(100) NOT NULL UNIQUE,
    invoice_type    VARCHAR(50) NOT NULL,                   -- incoming, outgoing
    contract_id     UUID REFERENCES contracts(id),
    acceptance_id   UUID REFERENCES contract_work_acceptances(id),
    issuer_id       UUID REFERENCES organizations(id),      -- кто выставил
    recipient_id    UUID REFERENCES organizations(id),       -- кому
    invoice_date    DATE NOT NULL,
    due_date        DATE,
    amount          NUMERIC(18,2) NOT NULL,
    tax_amount      NUMERIC(18,2) DEFAULT 0,
    tax_rate        NUMERIC(5,2) DEFAULT 0,                 -- НДС, %
    total_amount    NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'issued',  -- issued, sent, received, paid, overdue, cancelled
    paid_at         TIMESTAMPTZ,
    payment_ref     VARCHAR(200),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invoices_contract ON invoices(contract_id);
CREATE INDEX IF NOT EXISTS idx_invoices_date ON invoices(invoice_date);
CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
CREATE INDEX IF NOT EXISTS idx_invoices_type ON invoices(invoice_type);

-- ============================================================================
-- 5. План-факт анализ
-- ============================================================================
CREATE TABLE cost_control (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    report_date     DATE NOT NULL,
    report_type     VARCHAR(50) NOT NULL DEFAULT 'monthly', -- monthly, quarterly, annual, custom
    total_budget    NUMERIC(18,2),
    total_committed NUMERIC(18,2),
    total_actual    NUMERIC(18,2),
    total_forecast  NUMERIC(18,2),
    variance_amount NUMERIC(18,2),                          -- отклонение
    variance_pct    NUMERIC(5,2),                           -- отклонение, %
    earned_value    NUMERIC(18,2),                          -- освоенный объём
    planned_value   NUMERIC(18,2),                          -- плановый объём
    spi             NUMERIC(5,3),                           -- Schedule Performance Index
    cpi             NUMERIC(5,3),                           -- Cost Performance Index
    status          VARCHAR(50) NOT NULL DEFAULT 'green',   -- green, yellow, red
    notes           TEXT,
    created_by      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, report_date, report_type)
);

CREATE INDEX IF NOT EXISTS idx_cost_control_project ON cost_control(project_id);
CREATE INDEX IF NOT EXISTS idx_cost_control_date ON cost_control(report_date);

-- ============================================================================
-- 6. Банковские счета
-- ============================================================================
CREATE TABLE bank_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contractor_id   UUID REFERENCES organizations(id),
    account_name    VARCHAR(300) NOT NULL,
    account_number  VARCHAR(100) NOT NULL,
    bank_name       VARCHAR(200),
    bank_code       VARCHAR(50),                            -- BLZ, SWIFT, BIC
    iban            VARCHAR(50),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    account_type    VARCHAR(50) NOT NULL DEFAULT 'checking', -- checking, savings, credit, escrow
    is_active       BOOLEAN DEFAULT TRUE,
    balance         NUMERIC(18,2) DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 7. Транзакции по счетам
-- ============================================================================
CREATE TABLE bank_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id      UUID NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    transaction_date DATE NOT NULL,
    description     TEXT,
    amount          NUMERIC(18,2) NOT NULL,
    balance_after   NUMERIC(18,2),
    transaction_type VARCHAR(50) NOT NULL,                  -- debit, credit
    reference_type  VARCHAR(50),                            -- invoice, payment, transfer, fee
    reference_id    UUID,
    reconciled      BOOLEAN DEFAULT FALSE,
    reconciled_at   TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_tx_account ON bank_transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_bank_tx_date ON bank_transactions(transaction_date);

-- ============================================================================
-- 8. Налоги
-- ============================================================================
CREATE TABLE tax_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contractor_id   UUID REFERENCES organizations(id),
    tax_type        VARCHAR(100) NOT NULL,                  -- vat, income_tax, social_tax, property_tax
    tax_period      VARCHAR(50) NOT NULL,                   -- Q1-2026, 2026
    taxable_amount  NUMERIC(18,2),
    tax_rate        NUMERIC(5,2),
    tax_amount      NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'calculated', -- calculated, filed, paid, overdue
    due_date        DATE,
    paid_at         TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contractor_id, tax_type, tax_period)
);

-- ============================================================================
-- 9. Финансовые отчёты
-- ============================================================================
CREATE TABLE financial_reports (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id),
    report_type     VARCHAR(100) NOT NULL,                  -- pnl, balance_sheet, cash_flow, budget_vs_actual
    report_period   VARCHAR(50) NOT NULL,                   -- Q1-2026, 2026-annual
    report_data     JSONB,                                  -- данные отчёта
    total_revenue   NUMERIC(18,2),
    total_expense   NUMERIC(18,2),
    net_profit      NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',   -- draft, final, approved
    generated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    generated_by    VARCHAR(100),
    notes           TEXT
);

-- ============================================================================
-- 10. Валюты и курсы
-- ============================================================================
CREATE TABLE exchange_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency   VARCHAR(3) NOT NULL,
    to_currency     VARCHAR(3) NOT NULL,
    rate            NUMERIC(18,6) NOT NULL,
    rate_date       DATE NOT NULL,
    source          VARCHAR(100) DEFAULT 'manual',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(from_currency, to_currency, rate_date)
);

CREATE INDEX IF NOT EXISTS idx_rates_date ON exchange_rates(rate_date);

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE project_budgets IS 'Бюджеты проектов';
COMMENT ON TABLE budget_items IS 'Статьи бюджета (дерево)';
COMMENT ON TABLE cash_flow IS 'Денежные потоки (план/факт)';
COMMENT ON TABLE invoices IS 'Счета-фактуры';
COMMENT ON TABLE cost_control IS 'План-факт анализ / Earned Value';
COMMENT ON TABLE bank_accounts IS 'Банковские счета';
COMMENT ON TABLE bank_transactions IS 'Банковские транзакции';
COMMENT ON TABLE tax_records IS 'Налоговые записи';
COMMENT ON TABLE financial_reports IS 'Финансовые отчёты (P&L, Balance Sheet)';
COMMENT ON TABLE exchange_rates IS 'Курсы валют';
