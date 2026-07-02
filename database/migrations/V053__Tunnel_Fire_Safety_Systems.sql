-- ============================================================================
-- V053__Tunnel_Fire_Safety_Systems.sql
-- Пожарная сигнализация, эвакуация, огнезащита обделки
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Пожарные зоны тоннеля
-- ============================================================================
CREATE TABLE tunnel_fire_zones (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    zone_code           VARCHAR(100) NOT NULL,
    zone_name           VARCHAR(500) NOT NULL,
    ring_start          INTEGER NOT NULL,
    ring_end            INTEGER NOT NULL,
    fire_resistance_rating VARCHAR(50),                         -- R60, R90, R120, R180
    evacuation_distance_m NUMERIC(8,2),
    occupancy_max       INTEGER,
    fire_load_mj_m2     NUMERIC(10,2),                          -- пожарная нагрузка
    fire_class          VARCHAR(20),                            -- A, B, C, D, F
    suppression_system  VARCHAR(100),                           -- sprinkler, water_mist, foam, dry_pipe, gas
    detection_system    VARCHAR(100),                           -- smoke, heat, flame, multi_sensor, linear_heat
    status              VARCHAR(50) DEFAULT 'active',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tfz_project ON tunnel_fire_zones(project_id);

COMMENT ON TABLE tunnel_fire_zones IS 'Пожарные зоны тоннеля с классификацией';

-- ============================================================================
-- 2. Противопожарное оборудование
-- ============================================================================
CREATE TABLE tunnel_fire_equipment (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    zone_id             UUID NOT NULL REFERENCES tunnel_fire_zones(id) ON DELETE CASCADE,
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_code      VARCHAR(100) NOT NULL,
    equipment_type      VARCHAR(100) NOT NULL,                  -- hydrant, extinguisher, sprinkler, detector, alarm, emergency_light, fire_hose
    manufacturer        VARCHAR(300),
    model               VARCHAR(200),
    location_ring       INTEGER,
    location_description VARCHAR(500),
    quantity             INTEGER DEFAULT 1,
    inspection_freq_days INTEGER DEFAULT 30,
    last_inspection     DATE,
    next_inspection     DATE,
    expiry_date         DATE,
    status              VARCHAR(50) DEFAULT 'operational',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tfe_zone ON tunnel_fire_equipment(zone_id);
CREATE INDEX idx_tfe_inspection ON tunnel_fire_equipment(next_inspection) WHERE status='operational';

COMMENT ON TABLE tunnel_fire_equipment IS 'Противопожарное оборудование тоннеля';

-- ============================================================================
-- 3. Эвакуационные маршруты
-- ============================================================================
CREATE TABLE tunnel_evacuation_routes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    route_code          VARCHAR(100) NOT NULL,
    route_name          VARCHAR(500) NOT NULL,
    route_type          VARCHAR(50) NOT NULL,                  -- primary, secondary, emergency_exit, cross_passage
    ring_from           INTEGER NOT NULL,
    ring_to             INTEGER NOT NULL,
    tbm_id              UUID REFERENCES tbm(id) ON DELETE SET NULL,
    length_m            NUMERIC(8,2),
    width_m             NUMERIC(6,2),
    capacity_persons    INTEGER,
    estimated_time_s    INTEGER,                                -- расчётное время эвакуации (сек)
    lighting_type       VARCHAR(50),                            -- emergency, photoluminescent, led
    signage_type        VARCHAR(50),                            -- standard, photoluminescent, electronic
    communication_type  VARCHAR(100),                           -- intercom, phone, radio, none
    last_drill_date     DATE,
    drill_passed        BOOLEAN,
    status              VARCHAR(50) DEFAULT 'active',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ter_project ON tunnel_evacuation_routes(project_id);

COMMENT ON TABLE tunnel_evacuation_routes IS 'Эвакуационные маршруты тоннеля';

-- ============================================================================
-- 4. Журнал пожарных учений и проверок
-- ============================================================================
CREATE TABLE tunnel_fire_drills (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    zone_id             UUID REFERENCES tunnel_fire_zones(id) ON DELETE SET NULL,
    drill_type          VARCHAR(100) NOT NULL,                  -- fire_drill, evacuation, equipment_test, tabletop, full_scale
    drill_date          DATE NOT NULL,
    description         TEXT NOT NULL,
    participants_count  INTEGER,
    duration_min        INTEGER,
    scenario            TEXT,
    evacuation_time_actual_s INTEGER,
    evacuation_time_target_s INTEGER,
    issues_found        TEXT,
    corrective_actions  TEXT,
    passed              BOOLEAN,
    conducted_by        VARCHAR(300),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tfd_project ON tunnel_fire_drills(project_id);

COMMENT ON TABLE tunnel_fire_drills IS 'Журнал пожарных учений и тренировок';

-- ============================================================================
-- 5. Огнезащита обделки
-- ============================================================================
CREATE TABLE tunnel_fire_protection_lining (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    batch_id            UUID REFERENCES segment_production_batches(id) ON DELETE SET NULL,
    ring_number         INTEGER NOT NULL,
    application_type    VARCHAR(100) NOT NULL,                  -- intumescent_coating, concrete_cover, fire_board, spray_mortar, panel
    material            VARCHAR(300) NOT NULL,
    thickness_mm        NUMERIC(6,2) NOT NULL,
    fire_rating         VARCHAR(50),                            -- R60, R90, R120
    application_date    DATE,
    curing_time_hours   NUMERIC(8,2),
    inspector           VARCHAR(300),
    inspection_result   VARCHAR(50),                            -- pass, fail, conditional
    adhesion_test_result NUMERIC(6,2),                          -- МПа
    warranty_expiry     DATE,
    status              VARCHAR(50) DEFAULT 'planned',          -- planned, applied, cured, inspected, failed, repaired
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tfpl_ring ON tunnel_fire_protection_lining(ring_number);
CREATE INDEX idx_tfpl_project ON tunnel_fire_protection_lining(project_id);

COMMENT ON TABLE tunnel_fire_protection_lining IS 'Огнезащита обделки тоннеля';

-- ============================================================================
-- 6. View: готовность систем противопожарной защиты
-- ============================================================================
CREATE OR REPLACE VIEW tunnel_fire_safety_readiness AS
SELECT
    tfz.project_id,
    tfz.zone_code,
    tfz.zone_name,
    COUNT(DISTINCT tfe.id) as equipment_count,
    COUNT(DISTINCT CASE WHEN tfe.status='operational' THEN tfe.id END) as operational_count,
    COUNT(DISTINCT CASE WHEN tfe.next_inspection < CURRENT_DATE THEN tfe.id END) as overdue_inspections,
    COUNT(DISTINCT ter.id) as evacuation_routes,
    COUNT(DISTINCT tfd.id) as drills_conducted,
    MAX(tfd.drill_date) as last_drill_date
FROM tunnel_fire_zones tfz
LEFT JOIN tunnel_fire_equipment tfe ON tfe.zone_id = tfz.id
LEFT JOIN tunnel_evacuation_routes ter ON ter.project_id = tfz.project_id
LEFT JOIN tunnel_fire_drills tfd ON tfd.zone_id = tfz.id
GROUP BY tfz.id, tfz.project_id, tfz.zone_code, tfz.zone_name;

COMMENT ON VIEW tunnel_fire_safety_readiness IS 'Готовность систем противопожарной защиты по зонам';