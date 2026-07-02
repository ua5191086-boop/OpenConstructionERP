-- ============================================================================
-- V052__Tunnel_Ventilation_Design.sql
-- Расчёт вентиляции, датчики CO/NOx, аварийные режимы
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Вентиляционные зоны тоннеля
-- ============================================================================
CREATE TABLE tunnel_ventilation_zones (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    zone_code           VARCHAR(100) NOT NULL,
    zone_name           VARCHAR(500) NOT NULL,
    zone_type           VARCHAR(50) NOT NULL,                  -- heading, operational, cross_passage, emergency, shaft
    tbm_id              UUID REFERENCES tbm(id) ON DELETE SET NULL,
    ring_start          INTEGER NOT NULL,
    ring_end            INTEGER NOT NULL,
    length_m            NUMERIC(10,2) NOT NULL,
    cross_section_m2    NUMERIC(8,2),
    airflow_req_m3s     NUMERIC(10,2),                          -- требуемый расход воздуха м³/с
    temp_range_min      NUMERIC(5,2),                           -- °C
    temp_range_max      NUMERIC(5,2),
    humidity_max_pct    NUMERIC(5,2),
    dust_limit_mgm3     NUMERIC(8,2),                           -- мг/м³
    gas_limits          JSONB DEFAULT '{}'::JSONB,              -- {"co":25,"nox":5,"so2":2}
    fan_system_type     VARCHAR(50),                            -- jet, axial, centrifugal, booster
    status              VARCHAR(50) DEFAULT 'active',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tvz_project ON tunnel_ventilation_zones(project_id);
CREATE INDEX idx_tvz_tbm ON tunnel_ventilation_zones(tbm_id);
CREATE INDEX idx_tvz_type ON tunnel_ventilation_zones(zone_type);

COMMENT ON TABLE tunnel_ventilation_zones IS 'Вентиляционные зоны тоннеля';

-- ============================================================================
-- 2. Вентиляционное оборудование
-- ============================================================================
CREATE TABLE tunnel_ventilation_equipment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id             UUID NOT NULL REFERENCES tunnel_ventilation_zones(id) ON DELETE CASCADE,
    equipment_code      VARCHAR(100) NOT NULL,
    equipment_type      VARCHAR(100) NOT NULL,                  -- jet_fan, axial_fan, booster, air_duct, damper, sensor
    manufacturer        VARCHAR(300),
    model               VARCHAR(200),
    capacity_m3s        NUMERIC(10,2),                          -- м³/с (для вентиляторов)
    pressure_pa         NUMERIC(10,2),                          -- Па
    power_kw            NUMERIC(8,2),
    noise_db            NUMERIC(5,2),
    duct_diameter_mm    INTEGER,
    installation_date   DATE,
    last_maintenance    DATE,
    next_maintenance    DATE,
    status              VARCHAR(50) DEFAULT 'operational',      -- operational, maintenance, fault, standby
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tve_zone ON tunnel_ventilation_equipment(zone_id);

COMMENT ON TABLE tunnel_ventilation_equipment IS 'Вентиляционное оборудование тоннеля';

-- ============================================================================
-- 3. Показания датчиков воздуха
-- ============================================================================
CREATE TABLE tunnel_air_quality_readings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    zone_id             UUID REFERENCES tunnel_ventilation_zones(id) ON DELETE SET NULL,
    tbm_id              UUID REFERENCES tbm(id) ON DELETE SET NULL,
    sensor_id           VARCHAR(200),
    ring_number         INTEGER,
    reading_time        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    co_ppm              NUMERIC(8,2),                           -- угарный газ
    nox_ppm             NUMERIC(8,2),                           -- оксиды азота
    so2_ppm             NUMERIC(8,2),                           -- диоксид серы
    h2s_ppm             NUMERIC(8,2),                           -- сероводород
    ch4_ppm             NUMERIC(8,2),                           -- метан
    o2_pct              NUMERIC(5,2),                           -- кислород
    co2_ppm             NUMERIC(8,2),                           -- углекислый газ
    temperature_c       NUMERIC(5,2),                           -- температура
    humidity_pct        NUMERIC(5,2),
    airflow_m3s         NUMERIC(10,2),                          -- расход воздуха
    dust_pm10           NUMERIC(8,2),
    dust_pm25           NUMERIC(8,2),
    noise_db             NUMERIC(5,2),
    is_alarm            BOOLEAN DEFAULT FALSE,
    notes               TEXT
);

CREATE INDEX idx_tar_zone ON tunnel_air_quality_readings(zone_id);
CREATE INDEX idx_tar_tbm ON tunnel_air_quality_readings(tbm_id);
CREATE INDEX idx_tar_time ON tunnel_air_quality_readings(reading_time);
CREATE INDEX idx_tar_alarm ON tunnel_air_quality_readings(is_alarm) WHERE is_alarm = TRUE;

COMMENT ON TABLE tunnel_air_quality_readings IS 'Показания датчиков качества воздуха в тоннеле';

-- ============================================================================
-- 4. Аварийные сценарии вентиляции
-- ============================================================================
CREATE TABLE tunnel_ventilation_emergency_scenarios (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scenario_code       VARCHAR(100) NOT NULL,
    scenario_name       VARCHAR(500) NOT NULL,
    scenario_type       VARCHAR(50) NOT NULL,                  -- fire, gas_leak, equipment_failure, power_loss, flooding
    trigger_conditions  JSONB NOT NULL,                         -- условия активации
    response_actions    JSONB NOT NULL,                         -- действия: {fan_speed:100,opening_dampers:[],alarm_level:2}
    airflow_direction   VARCHAR(20),                            -- normal, reverse, bi_directional
    evacuation_routes   JSONB,
    communication_protocol TEXT,
    response_time_sec   INTEGER,
    last_drilled_date   DATE,
    next_drill_date     DATE,
    status              VARCHAR(50) DEFAULT 'active',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tves_project ON tunnel_ventilation_emergency_scenarios(project_id);

COMMENT ON TABLE tunnel_ventilation_emergency_scenarios IS 'Аварийные сценарии вентиляции';

-- ============================================================================
-- 5. View: сводка качества воздуха
-- ============================================================================
CREATE OR REPLACE VIEW tunnel_air_quality_summary AS
SELECT
    vp.project_id,
    vp.zone_code,
    vp.zone_name,
    COUNT(aq.id) as readings_count,
    AVG(aq.co_ppm) as avg_co,
    MAX(aq.co_ppm) as max_co,
    AVG(aq.nox_ppm) as avg_nox,
    MAX(aq.nox_ppm) as max_nox,
    AVG(aq.temperature_c) as avg_temp,
    MAX(aq.temperature_c) as max_temp,
    AVG(aq.humidity_pct) as avg_humidity,
    COUNT(CASE WHEN aq.is_alarm THEN 1 END) as alarms,
    MAX(aq.reading_time) as last_reading
FROM tunnel_ventilation_zones vp
LEFT JOIN tunnel_air_quality_readings aq ON aq.zone_id = vp.id
GROUP BY vp.id, vp.project_id, vp.zone_code, vp.zone_name;

COMMENT ON VIEW tunnel_air_quality_summary IS 'Сводка качества воздуха по вентзонам';

-- ============================================================================
-- 6. View: статус вентоборудования
-- ============================================================================
CREATE OR REPLACE VIEW tunnel_ventilation_equipment_status AS
SELECT
    ve.id,
    ve.zone_id,
    vz.zone_code,
    vz.zone_name,
    ve.equipment_code,
    ve.equipment_type,
    ve.status,
    ve.installation_date,
    ve.last_maintenance,
    ve.next_maintenance,
    CASE WHEN ve.next_maintenance < CURRENT_DATE THEN 'overdue'
         WHEN ve.next_maintenance < CURRENT_DATE + INTERVAL '7 days' THEN 'due_soon'
         ELSE 'ok' END as maintenance_status
FROM tunnel_ventilation_equipment ve
JOIN tunnel_ventilation_zones vz ON vz.id = ve.zone_id;

COMMENT ON VIEW tunnel_ventilation_equipment_status IS 'Статус вентиляционного оборудования с ТО';