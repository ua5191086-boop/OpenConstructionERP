-- OpenConstructionERP
-- V034: Variation Orders (SAD G-02) — изменение → оценка → утверждение → включение в бюджет.
-- Owner: core-py lane. Registered in docs/WORKSTREAMS.md.

CREATE TABLE variation_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    contract_id UUID REFERENCES contracts(id),
    number INTEGER NOT NULL,
    code VARCHAR(20) NOT NULL,                     -- 'VO-0001'
    title VARCHAR(255) NOT NULL,
    description TEXT,
    origin VARCHAR(30) NOT NULL DEFAULT 'client_instruction'
        CHECK (origin IN ('client_instruction','site_condition','design_change',
                          'claim_conversion','value_engineering','regulatory')),
    cost_impact NUMERIC(18,2) NOT NULL DEFAULT 0,  -- +/- к контрактной цене
    time_impact_days INTEGER NOT NULL DEFAULT 0,   -- +/- к сроку (EOT)
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','under_evaluation','approved',
                          'rejected','incorporated','void')),
    notice_ref VARCHAR(120),                       -- ссылка на notice по контрактному сроку
    submitted_at TIMESTAMPTZ,
    approved_at TIMESTAMPTZ,
    approved_by VARCHAR(120),
    incorporated_at TIMESTAMPTZ,
    budget_version_id UUID REFERENCES budget_versions(id),  -- снапшот при включении
    created_by VARCHAR(120),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, number),
    UNIQUE (project_id, code)
);
CREATE INDEX idx_vo_status ON variation_orders(project_id, status);

CREATE TABLE variation_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vo_id UUID NOT NULL REFERENCES variation_orders(id) ON DELETE CASCADE,
    boq_item_id UUID REFERENCES boq_items(id),     -- NULL = новая позиция
    description VARCHAR(255) NOT NULL,
    unit VARCHAR(20),
    quantity NUMERIC(20,4) NOT NULL DEFAULT 0,
    unit_price NUMERIC(20,6) NOT NULL DEFAULT 0,
    amount NUMERIC(18,2) GENERATED ALWAYS AS (round(quantity * unit_price, 2)) STORED,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_voi_vo ON variation_order_items(vo_id);

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('variation_order', 'Variation Order', 'delta', 'G-02')
ON CONFLICT (code) DO NOTHING;
