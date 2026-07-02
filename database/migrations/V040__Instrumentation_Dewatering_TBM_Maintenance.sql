-- ============================================================================
-- V040__Instrumentation_Dewatering_TBM_Maintenance.sql
-- Модули: Instrumentation, Dewatering, TBM Maintenance
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- PART 1: Instrumentation (Sensors & Monitoring)
-- ============================================================================
CREATE TABLE instrumentation_sensors (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sensor_code     VARCHAR(30) NOT NULL,
    sensor_name     VARCHAR(300) NOT NULL,
    sensor_type     VARCHAR(50) NOT NULL
        CHECK (sensor_type IN ('strain_gauge','load_cell','pressure_cell','thermistor','accelerometer','crack_meter','tiltmeter','vibration_meter','convergence','extensometer','piezometer','total_station','laser_scanner','acoustic')),
    chainage_m      NUMERIC(10,2),
    location        VARCHAR(300),
    install_date    DATE,
    manufacturer    VARCHAR(200),
    model           VARCHAR(200),
    serial_number   VARCHAR(100),
    reading_unit    VARCHAR(30) DEFAULT 'mm',
    reading_interval_sec INTEGER DEFAULT 3600,
    alarm_low       NUMERIC(12,4),
    alarm_high      NUMERIC(12,4),
    data_channel    VARCHAR(100),                               -- data acquisition channel
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, sensor_code)
);
CREATE INDEX idx_inst_sensor_project ON instrumentation_sensors(project_id);
CREATE INDEX idx_inst_sensor_type ON instrumentation_sensors(sensor_type);
COMMENT ON TABLE instrumentation_sensors IS 'Датчики и приборы мониторинга';

CREATE TABLE instrumentation_readings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sensor_id       UUID NOT NULL REFERENCES instrumentation_sensors(id) ON DELETE CASCADE,
    reading_time    TIMESTAMPTZ NOT NULL,
    value           NUMERIC(12,4) NOT NULL,
    unit            VARCHAR(30),
    temperature     NUMERIC(5,1),
    is_alarm        BOOLEAN DEFAULT FALSE,
    battery_level   NUMERIC(5,1),
    signal_strength INTEGER,
    raw_data        JSONB,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_inst_read_sensor ON instrumentation_readings(sensor_id);
CREATE INDEX idx_inst_read_time ON instrumentation_readings(reading_time DESC);
COMMENT ON TABLE instrumentation_readings IS 'Показания датчиков';

-- ============================================================================
-- PART 2: Dewatering
-- ============================================================================
CREATE TABLE dewatering_wells (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    well_code       VARCHAR(30) NOT NULL,
    well_name       VARCHAR(300) NOT NULL,
    well_type       VARCHAR(50) NOT NULL DEFAULT 'deep'
        CHECK (well_type IN ('deep','shallow','vacuum','horizontal','eductor','drain')),
    chainage_m      NUMERIC(10,2),
    lat             NUMERIC(10,7),
    lng             NUMERIC(10,7),
    depth_m         NUMERIC(8,2),
    diameter_mm     INTEGER,
    pump_capacity_m3h NUMERIC(10,2),                           -- m³/hour
    static_water_level_m NUMERIC(8,2),
    drawdown_m      NUMERIC(8,2),
    filter_type     VARCHAR(100),
    installation_date DATE,
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','maintenance','standby','decommissioned','abandoned')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, well_code)
);
CREATE INDEX idx_dew_well_project ON dewatering_wells(project_id);
CREATE INDEX idx_dew_well_chainage ON dewatering_wells(chainage_m);
COMMENT ON TABLE dewatering_wells IS 'Водопонизительные скважины';

CREATE TABLE dewatering_readings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    well_id         UUID NOT NULL REFERENCES dewatering_wells(id) ON DELETE CASCADE,
    reading_time    TIMESTAMPTZ NOT NULL,
    water_level_m   NUMERIC(8,2),
    flow_rate_m3h   NUMERIC(10,2),                             -- m³/hour
    pump_running    BOOLEAN DEFAULT TRUE,
    pump_pressure_bar NUMERIC(6,2),
    energy_kwh      NUMERIC(10,2),
    sediment_pct    NUMERIC(5,2),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_dew_read_well ON dewatering_readings(well_id);
CREATE INDEX idx_dew_read_time ON dewatering_readings(reading_time);
COMMENT ON TABLE dewatering_readings IS 'Замеры водопонижения';

-- ============================================================================
-- PART 3: TBM Maintenance
-- ============================================================================
CREATE TABLE tbm_maintenance_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tbm_id          VARCHAR(50) NOT NULL,
    task_code       VARCHAR(30) NOT NULL,
    task_name       VARCHAR(300) NOT NULL,
    task_type       VARCHAR(50) NOT NULL
        CHECK (task_type IN ('inspection','preventive','corrective','emergency','overhaul','replacement','calibration','cleaning','lubrication')),
    component       VARCHAR(200) NOT NULL,                      -- cutterhead, seal, gearbox, screw, erector, etc.
    priority        VARCHAR(20) DEFAULT 'medium'
        CHECK (priority IN ('low','medium','high','critical')),
    description     TEXT,
    interval_ring   INTEGER,                                   -- каждые N колец
    interval_days   INTEGER,
    last_done_at_ring INTEGER,
    last_done_date  DATE,
    estimated_hours NUMERIC(8,2),
    actual_hours    NUMERIC(8,2),
    assigned_crew   VARCHAR(200),
    required_parts  TEXT,
    status          VARCHAR(30) NOT NULL DEFAULT 'scheduled'
        CHECK (status IN ('scheduled','in_progress','completed','deferred','cancelled','overdue')),
    completion_date DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, task_code)
);
CREATE INDEX idx_tbm_maint_project ON tbm_maintenance_tasks(project_id);
CREATE INDEX idx_tbm_maint_tbm ON tbm_maintenance_tasks(tbm_id);
CREATE INDEX idx_tbm_maint_status ON tbm_maintenance_tasks(status);
COMMENT ON TABLE tbm_maintenance_tasks IS 'Задачи ТО ТБМ';

CREATE TABLE tbm_maintenance_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID NOT NULL REFERENCES tbm_maintenance_tasks(id) ON DELETE CASCADE,
    log_time        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ring_number     INTEGER,
    action          TEXT NOT NULL,
    duration_hours  NUMERIC(8,2),
    parts_used      TEXT,                                       -- список запчастей
    downtime_hours  NUMERIC(8,2),
    performed_by    VARCHAR(200),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tbm_log_task ON tbm_maintenance_logs(task_id);
COMMENT ON TABLE tbm_maintenance_logs IS 'Логи выполнения ТО ТБМ';

CREATE VIEW tunnel_services_summary AS
SELECT
    p.id AS project_id,
    COUNT(DISTINCT ins.id) AS active_sensors,
    COUNT(DISTINCT ir.id) AS sensor_readings,
    COUNT(DISTINCT ir.id) FILTER (WHERE ir.is_alarm = TRUE) AS sensor_alarms,
    COUNT(DISTINCT dw.id) FILTER (WHERE dw.status = 'active') AS active_dewatering_wells,
    COUNT(DISTINCT dr.id) AS dewatering_readings,
    COUNT(DISTINCT tmt.id) AS tbm_maintenance_tasks,
    COUNT(DISTINCT tmt.id) FILTER (WHERE tmt.status = 'scheduled') AS pending_maintenance
FROM projects p
LEFT JOIN instrumentation_sensors ins ON ins.project_id = p.id AND ins.is_active = TRUE
LEFT JOIN instrumentation_readings ir ON ir.sensor_id = ins.id
LEFT JOIN dewatering_wells dw ON dw.project_id = p.id
LEFT JOIN dewatering_readings dr ON dr.well_id = dw.id
LEFT JOIN tbm_maintenance_tasks tmt ON tmt.project_id = p.id
GROUP BY p.id;

COMMENT ON VIEW tunnel_services_summary IS 'Сводка по датчикам, водопонижению, ТО ТБМ';

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('tbm_maintenance', 'TBM Maintenance', 'tool',         'TBM'),
('tbm_maint_log',   'Maint. Log',      'clipboard',    'TBM'),
('instrument_sensor','Sensor',         'activity',     'TBM'),
('instrument_reading','Sensor Reading','bar-chart-2',   'TBM'),
('dewatering_well',  'Dewatering Well','droplet',      'TBM'),
('dewatering_reading','DW Reading',    'bar-chart-3',  'TBM')
ON CONFLICT (code) DO NOTHING;