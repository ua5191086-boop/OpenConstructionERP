-- ============================================================================
-- V042__Audit_Trail_Tax_Management.sql
-- Иммутабельный аудиторский след и налоговый учёт
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Иммутабельный аудиторский след (Audit Trail)
-- ============================================================================
CREATE TABLE audit_trail (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id) ON DELETE SET NULL,
    entity_type     VARCHAR(100) NOT NULL,                     -- contract, invoice, payment, po, budget, ipc
    entity_id       UUID NOT NULL,                             -- ID изменяемой сущности
    action          VARCHAR(50) NOT NULL,                       -- create, update, delete, approve, reject, void, revalue
    field_name      VARCHAR(200),                               -- имя изменённого поля (NULL для create/delete)
    old_value       TEXT,                                       -- предыдущее значение
    new_value       TEXT,                                       -- новое значение
    changed_by      VARCHAR(200) NOT NULL,
    changed_by_role VARCHAR(100),
    change_reason   TEXT,                                       -- причина изменения (обязательно для financial)
    financial_impact NUMERIC(18,2),                             -- финансовый эффект изменения
    currency        VARCHAR(3),
    is_financial    BOOLEAN DEFAULT FALSE,                      -- TRUE — финансово-значимое изменение
    checksum        VARCHAR(64) NOT NULL,                       -- SHA-256 предыдущей записи (иммутабельная цепочка)
    previous_checksum VARCHAR(64),
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска по сущности
CREATE INDEX idx_audit_entity ON audit_trail(entity_type, entity_id);
CREATE INDEX idx_audit_project ON audit_trail(project_id);
CREATE INDEX idx_audit_financial ON audit_trail(is_financial) WHERE is_financial = TRUE;
CREATE INDEX idx_audit_created ON audit_trail(created_at);
-- Индекс для проверки цепочки
CREATE INDEX idx_audit_checksum ON audit_trail(checksum);

COMMENT ON TABLE audit_trail IS 'Иммутабельный аудиторский след — SHA-256 цепочка всех финансово-значимых изменений';

-- Функция: вычисление SHA-256 для аудита
CREATE OR REPLACE FUNCTION audit_compute_checksum()
RETURNS TRIGGER AS $$
DECLARE
    prev_checksum VARCHAR(64);
    raw TEXT;
BEGIN
    -- Получаем checksum предыдущей записи для этой же сущности
    SELECT checksum INTO prev_checksum
    FROM audit_trail
    WHERE entity_type = NEW.entity_type AND entity_id = NEW.entity_id
    ORDER BY created_at DESC
    LIMIT 1;
    
    NEW.previous_checksum := prev_checksum;
    
    -- Собираем сырые данные для хэша
    raw := COALESCE(NEW.entity_type::TEXT, '') || '|' ||
           COALESCE(NEW.entity_id::TEXT, '') || '|' ||
           COALESCE(NEW.action, '') || '|' ||
           COALESCE(NEW.field_name, '') || '|' ||
           COALESCE(NEW.old_value, '') || '|' ||
           COALESCE(NEW.new_value, '') || '|' ||
           COALESCE(NEW.changed_by, '') || '|' ||
           COALESCE(prev_checksum, '');
    
    NEW.checksum := encode(sha256(raw::bytea), 'hex');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_checksum
    BEFORE INSERT ON audit_trail
    FOR EACH ROW
    EXECUTE FUNCTION audit_compute_checksum();

-- ============================================================================
-- 2. Налоговый учёт (Tax Management)
-- ============================================================================
CREATE TABLE tax_registrations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    country             VARCHAR(100) NOT NULL,
    tax_authority       VARCHAR(300) NOT NULL,
    registration_number VARCHAR(200) NOT NULL,
    tax_identifier      VARCHAR(200),                           -- ИНН/НДС номер
    registration_date   DATE NOT NULL,
    expiry_date         DATE,
    tax_regime          VARCHAR(100) NOT NULL DEFAULT 'standard', -- standard, simplified, exempt, special
    status              VARCHAR(50) NOT NULL DEFAULT 'active',   -- active, suspended, cancelled
    filing_frequency    VARCHAR(50) DEFAULT 'monthly',           -- monthly, quarterly, annually
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tax_reg_project ON tax_registrations(project_id);

COMMENT ON TABLE tax_registrations IS 'Налоговые регистрации проектов';

-- ============================================================================
-- 3. Налоговые накладные / счета-фактуры
-- ============================================================================
CREATE TABLE tax_invoices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    contract_id         UUID REFERENCES contracts(id) ON DELETE SET NULL,
    invoice_type        VARCHAR(50) NOT NULL,                   -- sales, purchase, credit_note, debit_note
    invoice_number      VARCHAR(200) NOT NULL,
    invoice_date        DATE NOT NULL,
    counterparty        VARCHAR(300) NOT NULL,
    counterparty_tax_id VARCHAR(200),
    gross_amount        NUMERIC(18,2) NOT NULL,                 -- с НДС
    net_amount          NUMERIC(18,2) NOT NULL,                 -- без НДС
    tax_amount          NUMERIC(18,2) NOT NULL,                 -- НДС
    tax_rate            NUMERIC(5,2) NOT NULL,                  -- ставка НДС (%)
    tax_code            VARCHAR(50),                            -- код налога
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    status              VARCHAR(50) NOT NULL DEFAULT 'issued',  -- issued, paid, overdue, cancelled, reversed
    due_date            DATE,
    paid_date           DATE,
    is_reverse_charge   BOOLEAN DEFAULT FALSE,                  -- обратный НДС (самообложение)
    fiscal_period       VARCHAR(50),                            -- налоговый период: YYYY-MM
    document_ref        VARCHAR(500),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tax_inv_project ON tax_invoices(project_id);
CREATE INDEX idx_tax_inv_contract ON tax_invoices(contract_id);
CREATE INDEX idx_tax_inv_fiscal ON tax_invoices(fiscal_period);
CREATE INDEX idx_tax_inv_status ON tax_invoices(status);
CREATE INDEX idx_tax_inv_date ON tax_invoices(invoice_date);

COMMENT ON TABLE tax_invoices IS 'Налоговые счета-фактуры с учётом НДС';

-- ============================================================================
-- 4. Налоговые отчёты и декларации
-- ============================================================================
CREATE TABLE tax_returns (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE SET NULL,
    tax_registration_id UUID REFERENCES tax_registrations(id) ON DELETE SET NULL,
    return_type         VARCHAR(100) NOT NULL,                  -- vat_return, income_tax, payroll_tax, withholding
    fiscal_period       VARCHAR(50) NOT NULL,                   -- YYYY-MM или YYYY
    period_start        DATE NOT NULL,
    period_end          DATE NOT NULL,
    total_taxable_amount NUMERIC(18,2) NOT NULL,
    total_tax_due       NUMERIC(18,2) NOT NULL,
    total_tax_credit    NUMERIC(18,2) DEFAULT 0,               -- налоговый кредит
    net_tax_payable     NUMERIC(18,2) NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    filing_date         DATE,
    due_date            DATE,
    paid_date           DATE,
    status              VARCHAR(50) NOT NULL DEFAULT 'draft',   -- draft, filed, paid, audited, amended
    filed_by            VARCHAR(200),
    amended_return_id   UUID REFERENCES tax_returns(id),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tax_ret_project ON tax_returns(project_id);
CREATE INDEX idx_tax_ret_period ON tax_returns(fiscal_period);
CREATE INDEX idx_tax_ret_status ON tax_returns(status);

COMMENT ON TABLE tax_returns IS 'Налоговые декларации и отчёты';

-- ============================================================================
-- 5. Трансфертное ценообразование
-- ============================================================================
CREATE TABLE transfer_pricing_records (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    transaction_type    VARCHAR(100) NOT NULL,                  -- intercompany_service, management_fee, royalty, loan
    related_party       VARCHAR(300) NOT NULL,
    related_party_country VARCHAR(100),
    transaction_value   NUMERIC(18,2) NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    arm_length_value    NUMERIC(18,2),                          -- рыночная цена для сравнения
    transfer_price_method VARCHAR(100),                         -- CUP, resale_minus, cost_plus, TNMM, profit_split
    documentation_ref   VARCHAR(500),
    fiscal_year         INTEGER NOT NULL,
    status              VARCHAR(50) DEFAULT 'pending_review',   -- pending_review, compliant, non_compliant
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tp_project ON transfer_pricing_records(project_id);
CREATE INDEX idx_tp_fiscal ON transfer_pricing_records(fiscal_year);

COMMENT ON TABLE transfer_pricing_records IS 'Учёт трансфертного ценообразования для финансов года';