-- ============================================================================
-- V025__EVM_Module.sql
-- Earned Value Management (EVM) — ANSI/EIA-748 compliant
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Контрольные счета (Control Accounts)
-- ============================================================================
CREATE TABLE evm_control_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ca_code         VARCHAR(100) NOT NULL,                  -- уникальный код CA
    ca_name         VARCHAR(500) NOT NULL,
    description     TEXT,
    wbs_code        VARCHAR(200),                            -- привязка к WBS
    responsible     VARCHAR(300),                            -- ответственный / CAM
    sort_order      INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, ca_code)
);

CREATE INDEX idx_evm_ca_project ON evm_control_accounts(project_id);
CREATE INDEX idx_evm_ca_wbs ON evm_control_accounts(wbs_code);

-- ============================================================================
-- 2. Базовые планы (Baselines)
-- ============================================================================
CREATE TABLE evm_baselines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    baseline_name   VARCHAR(500) NOT NULL,
    baseline_type   VARCHAR(50) NOT NULL DEFAULT 'target',   -- target, current, revised
    version         VARCHAR(50) NOT NULL DEFAULT '1.0',
    description     TEXT,
    is_approved     BOOLEAN DEFAULT FALSE,
    approved_by     VARCHAR(200),
    approved_at     TIMESTAMPTZ,
    is_active       BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, baseline_name)
);

CREATE INDEX idx_evm_bl_project ON evm_baselines(project_id);
CREATE INDEX idx_evm_bl_active ON evm_baselines(project_id, is_active) WHERE is_active = TRUE;

-- ============================================================================
-- 3. Плановые показатели по периодам (PV per period, P6-compatible)
-- ============================================================================
CREATE TABLE evm_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    control_account_id UUID REFERENCES evm_control_accounts(id),
    baseline_id     UUID REFERENCES evm_baselines(id),
    period_date     DATE NOT NULL,                           -- дата окончания периода
    period_type     VARCHAR(20) NOT NULL DEFAULT 'weekly',   -- weekly, monthly, quarterly
    planned_value   NUMERIC(18,2) DEFAULT 0,                 -- PV (BCWS) за период
    planned_hours   NUMERIC(12,2) DEFAULT 0,                 -- плановые часы
    planned_progress NUMERIC(5,2) DEFAULT 0,                 -- плановый % прогресса
    is_cumulative   BOOLEAN DEFAULT FALSE,                   -- накопленный итог
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, control_account_id, period_date, period_type)
);

CREATE INDEX idx_evm_periods_date ON evm_periods(project_id, period_date);
CREATE INDEX idx_evm_periods_ca ON evm_periods(control_account_id);

-- ============================================================================
-- 4. Фактические данные (Actuals — ACWP, hours, progress)
-- ============================================================================
CREATE TABLE evm_actuals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    control_account_id UUID REFERENCES evm_control_accounts(id),
    period_date     DATE NOT NULL,
    actual_cost     NUMERIC(18,2) DEFAULT 0,                 -- AC (ACWP)
    actual_hours    NUMERIC(12,2) DEFAULT 0,
    earned_value    NUMERIC(18,2) DEFAULT 0,                 -- EV (BCWP)
    progress_pct    NUMERIC(5,2) DEFAULT 0,                  -- фактический % завершения
    physical_pct    NUMERIC(5,2),                             -- physical % complete
    data_source     VARCHAR(100) DEFAULT 'manual',           -- manual, p6_sync, timesheet
    source_id       VARCHAR(255),                            -- ID из внешней системы
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, control_account_id, period_date)
);

CREATE INDEX idx_evm_actuals_date ON evm_actuals(project_id, period_date);
CREATE INDEX idx_evm_actuals_ca ON evm_actuals(control_account_id);

-- ============================================================================
-- 5. Расчётные метрики (PV, EV, AC, SV, CV, SPI, CPI, EAC, ETC, TCPI)
-- ============================================================================
CREATE TABLE evm_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    control_account_id UUID REFERENCES evm_control_accounts(id),
    period_date     DATE NOT NULL,
    -- Плановые
    pv              NUMERIC(18,2) DEFAULT 0,                 -- Planned Value (BCWS)
    ev              NUMERIC(18,2) DEFAULT 0,                 -- Earned Value (BCWP)
    ac              NUMERIC(18,2) DEFAULT 0,                 -- Actual Cost (ACWP)
    bac             NUMERIC(18,2) DEFAULT 0,                 -- Budget at Completion
    -- Отклонения
    sv              NUMERIC(18,2) DEFAULT 0,                 -- Schedule Variance (EV - PV)
    cv              NUMERIC(18,2) DEFAULT 0,                 -- Cost Variance (EV - AC)
    sv_pct          NUMERIC(8,2) DEFAULT 0,                  -- SV%
    cv_pct          NUMERIC(8,2) DEFAULT 0,                  -- CV%
    -- Индексы
    spi             NUMERIC(8,4) DEFAULT 0,                  -- Schedule Performance Index
    cpi             NUMERIC(8,4) DEFAULT 0,                  -- Cost Performance Index
    -- Прогнозы
    eac             NUMERIC(18,2) DEFAULT 0,                 -- Estimate at Completion
    etc             NUMERIC(18,2) DEFAULT 0,                 -- Estimate to Complete
    vac             NUMERIC(18,2) DEFAULT 0,                 -- Variance at Completion (BAC - EAC)
    tcpi            NUMERIC(8,4) DEFAULT 0,                  -- To-Complete Performance Index
    -- Агрегация
    metric_scope    VARCHAR(20) NOT NULL DEFAULT 'project',  -- project, ca, wbs
    is_cumulative   BOOLEAN DEFAULT FALSE,
    calculated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, control_account_id, period_date, metric_scope)
);

CREATE INDEX idx_evm_metrics_date ON evm_metrics(project_id, period_date);
CREATE INDEX idx_evm_metrics_ca ON evm_metrics(control_account_id);

-- ============================================================================
-- 6. Прогнозы (Forecasts — EAC, ETC, VAC)
-- ============================================================================
CREATE TABLE evm_forecasts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    control_account_id UUID REFERENCES evm_control_accounts(id),
    forecast_date   DATE NOT NULL,
    forecast_type   VARCHAR(50) NOT NULL DEFAULT 'eac',      -- eac, etc, vac, completion_date
    method          VARCHAR(50) NOT NULL DEFAULT 'cpi',      -- cpi, spi, composite, management, p6
    eac_value       NUMERIC(18,2),
    etc_value       NUMERIC(18,2),
    vac_value       NUMERIC(18,2),
    completion_date DATE,
    confidence_pct  NUMERIC(5,2),                            -- уровень уверенности
    notes           TEXT,
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evm_forecast_project ON evm_forecasts(project_id);
CREATE INDEX idx_evm_forecast_date ON evm_forecasts(forecast_date);

-- ============================================================================
-- 7. Правила освоения (Earned Rules)
-- ============================================================================
CREATE TABLE evm_earned_rules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    rule_name       VARCHAR(200) NOT NULL,
    rule_type       VARCHAR(50) NOT NULL,                    -- 0/100, 50/50, percent_complete, physical, custom
    description     TEXT,
    weight_pct      NUMERIC(5,2) DEFAULT 100,               -- вес правила
    config          JSONB,                                   -- доп. конфигурация
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, rule_name)
);

-- ============================================================================
-- 8. Привязка EVM к проектам
-- ============================================================================
CREATE TABLE evm_projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL UNIQUE REFERENCES projects(id) ON DELETE CASCADE,
    evm_enabled     BOOLEAN DEFAULT FALSE,
    default_baseline_id UUID REFERENCES evm_baselines(id),
    reporting_freq  VARCHAR(20) DEFAULT 'weekly',            -- weekly, monthly
    currency        VARCHAR(3) DEFAULT 'USD',
    threshold_spi   NUMERIC(5,2) DEFAULT 0.8,               -- порог SPI
    threshold_cpi   NUMERIC(5,2) DEFAULT 0.8,               -- порог CPI
    threshold_sv_pct NUMERIC(5,2) DEFAULT -10,               -- порог SV%
    threshold_cv_pct NUMERIC(5,2) DEFAULT -10,               -- порог CV%
    config          JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- Функция расчёта метрик EVM (вызывается триггером или вручную)
-- ============================================================================
CREATE OR REPLACE FUNCTION calculate_evm_metrics(
    p_project_id UUID,
    p_period_date DATE
) RETURNS UUID AS $$
DECLARE
    v_bac NUMERIC(18,2);
    v_pv NUMERIC(18,2);
    v_ev NUMERIC(18,2);
    v_ac NUMERIC(18,2);
    v_sv NUMERIC(18,2);
    v_cv NUMERIC(18,2);
    v_sv_pct NUMERIC(8,2);
    v_cv_pct NUMERIC(8,2);
    v_spi NUMERIC(8,4);
    v_cpi NUMERIC(8,4);
    v_eac NUMERIC(18,2);
    v_etc NUMERIC(18,2);
    v_vac NUMERIC(18,2);
    v_tcpi NUMERIC(8,4);
    v_metric_id UUID;
BEGIN
    -- BAC — сумма всех planned_value из baselines
    SELECT COALESCE(SUM(planned_value), 0) INTO v_bac
    FROM evm_periods WHERE project_id = p_project_id AND is_cumulative = FALSE;

    -- PV — cumulative planned_value до даты
    SELECT COALESCE(SUM(planned_value), 0) INTO v_pv
    FROM evm_periods WHERE project_id = p_project_id AND period_date <= p_period_date;

    -- EV — cumulative earned value
    SELECT COALESCE(SUM(earned_value), 0) INTO v_ev
    FROM evm_actuals WHERE project_id = p_project_id AND period_date <= p_period_date;

    -- AC — cumulative actual cost
    SELECT COALESCE(SUM(actual_cost), 0) INTO v_ac
    FROM evm_actuals WHERE project_id = p_project_id AND period_date <= p_period_date;

    -- Расчёт отклонений
    v_sv := v_ev - v_pv;
    v_cv := v_ev - v_ac;
    v_sv_pct := CASE WHEN v_pv > 0 THEN (v_sv / v_pv) * 100 ELSE 0 END;
    v_cv_pct := CASE WHEN v_ac > 0 THEN (v_cv / v_ac) * 100 ELSE 0 END;

    -- Индексы
    v_spi := CASE WHEN v_pv > 0 THEN v_ev / v_pv ELSE 1.0 END;
    v_cpi := CASE WHEN v_ac > 0 THEN v_ev / v_ac ELSE 1.0 END;

    -- Прогнозы (EAC = BAC / CPI)
    v_eac := CASE WHEN v_cpi > 0 THEN v_bac / v_cpi ELSE v_bac END;
    v_etc := v_eac - v_ac;
    v_vac := v_bac - v_eac;
    v_tcpi := CASE WHEN (v_bac - v_ev) > 0 THEN (v_bac - v_ev) / (v_bac - v_ac) ELSE 1.0 END;

    -- Upsert метрики
    INSERT INTO evm_metrics (
        project_id, period_date, pv, ev, ac, bac,
        sv, cv, sv_pct, cv_pct, spi, cpi,
        eac, etc, vac, tcpi, metric_scope, is_cumulative
    ) VALUES (
        p_project_id, p_period_date, v_pv, v_ev, v_ac, v_bac,
        v_sv, v_cv, v_sv_pct, v_cv_pct, v_spi, v_cpi,
        v_eac, v_etc, v_vac, v_tcpi, 'project', TRUE
    )
    ON CONFLICT (project_id, control_account_id, period_date, metric_scope)
    WHERE metric_scope = 'project' AND control_account_id IS NULL
    DO UPDATE SET
        pv = EXCLUDED.pv, ev = EXCLUDED.ev, ac = EXCLUDED.ac, bac = EXCLUDED.bac,
        sv = EXCLUDED.sv, cv = EXCLUDED.cv, sv_pct = EXCLUDED.sv_pct, cv_pct = EXCLUDED.cv_pct,
        spi = EXCLUDED.spi, cpi = EXCLUDED.cpi, eac = EXCLUDED.eac, etc = EXCLUDED.etc,
        vac = EXCLUDED.vac, tcpi = EXCLUDED.tcpi, calculated_at = NOW()
    RETURNING id INTO v_metric_id;

    RETURN v_metric_id;
END;
$$ LANGUAGE plpgsql;