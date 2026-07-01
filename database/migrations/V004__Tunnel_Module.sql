-- OpenConstructionERP
-- V004: Tunnel Module (minimum viable per SAD Tom 1 §6.3: L-01 TBM Operations,
--       L-03 Ring Register, L-04 Segment Tracking)
-- Owner: core-py lane. Registered in docs/WORKSTREAMS.md.

CREATE TABLE tbm (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,                  -- 'TBM-01'
    manufacturer VARCHAR(100),
    model VARCHAR(100),
    tbm_type VARCHAR(20) NOT NULL DEFAULT 'EPB'
        CHECK (tbm_type IN ('EPB','SLURRY','OPEN','MIXSHIELD','GRIPPER','MTBM')),
    diameter_mm INTEGER NOT NULL,
    commissioning_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'assembly'
        CHECK (status IN ('assembly','boring','standby','maintenance','breakthrough','demobilized')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, code)
);

CREATE TABLE tunnel_drives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tbm_id UUID REFERENCES tbm(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    method VARCHAR(20) NOT NULL DEFAULT 'TBM'
        CHECK (method IN ('TBM','NATM','DRILL_BLAST','MTBM','PIPE_JACKING')),
    chainage_from NUMERIC(10,2) NOT NULL,       -- пикетаж, м
    chainage_to   NUMERIC(10,2) NOT NULL,
    ring_width_mm INTEGER DEFAULT 1500,         -- ширина кольца
    design_rings INTEGER,                       -- проектное число колец
    status VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','boring','suspended','breakthrough','completed')),
    started_at DATE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, code)
);

CREATE INDEX idx_drives_project ON tunnel_drives(project_id);

CREATE TABLE tunnel_rings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drive_id UUID NOT NULL REFERENCES tunnel_drives(id) ON DELETE CASCADE,
    ring_no INTEGER NOT NULL,
    chainage NUMERIC(10,2),
    built_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    shift VARCHAR(10) CHECK (shift IN ('day','night','A','B','C')),
    ring_type VARCHAR(20) DEFAULT 'universal',
    key_position SMALLINT,                      -- позиция замкового блока (часовая, 1-12)
    advance_mm INTEGER,                         -- ход за кольцо
    grout_volume_m3 NUMERIC(8,2),
    grout_pressure_bar NUMERIC(6,2),
    attitude JSONB DEFAULT '{}'::jsonb,         -- крен/тангаж/отклонения от оси, мм
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (drive_id, ring_no)
);

CREATE INDEX idx_rings_drive ON tunnel_rings(drive_id, ring_no);
CREATE INDEX idx_rings_built ON tunnel_rings(drive_id, built_at);

CREATE TABLE tunnel_segments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    segment_type VARCHAR(10) NOT NULL,          -- A1..A6, B, K
    qr_code VARCHAR(64) UNIQUE,
    cast_at TIMESTAMPTZ,
    cast_batch VARCHAR(50),                     -- партия бетона / паспорт
    qc_status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (qc_status IN ('pending','passed','rejected','repaired')),
    ring_id UUID REFERENCES tunnel_rings(id),   -- NULL пока на складе
    position_in_ring SMALLINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_segments_ring ON tunnel_segments(ring_id);
CREATE INDEX idx_segments_qc ON tunnel_segments(project_id, qc_status);

-- Ontology object types for the tunnel domain
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('tbm',          'TBM',           'drill',  'L-01'),
('tunnel_drive', 'Tunnel Drive',  'route',  'L-01'),
('ring',         'Lining Ring',   'circle', 'L-03'),
('segment',      'Lining Segment','box',    'L-04')
ON CONFLICT (code) DO NOTHING;
