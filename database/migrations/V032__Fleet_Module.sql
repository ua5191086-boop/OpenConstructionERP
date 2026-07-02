-- ============================================================================
-- V032__Fleet_Module.sql
-- Fleet: Vehicles, Fuel, Maintenance, Tracking
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Fleet Vehicles (расширение equipment)
-- ============================================================================
CREATE TABLE fleet_vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_id    UUID REFERENCES equipment(id),
    vehicle_type    VARCHAR(100) NOT NULL,                    -- truck, crane, excavator, dozer, loader, grader, roller, van, bus, pickup, trailer
    make            VARCHAR(200),
    model           VARCHAR(200),
    year            INTEGER,
    vin             VARCHAR(100),
    license_plate   VARCHAR(50),
    registration_number VARCHAR(200),
    fuel_type       VARCHAR(50),                              -- diesel, petrol, electric, hybrid, cng
    engine_capacity NUMERIC(10,2),
    horsepower      INTEGER,
    weight_kg       NUMERIC(12,2),
    load_capacity_kg NUMERIC(12,2),
    status          VARCHAR(50) DEFAULT 'operational',        -- operational, under_maintenance, out_of_service, decommissioned
    assigned_driver UUID,
    location        VARCHAR(500),
    mileage_km      NUMERIC(12,2) DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_fleet_vehicles_project ON fleet_vehicles(project_id);
CREATE INDEX idx_fleet_vehicles_type ON fleet_vehicles(vehicle_type);
CREATE INDEX idx_fleet_vehicles_status ON fleet_vehicles(status);

-- ============================================================================
-- 2. Vehicle Drivers
-- ============================================================================
CREATE TABLE vehicle_drivers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    driver_name     VARCHAR(300) NOT NULL,
    license_number  VARCHAR(200),
    license_type    VARCHAR(100),
    license_expiry  DATE,
    contact_phone   VARCHAR(100),
    email           VARCHAR(300),
    certifications  JSONB,
    status          VARCHAR(50) DEFAULT 'active',             -- active, suspended, inactive
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_drivers_project ON vehicle_drivers(project_id);

-- ============================================================================
-- 3. Vehicle Fuel
-- ============================================================================
CREATE TABLE vehicle_fuel (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id),
    driver_id       UUID REFERENCES vehicle_drivers(id),
    fuel_date       DATE NOT NULL,
    fuel_type       VARCHAR(50),
    quantity_liters NUMERIC(12,2),
    unit_price      NUMERIC(10,4),
    total_cost      NUMERIC(14,2),
    currency        VARCHAR(3) DEFAULT 'USD',
    odometer_km     NUMERIC(12,2),
    station_name    VARCHAR(300),
    receipt_number  VARCHAR(200),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_fuel_vehicle ON vehicle_fuel(vehicle_id);
CREATE INDEX idx_vehicle_fuel_date ON vehicle_fuel(fuel_date);

-- ============================================================================
-- 4. Vehicle Maintenance
-- ============================================================================
CREATE TABLE vehicle_maintenance (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id),
    maintenance_type VARCHAR(200) NOT NULL,                   -- oil_change, tire_replacement, brake_service, engine_repair, transmission, inspection, body_repair
    description     TEXT,
    scheduled_date  DATE,
    completed_date  DATE,
    odometer_km     NUMERIC(12,2),
    cost_amount     NUMERIC(14,2),
    currency        VARCHAR(3) DEFAULT 'USD',
    vendor          VARCHAR(300),
    invoice_number  VARCHAR(200),
    status          VARCHAR(50) DEFAULT 'scheduled',          -- scheduled, in_progress, completed, cancelled
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_maint_vehicle ON vehicle_maintenance(vehicle_id);

-- ============================================================================
-- 5. Vehicle Accidents
-- ============================================================================
CREATE TABLE vehicle_accidents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id),
    driver_id       UUID REFERENCES vehicle_drivers(id),
    accident_date   TIMESTAMPTZ NOT NULL,
    location        VARCHAR(500),
    description     TEXT,
    severity        VARCHAR(50),                              -- minor, moderate, serious, fatal
    damages         TEXT,
    injuries        INTEGER DEFAULT 0,
    fatalities      INTEGER DEFAULT 0,
    police_report   VARCHAR(200),
    insurance_claim_id UUID REFERENCES insurance_claims(id),
    cost_estimate   NUMERIC(14,2),
    status          VARCHAR(50) DEFAULT 'reported',           -- reported, under_investigation, resolved
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_accidents_vehicle ON vehicle_accidents(vehicle_id);

-- ============================================================================
-- 6. Vehicle Tracking
-- ============================================================================
CREATE TABLE vehicle_tracking (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id),
    driver_id       UUID REFERENCES vehicle_drivers(id),
    track_date      DATE NOT NULL,
    start_time      TIMESTAMPTZ,
    end_time        TIMESTAMPTZ,
    start_location  VARCHAR(500),
    end_location    VARCHAR(500),
    distance_km     NUMERIC(12,2),
    duration_minutes INTEGER,
    purpose         VARCHAR(300),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_vehicle_tracking_vehicle ON vehicle_tracking(vehicle_id);
CREATE INDEX idx_vehicle_tracking_date ON vehicle_tracking(track_date);

-- ============================================================================
-- 7. Vehicle Telematics
-- ============================================================================
CREATE TABLE vehicle_telematics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id),
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    latitude        NUMERIC(10,7),
    longitude       NUMERIC(10,7),
    speed_kph       NUMERIC(8,2),
    heading         NUMERIC(5,2),
    altitude_m      NUMERIC(8,2),
    engine_temp     NUMERIC(8,2),
    fuel_level_pct  NUMERIC(5,2),
    battery_voltage NUMERIC(6,2),
    tire_pressure   JSONB,
    engine_rpm      INTEGER,
    odometer_km     NUMERIC(12,2),
    diagnostics     JSONB
);
CREATE INDEX idx_vehicle_telematics_vehicle ON vehicle_telematics(vehicle_id, recorded_at DESC);