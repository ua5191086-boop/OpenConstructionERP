-- ============================================================================
-- V038__Cross_Passage_Geology_Module.sql
-- Модули Cross Passage + Geology — пикеты, калотты, геология
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- PART 1: Cross Passage
-- ============================================================================
CREATE TABLE cross_passages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    cp_code         VARCHAR(30) NOT NULL,
    cp_name         VARCHAR(300) NOT NULL,
    chainage_m      NUMERIC(10,2) NOT NULL,                    -- пикет оси перехода
    tunnel_pair     VARCHAR(100),                               -- между какими тоннелями
    cross_section   VARCHAR(50) DEFAULT 'circular'
        CHECK (cross_section IN ('circular','horseshoe','rectangular','elliptical')),
    span_m          NUMERIC(8,2),                               -- ширина/диаметр
    height_m        NUMERIC(8,2),
    length_m        NUMERIC(8,2),                               -- длина перехода
    excavation_method VARCHAR(50) DEFAULT 'sequential'
        CHECK (excavation_method IN ('sequential','drill_blast','mechanical','pipe_roof','cut_and_cover')),
    lining_type     VARCHAR(50) DEFAULT 'shotcrete'
        CHECK (lining_type IN ('shotcrete','cast_in_situ','segmental','steel')),
    waterproofing   VARCHAR(100),
    ground_treatment VARCHAR(100),
    status          VARCHAR(30) NOT NULL DEFAULT 'design'
        CHECK (status IN ('design','ground_treatment','excavation','lining','waterproofing','backfill','completed','abandoned')),
    start_date      DATE,
    end_date        DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, cp_code)
);
CREATE INDEX idx_cp_project ON cross_passages(project_id);
CREATE INDEX idx_cp_chainage ON cross_passages(chainage_m);
COMMENT ON TABLE cross_passages IS 'Пикеты / Cross Passages между тоннелями';

CREATE TABLE cp_construction_stages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cp_id           UUID NOT NULL REFERENCES cross_passages(id) ON DELETE CASCADE,
    stage_number    INTEGER NOT NULL,
    stage_name      VARCHAR(300) NOT NULL,
    stage_type      VARCHAR(50) NOT NULL DEFAULT 'excavation'
        CHECK (stage_type IN ('ground_treatment','excavation','support','lining','waterproofing','testing','backfill')),
    volume_m3       NUMERIC(12,2),
    concrete_m3     NUMERIC(12,2),
    reinforcement_kg NUMERIC(12,2),
    start_date      DATE,
    end_date        DATE,
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','delayed')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cp_id, stage_number)
);
CREATE INDEX idx_cp_stage ON cp_construction_stages(cp_id);
COMMENT ON TABLE cp_construction_stages IS 'Этапы строительства пикета';

-- ============================================================================
-- PART 2: Geology
-- ============================================================================
CREATE TABLE geology_units (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    unit_code       VARCHAR(30) NOT NULL,
    unit_name       VARCHAR(300) NOT NULL,
    geology_type    VARCHAR(50) NOT NULL DEFAULT 'soil'
        CHECK (geology_type IN ('soil','rock','mixed','fill','water')),
    soil_class      VARCHAR(50),                                -- USC / ASTM classification
    rock_class      VARCHAR(50),                                -- RMR / Q / GSI
    description     TEXT,
    color           VARCHAR(30),
    density_kg_m3   NUMERIC(10,2),
    cohesion_kPa    NUMERIC(10,2),
    friction_angle  NUMERIC(5,2),
    modulus_mpa     NUMERIC(10,2),
    permeability_ms NUMERIC(12,8),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, unit_code)
);
CREATE INDEX idx_geo_unit_project ON geology_units(project_id);
COMMENT ON TABLE geology_units IS 'Геологические единицы / слои';

CREATE TABLE geology_boreholes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    borehole_code   VARCHAR(30) NOT NULL,
    location_name   VARCHAR(300),
    chainage_m      NUMERIC(10,2),
    offset_m        NUMERIC(8,2),                               -- смещение от оси
    lat             NUMERIC(10,7),
    lng             NUMERIC(10,7),
    ground_level_m  NUMERIC(8,2),
    total_depth_m   NUMERIC(8,2),
    water_table_m   NUMERIC(8,2),                               -- уровень грунтовых вод
    drilling_method VARCHAR(100),
    drilling_date   DATE,
    contractor      VARCHAR(200),
    log_file_path   VARCHAR(1000),
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','drilling','logging','testing','completed')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, borehole_code)
);
CREATE INDEX idx_geo_bh_project ON geology_boreholes(project_id);
CREATE INDEX idx_geo_bh_chainage ON geology_boreholes(chainage_m);
COMMENT ON TABLE geology_boreholes IS 'Геологические скважины';

CREATE TABLE geology_stratigraphy (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    borehole_id     UUID NOT NULL REFERENCES geology_boreholes(id) ON DELETE CASCADE,
    unit_id         UUID REFERENCES geology_units(id) ON DELETE SET NULL,
    depth_from_m    NUMERIC(8,2) NOT NULL,
    depth_to_m      NUMERIC(8,2) NOT NULL,
    thickness_m     NUMERIC(8,2) GENERATED ALWAYS AS (depth_to_m - depth_from_m) STORED,
    description     TEXT,
    sample_type     VARCHAR(50),
    spT_value       NUMERIC(6,1),                               -- SPT N-value
    rqd_pct         NUMERIC(5,1),                               -- Rock Quality Designation
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (borehole_id, depth_from_m)
);
CREATE INDEX idx_geo_strat_bh ON geology_stratigraphy(borehole_id);
COMMENT ON TABLE geology_stratigraphy IS 'Стратиграфия / разрезы скважин';

CREATE TABLE geology_face_mapping (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    mapping_date    DATE NOT NULL,
    chainage_from_m NUMERIC(10,2) NOT NULL,
    chainage_to_m   NUMERIC(10,2),
    unit_id         UUID REFERENCES geology_units(id),
    rock_class      VARCHAR(50),
    weathering      VARCHAR(50) CHECK (weathering IN ('fresh','slightly','moderately','highly','completely','residual')),
    fracture_count  INTEGER,                                     -- per meter
    water_inflow    VARCHAR(50) CHECK (water_inflow IN ('none','damp','dripping','flowing','gushing')),
    stand_up_time   VARCHAR(50),
    mapping_by      VARCHAR(200),
    photo_paths     JSONB DEFAULT '[]'::jsonb,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_geo_face_project ON geology_face_mapping(project_id);
CREATE INDEX idx_geo_face_chainage ON geology_face_mapping(chainage_from_m);
COMMENT ON TABLE geology_face_mapping IS 'Геологическая карта забоя (калотта)';

CREATE VIEW geology_summary AS
SELECT
    p.id AS project_id,
    COUNT(DISTINCT gb.id) AS boreholes,
    COUNT(DISTINCT gu.id) AS geology_units,
    COUNT(DISTINCT gfm.id) AS face_mappings,
    COUNT(DISTINCT cp.id) AS cross_passages,
    COUNT(DISTINCT cps.id) FILTER (WHERE cps.status = 'in_progress') AS active_cp_stages
FROM projects p
LEFT JOIN geology_boreholes gb ON gb.project_id = p.id
LEFT JOIN geology_units gu ON gu.project_id = p.id
LEFT JOIN geology_face_mapping gfm ON gfm.project_id = p.id
LEFT JOIN cross_passages cp ON cp.project_id = p.id
LEFT JOIN cp_construction_stages cps ON cps.cp_id = cp.id
GROUP BY p.id;

COMMENT ON VIEW geology_summary IS 'Сводка по геологии и пикетам';

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('cross_passage',       'Cross Passage',    'git-merge', 'TBM'),
('cp_stage',            'CP Stage',         'list',      'TBM'),
('geo_unit',            'Geology Unit',     'layers',    'TBM'),
('geo_borehole',        'Borehole',         'crosshair', 'TBM'),
('geo_stratigraphy',    'Stratigraphy',     'align-left','TBM'),
('geo_face',            'Face Map',         'map',       'TBM')
ON CONFLICT (code) DO NOTHING;