-- ============================================================================
-- V046__Performance_Benchmarking.sql
-- Эталонное сравнение производительности проектов
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Бенчмарки по типам проектов
-- ============================================================================
CREATE TABLE benchmark_templates (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(500) NOT NULL,
    description         TEXT,
    project_type        VARCHAR(100) NOT NULL,                  -- tunnel, building, road, bridge, industrial
    region              VARCHAR(100),                           -- global, europe, asia, cis, custom
    source              VARCHAR(200),                           -- industry_report, historical, custom, consultant
    reliability_score   NUMERIC(5,2),                           -- достоверность данных 0-100
    is_public           BOOLEAN DEFAULT TRUE,
    version             VARCHAR(20) DEFAULT '1.0',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_benchmark_templates_type ON benchmark_templates(project_type);

COMMENT ON TABLE benchmark_templates IS 'Шаблоны эталонных показателей по типам проектов';

-- ============================================================================
-- 2. Эталонные показатели (KPI benchmarks)
-- ============================================================================
CREATE TABLE benchmark_kpis (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id         UUID NOT NULL REFERENCES benchmark_templates(id) ON DELETE CASCADE,
    kpi_code            VARCHAR(100) NOT NULL,
    kpi_name            VARCHAR(500) NOT NULL,
    category            VARCHAR(100) NOT NULL,                  -- cost, schedule, productivity, safety, quality, sustainability
    unit                VARCHAR(50),
    p10_value           NUMERIC(18,4),                          -- лучший квартиль (top 10%)
    p25_value           NUMERIC(18,4),                          -- верхний квартиль
    p50_value           NUMERIC(18,4),                          -- медиана (типичное значение)
    p75_value           NUMERIC(18,4),                          -- нижний квартиль
    p90_value           NUMERIC(18,4),                          -- худший квартиль (bottom 10%)
    mean_value          NUMERIC(18,4),                          -- среднее
    std_dev             NUMERIC(18,4),                          -- стандартное отклонение
    sample_size         INTEGER,                                -- количество проектов в выборке
    period_from         DATE,
    period_to           DATE,
    formula_desc        TEXT,                                   -- описание формулы расчёта
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_benchmark_kpis_template ON benchmark_kpis(template_id);
CREATE INDEX idx_benchmark_kpis_category ON benchmark_kpis(category);
CREATE INDEX idx_benchmark_kpis_code ON benchmark_kpis(kpi_code);

COMMENT ON TABLE benchmark_kpis IS 'Эталонные значения KPI для сравнительного анализа';

-- ============================================================================
-- 3. Результаты сравнения проекта с бенчмарками
-- ============================================================================
CREATE TABLE benchmark_results (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    benchmark_kpi_id    UUID NOT NULL REFERENCES benchmark_kpis(id) ON DELETE CASCADE,
    template_id         UUID NOT NULL REFERENCES benchmark_templates(id) ON DELETE CASCADE,
    project_value       NUMERIC(18,4),                          -- фактическое значение проекта
    benchmark_value     NUMERIC(18,4),                          -- эталон (обычно P50)
    variance_pct        NUMERIC(8,2),                           -- отклонение в %
    percentile_rank     NUMERIC(5,2),                           -- процентиль проекта в распределении
    rating              VARCHAR(50),                            -- excellent, good, average, below_average, poor
    assessment_date     DATE NOT NULL,
    assessed_by         VARCHAR(200),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_benchmark_results_project ON benchmark_results(project_id);
CREATE INDEX idx_benchmark_results_kpi ON benchmark_results(benchmark_kpi_id);
CREATE INDEX idx_benchmark_results_date ON benchmark_results(assessment_date);

COMMENT ON TABLE benchmark_results IS 'Результаты сравнения проекта с отраслевыми бенчмарками';

-- ============================================================================
-- 4. Сводка бенчмарков по проекту (materialized view)
-- ============================================================================
CREATE MATERIALIZED VIEW project_benchmark_summary AS
SELECT
    br.project_id,
    p.code AS project_code,
    p.name AS project_name,
    bt.name AS benchmark_name,
    bt.project_type,
    bk.category,
    bk.kpi_code,
    bk.kpi_name,
    bk.unit,
    br.project_value,
    br.benchmark_value,
    br.variance_pct,
    br.percentile_rank,
    br.rating,
    bk.p10_value,
    bk.p50_value,
    bk.p90_value,
    br.assessment_date
FROM benchmark_results br
JOIN benchmark_kpis bk ON bk.id = br.benchmark_kpi_id
JOIN benchmark_templates bt ON bt.id = br.template_id
JOIN projects p ON p.id = br.project_id;

CREATE UNIQUE INDEX idx_pbs_unique ON project_benchmark_summary(project_id, kpi_code);

COMMENT ON MATERIALIZED VIEW project_benchmark_summary IS 'Сводка сравнения KPI проекта с отраслевыми эталонами';