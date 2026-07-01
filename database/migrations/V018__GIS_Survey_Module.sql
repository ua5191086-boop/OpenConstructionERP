-- ============================================================================
-- V018__GIS_Survey_Module.sql
-- Модуль GIS & Survey (GS) — Map layers, survey data, drone orthophoto
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. GIS Layers (слои карты)
-- ============================================================================
CREATE TABLE gis_layers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    layer_name      VARCHAR(300) NOT NULL,
    layer_type      VARCHAR(50) NOT NULL DEFAULT 'vector'
        CHECK (layer_type IN ('vector','raster','tile','wms','wfs','point_cloud','orthophoto','dem','other')),
    geometry_type   VARCHAR(30) NOT NULL DEFAULT 'polygon'
        CHECK (geometry_type IN ('point','linestring','polygon','multipoint','multilinestring','multipolygon','raster')),
    source              VARCHAR(300),
    source_type         VARCHAR(30) NOT NULL DEFAULT 'upload'
        CHECK (source_type IN ('upload','wms','tile_server','api','generated','external')),
    style_config        JSONB DEFAULT '{}'::jsonb,
    min_zoom            INTEGER DEFAULT 0,
    max_zoom            INTEGER DEFAULT 22,
    is_visible          BOOLEAN DEFAULT TRUE,
    is_base_map         BOOLEAN DEFAULT FALSE,
    description         TEXT,
    metadata            JSONB DEFAULT '{}'::jsonb,
    created_by          VARCHAR(200),
    status              VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','archived','pending')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, layer_name)
);

CREATE INDEX idx_gs_layers_project ON gis_layers(project_id);
CREATE INDEX idx_gs_layers_type ON gis_layers(project_id, layer_type);

COMMENT ON TABLE gis_layers IS 'GIS Layers — картографические слои проекта';

-- ============================================================================
-- 2. GIS Features (геопространственные объекты)
-- ============================================================================
CREATE TABLE gis_features (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    layer_id        UUID REFERENCES gis_layers(id) ON DELETE CASCADE,
    feature_name    VARCHAR(300),
    feature_type    VARCHAR(30) NOT NULL DEFAULT 'point'
        CHECK (feature_type IN ('point','linestring','polygon','multipoint','multilinestring','multipolygon')),
    geometry        JSONB NOT NULL,                          -- GeoJSON geometry
    properties      JSONB DEFAULT '{}'::jsonb,
    elevation       NUMERIC(10,2),
    area_sq_m       NUMERIC(14,2),
    length_m        NUMERIC(14,2),
    perimeter_m     NUMERIC(14,2),
    source          VARCHAR(100) DEFAULT 'manual'
        CHECK (source IN ('manual','survey','drone','import','api','generated')),
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','archived','draft')),
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gs_features_project ON gis_features(project_id);
CREATE INDEX idx_gs_features_layer ON gis_features(layer_id);
CREATE INDEX idx_gs_features_type ON gis_features(project_id, feature_type);
CREATE INDEX idx_gs_features_geom ON gis_features USING gin (geometry);

COMMENT ON TABLE gis_features IS 'GIS Features — геопространственные объекты на карте';

-- ============================================================================
-- 3. GIS Survey Points (съёмочные точки)
-- ============================================================================
CREATE TABLE gis_survey_points (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    point_number    INTEGER NOT NULL,
    point_code      VARCHAR(30) NOT NULL,
    point_name      VARCHAR(200),
    point_type      VARCHAR(30) NOT NULL DEFAULT 'control'
        CHECK (point_type IN ('control','topographic','boundary','benchmark','as_built','staking','monitoring','check')),
    latitude        NUMERIC(12,9) NOT NULL,
    longitude       NUMERIC(12,9) NOT NULL,
    elevation       NUMERIC(10,3),
    northing        NUMERIC(14,3),
    easting         NUMERIC(14,3),
    zone            VARCHAR(10),
    datum           VARCHAR(50) DEFAULT 'WGS84',
    accuracy_mm     NUMERIC(8,2),
    description     TEXT,
    surveyed_by     VARCHAR(200),
    survey_date     DATE,
    method          VARCHAR(30) NOT NULL DEFAULT 'gps'
        CHECK (method IN ('gps','total_station','level','drone','laser_scanner','photogrammetry','other')),
    status          VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','archived','rejected')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, point_code),
    UNIQUE (project_id, point_number)
);

CREATE INDEX idx_gs_sp_project ON gis_survey_points(project_id);
CREATE INDEX idx_gs_sp_type ON gis_survey_points(project_id, point_type);
CREATE INDEX idx_gs_sp_coords ON gis_survey_points(project_id, latitude, longitude);

COMMENT ON TABLE gis_survey_points IS 'GIS Survey Points — геодезические съёмочные точки';

-- ============================================================================
-- 4. GIS Survey Runs (рейсы съёмки)
-- ============================================================================
CREATE TABLE gis_survey_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    run_number      INTEGER NOT NULL,
    run_code        VARCHAR(30) NOT NULL,                    -- 'SR-0001'
    run_name        VARCHAR(300) NOT NULL,
    survey_type     VARCHAR(30) NOT NULL DEFAULT 'topographic'
        CHECK (survey_type IN ('topographic','control','as_built','staking','monitoring','boundary','hydrographic','drone','laser_scan','other')),
    scope           TEXT,
    area_covered    NUMERIC(14,2),
    unit            VARCHAR(20) DEFAULT 'sq_m',
    start_date      DATE NOT NULL,
    end_date        DATE,
    instrument      VARCHAR(300),
    instrument_sn   VARCHAR(100),
    crew_lead       VARCHAR(200),
    crew_members    TEXT,
    weather_conditions TEXT,
    reference_station VARCHAR(100),
    accuracy_statement TEXT,
    point_count     INTEGER DEFAULT 0,
    file_path       VARCHAR(500),
    status          VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','reviewed','approved','cancelled')),
    approved_by     VARCHAR(200),
    approved_at     TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, run_number),
    UNIQUE (project_id, run_code)
);

CREATE INDEX idx_gs_sr_project ON gis_survey_runs(project_id);
CREATE INDEX idx_gs_sr_type ON gis_survey_runs(project_id, survey_type);
CREATE INDEX idx_gs_sr_status ON gis_survey_runs(project_id, status);

COMMENT ON TABLE gis_survey_runs IS 'GIS Survey Runs — рейсы полевых геодезических работ';

-- ============================================================================
-- 5. GIS Survey Stations (станции хода)
-- ============================================================================
CREATE TABLE gis_survey_stations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    run_id          UUID REFERENCES gis_survey_runs(id) ON DELETE CASCADE,
    station_number  INTEGER NOT NULL,
    station_code    VARCHAR(30) NOT NULL,
    station_name    VARCHAR(200),
    station_type    VARCHAR(30) NOT NULL DEFAULT 'traverse'
        CHECK (station_type IN ('traverse','benchmark','temporary_benchmark','turning_point','control_point','setup')),
    northing        NUMERIC(14,3),
    easting         NUMERIC(14,3),
    elevation       NUMERIC(10,3),
    back_sight      NUMERIC(10,3),
    foresight       NUMERIC(10,3),
    horizontal_angle NUMERIC(10,5),
    vertical_angle  NUMERIC(10,5),
    slope_distance  NUMERIC(10,3),
    horizontal_distance NUMERIC(10,3),
    corrected_distance NUMERIC(10,3),
    instrument_height NUMERIC(8,3),
    target_height   NUMERIC(8,3),
    coordinate_north NUMERIC(14,3),
    coordinate_east  NUMERIC(14,3),
    coordinate_elev  NUMERIC(10,3),
    misclosure      NUMERIC(10,3),
    adjusted        BOOLEAN DEFAULT FALSE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, station_code),
    UNIQUE (run_id, station_number)
);

CREATE INDEX idx_gs_ss_project ON gis_survey_stations(project_id);
CREATE INDEX idx_gs_ss_run ON gis_survey_stations(run_id);
CREATE INDEX idx_gs_ss_type ON gis_survey_stations(project_id, station_type);

COMMENT ON TABLE gis_survey_stations IS 'GIS Survey Stations — станции геодезического хода';

-- ============================================================================
-- 6. GIS Alignments (оси трасс)
-- ============================================================================
CREATE TABLE gis_alignments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    alignment_code  VARCHAR(30) NOT NULL,
    alignment_name  VARCHAR(300) NOT NULL,
    alignment_type  VARCHAR(30) NOT NULL DEFAULT 'road'
        CHECK (alignment_type IN ('road','railway','pipeline','canal','tunnel','bridge','power_line','conveyor','other')),
    design_speed    NUMERIC(8,2),
    design_standard VARCHAR(200),
    start_chainage  NUMERIC(12,3) DEFAULT 0,
    end_chainage    NUMERIC(12,3),
    total_length    NUMERIC(12,3),
    geometry        JSONB NOT NULL,                          -- GeoJSON linestring
    horizontal_curves JSONB DEFAULT '[]'::jsonb,
    vertical_curves  JSONB DEFAULT '[]'::jsonb,
    crossfall_type  VARCHAR(30),
    datum           VARCHAR(50) DEFAULT 'WGS84',
    description     TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'design'
        CHECK (status IN ('design','approved','as_built','superseded')),
    created_by      VARCHAR(200),
    approved_by     VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, alignment_code)
);

CREATE INDEX idx_gs_aln_project ON gis_alignments(project_id);
CREATE INDEX idx_gs_aln_type ON gis_alignments(project_id, alignment_type);

COMMENT ON TABLE gis_alignments IS 'GIS Alignments — оси трасс (дороги, трубопроводы и т.д.)';

-- ============================================================================
-- 7. GIS Cross Sections (поперечные профили)
-- ============================================================================
CREATE TABLE gis_cross_sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    alignment_id    UUID REFERENCES gis_alignments(id) ON DELETE CASCADE,
    section_number  INTEGER NOT NULL,
    chainage        NUMERIC(12,3) NOT NULL,
    offset_left     NUMERIC(10,2),
    offset_right    NUMERIC(10,2),
    geometry        JSONB NOT NULL,                          -- GeoJSON linestring/profile
    points          JSONB DEFAULT '[]'::jsonb,               -- [{distance, elevation, description}]
    cut_area        NUMERIC(10,3),
    fill_area       NUMERIC(10,3),
    total_area      NUMERIC(10,3),
    design_elevation NUMERIC(10,3),
    ground_elevation NUMERIC(10,3),
    source          VARCHAR(30) DEFAULT 'survey'
        CHECK (source IN ('survey','design','drone','generated')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (alignment_id, section_number)
);

CREATE INDEX idx_gs_cs_project ON gis_cross_sections(project_id);
CREATE INDEX idx_gs_cs_alignment ON gis_cross_sections(alignment_id);
CREATE INDEX idx_gs_cs_chainage ON gis_cross_sections(alignment_id, chainage);

COMMENT ON TABLE gis_cross_sections IS 'GIS Cross Sections — поперечные профили трасс';

-- ============================================================================
-- 8. GIS Drone Flights (полёты дронов)
-- ============================================================================
CREATE TABLE gis_drone_flights (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    flight_number   INTEGER NOT NULL,
    flight_code     VARCHAR(30) NOT NULL,                    -- 'UAV-0001'
    flight_name     VARCHAR(300) NOT NULL,
    drone_model     VARCHAR(200) NOT NULL,
    drone_sn        VARCHAR(100),
    pilot           VARCHAR(200) NOT NULL,
    flight_date     DATE NOT NULL,
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    flight_duration_minutes INTEGER,
    altitude_m      NUMERIC(8,2),
    speed_ms        NUMERIC(8,2),
    area_covered_ha NUMERIC(10,4),
    gsd_cm          NUMERIC(8,3),                            -- ground sample distance
    overlap_pct     NUMERIC(5,2),
    sidelap_pct     NUMERIC(5,2),
    flight_plan     JSONB DEFAULT '{}'::jsonb,
    waypoints       JSONB DEFAULT '[]'::jsonb,
    camera_model    VARCHAR(200),
    sensor_type     VARCHAR(30) DEFAULT 'rgb'
        CHECK (sensor_type IN ('rgb','multispectral','thermal','lidar','nir')),
    images_count    INTEGER DEFAULT 0,
    processing_status VARCHAR(30) DEFAULT 'pending'
        CHECK (processing_status IN ('pending','processing','completed','failed','cancelled')),
    output_type     VARCHAR(50),
    output_files    JSONB DEFAULT '[]'::jsonb,
    weather_conditions TEXT,
    wind_speed_kmh  NUMERIC(6,2),
    temperature_c   NUMERIC(5,1),
    notes           TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, flight_number),
    UNIQUE (project_id, flight_code)
);

CREATE INDEX idx_gs_df_project ON gis_drone_flights(project_id);
CREATE INDEX idx_gs_df_date ON gis_drone_flights(project_id, flight_date DESC);
CREATE INDEX idx_gs_df_status ON gis_drone_flights(project_id, processing_status);

COMMENT ON TABLE gis_drone_flights IS 'GIS Drone Flights — полёты БПЛА и аэрофотосъёмка';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('gis_layer',         'GIS Layer',        'layers',        'GS'),
('gis_feature',       'GIS Feature',      'map-pin',       'GS'),
('gis_survey_point',  'Survey Point',     'crosshair',     'GS'),
('gis_survey_run',    'Survey Run',       'ruler',         'GS'),
('gis_survey_station','Survey Station',   'target',        'GS'),
('gis_alignment',     'Alignment',        'trending-up',   'GS'),
('gis_cross_section', 'Cross Section',    'bar-chart-2',   'GS'),
('gis_drone_flight',  'Drone Flight',     'drone',         'GS')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW gis_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM gis_layers WHERE project_id = p.id AND status = 'active') AS active_layers,
    (SELECT COUNT(*) FROM gis_features WHERE project_id = p.id AND status = 'active') AS active_features,
    (SELECT COUNT(*) FROM gis_survey_points WHERE project_id = p.id AND status = 'active') AS survey_points,
    (SELECT COUNT(*) FROM gis_survey_runs WHERE project_id = p.id AND status = 'completed') AS completed_surveys,
    (SELECT COUNT(*) FROM gis_survey_runs WHERE project_id = p.id AND status IN ('planned','in_progress')) AS pending_surveys,
    (SELECT COUNT(*) FROM gis_alignments WHERE project_id = p.id AND status NOT IN ('superseded')) AS active_alignments,
    (SELECT COUNT(*) FROM gis_cross_sections WHERE project_id = p.id) AS cross_sections,
    (SELECT COUNT(*) FROM gis_drone_flights WHERE project_id = p.id AND status = 'completed') AS completed_flights,
    (SELECT COUNT(*) FROM gis_drone_flights WHERE project_id = p.id AND processing_status = 'processing') AS processing_flights
FROM projects p;