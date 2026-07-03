-- ============================================================================
-- V041__Retention_Guarantees_MultiCurrency.sql
-- Удержания, гарантии и мультивалютный учёт
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Удержания (Retention) — гарантийное удержание из платежей
-- ============================================================================
CREATE TABLE retention_releases (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id         UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    retention_pct       NUMERIC(5,2) NOT NULL DEFAULT 5.0,      -- % удержания
    retention_amount    NUMERIC(18,2) NOT NULL,                  -- сумма удержания
    released_amount     NUMERIC(18,2) DEFAULT 0,                 -- сколько уже высвобождено
    release_condition   VARCHAR(100) NOT NULL DEFAULT 'acceptance', -- acceptance, warranty_expiry, milestone
    release_date        DATE,
    release_status      VARCHAR(50) NOT NULL DEFAULT 'held',     -- held, partially_released, fully_released
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_retention_contract ON retention_releases(contract_id);
CREATE INDEX IF NOT EXISTS idx_retention_project ON retention_releases(project_id);
CREATE INDEX IF NOT EXISTS idx_retention_status ON retention_releases(release_status);

COMMENT ON TABLE retention_releases IS 'Гарантийные удержания из платежей по контрактам';

-- ============================================================================
-- 2. Банковские гарантии (Guarantees)
-- ============================================================================
CREATE TABLE IF NOT EXISTS guarantees (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id         UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    guarantee_type      VARCHAR(50) NOT NULL,                    -- bid_bond, performance, advance_payment, warranty, retention
    guarantee_number    VARCHAR(200) NOT NULL,
    issuer_bank         VARCHAR(300) NOT NULL,
    beneficiary         VARCHAR(300) NOT NULL,                   -- в чью пользу
    amount              NUMERIC(18,2) NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    issue_date          DATE NOT NULL,
    expiry_date         DATE NOT NULL,
    claim_deadline      DATE,                                    -- срок предъявления требований
    status              VARCHAR(50) NOT NULL DEFAULT 'active',   -- active, expired, claimed, released, extended
    extended_to         DATE,                                    -- новая дата при продлении
    extension_count     INTEGER DEFAULT 0,
    document_ref        VARCHAR(500),                            -- ссылка на скан
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_guarantees_contract ON guarantees(contract_id);
CREATE INDEX IF NOT EXISTS idx_guarantees_project ON guarantees(project_id);
CREATE INDEX IF NOT EXISTS idx_guarantees_status ON guarantees(status);
CREATE INDEX IF NOT EXISTS idx_guarantees_expiry ON guarantees(expiry_date);

COMMENT ON TABLE guarantees IS 'Банковские гарантии по контрактам';

-- ============================================================================
-- 3. Кросс-курсы валют
-- ============================================================================
CREATE TABLE currency_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency   VARCHAR(3) NOT NULL,                        -- базовая валюта (например USD)
    target_currency VARCHAR(3) NOT NULL,                        -- целевая валюта (например KZT)
    rate            NUMERIC(18,6) NOT NULL,                     -- курс: 1 base = rate target
    rate_date       DATE NOT NULL,
    source          VARCHAR(50) DEFAULT 'manual',               -- manual, central_bank, provider
    is_historical   BOOLEAN DEFAULT FALSE,                      -- TRUE для исторических курсов
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(base_currency, target_currency, rate_date)
);

CREATE INDEX IF NOT EXISTS idx_currency_rates_date ON currency_rates(rate_date);
CREATE INDEX IF NOT EXISTS idx_currency_rates_pair ON currency_rates(base_currency, target_currency);

COMMENT ON TABLE currency_rates IS 'Кросс-курсы валют для мультивалютного учёта';

-- ============================================================================
-- 4. Мультивалютные транзакции
-- ============================================================================
CREATE TABLE multi_currency_transactions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    transaction_type    VARCHAR(50) NOT NULL,                   -- payment, receipt, revaluation, conversion
    source_currency     VARCHAR(3) NOT NULL,
    target_currency     VARCHAR(3) NOT NULL,
    source_amount       NUMERIC(18,2) NOT NULL,
    target_amount       NUMERIC(18,2) NOT NULL,
    exchange_rate       NUMERIC(18,6) NOT NULL,
    transaction_date    DATE NOT NULL,
    reference_type      VARCHAR(100),                           -- contract, invoice, guarantee
    reference_id        UUID,
    realized_gain_loss  NUMERIC(18,2),                          -- реализованная курсовая разница
    status              VARCHAR(50) NOT NULL DEFAULT 'posted',  -- draft, posted, reversed
    created_by          VARCHAR(200),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mct_project ON multi_currency_transactions(project_id);
CREATE INDEX IF NOT EXISTS idx_mct_date ON multi_currency_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_mct_type ON multi_currency_transactions(transaction_type);

COMMENT ON TABLE multi_currency_transactions IS 'Мультивалютные транзакции и курсовая переоценка';