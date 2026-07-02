-- OpenConstructionERP
-- V028: Quality & HSE minimum — M-03 NCR + N-01 Permit to Work
-- Owner: core-py lane. Registered in docs/WORKSTREAMS.md.

CREATE TABLE ncrs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    code VARCHAR(20) NOT NULL,                        -- 'NCR-0001'
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(10) NOT NULL DEFAULT 'minor'
        CHECK (severity IN ('minor','major','critical')),
    location VARCHAR(120),                            -- захватка/пикет/кольцо
    boq_item_id UUID REFERENCES boq_items(id),
    ring_id UUID REFERENCES tunnel_rings(id),
    raised_by VARCHAR(120),
    assigned_to VARCHAR(120),
    root_cause TEXT,
    corrective_action TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','disposition','corrective_action','verification','closed','void')),
    due_date DATE,
    raised_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, number),
    UNIQUE (project_id, code)
);
CREATE INDEX idx_ncrs_status ON ncrs(project_id, status);
CREATE INDEX idx_ncrs_severity ON ncrs(project_id, severity) WHERE status <> 'closed';

CREATE TABLE work_permits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    code VARCHAR(20) NOT NULL,                        -- 'PTW-0001'
    permit_type VARCHAR(30) NOT NULL
        CHECK (permit_type IN ('hot_work','confined_space','work_at_height',
                               'lifting','excavation','electrical','gas_hazard','general')),
    description TEXT NOT NULL,
    location VARCHAR(120) NOT NULL,
    contractor VARCHAR(160),
    issued_to VARCHAR(120) NOT NULL,                  -- ответственный исполнитель
    issued_by VARCHAR(120),                           -- выдал (HSE/нач. участка)
    valid_from TIMESTAMPTZ NOT NULL,
    valid_to TIMESTAMPTZ NOT NULL,
    precautions TEXT,                                 -- меры безопасности
    gas_test_required BOOLEAN DEFAULT false,
    status VARCHAR(15) NOT NULL DEFAULT 'issued'
        CHECK (status IN ('issued','active','suspended','closed','cancelled')),
    activated_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    closure_notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (valid_to > valid_from),
    UNIQUE (project_id, number),
    UNIQUE (project_id, code)
);
CREATE INDEX idx_ptw_active ON work_permits(project_id, status)
    WHERE status IN ('issued','active');
CREATE INDEX idx_ptw_validity ON work_permits(project_id, valid_to);

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('ncr',    'Non-Conformance Report', 'alert',  'M-03'),
('permit', 'Permit to Work',         'shield', 'N-01')
ON CONFLICT (code) DO NOTHING;
