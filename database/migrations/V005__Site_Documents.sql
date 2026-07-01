-- OpenConstructionERP
-- V005: Site documents minimum — D-03 RFI Management + C-05 Daily Reports
--       with work entries feeding physical progress against BOQ.
-- Owner: core-py lane. Registered in docs/WORKSTREAMS.md.

CREATE TABLE rfis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,                        -- sequential per project
    code VARCHAR(20) NOT NULL,                      -- 'RFI-0001'
    subject VARCHAR(255) NOT NULL,
    question TEXT NOT NULL,
    answer TEXT,
    discipline VARCHAR(30),                         -- civil/structural/MEP/track/geotech
    raised_by VARCHAR(120),
    assigned_to VARCHAR(120),
    status VARCHAR(15) NOT NULL DEFAULT 'open'
        CHECK (status IN ('open','answered','closed','void')),
    due_date DATE,
    raised_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, number),
    UNIQUE (project_id, code)
);

CREATE INDEX idx_rfis_status ON rfis(project_id, status);
CREATE INDEX idx_rfis_due ON rfis(project_id, due_date) WHERE status = 'open';

CREATE TABLE daily_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    report_date DATE NOT NULL,
    shift VARCHAR(10) NOT NULL DEFAULT 'day'
        CHECK (shift IN ('day','night','A','B','C')),
    weather VARCHAR(100),
    temp_c NUMERIC(4,1),
    manpower_total INTEGER,
    equipment_total INTEGER,
    narrative TEXT,                                 -- что делали
    hse_notes TEXT,                                 -- происшествия/наблюдения
    delays TEXT,                                    -- простои и причины
    author VARCHAR(120),
    status VARCHAR(15) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','submitted','approved')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, report_date, shift)
);

CREATE INDEX idx_dr_project_date ON daily_reports(project_id, report_date DESC);

-- Физобъёмы за смену против позиций BOQ -> физический прогресс
CREATE TABLE daily_work_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES daily_reports(id) ON DELETE CASCADE,
    boq_item_id UUID NOT NULL REFERENCES boq_items(id),
    qty_done NUMERIC(20,4) NOT NULL CHECK (qty_done >= 0),
    location VARCHAR(120),                          -- захватка/пикет
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_dwe_report ON daily_work_entries(report_id);
CREATE INDEX idx_dwe_item ON daily_work_entries(boq_item_id);

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('rfi',          'RFI',          'question', 'D-03'),
('daily_report', 'Daily Report', 'calendar', 'C-05')
ON CONFLICT (code) DO NOTHING;
