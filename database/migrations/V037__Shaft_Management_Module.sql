-- ============================================================================
-- V037__Shaft_Management_Module.sql
-- Модуль Shaft Management — строительство шахт / стволов
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

CREATE TABLE shaft_projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    shaft_code      VARCHAR(30) NOT NULL,
    shaft_name      VARCHAR(300) NOT NULL,
    shaft_type      VARCHAR(30) NOT NULL DEFAULT 'launch'
        CHECK (shaft_type IN ('launch','reception','intermediate','ventilation','access','emergency','construction')),
    construction_method VARCHAR(50) DEFAULT 'diaphragm_wall'
        CHECK (construction_method IN ('diaphragm_wall','secant_pile','sheet_pile','contiguous_pile','caisson','cut_and_cover','sink_and_float')),
    diameter_m      NUMERIC(8,2),
    depth_m         NUMERIC(8,2),
    wall_thickness_m NUMERIC(6,2),
    ground_level_m  NUMERIC(8,2),
    invert_level_m  NUMERIC(8,2),
    lining_type     VARCHAR(50) DEFAULT 'concrete'
        CHECK (lining_type IN ('concrete','steel','composite','shotcrete')),
    status          VARCHAR(30) NOT NULL DEFAULT 'design'
        CHECK (status IN ('design','permitting','excavation','lining','base_slab','completed','abandoned')),
    start_date      DATE,
    end_date        DATE,
    contractor_id   UUID REFERENCES organizations(id),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, shaft_code)
);
CREATE INDEX idx_shaft_project ON shaft_projects(project_id);
COMMENT ON TABLE shaft_projects IS 'Проекты шахт и стволов';

CREATE TABLE shaft_construction_sequences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shaft_id        UUID NOT NULL REFERENCES shaft_projects(id) ON DELETE CASCADE,
    sequence_number INTEGER NOT NULL,
    sequence_name   VARCHAR(300) NOT NULL,
    sequence_type   VARCHAR(50) NOT NULL DEFAULT 'excavation'
        CHECK (sequence_type IN ('excavation','support','dewatering','lining','base_slab','backfill','testing','other')),
    start_elevation_m NUMERIC(8,2),
    end_elevation_m   NUMERIC(8,2),
    volume_m3       NUMERIC(12,2),
    reinforcement_kg NUMERIC(12,2),
    concrete_m3     NUMERIC(12,2),
    start_date      DATE,
    end_date        DATE,
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','delayed','cancelled')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shaft_id, sequence_number)
);
CREATE INDEX idx_shaft_seq ON shaft_construction_sequences(shaft_id);
COMMENT ON TABLE shaft_construction_sequences IS 'Последовательность строительства шахты';

CREATE TABLE shaft_instrumentation (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shaft_id        UUID NOT NULL REFERENCES shaft_projects(id) ON DELETE CASCADE,
    instrument_code VARCHAR(50) NOT NULL,
    instrument_type VARCHAR(50) NOT NULL
        CHECK (instrument_type IN ('inclinometer','piezometer','strain_gauge','load_cell','settlement_marker','tiltmeter','extensometer','thermistor','convergence')),
    elevation_m     NUMERIC(8,2),
    depth_m         NUMERIC(8,2),
    install_date    DATE,
    reading_interval_hours NUMERIC(6,1) DEFAULT 24.0,
    alarm_threshold NUMERIC(12,4),
    alarm_direction VARCHAR(10) CHECK (alarm_direction IN ('above','below','both')),
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','damaged','removed','archived')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (shaft_id, instrument_code)
);
CREATE INDEX idx_shaft_inst ON shaft_instrumentation(shaft_id);
COMMENT ON TABLE shaft_instrumentation IS 'Приборы мониторинга шахты';

CREATE TABLE shaft_monitoring_readings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    instrument_id   UUID NOT NULL REFERENCES shaft_instrumentation(id) ON DELETE CASCADE,
    reading_time    TIMESTAMPTZ NOT NULL,
    value           NUMERIC(12,4) NOT NULL,
    unit            VARCHAR(30) DEFAULT 'mm',
    temperature     NUMERIC(5,1),
    is_alarm        BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_shaft_readings_inst ON shaft_monitoring_readings(instrument_id);
CREATE INDEX idx_shaft_readings_time ON shaft_monitoring_readings(reading_time);
COMMENT ON TABLE shaft_monitoring_readings IS 'Показания приборов мониторинга';

CREATE VIEW shaft_summary AS
SELECT
    p.id AS project_id,
    COUNT(DISTINCT sp.id) AS total_shafts,
    COUNT(DISTINCT sp.id) FILTER (WHERE sp.status IN ('excavation','lining','base_slab')) AS active_shafts,
    COUNT(DISTINCT sp.id) FILTER (WHERE sp.status = 'completed') AS completed_shafts,
    COALESCE(SUM(scs.volume_m3) FILTER (WHERE scs.status = 'completed'), 0) AS total_excavated_m3,
    COUNT(DISTINCT si.id) AS active_instruments,
    COUNT(DISTINCT smr.id) FILTER (WHERE smr.is_alarm = TRUE) AS active_alarms
FROM projects p
LEFT JOIN shaft_projects sp ON sp.project_id = p.id
LEFT JOIN shaft_construction_sequences scs ON scs.shaft_id = sp.id
LEFT JOIN shaft_instrumentation si ON si.shaft_id = sp.id AND si.status = 'active'
LEFT JOIN shaft_monitoring_readings smr ON smr.instrument_id = si.id
GROUP BY p.id;

COMMENT ON VIEW shaft_summary IS 'Сводка по шахтам';

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('shaft',               'Shaft',       'layers',           'TBM'),
('shaft_sequence',      'Shaft Seq.',  'list-ordered',     'TBM'),
('shaft_instrument',    'Instrument',  'activity',         'TBM'),
('shaft_reading',       'Reading',     'bar-chart-2',      'TBM')
ON CONFLICT (code) DO NOTHING;