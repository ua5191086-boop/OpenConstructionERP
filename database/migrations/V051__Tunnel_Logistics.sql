-- ============================================================================
-- V051__Tunnel_Logistics.sql
-- Логистика забоя — подача сегментов, рельсы, вагонетки, конвейеры
-- ============================================================================

CREATE TABLE tunnel_logistics_zones (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    zone_code VARCHAR(50) NOT NULL,
    zone_name VARCHAR(300),
    zone_type VARCHAR(50) NOT NULL, -- portal, launch_shaft, reception_shaft, tunnel_section, turnout
    chainage_from NUMERIC(12,4),
    chainage_to NUMERIC(12,4),
    logistics_capacity VARCHAR(100), -- segments_per_shift, m3_per_hour
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tunnel_logistics_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    delivery_type VARCHAR(50) NOT NULL, -- segment, mortar, reinforcement, rail, cable, vent_duct
    item_code VARCHAR(100),
    quantity NUMERIC(18,4) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    source_zone UUID REFERENCES tunnel_logistics_zones(id),
    destination_zone UUID REFERENCES tunnel_logistics_zones(id),
    delivery_date DATE NOT NULL,
    shift VARCHAR(20), -- day, night
    tbm_ring_id UUID REFERENCES tbm_ring_builds(id),
    transport_method VARCHAR(50), -- locomotive, conveyor, truck, crane
    batch_number VARCHAR(100),
    status VARCHAR(50) DEFAULT 'scheduled', -- scheduled, in_transit, delivered, delayed, cancelled
    received_by VARCHAR(200),
    received_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE tunnel_conveyor_monitoring (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    conveyor_id VARCHAR(100) NOT NULL,
    operating_hours NUMERIC(10,2),
    material_volume NUMERIC(18,4),
    belt_speed NUMERIC(10,2),
    power_consumption NUMERIC(10,2),
    temperature NUMERIC(8,2),
    vibration NUMERIC(8,4),
    status VARCHAR(50) DEFAULT 'running',
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tlz_project ON tunnel_logistics_zones(project_id);
CREATE INDEX idx_tld_project ON tunnel_logistics_deliveries(project_id);
CREATE INDEX idx_tld_date ON tunnel_logistics_deliveries(delivery_date);
CREATE INDEX idx_tld_ring ON tunnel_logistics_deliveries(tbm_ring_id);
CREATE INDEX idx_tcm_conveyor ON tunnel_conveyor_monitoring(conveyor_id, recorded_at);