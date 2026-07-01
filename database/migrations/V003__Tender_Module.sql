-- ============================================================================
-- V003__Tender_Module.sql
-- Модуль управления тендерами (Tender Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================
-- Зависимости: V001__BOQ_Module.sql (ссылается на boq_items, sections, contractors)
-- ============================================================================

-- ============================================================================
-- 1. Тендеры
-- ============================================================================
CREATE TABLE tenders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(50) NOT NULL UNIQUE,          -- T-2026-001
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    tender_type     VARCHAR(50) NOT NULL DEFAULT 'open',   -- open, limited, single_source, request_quote
    status          VARCHAR(50) NOT NULL DEFAULT 'draft', -- draft, published, in_progress, evaluation, awarded, cancelled, completed
    client_id       UUID REFERENCES organizations(id),
    project_id      UUID REFERENCES projects(id),
    budget_amount   NUMERIC(18,2),                        -- сметная стоимость
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    published_at    TIMESTAMPTZ,
    submission_deadline TIMESTAMPTZ,
    bid_open_date   TIMESTAMPTZ,
    award_date      TIMESTAMPTZ,
    contract_start  DATE,
    contract_end    DATE,
    bid_bond_pct    NUMERIC(5,2) DEFAULT 2.5,             -- обеспечение заявки, %
    performance_bond_pct NUMERIC(5,2) DEFAULT 10.0,      -- обеспечение исполнения, %
    advance_payment_pct  NUMERIC(5,2),                   -- аванс, %
    retention_pct   NUMERIC(5,2) DEFAULT 5.0,             -- гарантийное удержание, %
    retention_release_days INTEGER DEFAULT 365,           -- дней после завершения
    procurement_method VARCHAR(100),                      -- e-auction, competitive, negotiated
    funding_source  VARCHAR(200),                          -- бюджет, грант, частные
    notes           TEXT,
    created_by      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenders_status ON tenders(status);
CREATE INDEX IF NOT EXISTS idx_tenders_client ON tenders(client_id);
CREATE INDEX IF NOT EXISTS idx_tenders_project ON tenders(project_id);
CREATE INDEX IF NOT EXISTS idx_tenders_deadline ON tenders(submission_deadline);

-- ============================================================================
-- 2. Лоты тендера
-- ============================================================================
CREATE TABLE tender_lots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    lot_number      INTEGER NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    estimated_amount NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    section_id      UUID REFERENCES boq_sections(id),       -- участок строительства
    status          VARCHAR(50) NOT NULL DEFAULT 'active', -- active, awarded, cancelled
    award_decision  TEXT,                                  -- обоснование выбора победителя
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tender_id, lot_number)
);

CREATE INDEX IF NOT EXISTS idx_tender_lots_tender ON tender_lots(tender_id);

-- ============================================================================
-- 3. Позиции лота (привязка к BOQ)
-- ============================================================================
CREATE TABLE tender_lot_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id          UUID NOT NULL REFERENCES tender_lots(id) ON DELETE CASCADE,
    boq_item_id     UUID REFERENCES boq_items(id),
    item_code       VARCHAR(100),
    description     TEXT,
    unit            VARCHAR(20),
    quantity        NUMERIC(18,4),
    estimated_unit_price NUMERIC(18,2),
    estimated_total NUMERIC(18,2),
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tender_lot_items_lot ON tender_lot_items(lot_id);

-- ============================================================================
-- 4. Участники тендера
-- ============================================================================
CREATE TABLE tender_bidders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    lot_id          UUID REFERENCES tender_lots(id) ON DELETE CASCADE,
    contractor_id   UUID NOT NULL REFERENCES organizations(id),
    bid_number      VARCHAR(50),                          -- номер заявки
    status          VARCHAR(50) NOT NULL DEFAULT 'submitted', -- submitted, qualified, disqualified, withdrawn, winner, reserve
    bid_amount      NUMERIC(18,2),                        -- общая сумма заявки
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    bid_bond_amount NUMERIC(18,2),                        -- сумма обеспечения
    validity_days   INTEGER DEFAULT 90,                   -- срок действия заявки
    submission_date TIMESTAMPTZ,
    is_winner       BOOLEAN DEFAULT FALSE,
    award_amount    NUMERIC(18,2),                        -- сумма контракта (может отличаться от заявки)
    award_reason    TEXT,                                  -- обоснование присуждения
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tender_id, lot_id, contractor_id)
);

CREATE INDEX IF NOT EXISTS idx_tender_bidders_tender ON tender_bidders(tender_id);
CREATE INDEX IF NOT EXISTS idx_tender_bidders_lot ON tender_bidders(lot_id);
CREATE INDEX IF NOT EXISTS idx_tender_bidders_contractor ON tender_bidders(contractor_id);
CREATE INDEX IF NOT EXISTS idx_tender_bidders_winner ON tender_bidders(tender_id, is_winner) WHERE is_winner = TRUE;

-- ============================================================================
-- 5. Ценовые предложения участников (по позициям)
-- ============================================================================
CREATE TABLE tender_bid_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bidder_id       UUID NOT NULL REFERENCES tender_bidders(id) ON DELETE CASCADE,
    lot_item_id     UUID REFERENCES tender_lot_items(id) ON DELETE CASCADE,
    boq_item_id     UUID REFERENCES boq_items(id),
    item_code       VARCHAR(100),
    description     TEXT,
    unit            VARCHAR(20),
    quantity        NUMERIC(18,4),
    unit_price      NUMERIC(18,2),
    total_price     NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_alternative  BOOLEAN DEFAULT FALSE,                -- альтернативное предложение
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tender_bid_items_bidder ON tender_bid_items(bidder_id);

-- ============================================================================
-- 6. Оценка заявок
-- ============================================================================
CREATE TABLE tender_evaluations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    lot_id          UUID REFERENCES tender_lots(id) ON DELETE CASCADE,
    bidder_id       UUID NOT NULL REFERENCES tender_bidders(id) ON DELETE CASCADE,
    evaluator       VARCHAR(200),
    evaluation_type VARCHAR(50) NOT NULL DEFAULT 'technical', -- technical, financial, combined
    score           NUMERIC(5,2),                          -- оценка в баллах
    max_score       NUMERIC(5,2) DEFAULT 100.00,
    weight_pct      NUMERIC(5,2) DEFAULT 100.00,          -- вес критерия, %
    comments        TEXT,
    evaluated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tender_evaluations_tender ON tender_evaluations(tender_id);
CREATE INDEX IF NOT EXISTS idx_tender_evaluations_bidder ON tender_evaluations(bidder_id);

-- ============================================================================
-- 7. Критерии оценки
-- ============================================================================
CREATE TABLE tender_evaluation_criteria (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    weight_pct      NUMERIC(5,2) NOT NULL,                -- вес критерия, %
    max_score       NUMERIC(5,2) NOT NULL DEFAULT 100.00,
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tender_criteria_tender ON tender_evaluation_criteria(tender_id);

-- ============================================================================
-- 8. Документы тендера
-- ============================================================================
CREATE TABLE tender_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    document_type   VARCHAR(100) NOT NULL,                 -- rfp, addendum, bid, clarification, contract
    title           VARCHAR(500) NOT NULL,
    file_path       VARCHAR(1000),
    file_size       UUID,
    mime_type       VARCHAR(100),
    version         INTEGER DEFAULT 1,
    uploaded_by     VARCHAR(100),
    uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tender_documents_tender ON tender_documents(tender_id);

-- ============================================================================
-- 9. Вопросы и разъяснения
-- ============================================================================
CREATE TABLE tender_clarifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    lot_id          UUID REFERENCES tender_lots(id) ON DELETE CASCADE,
    bidder_id       UUID REFERENCES tender_bidders(id),
    question        TEXT NOT NULL,
    answer          TEXT,
    is_public       BOOLEAN DEFAULT TRUE,                  -- публичный ответ (всем участникам)
    asked_by        VARCHAR(200),
    answered_by     VARCHAR(200),
    asked_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_tender_clarifications_tender ON tender_clarifications(tender_id);

-- ============================================================================
-- 10. История статусов
-- ============================================================================
CREATE TABLE tender_status_history (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tender_id       UUID NOT NULL REFERENCES tenders(id) ON DELETE CASCADE,
    from_status     VARCHAR(50),
    to_status       VARCHAR(50) NOT NULL,
    changed_by      VARCHAR(100),
    reason          TEXT,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tender_status_history_tender ON tender_status_history(tender_id);

-- ============================================================================
-- 11. Вспомогательные функции
-- ============================================================================

-- Обновление updated_at
CREATE OR REPLACE FUNCTION update_tender_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tender_updated
    BEFORE UPDATE ON tenders
    FOR EACH ROW
    EXECUTE FUNCTION update_tender_timestamp();

-- Логирование смены статуса
CREATE OR REPLACE FUNCTION log_tender_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO tender_status_history(tender_id, from_status, to_status, changed_by)
        VALUES (NEW.id, OLD.status, NEW.status, current_user);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_tender_status_log
    AFTER UPDATE OF status ON tenders
    FOR EACH ROW
    EXECUTE FUNCTION log_tender_status_change();

-- ============================================================================
-- Комментарии к таблицам
-- ============================================================================
COMMENT ON TABLE tenders IS 'Тендеры / закупочные процедуры';
COMMENT ON TABLE tender_lots IS 'Лоты тендера';
COMMENT ON TABLE tender_lot_items IS 'Позиции лота (привязка к BOQ)';
COMMENT ON TABLE tender_bidders IS 'Участники тендера';
COMMENT ON TABLE tender_bid_items IS 'Ценовые предложения по позициям';
COMMENT ON TABLE tender_evaluations IS 'Оценка заявок';
COMMENT ON TABLE tender_evaluation_criteria IS 'Критерии оценки';
COMMENT ON TABLE tender_documents IS 'Документы тендера';
COMMENT ON TABLE tender_clarifications IS 'Вопросы и разъяснения';
COMMENT ON TABLE tender_status_history IS 'История изменения статусов';
