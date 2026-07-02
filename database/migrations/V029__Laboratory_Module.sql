-- ============================================================================
-- V029__Laboratory_Module.sql
-- Laboratory: Material Testing, Sampling, Lab Equipment
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Material Testing
-- ============================================================================
CREATE TABLE material_testing (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    test_number     VARCHAR(100) NOT NULL,
    material_type   VARCHAR(200) NOT NULL,                    -- concrete, steel, soil, aggregate, asphalt
    test_type       VARCHAR(300) NOT NULL,                    -- compression, tensile, sieve, proctor
    specification   VARCHAR(500),
    sample_id       VARCHAR(200),
    sampling_date   DATE,
    test_date       DATE,
    result          TEXT,
    status          VARCHAR(50) DEFAULT 'pending',            -- pending, in_progress, completed, rejected
    tested_by       VARCHAR(300),
    approved_by     VARCHAR(300),
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, test_number)
);
CREATE INDEX idx_material_testing_project ON material_testing(project_id);
CREATE INDEX idx_material_testing_type ON material_testing(material_type);

-- ============================================================================
-- 2. Concrete Tests
-- ============================================================================
CREATE TABLE concrete_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    material_test_id UUID REFERENCES material_testing(id),
    sample_id       VARCHAR(200),
    concrete_grade  VARCHAR(100),
    slump           NUMERIC(8,2),
    compressive_strength_7d  NUMERIC(10,2),
    compressive_strength_14d NUMERIC(10,2),
    compressive_strength_28d NUMERIC(10,2),
    flexural_strength NUMERIC(10,2),
    air_content     NUMERIC(8,2),
    temperature     NUMERIC(8,2),
    unit_weight     NUMERIC(10,2),
    curing_method   VARCHAR(100),
    test_date       DATE,
    result          VARCHAR(50),                              -- pass, fail, pending
    tested_by       VARCHAR(300),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_concrete_tests_project ON concrete_tests(project_id);

-- ============================================================================
-- 3. Soil Tests
-- ============================================================================
CREATE TABLE soil_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    material_test_id UUID REFERENCES material_testing(id),
    sample_id       VARCHAR(200),
    soil_type       VARCHAR(200),
    moisture_content NUMERIC(10,2),
    dry_density     NUMERIC(10,2),
    atterberg_limit_liquid NUMERIC(10,2),
    atterberg_limit_plastic NUMERIC(10,2),
    plasticity_index NUMERIC(10,2),
    compaction_pct  NUMERIC(8,2),
    cbr_value       NUMERIC(10,2),
    shear_strength  NUMERIC(10,2),
    test_date       DATE,
    result          VARCHAR(50),
    tested_by       VARCHAR(300),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_soil_tests_project ON soil_tests(project_id);

-- ============================================================================
-- 4. Steel Tests
-- ============================================================================
CREATE TABLE steel_tests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    material_test_id UUID REFERENCES material_testing(id),
    sample_id       VARCHAR(200),
    steel_grade     VARCHAR(100),
    diameter_mm     NUMERIC(8,2),
    yield_strength  NUMERIC(10,2),
    tensile_strength NUMERIC(10,2),
    elongation_pct  NUMERIC(8,2),
    bend_test_result VARCHAR(50),
    weld_test_result VARCHAR(50),
    test_date       DATE,
    result          VARCHAR(50),
    tested_by       VARCHAR(300),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_steel_tests_project ON steel_tests(project_id);

-- ============================================================================
-- 5. Lab Certificates
-- ============================================================================
CREATE TABLE lab_certificates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    certificate_number VARCHAR(200) NOT NULL,
    certificate_type VARCHAR(100) NOT NULL,                   -- calibration, accreditation, test_result
    issuing_body    VARCHAR(300),
    issue_date      DATE,
    expiry_date     DATE,
    description     TEXT,
    document_url    TEXT,
    status          VARCHAR(50) DEFAULT 'valid',              -- valid, expired, revoked
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_lab_certificates_project ON lab_certificates(project_id);

-- ============================================================================
-- 6. Lab Equipment
-- ============================================================================
CREATE TABLE lab_equipment (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    equipment_code  VARCHAR(100) NOT NULL,
    equipment_name  VARCHAR(500) NOT NULL,
    equipment_type  VARCHAR(200),
    manufacturer    VARCHAR(300),
    model           VARCHAR(200),
    serial_number   VARCHAR(200),
    calibration_due DATE,
    status          VARCHAR(50) DEFAULT 'operational',        -- operational, under_maintenance, decommissioned
    notes           TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_lab_equipment_project ON lab_equipment(project_id);

-- ============================================================================
-- 7. Sampling Log
-- ============================================================================
CREATE TABLE sampling_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sample_id       VARCHAR(200) NOT NULL,
    sample_type     VARCHAR(200) NOT NULL,                    -- concrete, soil, steel, aggregate
    location        VARCHAR(500),
    sampling_date   DATE NOT NULL,
    sampled_by      VARCHAR(300),
    material_test_id UUID REFERENCES material_testing(id),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, sample_id)
);
CREATE INDEX idx_sampling_log_project ON sampling_log(project_id);