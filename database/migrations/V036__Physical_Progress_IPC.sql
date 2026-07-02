-- ============================================================================
-- V033__Physical_Progress_IPC.sql
-- Физический прогресс по BOQ-позициям и IPC (Interim Payment Certificate)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Physical Progress — замеры физического прогресса с площадки
-- ============================================================================
CREATE TABLE physical_progress (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    contract_id         UUID REFERENCES contracts(id) ON DELETE SET NULL,
    boq_item_id         UUID REFERENCES boq_items(id) ON DELETE SET NULL,
    measurement_date    DATE NOT NULL,
    item_code           VARCHAR(100) NOT NULL,
    description         TEXT,
    unit                VARCHAR(20) NOT NULL,
    contract_quantity   NUMERIC(18,4) NOT NULL,               -- количество по договору
    prev_cumulative_qty NUMERIC(18,4) DEFAULT 0,             -- накоплено до этого замера
    current_qty         NUMERIC(18,4) NOT NULL,               -- выполнено в этом периоде
    total_cumulative_qty NUMERIC(18,4),                       -- всего выполнено (prev + current)
    completion_pct      NUMERIC(5,2) DEFAULT 0,               -- % выполнения
    unit_price          NUMERIC(18,2) NOT NULL,               -- единичная расценка
    ipc_amount          NUMERIC(18,2) DEFAULT 0,              -- сумма IPC за этот период (qty * price)
    source              VARCHAR(50) DEFAULT 'site_measurement', -- site_measurement, survey, bim, ai_estimate, manual
    verified_by         VARCHAR(200),
    verified_at         TIMESTAMPTZ,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_physical_progress_project ON physical_progress(project_id);
CREATE INDEX idx_physical_progress_contract ON physical_progress(contract_id);
CREATE INDEX idx_physical_progress_boq ON physical_progress(boq_item_id);
CREATE INDEX idx_physical_progress_date ON physical_progress(measurement_date);
CREATE INDEX idx_physical_progress_item ON physical_progress(item_code);

COMMENT ON TABLE physical_progress IS 'Физический прогресс по BOQ-позициям — замеры с площадки, survey, AI';

-- ============================================================================
-- 2. View: IPC vs Physical Progress по контрактам
-- ============================================================================
CREATE OR REPLACE VIEW ipc_vs_physical AS
SELECT
    pp.project_id,
    pp.contract_id,
    c.code AS contract_code,
    c.name AS contract_name,
    c.contract_amount,
    COUNT(DISTINCT pp.id) AS progress_entries,
    COALESCE(SUM(pp.ipc_amount), 0) AS total_ipc_from_progress,
    COALESCE(SUM(a.amount), 0) AS total_accepted_ipc,
    COALESCE(SUM(a.amount), 0) - COALESCE(SUM(pp.ipc_amount), 0) AS ipc_variance,
    CASE
        WHEN c.contract_amount > 0
        THEN (COALESCE(SUM(a.amount), 0) / c.contract_amount) * 100
        ELSE 0
    END AS ipc_pct_of_contract,
    CASE
        WHEN SUM(pp.contract_quantity) > 0
        THEN SUM(pp.completion_pct * pp.contract_quantity) / SUM(pp.contract_quantity)
        ELSE 0
    END AS avg_physical_completion_pct
FROM physical_progress pp
LEFT JOIN contracts c ON c.id = pp.contract_id
LEFT JOIN contract_work_acceptances a ON a.contract_id = pp.contract_id AND a.status IN ('approved', 'paid')
GROUP BY pp.project_id, pp.contract_id, c.code, c.name, c.contract_amount;

COMMENT ON VIEW ipc_vs_physical IS 'Сравнение IPC (акты КС-2/КС-3) и физического прогресса по замерам';

-- ============================================================================
-- 3. View: Сводка по прогрессу по проекту
-- ============================================================================
CREATE OR REPLACE VIEW project_progress_summary AS
SELECT
    p.id AS project_id,
    p.code AS project_code,
    p.name AS project_name,
    COUNT(DISTINCT pp.contract_id) AS contracts_with_progress,
    COALESCE(SUM(pp.ipc_amount), 0) AS total_physical_ipc,
    COALESCE(SUM(a.amount), 0) AS total_accepted_ipc,
    CASE
        WHEN SUM(pp.contract_quantity) > 0
        THEN ROUND((SUM(pp.completion_pct * pp.contract_quantity) / SUM(pp.contract_quantity))::numeric, 2)
        ELSE 0
    END AS weighted_physical_progress,
    MAX(pp.measurement_date) AS last_measurement_date
FROM projects p
LEFT JOIN physical_progress pp ON pp.project_id = p.id
LEFT JOIN contract_work_acceptances a ON a.contract_id IN (
    SELECT id FROM contracts WHERE project_id = p.id
) AND a.status IN ('approved', 'paid')
GROUP BY p.id, p.code, p.name;

COMMENT ON VIEW project_progress_summary IS 'Сводка прогресса по проекту — взвешенный физический % vs акты';

-- ============================================================================
-- 4. Register in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('physical_progress', 'Physical Progress', 'bar-chart-2', 'FINANCE')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON COLUMN physical_progress.unit_price IS 'Единичная расценка по BOQ';
COMMENT ON COLUMN physical_progress.ipc_amount IS 'Сумма Interim Payment Certificate за период';
COMMENT ON COLUMN physical_progress.source IS 'Источник данных: site_measurement, survey, bim, ai_estimate';