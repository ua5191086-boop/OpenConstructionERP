-- ============================================================================
-- V052__Tunnel_Ventilation_Design.sql
-- Расчёт вентиляции, датчики CO/NOx, аварийные режимы
-- ============================================================================

CREATE TABLE tunnel_ventilation_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    zone_code VARCHAR(50) NOT NULL,
    chainage_from NUMERIC(12,4),
    chainage_to NUMERIC(12,4),
    cross_section_area NUMERIC(10,4),
    ventilation_type VARCHAR(50), -- longitudinal, transverse, semi_transverse, jet_fan
    airflow_required NUMERIC(12,4), -- m3/s
    airflow_actual NUMERIC(12,4),
    air_velocity NUMERIC(8,4), -- m/s
    pressure_drop NUMERIC(10,4), -- Pa
    fan_count INTEGER DEFAULT 0,
    fan_total_power NUMERIC(10,2), -- kW
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tunnel_gas_sensors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    sensor_id VARCHAR(100) NOT NULL,
    zone_id UUID REFERENCES tunnel_ventilation_zones(id),
    chainage NUMERIC(12,4),
    sensor_type VARCHAR(50) NOT NULL, -- CO, NO2, CH4, H2S, O2, CO2, dust, VOC
    unit VARCHAR(20),
    current_value NUMERIC(12,4),
    threshold_low NUMERIC(12,4),
    threshold_high NUMERIC(12,4),
    alarm_status VARCHAR(50) DEFAULT 'normal', -- normal, warning, alarm, critical
    battery_level NUMERIC(5,2),
    last_calibrated_at TIMESTAMPTZ,
    calibration_due_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tunnel_ventilation_emergency (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    zone_id UUID REFERENCES tunnel_ventilation_zones(id),
    event_type VARCHAR(100), -- fire, gas_leak, power_failure, equipment_failure
    event_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    emergency_mode VARCHAR(100), -- full_exhaust, overpressure, smoke_extraction, reversal
    actions_taken TEXT,
    duration_minutes INTEGER,
    resolved_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    reported_by VARCHAR(200),
    notes TEXT
);

CREATE INDEX idx_tvz_project ON tunnel_ventilation_zones(project_id);
CREATE INDEX idx_tgs_zone ON tunnel_gas_sensors(zone_id);
CREATE INDEX idx_tgs_sensor ON tunnel_gas_sensors(sensor_id);
CREATE INDEX idx_tgs_alarm ON tunnel_gas_sensors(alarm_status) WHERE alarm_status != 'normal';
CREATE INDEX idx_tve_event ON tunnel_ventilation_emergency(event_time);