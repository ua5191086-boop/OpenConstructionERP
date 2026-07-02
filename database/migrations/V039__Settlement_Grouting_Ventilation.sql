-- ============================================================================
-- V039__Settlement_Grouting_Ventilation.sql
-- Модули: Settlement Monitoring, Grouting, Ventilation
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- PART 1: Settlement Monitoring
-- ============================================================================
CREATE TABLE settlement_monitoring_points (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    point_code      VARCHAR(30) NOT NULL,
    point_name      VARCHAR(300),
    point_type      VARCHAR(50) NOT NULL DEFAULT 'surface'
        CHECK (point_type IN ('surface','subsurface','building','utility','road','bridge','railway')),
    chainage_m      NUMERIC(10,2) NOT NULL,
    offset_m        NUMERIC(8,2) DEFAULT 0,
    lat             NUMERIC(10,7),
    lng             NUMERIC(10,7),
    initial_level_m NUMERIC(8,3) NOT NULL,                     -- начальная отметка
    trigger_alert_mm NUMERIC(6,2) DEFAULT 10.0,                -- порог оповещения (мм)
    trigger_urgent_mm NUMERIC(6,2) DEFAULT 25.0,               -- аварийный порог (мм)
    trigger_rate_mm_per_day NUMERIC(6,2) DEFAULT 2.0,          -- скорость осадки (мм/день)
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','completed','damaged','removed')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, point_code)
);
CREATE INDEX idx_settle_point_project ON settlement_monitoring_points(project_id);
CREATE INDEX idx_settle_point_chainage ON settlement_monitoring_points(chainage_m);
COMMENT ON TABLE settlement_monitoring_points IS 'Точки мониторинга осадок';

CREATE TABLE settlement_readings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    point_id        UUID NOT NULL REFERENCES settlement_monitoring_points(id) ON DELETE CASCADE,
    reading_time    TIMESTAMPTZ NOT NULL,
    level_m         NUMERIC(8,3) NOT NULL,                     -- текущая отметка
    settlement_mm   NUMERIC(8,2) NOT NULL,                     -- осадка от начальной (мм)
    rate_mm_per_day NUMERIC(6,2),                              -- скорость осадки
    is_alert        BOOLEAN DEFAULT FALSE,
    is_urgent       BOOLEAN DEFAULT FALSE,
    instrument      VARCHAR(100),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_settle_read_point ON settlement_readings(point_id);
CREATE INDEX idx_settle_read_time ON settlement_readings(reading_time);
CREATE INDEX idx_settle_read_alert ON settlement_readings(is_alert) WHERE is_alert = TRUE;
COMMENT ON TABLE settlement_readings IS 'Замеры осадок';

-- ============================================================================
-- PART 2: Grouting
-- ============================================================================
CREATE TABLE grouting_activities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    grout_code      VARCHAR(30) NOT NULL,
    grout_name      VARCHAR(300) NOT NULL,
    grout_type      VARCHAR(50) NOT NULL DEFAULT 'backfill'
        CHECK (grout_type IN ('backfill','consolidation','contact','preventive','curtain','compensation','anchorage','other')),
    chainage_from_m NUMERIC(10,2),
    chainage_to_m   NUMERIC(10,2),
    location        VARCHAR(200),                               -- ring / zone
    mix_design      VARCHAR(200),                               -- w/c, additives
    target_pressure_bar NUMERIC(6,2),
    max_pressure_bar   NUMERIC(6,2),
    target_volume_m3   NUMERIC(10,2),
    actual_volume_m3   NUMERIC(10,2),
    start_date      DATE,
    end_date        DATE,
    supervisor      VARCHAR(200),
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','verified','failed')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, grout_code)
);
CREATE INDEX idx_grout_project ON grouting_activities(project_id);
CREATE INDEX idx_grout_chainage ON grouting_activities(chainage_from_m);
COMMENT ON TABLE grouting_activities IS 'Тампонажные / инъекционные работы';

CREATE TABLE grouting_records (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grout_id        UUID NOT NULL REFERENCES grouting_activities(id) ON DELETE CASCADE,
    record_time     TIMESTAMPTZ NOT NULL,
    pressure_bar    NUMERIC(6,2),
    flow_rate_lpm   NUMERIC(8,2),                              -- litres per minute
    volume_m3       NUMERIC(8,3),
    density_kg_m3   NUMERIC(8,2),
    temperature     NUMERIC(5,1),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_grout_rec_activity ON grouting_records(grout_id);
COMMENT ON TABLE grouting_records IS 'Поточные записи тампонажа';

-- ============================================================================
-- PART 3: Ventilation
-- ============================================================================
CREATE TABLE ventilation_systems (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    system_code     VARCHAR(30) NOT NULL,
    system_name     VARCHAR(300) NOT NULL,
    vent_type       VARCHAR(50) NOT NULL DEFAULT 'forced'
        CHECK (vent_type IN ('forced','exhaust','combined','natural','jet_fan')),
    fan_count       INTEGER DEFAULT 1,
    fan_power_kw    NUMERIC(10,2),
    airflow_m3_s    NUMERIC(10,2),                              -- m³/sec
    duct_diameter_mm INTEGER,
    duct_length_m   NUMERIC(10,2),
    location        VARCHAR(300),
    chainage_m      NUMERIC(10,2),
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','maintenance','standby','decommissioned')),
    installed_date  DATE,
    last_maintenance DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, system_code)
);
CREATE INDEX idx_vent_project ON ventilation_systems(project_id);
COMMENT ON TABLE ventilation_systems IS 'Вентиляционные системы';

CREATE TABLE ventilation_readings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id       UUID NOT NULL REFERENCES ventilation_systems(id) ON DELETE CASCADE,
    reading_time    TIMESTAMPTZ NOT NULL,
    airflow_m3_s    NUMERIC(10,2),
    temperature_c   NUMERIC(5,1),
    humidity_pct    NUMERIC(5,1),
    co_ppm          NUMERIC(8,2),                               -- CO
    co2_ppm         NUMERIC(8,2),                               -- CO₂
    no2_ppm         NUMERIC(8,2),                               -- NO₂
    dust_mg_m3      NUMERIC(8,2),                               -- взвешенные частицы
    fan_speed_pct   NUMERIC(5,1),                               -- % от макс
    power_kw        NUMERIC(10,2),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vent_read_system ON ventilation_readings(system_id);
CREATE INDEX idx_vent_read_time ON ventilation_readings(reading_time);
COMMENT ON TABLE ventilation_readings IS 'Показатели вентиляции и空气质量';

CREATE VIEW settlement_grouting_summary AS
SELECT
    p.id AS project_id,
    COUNT(DISTINCT smp.id) AS settlement_points,
    COUNT(DISTINCT smp.id) FILTER (WHERE smp.status = 'active') AS active_settlement_points,
    COUNT(DISTINCT sr.id) AS settlement_readings,
    COUNT(DISTINCT sr.id) FILTER (WHERE sr.is_alert = TRUE) AS settlement_alerts,
    COUNT(DISTINCT sr.id) FILTER (WHERE sr.is_urgent = TRUE) AS settlement_urgent,
    COUNT(DISTINCT ga.id) AS grouting_activities,
    COUNT(DISTINCT ga.id) FILTER (WHERE ga.status = 'in_progress') AS active_grouting,
    COALESCE(SUM(ga.actual_volume_m3), 0) AS total_grout_volume_m3,
    COUNT(DISTINCT vs.id) AS ventilation_systems,
    COUNT(DISTINCT vr.id) AS ventilation_readings
FROM projects p
LEFT JOIN settlement_monitoring_points smp ON smp.project_id = p.id
LEFT JOIN settlement_readings sr ON sr.point_id = smp.id
LEFT JOIN grouting_activities ga ON ga.project_id = p.id
LEFT JOIN ventilation_systems vs ON vs.project_id = p.id
LEFT JOIN ventilation_readings vr ON vr.system_id = vs.id
GROUP BY p.id;

COMMENT ON VIEW settlement_grouting_summary IS 'Сводка по осадкам, тампонажу, вентиляции';

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('settlement_point',  'Settlement Point', 'crosshair',  'TBM'),
('settlement_reading','Settlement Read',  'activity',   'TBM'),
('grouting',          'Grouting',         'droplet',    'TBM'),
('grouting_record',   'Grout Record',     'bar-chart-3','TBM'),
('ventilation',       'Ventilation',      'wind',       'TBM'),
('ventilation_reading','Vent Reading',    'thermometer','TBM')
ON CONFLICT (code) DO NOTHING;