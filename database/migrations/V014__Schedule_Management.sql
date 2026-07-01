-- ============================================================================
-- V014__Schedule_Management.sql
-- Модуль Schedule Management (S-11) — Управление расписанием
-- P6-совместимость, CPM, critical path, resource loading, Gantt, baselines
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Schedules (основные расписания проектов)
-- ============================================================================
CREATE TABLE schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    schedule_code   VARCHAR(30) NOT NULL,
    schedule_name   VARCHAR(500) NOT NULL,
    schedule_type   VARCHAR(30) NOT NULL DEFAULT 'baseline'
        CHECK (schedule_type IN ('baseline','target','current','what_if','recovery','milestone_only')),
    calendar        VARCHAR(100) DEFAULT '5_day_week',
    data_date       DATE,
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','baselined','closed','archived')),
    total_float_pct NUMERIC(5,2),
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, schedule_code)
);

CREATE INDEX idx_schedules_project ON schedules(project_id);
CREATE INDEX idx_schedules_type ON schedules(project_id, schedule_type);

COMMENT ON TABLE schedules IS 'Schedules — расписания проектов (P6-совместимые)';

-- ============================================================================
-- 2. Schedule Activities (активности/операции расписания)
-- ============================================================================
CREATE TABLE schedule_activities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    activity_id     VARCHAR(50) NOT NULL,                    -- P6 activity ID
    wbs_code        VARCHAR(100),
    activity_name   VARCHAR(500) NOT NULL,
    activity_type   VARCHAR(30) NOT NULL DEFAULT 'task'
        CHECK (activity_type IN ('task','milestone','start_milestone','finish_milestone','level_of_effort','wbs_summary')),
    status          VARCHAR(20) NOT NULL DEFAULT 'not_started'
        CHECK (status IN ('not_started','in_progress','completed','suspended','resumed')),
    calendar        VARCHAR(100) DEFAULT '5_day_week',
    original_duration INTEGER NOT NULL DEFAULT 1,
    remaining_duration INTEGER NOT NULL DEFAULT 1,
    actual_duration    INTEGER DEFAULT 0,
    percent_complete   NUMERIC(5,2) DEFAULT 0.00,
    physical_complete  NUMERIC(5,2) DEFAULT 0.00,
    early_start     DATE,
    early_finish    DATE,
    late_start      DATE,
    late_finish     DATE,
    actual_start    DATE,
    actual_finish   DATE,
    start_date      DATE NOT NULL,
    finish_date     DATE NOT NULL,
    float_free      INTEGER,
    float_total     INTEGER,
    is_critical     BOOLEAN DEFAULT FALSE,
    is_driving      BOOLEAN DEFAULT FALSE,
    constraint_type VARCHAR(20) DEFAULT 'as_late_as_possible'
        CHECK (constraint_type IN ('as_late_as_possible','start_on','finish_on','mandatory_start','mandatory_finish','start_no_earlier','start_no_later','finish_no_earlier','finish_no_later')),
    constraint_date DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, activity_id)
);

CREATE INDEX idx_sa_schedule ON schedule_activities(schedule_id);
CREATE INDEX idx_sa_status ON schedule_activities(schedule_id, status);
CREATE INDEX idx_sa_critical ON schedule_activities(schedule_id, is_critical) WHERE is_critical = TRUE;
CREATE INDEX idx_sa_wbs ON schedule_activities(schedule_id, wbs_code);

COMMENT ON TABLE schedule_activities IS 'Schedule Activities — активности расписания с CPM-атрибутами';

-- ============================================================================
-- 3. Schedule Relationships (predecessors/successors)
-- ============================================================================
CREATE TABLE schedule_relationships (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    predecessor_id  UUID NOT NULL REFERENCES schedule_activities(id) ON DELETE CASCADE,
    successor_id    UUID NOT NULL REFERENCES schedule_activities(id) ON DELETE CASCADE,
    relation_type   VARCHAR(10) NOT NULL DEFAULT 'FS'
        CHECK (relation_type IN ('FS','SS','FF','SF')),
    lag_days        INTEGER DEFAULT 0,
    lag_calendar    VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, predecessor_id, successor_id, relation_type)
);

CREATE INDEX idx_sr_schedule ON schedule_relationships(schedule_id);
CREATE INDEX idx_sr_predecessor ON schedule_relationships(predecessor_id);
CREATE INDEX idx_sr_successor ON schedule_relationships(successor_id);

COMMENT ON TABLE schedule_relationships IS 'Schedule Relationships — связи предшествования/следования (CPM)';

-- ============================================================================
-- 4. Schedule Resources (resource loading)
-- ============================================================================
CREATE TABLE schedule_resources (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    activity_id     UUID NOT NULL REFERENCES schedule_activities(id) ON DELETE CASCADE,
    resource_type   VARCHAR(50) NOT NULL DEFAULT 'labor'
        CHECK (resource_type IN ('labor','material','equipment','cost','role')),
    resource_code   VARCHAR(100) NOT NULL,
    resource_name   VARCHAR(300) NOT NULL,
    units_per_day   NUMERIC(12,4) DEFAULT 1.0,
    total_units     NUMERIC(12,4) DEFAULT 1.0,
    unit_cost       NUMERIC(12,2) DEFAULT 0.00,
    total_cost      NUMERIC(12,2) DEFAULT 0.00,
    bid_price       NUMERIC(12,2),
    actual_units    NUMERIC(12,4) DEFAULT 0,
    actual_cost     NUMERIC(12,2) DEFAULT 0.00,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sr_activity ON schedule_resources(activity_id);
CREATE INDEX idx_sr_type ON schedule_resources(schedule_id, resource_type);

COMMENT ON TABLE schedule_resources IS 'Schedule Resources — ресурсная загрузка активностей';

-- ============================================================================
-- 5. Schedule Baselines (управление базовыми планами)
-- ============================================================================
CREATE TABLE schedule_baselines (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    baseline_number INTEGER NOT NULL,
    baseline_name   VARCHAR(300) NOT NULL,
    baseline_date   DATE NOT NULL,
    is_current      BOOLEAN DEFAULT FALSE,
    total_float_pct NUMERIC(5,2),
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, baseline_number)
);

CREATE INDEX idx_sb_schedule ON schedule_baselines(schedule_id);
CREATE INDEX idx_sb_current ON schedule_baselines(schedule_id, is_current) WHERE is_current = TRUE;

COMMENT ON TABLE schedule_baselines IS 'Schedule Baselines — базовые планы расписания';

-- ============================================================================
-- 6. Schedule Changes (изменения расписания)
-- ============================================================================
CREATE TABLE schedule_changes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    change_number   INTEGER NOT NULL,
    change_code     VARCHAR(30) NOT NULL,
    change_type     VARCHAR(50) NOT NULL DEFAULT 'scope'
        CHECK (change_type IN ('scope','duration','relationship','resource','constraint','calendar','logic','milestone')),
    description     TEXT NOT NULL,
    reason          TEXT,
    impact_days     INTEGER,
    impact_cost     NUMERIC(12,2),
    activity_id     UUID REFERENCES schedule_activities(id),
    baseline_id     UUID REFERENCES schedule_baselines(id),
    approved_by     VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'proposed'
        CHECK (status IN ('proposed','reviewed','approved','rejected','implemented','rolled_back')),
    proposed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (schedule_id, change_number),
    UNIQUE (schedule_id, change_code)
);

CREATE INDEX idx_sch_schedule ON schedule_changes(schedule_id);
CREATE INDEX idx_sch_status ON schedule_changes(schedule_id, status);

COMMENT ON TABLE schedule_changes IS 'Schedule Changes — управление изменениями расписания';

-- ============================================================================
-- 7. Critical Path Log (история расчётов критического пути)
-- ============================================================================
CREATE TABLE critical_path_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id     UUID NOT NULL REFERENCES schedules(id) ON DELETE CASCADE,
    run_number      INTEGER NOT NULL,
    run_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_activities INTEGER,
    critical_count  INTEGER,
    longest_path    INTEGER,                                 -- longest path duration (days)
    total_float_min INTEGER,
    total_float_max INTEGER,
    total_float_avg NUMERIC(6,2),
    critical_path   TEXT,                                    -- JSON array of activity_ids on critical path
    duration        INTEGER,                                 -- calculation duration (ms)
    status          VARCHAR(20) NOT NULL DEFAULT 'completed'
        CHECK (status IN ('running','completed','failed','invalidated')),
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cpl_schedule ON critical_path_log(schedule_id);
CREATE INDEX idx_cpl_run ON critical_path_log(schedule_id, run_number DESC);

COMMENT ON TABLE critical_path_log IS 'Critical Path Log — история расчётов критического пути (CPM)';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('schedule',            'Schedule',             'calendar',     'S-11'),
('schedule_activity',   'Schedule Activity',    'checklist',    'S-11'),
('schedule_relationship','Schedule Relation',   'link',         'S-11'),
('schedule_resource',   'Schedule Resource',    'users',        'S-11'),
('schedule_baseline',   'Schedule Baseline',    'flag',         'S-11'),
('schedule_change',     'Schedule Change',      'edit',         'S-11'),
('critical_path',       'Critical Path',        'route',        'S-11')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW schedule_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM schedules WHERE project_id = p.id) AS total_schedules,
    (SELECT COUNT(*) FROM schedules WHERE project_id = p.id AND status = 'active') AS active_schedules,
    (SELECT COUNT(*) FROM schedule_activities sa JOIN schedules s ON sa.schedule_id = s.id WHERE s.project_id = p.id) AS total_activities,
    (SELECT COUNT(*) FROM schedule_activities sa JOIN schedules s ON sa.schedule_id = s.id WHERE s.project_id = p.id AND sa.status = 'not_started') AS not_started,
    (SELECT COUNT(*) FROM schedule_activities sa JOIN schedules s ON sa.schedule_id = s.id WHERE s.project_id = p.id AND sa.status = 'in_progress') AS in_progress,
    (SELECT COUNT(*) FROM schedule_activities sa JOIN schedules s ON sa.schedule_id = s.id WHERE s.project_id = p.id AND sa.status = 'completed') AS completed,
    (SELECT COUNT(*) FROM schedule_activities sa JOIN schedules s ON sa.schedule_id = s.id WHERE s.project_id = p.id AND sa.is_critical = TRUE) AS critical_activities,
    (SELECT COUNT(*) FROM schedule_relationships sr JOIN schedules s ON sr.schedule_id = s.id WHERE s.project_id = p.id) AS total_relationships,
    (SELECT COUNT(*) FROM schedule_resources sr JOIN schedules s ON sr.schedule_id = s.id WHERE s.project_id = p.id) AS total_resources,
    (SELECT COUNT(*) FROM schedule_baselines sb JOIN schedules s ON sb.schedule_id = s.id WHERE s.project_id = p.id) AS total_baselines,
    (SELECT COUNT(*) FROM schedule_changes sc JOIN schedules s ON sc.schedule_id = s.id WHERE s.project_id = p.id AND sc.status IN ('proposed','reviewed')) AS pending_changes,
    (SELECT COUNT(*) FROM critical_path_log cpl JOIN schedules s ON cpl.schedule_id = s.id WHERE s.project_id = p.id) AS cpm_runs
FROM projects p;