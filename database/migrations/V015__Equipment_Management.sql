-- ============================================================================
-- V015__Equipment_Management.sql
-- Модуль Equipment Management (E-12) — Управление оборудованием
-- TBM, cranes, fleet, predictive maintenance, telemetry
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Equipment Categories
-- ============================================================================
CREATE TABLE equipment_categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_code   VARCHAR(30) NOT NULL UNIQUE,
    category_name   VARCHAR(200) NOT NULL,
    description     TEXT,
    parent_id       UUID REFERENCES equipment_categories(id),
    equipment_type  VARCHAR(50) NOT NULL DEFAULT 'general'
        CHECK (equipment_type IN ('tbm','crane','fleet','heavy','light','specialty','general')),
    icon            VARCHAR(50),
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ec_parent ON equipment_categories(parent_id);
CREATE INDEX idx_ec_type ON equipment_categories(equipment_type);

COMMENT ON TABLE equipment_categories IS 'Equipment Categories — категории оборудования';

-- ============================================================================
-- 2. Equipment (основная таблица)
-- ============================================================================
CREATE TABLE equipment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_code  VARCHAR(50) NOT NULL,
    equipment_name  VARCHAR(500) NOT NULL,
    category_id     UUID REFERENCES equipment_categories(id),
    equipment_type  VARCHAR(50) NOT NULL DEFAULT 'general'
        CHECK (equipment_type IN ('tbm','crane','fleet_vehicle','heavy_machine','light_equipment','specialty','general','pump','generator','compressor','welding')),
    manufacturer    VARCHAR(200),
    model           VARCHAR(200),
    serial_number   VARCHAR(200),
    year_manufactured INTEGER,
    capacity        VARCHAR(100),                            -- грузоподъёмность/мощность
    capacity_unit   VARCHAR(20),                             -- tons, kW, m3, etc
    status          VARCHAR(30) NOT NULL DEFAULT 'available'
        CHECK (status IN ('available','in_use','under_maintenance','out_of_service','retired','reserved','transferred')),
    location        VARCHAR(300),
    gps_coordinates POINT,
    purchase_date   DATE,
    purchase_cost   NUMERIC(12,2),
    current_value   NUMERIC(12,2),
    fuel_type       VARCHAR(30),                             -- diesel, petrol, electric, hybrid, lpg, cng
    fuel_capacity   NUMERIC(8,2),
    hourly_rate     NUMERIC(10,2),
    meter_type      VARCHAR(30) DEFAULT 'hours'
        CHECK (meter_type IN ('hours','km','miles','cycles')),
    meter_reading   NUMERIC(12,2) DEFAULT 0,
    operator_required BOOLEAN DEFAULT TRUE,
    insurance_policy VARCHAR(200),
    insurance_expiry DATE,
    inspection_date DATE,
    last_service_date DATE,
    next_service_date DATE,
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, equipment_code)
);

CREATE INDEX idx_equip_project ON equipment(project_id);
CREATE INDEX idx_equip_status ON equipment(project_id, status);
CREATE INDEX idx_equip_type ON equipment(project_id, equipment_type);
CREATE INDEX idx_equip_category ON equipment(category_id);
CREATE INDEX idx_equip_active ON equipment(project_id, is_active) WHERE is_active = TRUE;

COMMENT ON TABLE equipment IS 'Equipment — учёт строительной техники и оборудования';

-- ============================================================================
-- 3. Equipment Maintenance (журнал обслуживания)
-- ============================================================================
CREATE TABLE equipment_maintenance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    maintenance_code VARCHAR(30) NOT NULL,
    maintenance_type VARCHAR(50) NOT NULL DEFAULT 'preventive'
        CHECK (maintenance_type IN ('preventive','corrective','predictive','condition_based','emergency','overhaul','inspection')),
    description     TEXT NOT NULL,
    priority        VARCHAR(15) DEFAULT 'normal'
        CHECK (priority IN ('low','normal','high','critical')),
    status          VARCHAR(20) NOT NULL DEFAULT 'scheduled'
        CHECK (status IN ('scheduled','in_progress','completed','cancelled','deferred')),
    meter_at_service NUMERIC(12,2),
    cost_estimated  NUMERIC(12,2),
    cost_actual     NUMERIC(12,2),
    downtime_hours  NUMERIC(6,2),
    technician      VARCHAR(200),
    vendor          VARCHAR(200),
    parts_used      TEXT,
    findings        TEXT,
    recommendations TEXT,
    scheduled_date  DATE,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    next_service_meter NUMERIC(12,2),
    next_service_date  DATE,
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (equipment_id, maintenance_code)
);

CREATE INDEX idx_em_equipment ON equipment_maintenance(equipment_id);
CREATE INDEX idx_em_status ON equipment_maintenance(equipment_id, status);
CREATE INDEX idx_em_type ON equipment_maintenance(equipment_id, maintenance_type);
CREATE INDEX idx_em_scheduled ON equipment_maintenance(scheduled_date) WHERE status IN ('scheduled','in_progress');

COMMENT ON TABLE equipment_maintenance IS 'Equipment Maintenance — журнал технического обслуживания';

-- ============================================================================
-- 4. Maintenance Schedules (регулярные графики ТО)
-- ============================================================================
CREATE TABLE maintenance_schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    schedule_name   VARCHAR(300) NOT NULL,
    interval_type   VARCHAR(20) NOT NULL DEFAULT 'calendar'
        CHECK (interval_type IN ('calendar','meter','both')),
    interval_days   INTEGER,
    interval_meter  NUMERIC(12,2),
    task_list       TEXT,
    estimated_hours NUMERIC(6,2),
    required_skills TEXT,
    spare_parts     TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ms_equipment ON maintenance_schedules(equipment_id);
CREATE INDEX idx_ms_active ON maintenance_schedules(equipment_id, is_active) WHERE is_active = TRUE;

COMMENT ON TABLE maintenance_schedules IS 'Maintenance Schedules — регламентные графики обслуживания';

-- ============================================================================
-- 5. Equipment Telemetry (телеметрия)
-- ============================================================================
CREATE TABLE equipment_telemetry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    meter_value     NUMERIC(12,2),
    fuel_level_pct  NUMERIC(5,2),
    engine_temp_c   NUMERIC(5,1),
    oil_pressure_bar NUMERIC(5,2),
    rpm             INTEGER,
    speed_kph       NUMERIC(6,2),
    gps_lat         NUMERIC(10,7),
    gps_lon         NUMERIC(10,7),
    battery_voltage NUMERIC(5,2),
    error_codes     TEXT,                                    -- JSON array
    vibration_x     NUMERIC(6,3),
    vibration_y     NUMERIC(6,3),
    vibration_z     NUMERIC(6,3),
    is_operating    BOOLEAN DEFAULT FALSE,
    data_source     VARCHAR(50) DEFAULT 'manual'
        CHECK (data_source IN ('manual','iot_sensor','api','gps_tracker','can_bus'))
);

CREATE INDEX idx_tel_equipment ON equipment_telemetry(equipment_id);
CREATE INDEX idx_tel_time ON equipment_telemetry(equipment_id, recorded_at DESC);
CREATE INDEX idx_tel_recent ON equipment_telemetry(equipment_id, recorded_at DESC) WHERE is_operating = TRUE;

COMMENT ON TABLE equipment_telemetry IS 'Equipment Telemetry — телеметрия оборудования (IoT)';

-- ============================================================================
-- 6. Equipment Fuel (учёт топлива)
-- ============================================================================
CREATE TABLE equipment_fuel (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    refuel_date     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    fuel_type       VARCHAR(30) NOT NULL DEFAULT 'diesel'
        CHECK (fuel_type IN ('diesel','petrol','electric','adblue','lpg','cng')),
    quantity_liters NUMERIC(10,2) NOT NULL,
    cost_per_liter  NUMERIC(8,2),
    total_cost      NUMERIC(10,2),
    meter_reading   NUMERIC(12,2),
    operator        VARCHAR(200),
    vendor          VARCHAR(200),
    receipt_number  VARCHAR(100),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ef_equipment ON equipment_fuel(equipment_id);
CREATE INDEX idx_ef_date ON equipment_fuel(equipment_id, refuel_date DESC);

COMMENT ON TABLE equipment_fuel IS 'Equipment Fuel — учёт топлива';

-- ============================================================================
-- 7. Equipment Operators (назначение операторов)
-- ============================================================================
CREATE TABLE equipment_operators (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    employee_id     UUID NOT NULL,                           -- связь с HR-модулем
    full_name       VARCHAR(300) NOT NULL,
    certification   VARCHAR(200),
    certification_expiry DATE,
    assigned_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    end_date        DATE,
    shift           VARCHAR(20) DEFAULT 'day'
        CHECK (shift IN ('day','night','A','B','C','rotating')),
    is_primary      BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (equipment_id, employee_id, COALESCE(end_date, '9999-12-31'))
);

CREATE INDEX idx_eo_equipment ON equipment_operators(equipment_id);
CREATE INDEX idx_eo_employee ON equipment_operators(employee_id);

COMMENT ON TABLE equipment_operators IS 'Equipment Operators — операторы оборудования';

-- ============================================================================
-- 8. Equipment Downtime (учёт простоев)
-- ============================================================================
CREATE TABLE equipment_downtime (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    downtime_type   VARCHAR(50) NOT NULL DEFAULT 'breakdown'
        CHECK (downtime_type IN ('breakdown','maintenance','fueling','operator_unavailable','no_work','weather','mobilization','other')),
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ,
    duration_hours  NUMERIC(6,2),
    reason          TEXT,
    impact          TEXT,
    cost_impact     NUMERIC(12,2),
    reported_by     VARCHAR(200),
    status          VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','resolved','closed')),
    resolution      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ed_equipment ON equipment_downtime(equipment_id);
CREATE INDEX idx_ed_type ON equipment_downtime(equipment_id, downtime_type);
CREATE INDEX idx_ed_time ON equipment_downtime(equipment_id, start_time DESC);

COMMENT ON TABLE equipment_downtime IS 'Equipment Downtime — учёт простоев оборудования';

-- ============================================================================
-- 9. Equipment Spare Parts (запчасти)
-- ============================================================================
CREATE TABLE equipment_spare_parts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_id    UUID REFERENCES equipment(id),            -- NULL = общая запчасть
    part_code       VARCHAR(50) NOT NULL,
    part_name       VARCHAR(300) NOT NULL,
    part_number     VARCHAR(200),                            -- manufacturer part number
    category        VARCHAR(100),
    unit            VARCHAR(20) NOT NULL DEFAULT 'pcs',
    quantity_on_hand NUMERIC(10,2) DEFAULT 0,
    min_stock_level  NUMERIC(10,2) DEFAULT 0,
    unit_cost       NUMERIC(10,2) DEFAULT 0.00,
    supplier        VARCHAR(300),
    lead_time_days  INTEGER,
    storage_location VARCHAR(200),
    last_restocked  DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (equipment_id, part_code)
);

CREATE INDEX idx_esp_equipment ON equipment_spare_parts(equipment_id);
CREATE INDEX idx_esp_stock ON equipment_spare_parts(quantity_on_hand) WHERE quantity_on_hand <= min_stock_level;

COMMENT ON TABLE equipment_spare_parts IS 'Equipment Spare Parts — запчасти для оборудования';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('equipment',           'Equipment',            'truck',        'E-12'),
('equipment_category',  'Equipment Category',   'folder',       'E-12'),
('equipment_maintenance','Equipment Maint.',    'wrench',       'E-12'),
('maintenance_schedule','Maint. Schedule',      'clock',        'E-12'),
('equipment_telemetry', 'Equipment Telemetry',  'activity',     'E-12'),
('equipment_fuel',      'Equipment Fuel',       'droplet',      'E-12'),
('equipment_operator',  'Equipment Operator',   'user',         'E-12'),
('equipment_downtime',  'Equipment Downtime',   'pause',        'E-12'),
('equipment_spare_part','Spare Part',           'package',      'E-12')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW equipment_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id) AS total_equipment,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND status = 'available') AS available,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND status = 'in_use') AS in_use,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND status = 'under_maintenance') AS under_maintenance,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND status = 'out_of_service') AS out_of_service,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND equipment_type = 'tbm') AS tbms,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND equipment_type = 'crane') AS cranes,
    (SELECT COUNT(*) FROM equipment WHERE project_id = p.id AND equipment_type = 'fleet_vehicle') AS fleet,
    (SELECT COUNT(*) FROM equipment_maintenance em JOIN equipment e ON em.equipment_id = e.id WHERE e.project_id = p.id AND em.status IN ('scheduled','in_progress')) AS pending_maintenance,
    (SELECT COUNT(*) FROM equipment_downtime ed JOIN equipment e ON ed.equipment_id = e.id WHERE e.project_id = p.id AND ed.status = 'open') AS active_downtime,
    (SELECT COALESCE(SUM(ed.duration_hours),0) FROM equipment_downtime ed JOIN equipment e ON ed.equipment_id = e.id WHERE e.project_id = p.id AND ed.start_time >= CURRENT_DATE - 30) AS downtime_30d,
    (SELECT COUNT(*) FROM equipment_fuel ef JOIN equipment e ON ef.equipment_id = e.id WHERE e.project_id = p.id AND ef.refuel_date >= CURRENT_DATE - 30) AS refuels_30d
FROM projects p;