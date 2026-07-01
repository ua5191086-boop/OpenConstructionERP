-- OpenConstructionERP
-- V002: Ontology core (minimal, per SAD Tom 1 §3 / ADR-001) + regional coefficients
-- Prototype scope: object types, objects, links. Bitemporal versioning and actions come later.

CREATE TABLE object_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    icon VARCHAR(50),
    schema JSONB NOT NULL DEFAULT '{}'::jsonb,   -- JSON Schema for props validation
    module_owner VARCHAR(20),                    -- SAD module code, e.g. 'B-04'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type_id UUID NOT NULL REFERENCES object_types(id),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,  -- NULL = holding-level object
    props JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_table VARCHAR(63),   -- mirror provenance: which domain table owns the record
    source_id UUID,             -- PK in the domain table
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_objects_type ON objects(type_id);
CREATE INDEX idx_objects_project ON objects(project_id, type_id);
CREATE INDEX idx_objects_props ON objects USING GIN(props);
CREATE UNIQUE INDEX uq_objects_source ON objects(source_table, source_id) WHERE source_table IS NOT NULL;

CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_type VARCHAR(50) NOT NULL,              -- e.g. 'belongs_to', 'costed_by', 'located_in'
    from_object UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    to_object UUID NOT NULL REFERENCES objects(id) ON DELETE CASCADE,
    props JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(link_type, from_object, to_object)
);

CREATE INDEX idx_links_from ON links(from_object, link_type);
CREATE INDEX idx_links_to ON links(to_object, link_type);

-- Regional coefficients (Cost Intelligence: live regional factors applied to estimates)
CREATE TABLE regional_coefficients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_code VARCHAR(10) NOT NULL UNIQUE,     -- 'UZ', 'KZ', 'LT', 'AT', 'TM', 'GE', 'UA'
    region_name VARCHAR(100) NOT NULL,
    labour_factor NUMERIC(6,3) NOT NULL DEFAULT 1.0,
    material_factor NUMERIC(6,3) NOT NULL DEFAULT 1.0,
    equipment_factor NUMERIC(6,3) NOT NULL DEFAULT 1.0,
    overall_factor NUMERIC(6,3) NOT NULL DEFAULT 1.0,  -- used when item has no component split
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    notes TEXT,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed: base object types
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('project',      'Project',            'folder',   'C-01'),
('organization', 'Organization',       'building', 'A-03'),
('contract',     'Contract',           'file',     'G-01'),
('boq_item',     'BOQ Item',           'list',     'B-04'),
('cbs_chapter',  'CBS Chapter',        'tree',     'C-02'),
('boq_section',  'BOQ Section',        'layers',   'B-04'),
('document',     'Document',           'doc',      'D-01');

-- Seed: regional coefficients (baseline 1.0 — real factors loaded from Cost Intelligence)
INSERT INTO regional_coefficients (region_code, region_name, overall_factor, labour_factor, material_factor, equipment_factor, notes) VALUES
('BASE', 'Baseline (no adjustment)', 1.000, 1.000, 1.000, 1.000, 'Reference'),
('UZ', 'Uzbekistan',   0.780, 0.450, 0.920, 0.850, 'Placeholder — replace with Cost Intelligence live values'),
('KZ', 'Kazakhstan',   0.850, 0.550, 0.950, 0.900, 'Placeholder'),
('LT', 'Lithuania',    1.050, 0.900, 1.020, 1.000, 'Placeholder'),
('AT', 'Austria',      1.350, 1.600, 1.150, 1.100, 'Placeholder'),
('TM', 'Turkmenistan', 0.820, 0.400, 1.050, 0.950, 'Placeholder'),
('GE', 'Georgia',      0.800, 0.500, 0.980, 0.900, 'Placeholder'),
('UA', 'Ukraine',      0.750, 0.400, 0.900, 0.850, 'Placeholder');
