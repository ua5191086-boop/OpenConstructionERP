-- ============================================================================
-- V053__Tunnel_Fire_Safety.sql
-- Пожарная сигнализация, эвакуация, огнезащита обделки
-- ============================================================================

CREATE TABLE tunnel_fire_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    zone_code VARCHAR(50) NOT NULL,
    chainage_from NUMERIC(12,4),
    chainage_to NUMERIC(12,4),
    fire_risk_category VARCHAR(50), -- low, medium, high, extreme
    fire_resistance_rating VARCHAR(50), -- R60, R90, R120, R180
    evacuation_distance NUMERIC(10,2), -- meters
    cross_passage_count INTEGER,
    fire_extinguishers INTEGER,
    hose_reels INTEGER,
    sprinkler_system BOOLEAN DEFAULT FALSE,
    smoke_detectors INTEGER,
    heat_detectors INTEGER,
    manual_call_points INTEGER,
    emergency_lighting BOOLEAN DEFAULT TRUE,
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tunnel_fire_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    zone_id UUID REFERENCES tunnel_fire_zones(id),
    incident_type VARCHAR(100) NOT NULL, -- fire, smoke, alarm_false, equipment_fire, electrical
    alarm_time TIMESTAMPTZ NOT NULL,
    arrival_time TIMESTAMPTZ,
    containment_time TIMESTAMPTZ,
    extinguished_time TIMESTAMPTZ,
    severity VARCHAR(50), -- minor, moderate, major, critical
    cause TEXT,
    damage_assessment TEXT,
    injuries INTEGER DEFAULT 0,
    fatalities INTEGER DEFAULT 0,
    evacuation_used BOOLEAN DEFAULT FALSE,
    evacuated_count INTEGER,
    equipment_damaged TEXT,
    estimated_loss NUMERIC(18,2),
    insurance_claim_ref VARCHAR(200),
    report_document VARCHAR(500),
    status VARCHAR(50) DEFAULT 'reported',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tunnel_fire_drills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    drill_date DATE NOT NULL,
    drill_type VARCHAR(100), -- full_evacuation, tabletop, equipment_test, communication
    participants INTEGER,
    duration_minutes INTEGER,
    scenario TEXT,
    objectives TEXT,
    results TEXT,
    improvements TEXT,
    conducted_by VARCHAR(300),
    status VARCHAR(50) DEFAULT 'completed',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tfz_project ON tunnel_fire_zones(project_id);
CREATE INDEX idx_tfi_zone ON tunnel_fire_incidents(zone_id);
CREATE INDEX idx_tfi_time ON tunnel_fire_incidents(alarm_time);