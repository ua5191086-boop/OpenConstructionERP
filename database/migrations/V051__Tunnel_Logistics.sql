-- ============================================================================
-- V051__Tunnel_Logistics.sql
-- Логистика тоннельного забоя — подача сегментов, рельсы, вагонетки, конвейеры
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Логистические маршруты забоя
-- ============================================================================
CREATE TABLE tunnel_logistics_routes (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    route_code          VARCHAR(100) NOT NULL,
    route_name          VARCHAR(500) NOT NULL,
    route_type          VARCHAR(50) NOT NULL,                  -- segment_delivery, muck_removal, material_supply, personnel
    tbm_id              UUID REFERENCES tbm(id) ON DELETE SET NULL,
    shaft_id            UUID REFERENCES shaft_projects(id) ON DELETE SET NULL,
    ring_range_start    INTEGER,
    ring_range_end      INTEGER,
    distance_m          NUMERIC(10,2),
    avg_travel_time_min NUMERIC(8,2),
    transport_mode      VARCHAR(50),                            -- train, conveyor, vehicle, crane
    capacity_per_trip   NUMERIC(10,2),
    unit                VARCHAR(20),                            -- tons, m3, segments, personnel
    status              VARCHAR(50) DEFAULT 'active',
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tnlr_project ON tunnel_logistics_routes(project_id);
CREATE INDEX idx_tnlr_tbm ON tunnel_logistics_routes(tbm_id);
CREATE INDEX idx_tnlr_type ON tunnel_logistics_routes(route_type);

COMMENT ON TABLE tunnel_logistics_routes IS 'Логистические маршруты подачи материалов в забой';

-- ============================================================================
-- 2. Циклограммы подачи (графики поставок)
-- ============================================================================
CREATE TABLE tunnel_delivery_schedules (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    route_id            UUID NOT NULL REFERENCES tunnel_logistics_routes(id) ON DELETE CASCADE,
    ring_number         INTEGER NOT NULL,
    delivery_date       DATE NOT NULL,
    shift               VARCHAR(20),                            -- day, night, graveyard
    planned_trips       INTEGER NOT NULL,
    actual_trips        INTEGER,
    planned_quantity    NUMERIC(12,2),
    actual_quantity     NUMERIC(12,2),
    material_type       VARCHAR(100),                           -- segments, mortar, bolts, muck, equipment
    status              VARCHAR(50) DEFAULT 'planned',          -- planned, in_transit, delivered, delayed, cancelled
    delay_reason        TEXT,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tds_route ON tunnel_delivery_schedules(route_id);
CREATE INDEX idx_tds_ring ON tunnel_delivery_schedules(ring_number);
CREATE INDEX idx_tds_date ON tunnel_delivery_schedules(delivery_date);
CREATE INDEX idx_tds_status ON tunnel_delivery_schedules(status);

COMMENT ON TABLE tunnel_delivery_schedules IS 'Графики подачи материалов к кольцам забоя';

-- ============================================================================
-- 3. Инвентаризация забоя (что сейчас в забое)
-- ============================================================================
CREATE TABLE tunnel_face_inventory (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tbm_id              UUID REFERENCES tbm(id) ON DELETE SET NULL,
    ring_number         INTEGER NOT NULL,
    inventory_date      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    segments_stored     INTEGER DEFAULT 0,
    bolts_stored        INTEGER DEFAULT 0,
    mortar_kg           NUMERIC(10,2) DEFAULT 0,
    gaskets_stored      INTEGER DEFAULT 0,
    packers_stored      INTEGER DEFAULT 0,
    muck_volume_m3      NUMERIC(10,2),
    water_l_min         NUMERIC(10,2),
    air_pressure_bar    NUMERIC(5,2),
    recorded_by         VARCHAR(200),
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tfi_tbm ON tunnel_face_inventory(tbm_id);
CREATE INDEX idx_tfi_ring ON tunnel_face_inventory(ring_number);

COMMENT ON TABLE tunnel_face_inventory IS 'Текущая инвентаризация материалов в забое';

-- ============================================================================
-- 4. Логи событий логистики
-- ============================================================================
CREATE TABLE tunnel_logistics_events (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    route_id            UUID REFERENCES tunnel_logistics_routes(id) ON DELETE SET NULL,
    tbm_id              UUID REFERENCES tbm(id) ON DELETE SET NULL,
    event_type          VARCHAR(50) NOT NULL,                   -- departure, arrival, delay, breakdown, overload, accident
    event_time          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    location_from       VARCHAR(300),
    location_to         VARCHAR(300),
    transport_unit      VARCHAR(200),
    operator            VARCHAR(200),
    load_description    TEXT,
    load_quantity       NUMERIC(12,2),
    load_unit           VARCHAR(20),
    duration_min        INTEGER,
    delay_min           INTEGER,
    delay_reason        TEXT,
    severity            VARCHAR(20) DEFAULT 'info',             -- info, warning, critical
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tle_route ON tunnel_logistics_events(route_id);
CREATE INDEX idx_tle_event_time ON tunnel_logistics_events(event_time);
CREATE INDEX idx_tle_type ON tunnel_logistics_events(event_type);
CREATE INDEX idx_tle_severity ON tunnel_logistics_events(severity);

COMMENT ON TABLE tunnel_logistics_events IS 'Оперативные события логистики забоя';

-- ============================================================================
-- 5. View: эффективность логистики
-- ============================================================================
CREATE OR REPLACE VIEW tunnel_logistics_efficiency AS
SELECT
    lr.project_id,
    lr.route_code,
    lr.route_name,
    lr.route_type,
    COUNT(tle.id) as total_trips,
    SUM(tle.delay_min) as total_delay_min,
    AVG(tle.duration_min) as avg_duration_min,
    COUNT(CASE WHEN tle.severity = 'critical' THEN 1 END) as critical_events,
    AVG(tle.load_quantity) as avg_load,
    MAX(tle.event_time) as last_event
FROM tunnel_logistics_routes lr
LEFT JOIN tunnel_logistics_events tle ON tle.route_id = lr.id
GROUP BY lr.id, lr.project_id, lr.route_code, lr.route_name, lr.route_type;

COMMENT ON VIEW tunnel_logistics_efficiency IS 'Показатели эффективности логистики по маршрутам';