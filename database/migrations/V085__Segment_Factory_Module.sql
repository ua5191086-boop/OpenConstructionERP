-- ============================================================================
-- V036__Segment_Factory_Module.sql
-- Модуль Segment Factory — производство колец и сегментов
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Производственные линии / формы
-- ============================================================================
CREATE TABLE segment_factory_lines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    line_code       VARCHAR(30) NOT NULL,
    line_name       VARCHAR(300) NOT NULL,
    line_type       VARCHAR(50) NOT NULL DEFAULT 'carousel'
        CHECK (line_type IN ('carousel','stationary','battery_mould','tunnel_form','other')),
    capacity_per_day INTEGER NOT NULL DEFAULT 4,              -- колец/сегментов в сутки
    mould_count     INTEGER DEFAULT 1,
    curing_method   VARCHAR(50) DEFAULT 'steam'
        CHECK (curing_method IN ('steam','water','air','accelerated','natural')),
    curing_hours    NUMERIC(5,1) DEFAULT 12.0,
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','maintenance','idle','decommissioned')),
    location        VARCHAR(300),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, line_code)
);

CREATE INDEX idx_sf_lines_project ON segment_factory_lines(project_id);
COMMENT ON TABLE segment_factory_lines IS 'Производственные линии для сегментов';

-- ============================================================================
-- 2. План производства
-- ============================================================================
CREATE TABLE segment_production_plans (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    plan_code       VARCHAR(30) NOT NULL,
    plan_name       VARCHAR(300) NOT NULL,
    plan_date       DATE NOT NULL,
    line_id         UUID REFERENCES segment_factory_lines(id) ON DELETE SET NULL,
    ring_type       VARCHAR(50) NOT NULL DEFAULT 'standard'
        CHECK (ring_type IN ('standard','left','right','key','closure','transition','special')),
    concrete_grade  VARCHAR(30) DEFAULT 'C50/60',
    segment_count   INTEGER NOT NULL DEFAULT 7,               -- сегментов в кольце
    planned_rings   INTEGER NOT NULL,
    planned_segments INTEGER GENERATED ALWAYS AS (planned_rings * segment_count) STORED,
    produced_rings  INTEGER DEFAULT 0,
    rejected_rings  INTEGER DEFAULT 0,
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','paused','completed','cancelled')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, plan_code)
);

CREATE INDEX idx_sf_plans_project ON segment_production_plans(project_id);
CREATE INDEX idx_sf_plans_line ON segment_production_plans(line_id);
COMMENT ON TABLE segment_production_plans IS 'Планы производства колец и сегментов';

-- ============================================================================
-- 3. Производственные циклы (замес → формовка → термообработка → склад)
-- ============================================================================
CREATE TABLE segment_production_batches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    plan_id         UUID REFERENCES segment_production_plans(id) ON DELETE SET NULL,
    line_id         UUID REFERENCES segment_factory_lines(id),
    batch_number    VARCHAR(50) NOT NULL,
    ring_number     INTEGER NOT NULL,
    segment_number  INTEGER NOT NULL,                         -- 1..segment_count
    segment_type    VARCHAR(30) NOT NULL DEFAULT 'A'
        CHECK (segment_type IN ('A','B','C','D','E','F','G','K','special')),
    concrete_grade  VARCHAR(30) DEFAULT 'C50/60',
    mould_id        VARCHAR(50),
    pour_time       TIMESTAMPTZ,                               -- время заливки
    pour_operator   VARCHAR(200),
    pour_temp       NUMERIC(5,1),                             -- температура бетона
    pour_volume     NUMERIC(10,2),                            -- объём заливки (m³)
    curing_start    TIMESTAMPTZ,                               -- начало термообработки
    curing_end      TIMESTAMPTZ,                               -- конец термообработки
    curing_temp     NUMERIC(5,1),                             -- температура термообработки
    stripping_time  TIMESTAMPTZ,                               -- распалубка
    qc_passed       BOOLEAN,                                   -- контроль качества
    qc_checked_by   VARCHAR(200),
    qc_checked_at   TIMESTAMPTZ,
    qc_defects      JSONB DEFAULT '[]'::jsonb,                 -- [{type, severity, location}]
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','pouring','curing','stripping','qc_passed','qc_failed','stocked','scrapped')),
    stocked_at      TIMESTAMPTZ,                               -- передача на склад
    stock_location  VARCHAR(200),
    ring_id         UUID,                                        -- FK removed — rings not guaranteed to exist
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, batch_number),
    UNIQUE (project_id, ring_number, segment_number)
);

CREATE INDEX idx_sf_batches_project ON segment_production_batches(project_id);
CREATE INDEX idx_sf_batches_plan ON segment_production_batches(plan_id);
CREATE INDEX idx_sf_batches_line ON segment_production_batches(line_id);
CREATE INDEX idx_sf_batches_ring ON segment_production_batches(ring_number);
CREATE INDEX idx_sf_batches_status ON segment_production_batches(status);

COMMENT ON TABLE segment_production_batches IS 'Производственные циклы сегментов';

-- ============================================================================
-- 4. Склад готовой продукции (сегменты на складе)
-- ============================================================================
CREATE TABLE segment_stock (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    batch_id        UUID REFERENCES segment_production_batches(id) ON DELETE SET NULL,
    ring_number     INTEGER NOT NULL,
    segment_number  INTEGER NOT NULL,
    segment_type    VARCHAR(30) NOT NULL,
    concrete_grade  VARCHAR(30),
    production_date DATE NOT NULL,
    stock_date      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    location        VARCHAR(200),                              -- стеллаж / зона
    status          VARCHAR(30) NOT NULL DEFAULT 'in_stock'
        CHECK (status IN ('in_stock','reserved','shipped','installed','scrapped')),
    shipped_date    TIMESTAMPTZ,
    destination     VARCHAR(300),                              -- туннель / участок
    installed_ring_id UUID,                                    -- FK skipped — rings table not guaranteed
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ring_number, segment_number)
);

CREATE INDEX idx_sf_stock_project ON segment_stock(project_id);
CREATE INDEX idx_sf_stock_status ON segment_stock(status);
CREATE INDEX idx_sf_stock_ring ON segment_stock(ring_number);

COMMENT ON TABLE segment_stock IS 'Склад готовых сегментов';

-- ============================================================================
-- 5. Контроль качества сегментов
-- ============================================================================
CREATE TABLE segment_qc_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    batch_id        UUID REFERENCES segment_production_batches(id) ON DELETE CASCADE,
    ring_number     INTEGER NOT NULL,
    segment_number  INTEGER NOT NULL,
    check_type      VARCHAR(50) NOT NULL
        CHECK (check_type IN ('dimensional','compressive_strength','reinforcement','cover','surface','waterproofing','whole_ring')),
    result          VARCHAR(20) NOT NULL
        CHECK (result IN ('pass','fail','conditional')),
    measured_value  NUMERIC(12,4),
    tolerance_min   NUMERIC(12,4),
    tolerance_max   NUMERIC(12,4),
    deviation_pct   NUMERIC(6,2),
    defect_type     VARCHAR(100),
    defect_severity VARCHAR(20) CHECK (defect_severity IN ('minor','major','critical')),
    defect_location VARCHAR(200),
    corrective_action TEXT,
    checked_by      VARCHAR(200),
    checked_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, ring_number, segment_number, check_type)
);

CREATE INDEX idx_sf_qc_batch ON segment_qc_records(batch_id);
CREATE INDEX idx_sf_qc_ring ON segment_qc_records(ring_number);
COMMENT ON TABLE segment_qc_records IS 'Контроль качества сегментов';

-- ============================================================================
-- 6. Сводка по заводскому производству
-- ============================================================================
CREATE VIEW segment_factory_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM segment_factory_lines WHERE project_id = p.id AND status = 'active') AS active_lines,
    (SELECT COUNT(*) FROM segment_production_plans WHERE project_id = p.id AND status IN ('planned','in_progress')) AS active_plans,
    (SELECT SUM(planned_rings) FROM segment_production_plans WHERE project_id = p.id) AS total_planned_rings,
    (SELECT SUM(produced_rings) FROM segment_production_plans WHERE project_id = p.id) AS total_produced_rings,
    (SELECT COUNT(*) FROM segment_production_batches WHERE project_id = p.id AND status = 'qc_passed') AS qc_passed,
    (SELECT COUNT(*) FROM segment_production_batches WHERE project_id = p.id AND status = 'qc_failed') AS qc_failed,
    (SELECT COUNT(*) FROM segment_production_batches WHERE project_id = p.id AND status = 'scrapped') AS scrapped,
    (SELECT COUNT(*) FROM segment_stock WHERE project_id = p.id AND status = 'in_stock') AS in_stock,
    (SELECT COUNT(*) FROM segment_stock WHERE project_id = p.id AND status = 'shipped') AS shipped
FROM projects p;

COMMENT ON VIEW segment_factory_summary IS 'Сводка по заводскому производству сегментов';

-- ============================================================================
-- Register in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('sf_line',    'Factory Line',  'factory',    'TBM'),
('sf_plan',    'Production Plan','clipboard', 'TBM'),
('sf_batch',   'Prod. Batch',   'layers',     'TBM'),
('sf_stock',   'Segment Stock', 'package',    'TBM'),
('sf_qc',      'Segment QC',    'check-circle','TBM')
ON CONFLICT (code) DO NOTHING;