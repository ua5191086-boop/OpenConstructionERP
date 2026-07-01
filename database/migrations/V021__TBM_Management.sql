-- OpenConstructionERP
-- V021: TBM Management — real-time telemetry, alarms, operator logs, consumables, performance
-- Owner: core Go lane. Extends V004 Tunnel Module.

-- ============================================================================
-- 1. TBM Telemetry (real-time sensor data)
-- ============================================================================
CREATE TABLE IF NOT EXISTS tbm_telemetry (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tbm_id UUID NOT NULL REFERENCES tbm(id) ON DELETE CASCADE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- EPB parameters
    epb_face_pressure_bar NUMERIC(6,2),
    epb_screw_speed_rpm  NUMERIC(6,2),
    epb_screw_torque_kNm NUMERIC(8,2),
    epb_chamber_pressure_bar NUMERIC(6,2),

    -- Slurry parameters
    slurry_density_kgm3   NUMERIC(8,2),
    slurry_flow_in_m3h    NUMERIC(8,2),
    slurry_flow_out_m3h   NUMERIC(8,2),
    slurry_pressure_bar   NUMERIC(6,2),

    -- Thrust & torque
    thrust_force_kN    NUMERIC(10,2),
    thrust_speed_mmmin NUMERIC(8,2),
    torque_kNm         NUMERIC(10,2),
    torque_pct         NUMERIC(5,2),

    -- Advance
    advance_rate_mmmin NUMERIC(8,2),
    advance_mm         NUMERIC(8,2),           -- total advance since start
    face_pressure_bar  NUMERIC(6,2),           -- overall face support pressure

    -- Cutterhead
    cutterhead_rpm     NUMERIC(6,2),
    cutterhead_torque_kNm NUMERIC(10,2),
    cutterhead_wear_mm NUMERIC(6,2),

    -- Tail skin & articulation
    tail_skin_grease_bar NUMERIC(6,2),
    articulation_angle_deg NUMERIC(5,2),

    -- Miscellaneous
    belt_weight_kg     NUMERIC(10,2),
    total_power_kw     NUMERIC(10,2),
    data_source        VARCHAR(20) DEFAULT 'plc',
    is_valid           BOOLEAN DEFAULT TRUE,
    raw_payload        JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_tbm_telemetry_tbm_time ON tbm_telemetry(tbm_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_tbm_telemetry_day ON tbm_telemetry(tbm_id, (recorded_at::date));

-- ============================================================================
-- 2. TBM Alarms
-- ============================================================================
CREATE TABLE IF NOT EXISTS tbm_alarms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tbm_id UUID NOT NULL REFERENCES tbm(id) ON DELETE CASCADE,
    alarm_code VARCHAR(32) NOT NULL,
    alarm_severity VARCHAR(16) NOT NULL DEFAULT 'warning'
        CHECK (alarm_severity IN ('info','warning','critical','emergency')),
    alarm_name VARCHAR(255) NOT NULL,
    description TEXT,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by VARCHAR(128),
    cleared_at TIMESTAMPTZ,
    cleared_by VARCHAR(128),
    param_value NUMERIC(10,2),
    threshold_value NUMERIC(10,2),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tbm_alarms_active ON tbm_alarms(tbm_id, is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_tbm_alarms_time ON tbm_alarms(tbm_id, triggered_at DESC);

-- ============================================================================
-- 3. TBM Operators
-- ============================================================================
CREATE TABLE IF NOT EXISTS tbm_operators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id VARCHAR(64) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    qualification VARCHAR(100),
    certification_number VARCHAR(64),
    certification_expiry DATE,
    tbm_types VARCHAR(100),          -- comma-separated: EPB,SLURFY,MTBM
    phone VARCHAR(32),
    email VARCHAR(128),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- 4. TBM Shifts
-- ============================================================================
CREATE TABLE IF NOT EXISTS tbm_shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tbm_id UUID NOT NULL REFERENCES tbm(id) ON DELETE CASCADE,
    shift_date DATE NOT NULL,
    shift_label VARCHAR(10) NOT NULL CHECK (shift_label IN ('A','B','C','day','night')),
    operator_id UUID REFERENCES tbm_operators(id),
    assistant_id UUID REFERENCES tbm_operators(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    rings_built INTEGER DEFAULT 0,
    advance_mm INTEGER DEFAULT 0,
    downtime_minutes INTEGER DEFAULT 0,
    downtime_reason TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (tbm_id, shift_date, shift_label)
);

CREATE INDEX IF NOT EXISTS idx_tbm_shifts_date ON tbm_shifts(tbm_id, shift_date DESC);

-- ============================================================================
-- 5. TBM Consumables
-- ============================================================================
CREATE TABLE IF NOT EXISTS tbm_consumables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tbm_id UUID NOT NULL REFERENCES tbm(id) ON DELETE CASCADE,
    consumable_type VARCHAR(32) NOT NULL
        CHECK (consumable_type IN ('cutterhead','seals','foam','bentonite','grease','grout','wear_parts','hydraulic_oil','gear_oil')),
    item_name VARCHAR(255) NOT NULL,
    item_code VARCHAR(64),
    unit VARCHAR(20) NOT NULL DEFAULT 'kg',
    quantity_used NUMERIC(12,2) DEFAULT 0,
    quantity_remaining NUMERIC(12,2) DEFAULT 0,
    unit_price NUMERIC(12,2),
    used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    shift_id UUID REFERENCES tbm_shifts(id),
    recorded_by VARCHAR(128),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tbm_cons_type ON tbm_consumables(tbm_id, consumable_type, used_at DESC);

-- ============================================================================
-- 6. TBM Performance Metrics
-- ============================================================================
CREATE TABLE IF NOT EXISTS tbm_performance_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tbm_id UUID NOT NULL REFERENCES tbm(id) ON DELETE CASCADE,
    -- Daily / shift-level aggregates
    metric_date DATE NOT NULL,
    shift_label VARCHAR(10) CHECK (shift_label IN ('A','B','C','day','night')),
    rings_built INTEGER DEFAULT 0,
    advance_mm INTEGER DEFAULT 0,
    avg_advance_rate_mmmin NUMERIC(8,2),
    max_advance_rate_mmmin NUMERIC(8,2),
    avg_thrust_force_kN NUMERIC(10,2),
    max_thrust_force_kN NUMERIC(10,2),
    avg_torque_kNm NUMERIC(10,2),
    avg_face_pressure_bar NUMERIC(6,2),
    total_downtime_minutes INTEGER DEFAULT 0,
    utilisation_pct NUMERIC(5,2),     -- boring time / total shift time
    tbm_availability_pct NUMERIC(5,2),
    performance_factor NUMERIC(5,2),   -- actual / theoretical advance
    grout_volume_m3 NUMERIC(10,2),
    grout_pressure_avg_bar NUMERIC(6,2),
    cutterhead_wear_avg_mm NUMERIC(6,2),
    foam_consumption_kg NUMERIC(10,2),
    bentonite_consumption_kg NUMERIC(10,2),
    data_points INTEGER DEFAULT 0,      -- number of telemetry readings used
    calculated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (tbm_id, metric_date, shift_label)
);

CREATE INDEX IF NOT EXISTS idx_tbm_perf_date ON tbm_performance_metrics(tbm_id, metric_date DESC);

-- ============================================================================
-- Ontology entries for TBM domain
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('tbm_telemetry', 'TBM Telemetry', 'activity', 'V021'),
('tbm_alarm',     'TBM Alarm',     'alert-triangle', 'V021'),
('tbm_operator',  'TBM Operator',  'users',  'V021'),
('tbm_shift',     'TBM Shift',     'clock',  'V021'),
('tbm_consumable','TBM Consumable','package','V021'),
('tbm_perf_metric','TBM Performance','trending-up','V021')
ON CONFLICT (code) DO NOTHING;