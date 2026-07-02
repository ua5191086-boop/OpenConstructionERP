-- ============================================================================
-- V044__Reporting_Builder.sql
-- Кастомный конструктор отчётов
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Шаблоны отчётов
-- ============================================================================
CREATE TABLE report_templates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(500) NOT NULL,
    description         TEXT,
    category            VARCHAR(100) NOT NULL,                  -- financial, progress, quality, hse, tunnel, schedule, procurement
    report_type         VARCHAR(100) NOT NULL,                  -- table, chart, cross_tab, summary, pivot
    data_source         VARCHAR(100) NOT NULL,                  -- sql, view, api, csv_upload
    query_text          TEXT,                                    -- SQL запрос (для sql/data_source)
    parameters          JSONB DEFAULT '[]'::JSONB,              -- параметры отчёта: [{"name":"project_id","type":"uuid","label":"Проект"}]
    columns_config      JSONB DEFAULT '[]'::JSONB,              -- колонки: [{"field":"name","label":"Название","type":"string","width":200}]
    chart_config        JSONB,                                  -- настройки графика: {type:"bar",x:"date",y:"amount"}
    aggregation         JSONB,                                  -- {group_by:["project"],aggregates:[{field:"amount",func:"sum"}]}
    filters             JSONB DEFAULT '[]'::JSONB,              -- фильтры по умолчанию
    sort_config         JSONB,                                  -- {field:"date",direction:"desc"}
    export_formats      JSONB DEFAULT '["csv","xlsx","pdf"]'::JSONB,
    is_system           BOOLEAN DEFAULT FALSE,                  -- системный шаблон (не редактируется)
    is_public           BOOLEAN DEFAULT FALSE,                  -- доступен всем пользователям
    owner_id            VARCHAR(200),                            -- владелец
    version             INTEGER DEFAULT 1,
    last_run_at         TIMESTAMPTZ,
    last_run_by         VARCHAR(200),
    run_count           INTEGER DEFAULT 0,
    avg_exec_time_ms    NUMERIC(10,2),
    status              VARCHAR(50) DEFAULT 'active',           -- active, archived, disabled
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_templates_category ON report_templates(category);
CREATE INDEX idx_report_templates_type ON report_templates(report_type);
CREATE INDEX idx_report_templates_owner ON report_templates(owner_id);
CREATE INDEX idx_report_templates_status ON report_templates(status);

COMMENT ON TABLE report_templates IS 'Шаблоны кастомных отчётов';

-- ============================================================================
-- 2. Сохранённые отчёты (с параметрами)
-- ============================================================================
CREATE TABLE saved_reports (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id         UUID NOT NULL REFERENCES report_templates(id) ON DELETE CASCADE,
    name                VARCHAR(500) NOT NULL,
    description         TEXT,
    parameter_values    JSONB DEFAULT '{}'::JSONB,              -- значения параметров
    schedule            JSONB,                                  -- расписание: {cron:"0 8 * * 1",enabled:true,last_sent:...}
    recipients          JSONB DEFAULT '[]'::JSONB,              -- получатели: [{"email":"...","telegram":"..."}]
    output_format       VARCHAR(50) DEFAULT 'pdf',              -- pdf, csv, xlsx, html
    last_generated_at   TIMESTAMPTZ,
    last_generated_by   VARCHAR(200),
    is_favorite         BOOLEAN DEFAULT FALSE,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_saved_reports_template ON saved_reports(template_id);
CREATE INDEX idx_saved_reports_fav ON saved_reports(is_favorite) WHERE is_favorite = TRUE;

COMMENT ON TABLE saved_reports IS 'Сохранённые отчёты с фиксированными параметрами';

-- ============================================================================
-- 3. История запусков отчётов
-- ============================================================================
CREATE TABLE report_execution_log (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id         UUID NOT NULL REFERENCES report_templates(id) ON DELETE CASCADE,
    saved_report_id     UUID REFERENCES saved_reports(id) ON DELETE SET NULL,
    parameter_values    JSONB,
    row_count           INTEGER,
    exec_time_ms        NUMERIC(10,2),
    result_size_bytes   BIGINT,
    output_format       VARCHAR(50),
    output_path         VARCHAR(500),
    triggered_by        VARCHAR(200),                           -- manual, schedule, api
    triggered_by_user   VARCHAR(200),
    status              VARCHAR(50) DEFAULT 'running',          -- running, completed, failed, cancelled
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_report_log_template ON report_execution_log(template_id);
CREATE INDEX idx_report_log_status ON report_execution_log(status);
CREATE INDEX idx_report_log_created ON report_execution_log(created_at);

COMMENT ON TABLE report_execution_log IS 'Лог выполнения отчётов';

-- ============================================================================
-- 4. Пользовательские дашборды
-- ============================================================================
CREATE TABLE custom_dashboards (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    name                VARCHAR(500) NOT NULL,
    description         TEXT,
    layout_config       JSONB NOT NULL DEFAULT '[]'::JSONB,     -- виджеты: [{"type":"chart","report_id":"...","x":0,"y":0,"w":6,"h":4}]
    is_default          BOOLEAN DEFAULT FALSE,
    is_public           BOOLEAN DEFAULT FALSE,
    owner_id            VARCHAR(200),
    auto_refresh_sec    INTEGER DEFAULT 0,                      -- 0 = нет автообновления
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_custom_dashboards_project ON custom_dashboards(project_id);
CREATE INDEX idx_custom_dashboards_owner ON custom_dashboards(owner_id);

COMMENT ON TABLE custom_dashboards IS 'Пользовательские дашборды с настраиваемой сеткой виджетов';