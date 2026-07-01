-- ============================================================================
-- V003__Contract_Module.sql
-- Модуль управления договорами (Contract Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================
-- Зависимости: V001__BOQ_Module.sql, V002__Tender_Module.sql
-- ============================================================================

-- ============================================================================
-- 1. Договоры
-- ============================================================================
CREATE TABLE contracts (
    id              BIGSERIAL PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,          -- C-2026-001
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    contract_type   VARCHAR(50) NOT NULL DEFAULT 'lump_sum', -- lump_sum, unit_price, cost_plus, time_material, design_build, epc, epcm
    status          VARCHAR(50) NOT NULL DEFAULT 'draft', -- draft, negotiation, signed, active, suspended, completed, terminated
    tender_id       BIGINT REFERENCES tenders(id),
    lot_id          BIGINT REFERENCES tender_lots(id),
    client_id       BIGINT NOT NULL REFERENCES contractors(id),
    contractor_id   BIGINT NOT NULL REFERENCES contractors(id),
    project_id      BIGINT REFERENCES projects(id),
    
    -- Суммы
    contract_amount NUMERIC(18,2) NOT NULL,               -- сумма договора
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    advance_amount  NUMERIC(18,2),                        -- аванс
    advance_pct     NUMERIC(5,2),                          -- аванс, %
    
    -- Сроки
    signed_at       DATE,
    start_date      DATE,
    end_date        DATE,
    duration_days   INTEGER,                               -- продолжительность
    
    -- Обеспечение
    performance_bond_amount NUMERIC(18,2),
    performance_bond_pct    NUMERIC(5,2),
    warranty_period_days    INTEGER DEFAULT 730,           -- гарантийный срок, дней
    retention_pct           NUMERIC(5,2) DEFAULT 5.0,
    retention_release_days  INTEGER DEFAULT 365,
    
    -- Штрафы
    penalty_rate_daily      NUMERIC(5,3) DEFAULT 0.05,     -- % в день за просрочку
    penalty_max_pct         NUMERIC(5,2) DEFAULT 10.0,     -- макс % штрафа
    liquidated_damages      NUMERIC(18,2),                  -- неустойка
    
    -- Финансирование
    funding_source          VARCHAR(200),
    payment_terms           TEXT,                           -- условия оплаты
    payment_terms_type      VARCHAR(50) DEFAULT 'monthly',  -- monthly, milestone, advance, completion
    
    -- Документы
    document_path           VARCHAR(1000),
    
    notes                   TEXT,
    created_by              VARCHAR(100),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_client ON contracts(client_id);
CREATE INDEX idx_contracts_contractor ON contracts(contractor_id);
CREATE INDEX idx_contracts_project ON contracts(project_id);
CREATE INDEX idx_contracts_tender ON contracts(tender_id);
CREATE INDEX idx_contracts_dates ON contracts(start_date, end_date);

-- ============================================================================
-- 2. Этапы / milestones
-- ============================================================================
CREATE TABLE contract_milestones (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    milestone_number INTEGER NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    milestone_type  VARCHAR(50) NOT NULL DEFAULT 'payment', -- payment, delivery, completion, approval
    planned_date    DATE,
    actual_date     DATE,
    amount          NUMERIC(18,2),                         -- сумма этапа
    amount_pct      NUMERIC(5,2),                          -- % от суммы договора
    status          VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, completed, delayed, cancelled
    completion_pct  NUMERIC(5,2) DEFAULT 0,                 -- % выполнения
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contract_id, milestone_number)
);

CREATE INDEX idx_contract_milestones_contract ON contract_milestones(contract_id);
CREATE INDEX idx_contract_milestones_dates ON contract_milestones(planned_date);

-- ============================================================================
-- 3. Акты выполненных работ (КС-2 / КС-3)
-- ============================================================================
CREATE TABLE contract_work_acceptances (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    milestone_id    BIGINT REFERENCES contract_milestones(id),
    acceptance_number VARCHAR(50) NOT NULL,
    acceptance_date DATE NOT NULL,
    period_from     DATE,
    period_to       DATE,
    amount          NUMERIC(18,2) NOT NULL,                -- сумма акта
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',  -- draft, submitted, approved, paid, disputed
    approved_by     VARCHAR(200),
    approved_at     TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    payment_ref     VARCHAR(200),                          -- платёжное поручение
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contract_id, acceptance_number)
);

CREATE INDEX idx_contract_acceptances_contract ON contract_work_acceptances(contract_id);
CREATE INDEX idx_contract_acceptances_date ON contract_work_acceptances(acceptance_date);

-- ============================================================================
-- 4. Позиции акта (привязка к BOQ)
-- ============================================================================
CREATE TABLE contract_acceptance_items (
    id              BIGSERIAL PRIMARY KEY,
    acceptance_id   BIGINT NOT NULL REFERENCES contract_work_acceptances(id) ON DELETE CASCADE,
    boq_item_id     BIGINT REFERENCES boq_items(id),
    item_code       VARCHAR(100),
    description     TEXT,
    unit            VARCHAR(20),
    contract_quantity   NUMERIC(18,4),                     -- количество по договору
    prev_quantity       NUMERIC(18,4) DEFAULT 0,           -- выполнено ранее
    current_quantity    NUMERIC(18,4) NOT NULL,             -- выполнено в этом периоде
    total_quantity      NUMERIC(18,4),                     -- всего выполнено
    unit_price          NUMERIC(18,2) NOT NULL,
    current_amount      NUMERIC(18,2) NOT NULL,
    total_amount        NUMERIC(18,2),
    sort_order          INTEGER DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contract_acceptance_items_acc ON contract_acceptance_items(acceptance_id);

-- ============================================================================
-- 5. Дополнительные соглашения
-- ============================================================================
CREATE TABLE contract_addendums (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    addendum_number INTEGER NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    addendum_type   VARCHAR(50) NOT NULL DEFAULT 'variation', -- variation, extension, price_adjustment, termination
    amount_change   NUMERIC(18,2) DEFAULT 0,               -- изменение суммы (+/-)
    days_change     INTEGER DEFAULT 0,                      -- изменение сроков (+/-)
    new_end_date    DATE,
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',   -- draft, signed, rejected
    signed_at       DATE,
    document_path   VARCHAR(1000),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contract_id, addendum_number)
);

CREATE INDEX idx_contract_addendums_contract ON contract_addendums(contract_id);

-- ============================================================================
-- 6. Платежи
-- ============================================================================
CREATE TABLE contract_payments (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    acceptance_id   BIGINT REFERENCES contract_work_acceptances(id),
    milestone_id    BIGINT REFERENCES contract_milestones(id),
    payment_number  VARCHAR(50) NOT NULL,
    payment_date    DATE NOT NULL,
    amount          NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_type    VARCHAR(50) NOT NULL DEFAULT 'progress', -- advance, progress, milestone, retention, final
    payment_method  VARCHAR(50) DEFAULT 'bank_transfer',     -- bank_transfer, letter_of_credit, cash
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',   -- pending, processed, confirmed, rejected
    bank_ref        VARCHAR(200),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contract_payments_contract ON contract_payments(contract_id);
CREATE INDEX idx_contract_payments_date ON contract_payments(payment_date);

-- ============================================================================
-- 7. Претензии / Claims
-- ============================================================================
CREATE TABLE contract_claims (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    claim_number    VARCHAR(50) NOT NULL,
    claim_type      VARCHAR(50) NOT NULL,                  -- extension, additional_cost, delay_damages, quality
    description     TEXT NOT NULL,
    amount_claimed  NUMERIC(18,2),
    amount_approved NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'submitted', -- submitted, review, negotiation, approved, rejected, withdrawn
    submitted_by    VARCHAR(200),
    submitted_at    TIMESTAMPTZ,
    resolved_at     TIMESTAMPTZ,
    resolution      TEXT,
    document_path   VARCHAR(1000),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(contract_id, claim_number)
);

CREATE INDEX idx_contract_claims_contract ON contract_claims(contract_id);

-- ============================================================================
-- 8. История статусов
-- ============================================================================
CREATE TABLE contract_status_history (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    from_status     VARCHAR(50),
    to_status       VARCHAR(50) NOT NULL,
    changed_by      VARCHAR(100),
    reason          TEXT,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_contract_status_history_contract ON contract_status_history(contract_id);

-- ============================================================================
-- 9. Триггеры
-- ============================================================================
CREATE OR REPLACE FUNCTION update_contract_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_contract_updated
    BEFORE UPDATE ON contracts
    FOR EACH ROW
    EXECUTE FUNCTION update_contract_timestamp();

CREATE OR REPLACE FUNCTION log_contract_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO contract_status_history(contract_id, from_status, to_status, changed_by)
        VALUES (NEW.id, OLD.status, NEW.status, current_user);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_contract_status_log
    AFTER UPDATE OF status ON contracts
    FOR EACH ROW
    EXECUTE FUNCTION log_contract_status_change();

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE contracts IS 'Договоры / контракты';
COMMENT ON TABLE contract_milestones IS 'Этапы / milestones договора';
COMMENT ON TABLE contract_work_acceptances IS 'Акты выполненных работ (КС-2, КС-3)';
COMMENT ON TABLE contract_acceptance_items IS 'Позиции акта выполненных работ';
COMMENT ON TABLE contract_addendums IS 'Дополнительные соглашения';
COMMENT ON TABLE contract_payments IS 'Платежи по договору';
COMMENT ON TABLE contract_claims IS 'Претензии / claims';
COMMENT ON TABLE contract_status_history IS 'История статусов договора';
