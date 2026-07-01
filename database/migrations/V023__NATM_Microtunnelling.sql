-- OpenConstructionERP
-- V023: NATM & Microtunnelling — sequential excavation, shotcrete, rock bolts,
--       steel sets, face mapping, MTBM pipe jacking, shafts, cross passages,
--       grouting, settlement monitoring
-- Owner: core Go lane. Extends V004 Tunnel Module.

-- ============================================================================
-- NATM: Excavation Log
-- ============================================================================
CREATE TABLE IF NOT EXISTS natm_excavation_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    round_no INTEGER NOT NULL,
    chainage_from NUMERIC(10,2) NOT NULL,
    chainage_to   NUMERIC(10,2) NOT NULL,
    excavation_date DATE NOT NULL DEFAULT CURRENT_DATE,
    shift VARCHAR(10) CHECK (shift IN ('A','B','C','day','night')),
    method VARCHAR(32) NOT NULL DEFAULT 'drill_blast'
        CHECK (method IN ('drill_blast','mechanical','hydraulic_breaker','top_heading','bench','full_face')),
    round_length_m NUMERIC(5,2),
    excavated_volume_m3 NUMERIC(10,2),
    geotech_class VARCHAR(20),                 -- RMR / Q-class / GSI
    water_inflow_lmin NUMERIC(10,2),
    support_class VARCHAR(20),
    standup_time_hours NUMERIC(8,2),
    delay_minutes INTEGER DEFAULT 0,
    delay_reason TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (drive_id, round_no)
);

CREATE INDEX IF NOT EXISTS idx_natm_exc_drive ON natm_excavation_log(drive_id, round_no);

-- ============================================================================
-- NATM: Shotcrete (sprayed concrete)
-- ============================================================================
CREATE TABLE IF NOT EXISTS natm_shotcrete (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    round_id UUID REFERENCES natm_excavation_log(id),
    application_date DATE NOT NULL DEFAULT CURRENT_DATE,
    location_type VARCHAR(20) NOT NULL DEFAULT 'invert'
        CHECK (location_type IN ('invert','arch','wall','bench','face','complete')),
    shotcrete_type VARCHAR(32) NOT NULL DEFAULT 'dry_mix'
        CHECK (shotcrete_type IN ('dry_mix','wet_mix','fiber_reinforced','steel_fiber','polypropylene_fiber')),
    design_class VARCHAR(20) DEFAULT 'C25/30',
    thickness_mm NUMERIC(6,2) NOT NULL,
    area_m2 NUMERIC(10,2),
    volume_m3 NUMERIC(10,2),
    compressive_strength_mpa NUMERIC(6,2),
    fiber_content_kgm3 NUMERIC(8,2),
    accelerator_type VARCHAR(64),
    accelerator_dosage_pct NUMERIC(5,2),
    rebound_pct NUMERIC(5,2),
    application_temp_c NUMERIC(5,1),
    nozzleman VARCHAR(128),
    qc_status VARCHAR(20) DEFAULT 'pending'
        CHECK (qc_status IN ('pending','passed','failed','retest')),
    test_core_28d_mpa NUMERIC(6,2),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_natm_shotcrete_drive ON natm_shotcrete(drive_id);

-- ============================================================================
-- NATM: Rock Bolts
-- ============================================================================
CREATE TABLE IF NOT EXISTS natm_rock_bolts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    round_id UUID REFERENCES natm_excavation_log(id),
    bolt_type VARCHAR(32) NOT NULL DEFAULT 'expansion'
        CHECK (bolt_type IN ('expansion','resin','mechanical','swellex','self_drilling','friction','tensioned','grouted')),
    bolt_diameter_mm INTEGER NOT NULL,
    bolt_length_mm INTEGER NOT NULL,
    bolt_grade VARCHAR(20) DEFAULT '500',
    spacing_longitudinal_m NUMERIC(5,2),
    spacing_transverse_m NUMERIC(5,2),
    quantity_installed INTEGER NOT NULL DEFAULT 0,
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pretension_kN NUMERIC(8,2),
    pullout_test_kN NUMERIC(8,2),
    grout_volume_l NUMERIC(8,2),
    pattern_type VARCHAR(20) DEFAULT 'systematic'
        CHECK (pattern_type IN ('systematic','spot','pattern','random')),
    installed_by VARCHAR(128),
    qc_status VARCHAR(20) DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_natm_bolts_drive ON natm_rock_bolts(drive_id);

-- ============================================================================
-- NATM: Steel Sets
-- ============================================================================
CREATE TABLE IF NOT EXISTS natm_steel_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    round_id UUID REFERENCES natm_excavation_log(id),
    set_number INTEGER NOT NULL,
    chainage NUMERIC(10,2),
    set_type VARCHAR(32) NOT NULL DEFAULT 'TH'
        CHECK (set_type IN ('TH','TH-44','TH-58','GRI','HEB','IPE','Lattice_girder','UMC')),
    section_name VARCHAR(64),
    spacing_m NUMERIC(5,2),
    steel_grade VARCHAR(20) DEFAULT 'S355',
    weight_kg_m NUMERIC(8,2),
    quantity_arches INTEGER NOT NULL DEFAULT 1,
    installed_at TIMESTAMPTZ,
    connection_type VARCHAR(32) DEFAULT 'bolted',
    lagging_type VARCHAR(64),
    qc_status VARCHAR(20) DEFAULT 'pending',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (drive_id, set_number)
);

CREATE INDEX IF NOT EXISTS idx_natm_steel_drive ON natm_steel_sets(drive_id);

-- ============================================================================
-- NATM: Convergence Monitoring
-- ============================================================================
CREATE TABLE IF NOT EXISTS natm_convergence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    round_id UUID REFERENCES natm_excavation_log(id),
    measurement_point VARCHAR(64) NOT NULL,
    chainage NUMERIC(10,2),
    measured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Displacement (mm)
    displacement_vertical_mm NUMERIC(8,2),
    displacement_horizontal_mm NUMERIC(8,2),
    displacement_longitudinal_mm NUMERIC(8,2),
    -- Convergence rate
    convergence_rate_mmday NUMERIC(8,4),
    cumulative_displacement_mm NUMERIC(8,2),
    -- Instrument
    instrument_type VARCHAR(32) DEFAULT 'total_station'
        CHECK (instrument_type IN ('total_station','extensometer','convergence_tape','inclinometer','laser','photogrammetry','radar')),
    distance_from_face_m NUMERIC(6,2),
    temperature_c NUMERIC(5,1),
    alarm_triggered BOOLEAN DEFAULT FALSE,
    reading_by VARCHAR(128),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_natm_conv_drive ON natm_convergence(drive_id, chainage);
CREATE INDEX IF NOT EXISTS idx_natm_conv_point ON natm_convergence(measurement_point, measured_at DESC);

-- ============================================================================
-- NATM: Face Mapping (geological face logging)
-- ============================================================================
CREATE TABLE IF NOT EXISTS natm_face_mapping (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    round_id UUID REFERENCES natm_excavation_log(id),
    chainage NUMERIC(10,2),
    mapped_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    rock_type VARCHAR(100),
    weathering_grade VARCHAR(20),               -- I-VI / fresh to residual
    rmr_score NUMERIC(5,1),                    -- Rock Mass Rating
    q_score NUMERIC(8,2),                      -- Q-value (Barton)
    gsi_value NUMERIC(5,1),                    -- Geological Strength Index
    joint_count INTEGER,
    joint_spacing_m NUMERIC(5,2),
    joint_condition VARCHAR(100),
    groundwater_condition VARCHAR(50),
    fault_zone BOOLEAN DEFAULT FALSE,
    water_inflow_estimated_lmin NUMERIC(10,2),
    standup_time_est_hours NUMERIC(8,2),
    support_recommendation TEXT,
    photo_url VARCHAR(512),
    mapped_by VARCHAR(128),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_natm_face_drive ON natm_face_mapping(drive_id, chainage);

-- ============================================================================
-- Microtunnelling: MTBM Drives
-- ============================================================================
CREATE TABLE IF NOT EXISTS mtbm_drives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    drive_code VARCHAR(50) NOT NULL,
    drive_name VARCHAR(255) NOT NULL,
    mtbm_id VARCHAR(64),
    pipe_type VARCHAR(32) DEFAULT 'concrete'
        CHECK (pipe_type IN ('concrete','steel','ductile_iron','HDPE','GRP','vitrified_clay')),
    pipe_diameter_mm INTEGER NOT NULL,
    pipe_length_mm INTEGER DEFAULT 3000,
    wall_thickness_mm INTEGER,
    design_length_m NUMERIC(10,2),
    max_jacking_force_kN NUMERIC(10,2),
    intermediate_jack_stations INTEGER DEFAULT 0,
    lubrication_type VARCHAR(32) DEFAULT 'bentonite',
    chainage_from NUMERIC(10,2),
    chainage_to NUMERIC(10,2),
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','pipeline_installation','jacking','intermediate','breakthrough','completed','abandoned')),
    start_date DATE,
    breakthrough_date DATE,
    contractor VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, drive_code)
);

-- ============================================================================
-- Microtunnelling: Thrust Log
-- ============================================================================
CREATE TABLE IF NOT EXISTS mtbm_thrust_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mtbm_drive_id UUID NOT NULL REFERENCES mtbm_drives(id) ON DELETE CASCADE,
    pipe_no INTEGER NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    thrust_force_kN NUMERIC(10,2),
    thrust_pressure_bar NUMERIC(6,2),
    push_ram_extent_mm INTEGER,
    advance_speed_mmmin NUMERIC(8,2),
    torque_kNm NUMERIC(10,2),
    torque_pct NUMERIC(5,2),
    slurry_pressure_bar NUMERIC(6,2),
    slurry_flow_m3h NUMERIC(8,2),
    face_pressure_bar NUMERIC(6,2),
    penetration_rate_mm_min NUMERIC(8,2),
    alignment_vertical_mm NUMERIC(6,2),         -- deviation
    alignment_horizontal_mm NUMERIC(6,2),
    rod_count INTEGER,
    intermediate_jack_force_kN NUMERIC(10,2),
    water_inflow_lmin NUMERIC(10,2),
    notes TEXT,
    UNIQUE (mtbm_drive_id, pipe_no)
);

CREATE INDEX IF NOT EXISTS idx_mtbm_thrust_drive ON mtbm_thrust_log(mtbm_drive_id, pipe_no);

-- ============================================================================
-- Microtunnelling: Lubrication
-- ============================================================================
CREATE TABLE IF NOT EXISTS mtbm_lubrication (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mtbm_drive_id UUID NOT NULL REFERENCES mtbm_drives(id) ON DELETE CASCADE,
    pipe_no INTEGER,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lubricant_type VARCHAR(32) NOT NULL DEFAULT 'bentonite'
        CHECK (lubricant_type IN ('bentonite','polymer','foam','combined','other')),
    injection_pressure_bar NUMERIC(6,2),
    flow_rate_lmin NUMERIC(10,2),
    total_volume_m3 NUMERIC(10,2),
    density_kgm3 NUMERIC(8,2),
    viscosity_cp NUMERIC(10,2),
    marsh_viscosity_sec NUMERIC(6,2),
    filtrate_loss_ml NUMERIC(6,2),
    ph_level NUMERIC(4,1),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mtbm_lube_drive ON mtbm_lubrication(mtbm_drive_id, pipe_no);

-- ============================================================================
-- Microtunnelling: Survey
-- ============================================================================
CREATE TABLE IF NOT EXISTS mtbm_survey (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mtbm_drive_id UUID NOT NULL REFERENCES mtbm_drives(id) ON DELETE CASCADE,
    pipe_no INTEGER NOT NULL,
    surveyed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- As-built position
    northing_m NUMERIC(12,4),
    easting_m NUMERIC(12,4),
    elevation_m NUMERIC(10,4),
    chainage_m NUMERIC(10,2),
    -- Deviation from design
    deviation_vertical_mm NUMERIC(8,2),
    deviation_horizontal_mm NUMERIC(8,2),
    deviation_roll_deg NUMERIC(6,2),
    deviation_yaw_deg NUMERIC(6,2),
    deviation_pitch_deg NUMERIC(6,2),
    -- Target
    target_northing_m NUMERIC(12,4),
    target_easting_m NUMERIC(12,4),
    target_elevation_m NUMERIC(10,4),
    instrument_type VARCHAR(32) DEFAULT 'gyro',
    survey_by VARCHAR(128),
    adjustment_applied VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mtbm_surv_drive ON mtbm_survey(mtbm_drive_id, pipe_no);

-- ============================================================================
-- Shaft Construction
-- ============================================================================
CREATE TABLE IF NOT EXISTS shaft_construction (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    shaft_code VARCHAR(50) NOT NULL,
    shaft_name VARCHAR(255) NOT NULL,
    shaft_type VARCHAR(32) NOT NULL DEFAULT 'launch'
        CHECK (shaft_type IN ('launch','reception','intermediate','ventilation','access','pumping')),
    construction_method VARCHAR(32) DEFAULT 'secant_pile'
        CHECK (construction_method IN ('secant_pile','sheet_pile','diaphragm_wall','caisson','cut_and_cover','soldier_pile','sprayed_concrete')),
    diameter_m NUMERIC(8,2),
    depth_m NUMERIC(8,2),
    wall_thickness_mm INTEGER,
    excavation_method VARCHAR(32) DEFAULT 'mechanical',
    support_system VARCHAR(64),
    dewatering_method VARCHAR(64),
    chainage NUMERIC(10,2),
    start_date DATE,
    completion_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','excavation','support','base_slab','equipment_install','completed','backfilled')),
    contractor VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, shaft_code)
);

-- ============================================================================
-- Shaft Equipment Installation
-- ============================================================================
CREATE TABLE IF NOT EXISTS shaft_equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shaft_id UUID NOT NULL REFERENCES shaft_construction(id) ON DELETE CASCADE,
    equipment_type VARCHAR(64) NOT NULL,
    equipment_name VARCHAR(255) NOT NULL,
    manufacturer VARCHAR(128),
    model VARCHAR(128),
    serial_number VARCHAR(128),
    installed_at TIMESTAMPTZ,
    commissioned_at TIMESTAMPTZ,
    rated_capacity VARCHAR(64),
    power_kw NUMERIC(8,2),
    weight_kg NUMERIC(10,2),
    location_in_shaft VARCHAR(128),
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','installed','commissioned','operational','maintenance','decommissioned')),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shaft_eqp_shaft ON shaft_equipment(shaft_id);

-- ============================================================================
-- Cross Passages
-- ============================================================================
CREATE TABLE IF NOT EXISTS cross_passages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    passage_code VARCHAR(50) NOT NULL,
    passage_name VARCHAR(255) NOT NULL,
    chainage NUMERIC(10,2),
    construction_method VARCHAR(32) DEFAULT 'NATM'
        CHECK (construction_method IN ('NATM','SCL','drill_blast','pipe_jacking','precast')),
    length_m NUMERIC(8,2),
    width_m NUMERIC(6,2),
    height_m NUMERIC(6,2),
    lining_type VARCHAR(64) DEFAULT 'shotcrete',
    water_proofing VARCHAR(64),
    start_date DATE,
    completion_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','excavation','lining','waterproofing','completion','completed')),
    contractor VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, passage_code)
);

-- ============================================================================
-- Grouting Records
-- ============================================================================
CREATE TABLE IF NOT EXISTS grouting_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    grouting_type VARCHAR(32) NOT NULL DEFAULT 'contact'
        CHECK (grouting_type IN ('contact','void','consolidation','curtain','compensation','backfill','annulus','pre_excavation')),
    location_type VARCHAR(20) DEFAULT 'TBM'
        CHECK (location_type IN ('TBM','NATM','shaft','cross_passage','MTBM','other')),
    location_id UUID,                            -- flexible: tbm_id / drive_id / shaft_id / passage_id
    chainage NUMERIC(10,2),
    grout_date DATE NOT NULL DEFAULT CURRENT_DATE,
    grout_mix_type VARCHAR(64) DEFAULT 'cement_bentonite',
    grout_density_kgm3 NUMERIC(8,2),
    wc_ratio NUMERIC(5,2),
    pressure_bar NUMERIC(6,2),
    flow_rate_lmin NUMERIC(10,2),
    volume_planned_m3 NUMERIC(10,2),
    volume_actual_m3 NUMERIC(10,2),
    injection_point VARCHAR(128),
    number_of_holes INTEGER,
    spacing_m NUMERIC(5,2),
    refusal_pressure_bar NUMERIC(6,2),
    take_kgm NUMERIC(10,2),
    supervisor VARCHAR(128),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_grouting_project ON grouting_records(project_id, grouting_type);
CREATE INDEX IF NOT EXISTS idx_grouting_chainage ON grouting_records(chainage);

-- ============================================================================
-- Settlement Monitoring
-- ============================================================================
CREATE TABLE IF NOT EXISTS settlement_monitoring (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    point_id VARCHAR(64) NOT NULL,
    point_type VARCHAR(32) NOT NULL DEFAULT 'surface'
        CHECK (point_type IN ('surface','subsurface','building','utility','pavement','track','bridge')),
    northing_m NUMERIC(12,4),
    easting_m NUMERIC(12,4),
    elevation_ref_m NUMERIC(10,4),
    chainage NUMERIC(10,2),
    offset_m NUMERIC(8,2),                     -- distance from tunnel centerline
    monitored_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settlement_mm NUMERIC(8,2),                -- positive = settlement, negative = heave
    cumulative_settlement_mm NUMERIC(8,2),
    settlement_rate_mmday NUMERIC(8,4),
    horizontal_displacement_mm NUMERIC(8,2),
    tilt_ratio NUMERIC(8,6),                   -- angular distortion
    strain_micron NUMERIC(10,2),
    instrument_type VARCHAR(32) DEFAULT 'leveling'
        CHECK (instrument_type IN ('leveling','total_station','inclinometer','piezometer','extensometer','tiltmeter','crack_gauges','insar','lidar')),
    reading_accuracy_mm NUMERIC(6,3),
    alarm_threshold_mm NUMERIC(8,2),
    alarm_triggered BOOLEAN DEFAULT FALSE,
    reading_by VARCHAR(128),
    weather_conditions VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_settlement_project ON settlement_monitoring(project_id, point_id, monitored_at DESC);
CREATE INDEX IF NOT EXISTS idx_settlement_chainage ON settlement_monitoring(chainage, monitored_at DESC);
CREATE INDEX IF NOT EXISTS idx_settlement_alarm ON settlement_monitoring(project_id, alarm_triggered) WHERE alarm_triggered = TRUE;

-- ============================================================================
-- Ontology entries
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('natm_excavation',  'NATM Excavation',    'drill',        'V023'),
('natm_shotcrete',   'NATM Shotcrete',     'droplets',     'V023'),
('natm_rock_bolt',   'Rock Bolt',          'slash',        'V023'),
('natm_steel_set',   'Steel Set',          'grid',         'V023'),
('natm_convergence', 'NATM Convergence',   'activity',     'V023'),
('natm_face_mapping','Face Mapping',       'map',          'V023'),
('mtbm_drive',       'MTBM Drive',         'cpu',          'V023'),
('mtbm_thrust',      'MTBM Thrust',        'arrow-up',     'V023'),
('mtbm_lubrication',  'MTBM Lubrication',  'droplet',      'V023'),
('mtbm_survey',      'MTBM Survey',        'crosshair',    'V023'),
('shaft_construct',  'Shaft Construction', 'layers',       'V023'),
('shaft_equipment',  'Shaft Equipment',    'tool',         'V023'),
('cross_passage',    'Cross Passage',      'git-branch',   'V023'),
('grouting',         'Grouting Record',    'inject',       'V023'),
('settlement',       'Settlement Monitor', 'trending-down','V023')
ON CONFLICT (code) DO NOTHING;