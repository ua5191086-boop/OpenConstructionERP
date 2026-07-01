-- OpenConstructionERP
-- V022: Ring Builder & Segment Tracking — ring design, segment production→curing→transport→install, QC, convergence
-- Owner: core Go lane. Extends V004 Tunnel Module.

-- ============================================================================
-- 1. Ring Designs
-- ============================================================================
CREATE TABLE IF NOT EXISTS ring_designs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    design_code VARCHAR(50) NOT NULL,
    design_name VARCHAR(255) NOT NULL,
    ring_type VARCHAR(32) NOT NULL DEFAULT 'universal'
        CHECK (ring_type IN ('universal','tapered','straight','bolted','key','special')),
    inner_diameter_mm INTEGER NOT NULL,
    outer_diameter_mm INTEGER NOT NULL,
    ring_width_mm INTEGER NOT NULL,
    taper_mm INTEGER DEFAULT 0,
    segment_count INTEGER NOT NULL,
    key_position VARCHAR(10) DEFAULT 'K',     -- замковый блок
    segment_mapping JSONB DEFAULT '{}'::jsonb, -- позиции сегментов и типы
    concrete_grade VARCHAR(20) DEFAULT 'C45/55',
    reinforcement_type VARCHAR(20) DEFAULT 'steel',
    weight_kg NUMERIC(10,2),
    design_file_url VARCHAR(512),
    revision VARCHAR(20) DEFAULT 'A',
    status VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','archived','superseded')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, design_code)
);

CREATE INDEX IF NOT EXISTS idx_ring_designs_project ON ring_designs(project_id);

-- ============================================================================
-- 2. Segment Production (casting yard)
-- ============================================================================
CREATE TABLE IF NOT EXISTS segment_production (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    design_id UUID REFERENCES ring_designs(id),
    segment_code VARCHAR(64) NOT NULL UNIQUE,   -- e.g. SEG-01-A1
    segment_type VARCHAR(10) NOT NULL,           -- A1..A6, B, K
    ring_designation VARCHAR(10),               -- ring number reference
    mold_id VARCHAR(64),
    cast_batch VARCHAR(64),                      -- бетонная партия
    concrete_grade VARCHAR(20),
    concrete_volume_m3 NUMERIC(8,2),
    steel_weight_kg NUMERIC(8,2),
    cast_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cast_by VARCHAR(128),
    curing_start_at TIMESTAMPTZ,
    curing_end_at TIMESTAMPTZ,
    curing_method VARCHAR(32) DEFAULT 'steam'
        CHECK (curing_method IN ('steam','water','air','accelerated')),
    curing_temp_c NUMERIC(5,1),
    demold_at TIMESTAMPTZ,
    transport_at TIMESTAMPTZ,
    install_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'cast'
        CHECK (status IN ('cast','curing','demolded','transport','in_stock','installed','rejected','quarantine')),
    qc_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (qc_status IN ('pending','passed','conditional','failed')),
    qr_code VARCHAR(128) UNIQUE,
    rfid_tag VARCHAR(64),
    location VARCHAR(255),                       -- склад / транспорт / забои
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seg_prod_status ON segment_production(project_id, status);
CREATE INDEX IF NOT EXISTS idx_seg_prod_batch ON segment_production(cast_batch);
CREATE INDEX IF NOT EXISTS idx_seg_prod_qr ON segment_production(qr_code);

-- ============================================================================
-- 3. Segment Curing Records
-- ============================================================================
CREATE TABLE IF NOT EXISTS segment_curing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL REFERENCES segment_production(id) ON DELETE CASCADE,
    curing_stage VARCHAR(20) NOT NULL
        CHECK (curing_stage IN ('initial_set','steam','cooling','final_set','demold')),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    temp_target_c NUMERIC(5,1),
    temp_actual_c NUMERIC(5,1),
    humidity_target_pct NUMERIC(5,1),
    humidity_actual_pct NUMERIC(5,1),
    gradient_rate_cph NUMERIC(5,2),             -- °C/hour heating/cooling
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seg_curing_seg ON segment_curing(segment_id);

-- ============================================================================
-- 4. Segment Transport
-- ============================================================================
CREATE TABLE IF NOT EXISTS segment_transport (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL REFERENCES segment_production(id) ON DELETE CASCADE,
    transport_date DATE NOT NULL,
    transport_mode VARCHAR(32) NOT NULL DEFAULT 'truck'
        CHECK (transport_mode IN ('truck','train','flatbed','crane','gantry','other')),
    vehicle_number VARCHAR(64),
    driver_name VARCHAR(128),
    from_location VARCHAR(255) NOT NULL,
    to_location VARCHAR(255) NOT NULL,
    departure_time TIMESTAMPTZ,
    arrival_time TIMESTAMPTZ,
    distance_km NUMERIC(8,2),
    damage_reported BOOLEAN DEFAULT FALSE,
    damage_notes TEXT,
    temperature_c NUMERIC(5,1),
    transport_cost NUMERIC(12,2),
    created_by VARCHAR(128),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seg_transport_seg ON segment_transport(segment_id);

-- ============================================================================
-- 5. Segment Installation (in tunnel)
-- ============================================================================
CREATE TABLE IF NOT EXISTS segment_installation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL REFERENCES segment_production(id) ON DELETE CASCADE,
    ring_id UUID REFERENCES tunnel_rings(id),
    erector_cycle_time_sec INTEGER,           -- время установки эректором
    bolt_count INTEGER DEFAULT 0,
    bolt_torque_nm NUMERIC(8,2),
    packer_type VARCHAR(64),
    gap_mm NUMERIC(5,1),                      -- зазор между сегментами
    offset_radial_mm NUMERIC(5,1),            -- радиальное смещение
    offset_longitudinal_mm NUMERIC(5,1),       -- продольное смещение
    installed_by VARCHAR(128),
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seg_install_ring ON segment_installation(ring_id);
CREATE INDEX IF NOT EXISTS idx_seg_install_seg ON segment_installation(segment_id);

-- ============================================================================
-- 6. Segment QC (Quality Control)
-- ============================================================================
CREATE TABLE IF NOT EXISTS segment_qc (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL REFERENCES segment_production(id) ON DELETE CASCADE,
    -- Dimensional checks
    length_mm NUMERIC(6,2),
    width_mm NUMERIC(6,2),
    thickness_mm NUMERIC(6,2),
    diagonal_diff_mm NUMERIC(6,2),
    -- Surface
    surface_defects VARCHAR(255),
    honeycomb_pct NUMERIC(5,1),
    spalling_pct NUMERIC(5,1),
    cracking VARCHAR(255),
    -- Reinforcement cover
    cover_min_mm NUMERIC(5,1),
    cover_max_mm NUMERIC(5,1),
    -- Strength
    compressive_strength_mpa NUMERIC(6,2),
    water_absorption_pct NUMERIC(5,2),
    -- Embedded items
    bolt_socket_present BOOLEAN DEFAULT TRUE,
    lifting_anchor_present BOOLEAN DEFAULT TRUE,
    -- Overall
    qc_result VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (qc_result IN ('pending','pass','conditional_pass','fail','rework')),
    qc_inspector VARCHAR(128),
    qc_date DATE NOT NULL DEFAULT CURRENT_DATE,
    corrective_action TEXT,
    retest_result VARCHAR(20),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_seg_qc_seg ON segment_qc(segment_id);
CREATE INDEX IF NOT EXISTS idx_seg_qc_result ON segment_qc(qc_result);

-- ============================================================================
-- 7. Segment Inventory
-- ============================================================================
CREATE TABLE IF NOT EXISTS segment_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    design_id UUID REFERENCES ring_designs(id),
    segment_type VARCHAR(10) NOT NULL,
    quantity_planned INTEGER NOT NULL DEFAULT 0,
    quantity_produced INTEGER NOT NULL DEFAULT 0,
    quantity_passed_qc INTEGER NOT NULL DEFAULT 0,
    quantity_installed INTEGER NOT NULL DEFAULT 0,
    quantity_defective INTEGER NOT NULL DEFAULT 0,
    quantity_in_transit INTEGER NOT NULL DEFAULT 0,
    quantity_in_stock INTEGER GENERATED ALWAYS AS (quantity_passed_qc - quantity_installed - quantity_defective - quantity_in_transit) STORED,
    stock_location VARCHAR(255),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, design_id, segment_type)
);

CREATE INDEX IF NOT EXISTS idx_seg_inv_project ON segment_inventory(project_id);

-- ============================================================================
-- 8. Ring Measurements (convergence, ovality, deformation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS ring_measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ring_id UUID NOT NULL REFERENCES tunnel_rings(id) ON DELETE CASCADE,
    measured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Convergence (horizontal & vertical displacement, mm)
    horizontal_convergence_mm NUMERIC(8,2),
    vertical_convergence_mm NUMERIC(8,2),
    diagonal_1_mm NUMERIC(8,2),
    diagonal_2_mm NUMERIC(8,2),
    -- Ovality
    ovality_pct NUMERIC(6,3),                  -- (Dmax - Dmin) / Ddesign * 100
    ovality_mm NUMERIC(8,2),
    -- Deformation
    deformation_vertical_mm NUMERIC(8,2),
    deformation_horizontal_mm NUMERIC(8,2),
    settlement_mm NUMERIC(8,2),
    -- Profile
    profile_chainage NUMERIC(10,2),
    section_area_loss_pct NUMERIC(6,3),
    -- Monitoring
    instrument_type VARCHAR(64) DEFAULT 'total_station'
        CHECK (instrument_type IN ('total_station','laser_scanner','photogrammetry','convergence_tape','distometer')),
    measured_by VARCHAR(128),
    weather_conditions VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ring_meas_ring ON ring_measurements(ring_id, measured_at DESC);

-- ============================================================================
-- Ontology entries
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('ring_design',       'Ring Design',           'ruler',           'V022'),
('segment_production','Segment Production',    'factory',         'V022'),
('segment_curing',    'Segment Curing',        'thermometer',     'V022'),
('segment_transport', 'Segment Transport',     'truck',           'V022'),
('segment_install',   'Segment Installation',  'tool',            'V022'),
('segment_qc',        'Segment QC',            'check-circle',    'V022'),
('segment_inventory', 'Segment Inventory',     'archive',         'V022'),
('ring_measurement',  'Ring Measurement',      'activity',        'V022')
ON CONFLICT (code) DO NOTHING;